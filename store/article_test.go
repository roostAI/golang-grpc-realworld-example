package store

import (
	"testing"
	"github.com/jinzhu/gorm"
	"github.com/DATA-DOG/go-sqlmock"
	"sync"
	"github.com/raahii/golang-grpc-realworld-example/model"
	"github.com/raahii/golang-grpc-realworld-example/store"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"errors"
	"github.com/stretchr/testify/assert"
	"fmt"
	"github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm/dialects/postgres"
	"database/sql/driver"
	"context"
	"database/sql"
)

const errDBInit = "DB initialization error: %v"type T struct {
	common
	isParallel bool
	isEnvSet   bool
	context    *testContext
}
type ExpectedBegin struct {
	commonExpectation
	delay time.Duration
}
type ExpectedCommit struct {
	commonExpectation
}
type ExpectedExec struct {
	queryBasedExpectation
	result driver.Result
	delay  time.Duration
}
type ExpectedRollback struct {
	commonExpectation
}
type ExpectedQuery struct {
	queryBasedExpectation
	rows             driver.Rows
	delay            time.Duration
	rowsMustBeClosed bool
	rowsWereClosed   bool
}
type Rows struct {
	converter driver.ValueConverter
	cols      []string
	def       []*Column
	rows      [][]driver.Value
	pos       int
	nextErr   map[int]error
	closeErr  error
}
/*
ROOST_METHOD_HASH=NewArticleStore_6be2824012
ROOST_METHOD_SIG_HASH=NewArticleStore_3fe6f79a92


 */
func TestNewArticleStore(t *testing.T) {

	t.Run("Scenario 1: Successful Initialization of ArticleStore with a Valid Database Connection", func(t *testing.T) {

		db, _, err := sqlmock.New()
		if err != nil {
			t.Fatalf(errDBInit, err)
		}
		defer db.Close()

		gormDB, err := gorm.Open("sqlite3", db)
		if err != nil {
			t.Fatalf(errDBInit, err)
		}

		articleStore := NewArticleStore(gormDB)

		if articleStore == nil {
			t.Error("Expected ArticleStore not to be nil")
		}
		if articleStore.db != gormDB {
			t.Errorf("Expected db field in ArticleStore to match the provided gorm.DB object. Got different values.")
		}

		t.Log("Test Scenario 1 passed: ArticleStore initialized successfully with a valid DB connection.")
	})

	t.Run("Scenario 2: Initialization of ArticleStore with a Nil Database Connection", func(t *testing.T) {

		var nilDB *gorm.DB

		articleStore := NewArticleStore(nilDB)

		if articleStore.db != nil {
			t.Errorf("Expected db field to be nil. Got a non-nil db.")
		}

		t.Log("Test Scenario 2 passed: ArticleStore initialized with a nil DB connection without panicking.")
	})

	t.Run("Scenario 3: Thread Safety of ArticleStore Initialization", func(t *testing.T) {

		db, _, err := sqlmock.New()
		if err != nil {
			t.Fatalf(errDBInit, err)
		}
		defer db.Close()

		gormDB, err := gorm.Open("sqlite3", db)
		if err != nil {
			t.Fatalf(errDBInit, err)
		}

		var wg sync.WaitGroup
		const concurrency = 10
		stores := make(chan *ArticleStore, concurrency)

		for i := 0; i < concurrency; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				store := NewArticleStore(gormDB)
				stores <- store
			}()
		}

		wg.Wait()
		close(stores)

		for store := range stores {
			if store == nil || store.db != gormDB {
				t.Error("Expected all ArticleStore instances to have the same valid db value and be non-nil")
			}
		}

		t.Log("Test Scenario 3 passed: ArticleStore initializations are thread-safe.")
	})
}


/*
ROOST_METHOD_HASH=Create_0a911e138d
ROOST_METHOD_SIG_HASH=Create_723c594377


 */
func TestArticleStoreCreate(t *testing.T) {
	tests := []struct {
		name    string
		article model.Article
		mock    func(mock sqlmock.Sqlmock)
		wantErr bool
	}{
		{
			name: "Successfully Create a Valid Article",
			article: model.Article{
				Title:       "A Valid Title",
				Description: "A valid description",
				Body:        "Valid article body",
				UserID:      1,
			},
			mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO \"articles\"").
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
			wantErr: false,
		},
		{
			name: "Fail to Create an Article Due to Missing Title",
			article: model.Article{
				Description: "Description without title",
				Body:        "Body without title",
				UserID:      1,
			},
			mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO \"articles\"").
					WillReturnError(gorm.ErrInvalidData)
				mock.ExpectRollback()
			},
			wantErr: true,
		},
		{
			name: "Database Error During Article Creation",
			article: model.Article{
				Title:       "Another Valid Title",
				Description: "Another valid description",
				Body:        "Another valid article body",
				UserID:      2,
			},
			mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO \"articles\"").
					WillReturnError(gorm.ErrCantStartTransaction)
				mock.ExpectRollback()
			},
			wantErr: true,
		},
		{
			name: "Ensure Default FavoritesCount Set to Zero",
			article: model.Article{
				Title:       "Title With Default",
				Description: "Description With Default",
				Body:        "Body With Default",
				UserID:      3,
			},
			mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO \"articles\"").
					WillReturnResult(sqlmock.NewResult(2, 1))
				mock.ExpectCommit()
			},
			wantErr: false,
		},
		{
			name: "Handle Foreign Key Constraint Violation",
			article: model.Article{
				Title:       "Title With FK Issue",
				Description: "Description With FK Issue",
				Body:        "Body With FK Issue",
				UserID:      999,
			},
			mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO \"articles\"").
					WillReturnError(gorm.ErrForeignKeyViolation)
				mock.ExpectRollback()
			},
			wantErr: true,
		},
		{
			name: "Store Article With Tags and Comments",
			article: model.Article{
				Title:       "Complex Article",
				Description: "Complex Description",
				Body:        "Full Body",
				UserID:      4,
				Tags: []model.Tag{
					{Name: "Tech"},
					{Name: "Golang"},
				},
				Comments: []model.Comment{
					{Body: "Interesting comment"},
					{Body: "Another comment"},
				},
			},
			mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO \"articles\"").
					WillReturnResult(sqlmock.NewResult(3, 1))
				mock.ExpectCommit()
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("error initializing mock database: %v", err)
			}
			defer db.Close()

			dialector := postgres.New(postgres.Config{
				Conn: db,
			})
			gormDB, err := gorm.Open(dialector, &gorm.Config{})
			if err != nil {
				t.Fatalf("failed to open gorm DB: %v", err)
			}

			tt.mock(mock)

			store := &store.ArticleStore{DB: gormDB}
			err = store.Create(&tt.article)
			if (err != nil) != tt.wantErr {
				t.Errorf("Create() error = %v, wantErr %v", err, tt.wantErr)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}


/*
ROOST_METHOD_HASH=CreateComment_58d394e2c6
ROOST_METHOD_SIG_HASH=CreateComment_28b95f60a6


 */
func TestArticleStoreCreateComment(t *testing.T) {
	tests := []struct {
		name       string
		comment    *model.Comment
		setupMocks func(mock sqlmock.Sqlmock)
		expectErr  bool
	}{
		{
			name: "Successfully Create a Valid Comment",
			comment: &model.Comment{
				Body:      "This is a comment",
				UserID:    1,
				ArticleID: 1,
			},
			setupMocks: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec(`INSERT INTO "comments"`).
					WithArgs(sqlmock.AnyArg(), "This is a comment", 1, 1).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
			expectErr: false,
		},
		{
			name: "Attempt to Create a Comment with Missing Required Fields",
			comment: &model.Comment{
				Body:   "",
				UserID: 0,
			},
			setupMocks: func(mock sqlmock.Sqlmock) {

			},
			expectErr: true,
		},
		{
			name: "Handle Database Connection Error When Creating Comment",
			comment: &model.Comment{
				Body:      "Test comment",
				UserID:    1,
				ArticleID: 1,
			},
			setupMocks: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec(`INSERT INTO "comments"`).
					WithArgs(sqlmock.AnyArg(), "Test comment", 1, 1).
					WillReturnError(errors.New("connection error"))
				mock.ExpectRollback()
			},
			expectErr: true,
		},
		{
			name: "Create a Comment with a Non-Existing ArticleID",
			comment: &model.Comment{
				Body:      "Another comment",
				UserID:    1,
				ArticleID: 999,
			},
			setupMocks: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec(`INSERT INTO "comments"`).
					WithArgs(sqlmock.AnyArg(), "Another comment", 1, 999).
					WillReturnError(errors.New("foreign key constraint fails"))
				mock.ExpectRollback()
			},
			expectErr: true,
		},
		{
			name: "Simulate High Concurrency with Simultaneous Comment Creations",
			comment: &model.Comment{
				Body:      "Concurrent comment",
				UserID:    1,
				ArticleID: 1,
			},
			setupMocks: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec(`INSERT INTO "comments"`).
					WithArgs(sqlmock.AnyArg(), "Concurrent comment", 1, 1).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("Failed to open sqlmock database: %s", err)
			}
			defer db.Close()

			tt.setupMocks(mock)

			sqlDB, err := gorm.Open("postgres", db)
			if err != nil {
				t.Fatalf("Failed to initialize gorm database: %s", err)
			}

			store := &ArticleStore{db: sqlDB}

			if tt.name == "Simulate High Concurrency with Simultaneous Comment Creations" {
				var wg sync.WaitGroup
				for i := 0; i < 10; i++ {
					wg.Add(1)
					go func() {
						defer wg.Done()
						err := store.CreateComment(tt.comment)
						if (err != nil) != tt.expectErr {
							t.Errorf("%s: expected error: %v, got: %v", tt.name, tt.expectErr, err)
						}
					}()
				}
				wg.Wait()
			} else {
				err := store.CreateComment(tt.comment)
				if (err != nil) != tt.expectErr {
					t.Errorf("%s: expected error: %v, got: %v", tt.name, tt.expectErr, err)
				}
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("There were unfulfilled expectations: %s", err)
			}
		})
		t.Logf("Successfully tested %s", tt.name)
	}
}


/*
ROOST_METHOD_HASH=Delete_a8dc14c210
ROOST_METHOD_SIG_HASH=Delete_a4cc8044b1


 */
func TestArticleStoreDelete(t *testing.T) {

	tests := []struct {
		name          string
		article       *model.Article
		mockSetup     func(mock sqlmock.Sqlmock)
		expectedError error
	}{
		{
			name: "Successfully delete a valid article",
			article: &model.Article{
				Model: gorm.Model{ID: 1},
			},
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("DELETE FROM \"articles\" WHERE \"articles\".\"id\" = ?").
					WithArgs(1).
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
			expectedError: nil,
		},
		{
			name: "Attempt to delete a non-existent article",
			article: &model.Article{
				Model: gorm.Model{ID: 99},
			},
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("DELETE FROM \"articles\" WHERE \"articles\".\"id\" = ?").
					WithArgs(99).
					WillReturnResult(sqlmock.NewResult(1, 0))
			},
			expectedError: gorm.ErrRecordNotFound,
		},
		{
			name: "Delete an article with associated comments",
			article: &model.Article{
				Model:    gorm.Model{ID: 2},
				Comments: []model.Comment{{Body: "Sample comment"}},
			},
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("DELETE FROM \"articles\" WHERE \"articles\".\"id\" = ?").
					WithArgs(2).
					WillReturnResult(sqlmock.NewResult(1, 1))

			},
			expectedError: nil,
		},
		{
			name: "Handle database connection error during deletion",
			article: &model.Article{
				Model: gorm.Model{ID: 3},
			},
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("DELETE FROM \"articles\" WHERE \"articles\".\"id\" = ?").
					WithArgs(3).
					WillReturnError(errors.New("database connection error"))
			},
			expectedError: errors.New("database connection error"),
		},
		{
			name: "Attempt to delete an article with invalid data",
			article: &model.Article{
				Model: gorm.Model{ID: 4},
				Title: "",
			},
			mockSetup: func(mock sqlmock.Sqlmock) {

			},
			expectedError: errors.New("invalid data"),
		},
		{
			name: "Simultaneous deletions of the same article",
			article: &model.Article{
				Model: gorm.Model{ID: 5},
			},
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("DELETE FROM \"articles\" WHERE \"articles\".\"id\" = ?").
					WithArgs(5).
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
			expectedError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("Failed to open a stub database connection: %v", err)
			}
			defer db.Close()

			gormDB, err := gorm.Open("postgres", db)
			if err != nil {
				t.Fatalf("Failed to initialize GORM with sqlmock: %v", err)
			}

			store := ArticleStore{db: gormDB}

			tt.mockSetup(mock)

			err = store.Delete(tt.article)

			if err != nil && tt.expectedError == nil {
				t.Errorf("unexpected error: %v", err)
			} else if err == nil && tt.expectedError != nil {
				t.Errorf("expected error: %v, got nil", tt.expectedError)
			} else if err != nil && tt.expectedError != nil {
				if err.Error() != tt.expectedError.Error() {
					t.Errorf("expected error: %v, got: %v", tt.expectedError, err)
				}
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}


/*
ROOST_METHOD_HASH=DeleteComment_b345e525a7
ROOST_METHOD_SIG_HASH=DeleteComment_732762ff12


 */
func TestArticleStoreDeleteComment(t *testing.T) {
	type args struct {
		comment *model.Comment
	}
	tests := []struct {
		name     string
		mockFunc func(sqlmock.Sqlmock)
		args     args
		wantErr  bool
	}{
		{
			name: "Successful Deletion of an Existing Comment",
			mockFunc: func(mock sqlmock.Sqlmock) {

				mock.ExpectExec("DELETE FROM `comments` WHERE `comments`.`deleted_at` IS NULL AND \\(`comments`.`id` = \\?\\)").WithArgs(1).WillReturnResult(sqlmock.NewResult(1, 1))
			},
			args: args{
				comment: &model.Comment{Model: gorm.Model{ID: 1}, Body: "Test Comment", UserID: 1, ArticleID: 1},
			},
			wantErr: false,
		},
		{
			name: "Attempt to Delete a Non-Existent Comment",
			mockFunc: func(mock sqlmock.Sqlmock) {

				mock.ExpectExec("DELETE FROM `comments` WHERE `comments`.`deleted_at` IS NULL AND \\(`comments`.`id` = \\?\\)").WithArgs(999).WillReturnResult(sqlmock.NewResult(0, 0))
			},
			args: args{
				comment: &model.Comment{Model: gorm.Model{ID: 999}, Body: "Non-existent Comment", UserID: 999, ArticleID: 999},
			},
			wantErr: true,
		},
		{
			name: "Deletion with Database Error",
			mockFunc: func(mock sqlmock.Sqlmock) {

				mock.ExpectExec("DELETE FROM `comments` WHERE `comments`.`deleted_at` IS NULL AND \\(`comments`.`id` = \\?\\)").WithArgs(2).WillReturnError(errors.New("database error"))
			},
			args: args{
				comment: &model.Comment{Model: gorm.Model{ID: 2}, Body: "Error-prone Comment", UserID: 2, ArticleID: 2},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			assert.NoError(t, err)
			defer db.Close()

			gormDB, err := gorm.Open("sqlite3", db)
			assert.NoError(t, err)

			store := &ArticleStore{db: gormDB}

			tt.mockFunc(mock)

			err = store.DeleteComment(tt.args.comment)
			assert.Equal(t, tt.wantErr, err != nil)

			err = mock.ExpectationsWereMet()
			assert.NoError(t, err)
		})
	}
}


/*
ROOST_METHOD_HASH=GetCommentByID_4bc82104a6
ROOST_METHOD_SIG_HASH=GetCommentByID_333cab101b


 */
func TestArticleStoreGetCommentByID(t *testing.T) {

	t.Run("Fetch Existing Comment", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		assert.NoError(t, err)
		defer db.Close()

		gormDB, err := gorm.Open("postgres", db)
		assert.NoError(t, err)

		comment := model.Comment{
			Model:     gorm.Model{ID: 1},
			Body:      "Test Comment",
			UserID:    1,
			ArticleID: 1,
		}

		mock.ExpectQuery(`SELECT \* FROM "comments" WHERE "comments"\."deleted_at" IS NULL AND "comments"\."id" = \$1`).
			WithArgs(1).
			WillReturnRows(sqlmock.NewRows([]string{"id", "body", "user_id", "article_id"}).
				AddRow(comment.ID, comment.Body, comment.UserID, comment.ArticleID))

		store := ArticleStore{db: gormDB}
		result, err := store.GetCommentByID(1)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, comment.ID, result.ID)

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Log("Expectations were not met:", err)
		}
	})

	t.Run("Fetch Non-Existing Comment", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		assert.NoError(t, err)
		defer db.Close()

		gormDB, err := gorm.Open("postgres", db)
		assert.NoError(t, err)

		mock.ExpectQuery(`SELECT \* FROM "comments" WHERE "comments"\."deleted_at" IS NULL AND "comments"\."id" = \$1`).
			WithArgs(2).WillReturnError(gorm.ErrRecordNotFound)

		store := ArticleStore{db: gormDB}
		result, err := store.GetCommentByID(2)
		assert.Error(t, err)
		assert.Nil(t, result)

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Log("Expectations were not met:", err)
		}
	})

	t.Run("Database Connection Error", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		assert.NoError(t, err)
		defer db.Close()

		gormDB, err := gorm.Open("postgres", db)
		assert.NoError(t, err)

		mock.ExpectQuery(`SELECT \* FROM "comments" WHERE "comments"\."deleted_at" IS NULL AND "comments"\."id" = \$1`).
			WithArgs(1).WillReturnError(errors.New("database connection error"))

		store := ArticleStore{db: gormDB}
		result, err := store.GetCommentByID(1)
		assert.Error(t, err)
		assert.Nil(t, result)

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Log("Expectations were not met:", err)
		}
	})

	t.Run("Fetch Comment with ID Zero", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		assert.NoError(t, err)
		defer db.Close()

		gormDB, err := gorm.Open("postgres", db)
		assert.NoError(t, err)

		mock.ExpectQuery(`SELECT \* FROM "comments" WHERE "comments"\."deleted_at" IS NULL AND "comments"\."id" = \$1`).
			WithArgs(0).WillReturnError(gorm.ErrRecordNotFound)

		store := ArticleStore{db: gormDB}
		result, err := store.GetCommentByID(0)
		assert.Error(t, err)
		assert.Nil(t, result)

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Log("Expectations were not met:", err)
		}
	})

	t.Run("Specific Error for Non-Existing ID", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		assert.NoError(t, err)
		defer db.Close()

		gormDB, err := gorm.Open("postgres", db)
		assert.NoError(t, err)

		mock.ExpectQuery(`SELECT \* FROM "comments" WHERE "comments"\."deleted_at" IS NULL AND "comments"\."id" = \$1`).
			WithArgs(3).WillReturnError(gorm.ErrRecordNotFound)

		store := ArticleStore{db: gormDB}
		result, err := store.GetCommentByID(3)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, gorm.ErrRecordNotFound))
		assert.Nil(t, result)

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Log("Expectations were not met:", err)
		}
	})
}


/*
ROOST_METHOD_HASH=GetTags_ac049ebded
ROOST_METHOD_SIG_HASH=GetTags_25034b82b0


 */
func TestArticleStoreGetTags(t *testing.T) {
	tests := []struct {
		name           string
		mockSetup      func(sqlmock.Sqlmock)
		expectedTags   []model.Tag
		expectedError  error
		concurrentTest bool
	}{
		{
			name: "Retrieving Tags Successfully",
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "name"}).
					AddRow(1, "go").
					AddRow(2, "programming")
				mock.ExpectQuery("^SELECT (.+) FROM `tags`").WillReturnRows(rows)
			},
			expectedTags: []model.Tag{
				{Model: gorm.Model{ID: 1}, Name: "go"},
				{Model: gorm.Model{ID: 2}, Name: "programming"},
			},
			expectedError:  nil,
			concurrentTest: false,
		},
		{
			name: "Database Error While Retrieving Tags",
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("^SELECT (.+) FROM `tags`").WillReturnError(gorm.ErrInvalidSQL)
			},
			expectedTags:   nil,
			expectedError:  gorm.ErrInvalidSQL,
			concurrentTest: false,
		},
		{
			name: "No Tags Available in the Database",
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "name"})
				mock.ExpectQuery("^SELECT (.+) FROM `tags`").WillReturnRows(rows)
			},
			expectedTags:   []model.Tag{},
			expectedError:  nil,
			concurrentTest: false,
		},
		{
			name: "Database Connection Issue",
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("^SELECT (.+) FROM `tags`").WillReturnError(gorm.ErrInvalidTransaction)
			},
			expectedTags:   nil,
			expectedError:  gorm.ErrInvalidTransaction,
			concurrentTest: false,
		},
		{
			name: "Concurrent Access to the Tags Table",
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "name"}).
					AddRow(1, "go").
					AddRow(2, "programming")
				mock.ExpectQuery("^SELECT (.+) FROM `tags`").WillReturnRows(rows)
			},
			expectedTags: []model.Tag{
				{Model: gorm.Model{ID: 1}, Name: "go"},
				{Model: gorm.Model{ID: 2}, Name: "programming"},
			},
			expectedError:  nil,
			concurrentTest: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			db, mock, err := sqlmock.New()
			assert.NoError(t, err)
			defer db.Close()

			gormDB, err := gorm.Open("mysql", db)
			assert.NoError(t, err)
			store := &ArticleStore{db: gormDB}

			tt.mockSetup(mock)

			if tt.concurrentTest {
				var wg sync.WaitGroup
				results := make(chan struct {
					tags []model.Tag
					err  error
				}, 5)

				for i := 0; i < 5; i++ {
					wg.Add(1)
					go func() {
						defer wg.Done()
						tags, err := store.GetTags()
						results <- struct {
							tags []model.Tag
							err  error
						}{tags, err}
					}()
				}

				wg.Wait()
				close(results)

				for res := range results {
					assert.Equal(t, tt.expectedTags, res.tags)
					assert.Equal(t, tt.expectedError, res.err)
				}
			} else {

				tags, err := store.GetTags()

				assert.Equal(t, tt.expectedTags, tags)
				assert.Equal(t, tt.expectedError, err)

				if tt.expectedError == nil {
					t.Logf("Passed: %s", tt.name)
				} else {
					t.Logf("Error caught as expected in: %s - %v", tt.name, err)
				}
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("Expectations were not met: %v", err)
			}
		})
	}
}


/*
ROOST_METHOD_HASH=GetByID_36e92ad6eb
ROOST_METHOD_SIG_HASH=GetByID_9616e43e52


 */
func TestArticleStoreGetByID(t *testing.T) {
	type testCase struct {
		name            string
		prepare         func(mock sqlmock.Sqlmock)
		id              uint
		expectedError   error
		expectedArticle *model.Article
	}

	tests := []testCase{
		{
			name: "Successfully Retrieve an Article by a Valid ID",
			prepare: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`^SELECT \* FROM "articles"`).
					WithArgs(1).
					WillReturnRows(sqlmock.NewRows([]string{"id", "title", "description", "body", "user_id"}).
						AddRow(1, "Sample Title", "Sample Description", "Sample Body", 1))
				mock.ExpectQuery(`^SELECT \* FROM "tags"`).
					WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).
						AddRow(1, "Tag1"))
				mock.ExpectQuery(`^SELECT \* FROM "users"`).
					WillReturnRows(sqlmock.NewRows([]string{"id", "username"}).
						AddRow(1, "AuthorName"))
			},
			id: 1,
			expectedArticle: &model.Article{
				Model:       gorm.Model{ID: 1},
				Title:       "Sample Title",
				Description: "Sample Description",
				Body:        "Sample Body",
				UserID:      1,
				Tags:        []model.Tag{{Model: gorm.Model{ID: 1}, Name: "Tag1"}},
				Author:      model.User{Model: gorm.Model{ID: 1}, Username: "AuthorName"},
			},
			expectedError: nil,
		},
		{
			name: "Return Error for Non-Existent Article ID",
			prepare: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`^SELECT \* FROM "articles"`).
					WithArgs(99).
					WillReturnError(gorm.ErrRecordNotFound)
			},
			id:              99,
			expectedArticle: nil,
			expectedError:   gorm.ErrRecordNotFound,
		},
		{
			name: "Ensure Proper Error Handling for Database Access Failures",
			prepare: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`^SELECT \* FROM "articles"`).
					WithArgs(1).
					WillReturnError(errors.New("database connection error"))
			},
			id:              1,
			expectedArticle: nil,
			expectedError:   errors.New("database connection error"),
		},
		{
			name: "Retrieve an Article with No Associated Tags",
			prepare: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`^SELECT \* FROM "articles"`).
					WithArgs(2).
					WillReturnRows(sqlmock.NewRows([]string{"id", "title", "description", "body", "user_id"}).
						AddRow(2, "Title Without Tags", "Description", "Body", 1))
				mock.ExpectQuery(`^SELECT \* FROM "tags"`).WillReturnRows(sqlmock.NewRows(nil))
				mock.ExpectQuery(`^SELECT \* FROM "users"`).
					WillReturnRows(sqlmock.NewRows([]string{"id", "username"}).
						AddRow(1, "AuthorName"))
			},
			id: 2,
			expectedArticle: &model.Article{
				Model:       gorm.Model{ID: 2},
				Title:       "Title Without Tags",
				Description: "Description",
				Body:        "Body",
				UserID:      1,
				Tags:        []model.Tag{},
				Author:      model.User{Model: gorm.Model{ID: 1}, Username: "AuthorName"},
			},
			expectedError: nil,
		},
		{
			name: "Retrieve an Article with Maximal Field Values",
			prepare: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`^SELECT \* FROM "articles"`).
					WithArgs(3).
					WillReturnRows(sqlmock.NewRows([]string{"id", "title", "description", "body", "user_id"}).
						AddRow(3, "Maximal Title", "Maximal Description", "Maximal Body", 1))
				mock.ExpectQuery(`^SELECT \* FROM "tags"`).
					WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).
						AddRow(1, "Tag1").AddRow(2, "Tag2"))
				mock.ExpectQuery(`^SELECT \* FROM "users"`).
					WillReturnRows(sqlmock.NewRows([]string{"id", "username"}).
						AddRow(1, "AuthorName"))
			},
			id: 3,
			expectedArticle: &model.Article{
				Model:       gorm.Model{ID: 3},
				Title:       "Maximal Title",
				Description: "Maximal Description",
				Body:        "Maximal Body",
				UserID:      1,
				Tags:        []model.Tag{{Model: gorm.Model{ID: 1}, Name: "Tag1"}, {Model: gorm.Model{ID: 2}, Name: "Tag2"}},
				Author:      model.User{Model: gorm.Model{ID: 1}, Username: "AuthorName"},
			},
			expectedError: nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			assert.NoError(t, err)

			gormDB, err := gorm.Open("postgres", db)
			assert.NoError(t, err)

			store := &ArticleStore{db: gormDB}
			tc.prepare(mock)

			article, err := store.GetByID(tc.id)

			assert.Equal(t, tc.expectedError, err)
			assert.Equal(t, tc.expectedArticle, article)

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}


/*
ROOST_METHOD_HASH=Update_51145aa965
ROOST_METHOD_SIG_HASH=Update_6c1b5471fe


 */
func TestArticleStoreUpdate(t *testing.T) {
	tests := []struct {
		name       string
		setupMock  func(mock sqlmock.Sqlmock, article *model.Article)
		input      *model.Article
		wantError  bool
		errorCheck func(err error) bool
	}{
		{
			name: "Successful Update of an Article",
			setupMock: func(mock sqlmock.Sqlmock, article *model.Article) {
				mock.ExpectBegin()
				mock.ExpectExec("UPDATE articles").WithArgs(article.Title, article.Description, article.Body, article.UserID, article.FavoritesCount, article.ID).
					WillReturnResult(sqlmock.NewResult(0, 1))
				mock.ExpectCommit()
			},
			input: &model.Article{
				Model:          gorm.Model{ID: 1},
				Title:          "Updated Title",
				Description:    "Updated Description",
				Body:           "Updated Body",
				UserID:         1,
				FavoritesCount: 5,
			},
			wantError:  false,
			errorCheck: nil,
		},
		{
			name: "Update of Non-Existent Article",
			setupMock: func(mock sqlmock.Sqlmock, article *model.Article) {
				mock.ExpectBegin()
				mock.ExpectExec("UPDATE articles").WithArgs(article.Title, article.Description, article.Body, article.UserID, article.FavoritesCount, article.ID).
					WillReturnResult(sqlmock.NewResult(0, 0))
				mock.ExpectCommit()
			},
			input: &model.Article{
				Model:          gorm.Model{ID: 9999},
				Title:          "Non-existent Article",
				Description:    "No Description",
				Body:           "No Body",
				UserID:         1,
				FavoritesCount: 0,
			},
			wantError: true,
			errorCheck: func(err error) bool {
				return gorm.IsRecordNotFoundError(err)
			},
		},
		{
			name: "Update with Invalid Data",
			setupMock: func(mock sqlmock.Sqlmock, article *model.Article) {
				mock.ExpectBegin()
				mock.ExpectExec("UPDATE articles").WithArgs(article.Title, article.Description, article.Body, article.UserID, article.FavoritesCount, article.ID).
					WillReturnError(gorm.ErrInvalidTransaction)
				mock.ExpectCommit()
			},
			input: &model.Article{
				Model:          gorm.Model{ID: 1},
				Title:          "",
				Description:    "Description",
				Body:           "Body",
				UserID:         1,
				FavoritesCount: 0,
			},
			wantError: true,
			errorCheck: func(err error) bool {
				return err == gorm.ErrInvalidTransaction
			},
		},
		{
			name: "Database Connection Failure",
			setupMock: func(mock sqlmock.Sqlmock, article *model.Article) {
				mock.ExpectBegin()
				mock.ExpectExec("UPDATE articles").WithArgs(article.Title, article.Description, article.Body, article.UserID, article.FavoritesCount, article.ID).
					WillReturnError(gorm.ErrCantStartTransaction)
				mock.ExpectCommit()
			},
			input: &model.Article{
				Model:          gorm.Model{ID: 1},
				Title:          "Title",
				Description:    "Description",
				Body:           "Body",
				UserID:         1,
				FavoritesCount: 0,
			},
			wantError: true,
			errorCheck: func(err error) bool {
				return err == gorm.ErrCantStartTransaction
			},
		},
		{
			name: "Partial Data Update",
			setupMock: func(mock sqlmock.Sqlmock, article *model.Article) {
				mock.ExpectBegin()
				mock.ExpectExec("UPDATE articles").WithArgs(article.Title, article.ID).
					WillReturnResult(sqlmock.NewResult(0, 1))
				mock.ExpectCommit()
			},
			input: &model.Article{
				Model: gorm.Model{ID: 1},
				Title: "Partially Updated Title",
			},
			wantError:  false,
			errorCheck: nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {

			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
			}
			defer db.Close()

			gormDB, err := gorm.Open("postgres", db)
			if err != nil {
				t.Fatalf("an error '%s' was not expected when initializing gorm", err)
			}
			defer gormDB.Close()

			store := &ArticleStore{
				db: gormDB,
			}

			tc.setupMock(mock, tc.input)

			err = store.Update(tc.input)

			if tc.wantError {
				if err == nil {
					t.Errorf("expected an error but got none")
					return
				}
				if tc.errorCheck != nil && !tc.errorCheck(err) {
					t.Errorf("unexpected error type: %v", err)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}

			t.Logf("Scenario '%s': executed and validated successfully", tc.name)
		})
	}
}


/*
ROOST_METHOD_HASH=GetComments_e24a0f1b73
ROOST_METHOD_SIG_HASH=GetComments_fa6661983e


 */
func TestArticleStoreGetComments(t *testing.T) {

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("unexpected error when opening a stub database connection: %s", err)
	}
	defer db.Close()

	gormDB, err := gorm.Open("mysql", db)
	if err != nil {
		t.Fatalf("unexpected error when opening a gorm DB: %s", err)
	}
	defer gormDB.Close()

	store := &ArticleStore{db: gormDB}

	type testCase struct {
		name             string
		setupMock        func()
		article          *model.Article
		expectedComments []model.Comment
		expectedError    error
	}

	testCases := []testCase{
		{
			name: "Scenario 1: Normal Operation with Comments Present",
			setupMock: func() {
				rows := sqlmock.NewRows([]string{"id", "body", "user_id", "article_id"}).
					AddRow(1, "comment 1", 1, 1).
					AddRow(2, "comment 2", 1, 1)
				mock.ExpectQuery("^SELECT \\* FROM `comments` WHERE article_id = ?").
					WillReturnRows(rows)
			},
			article: &model.Article{Model: gorm.Model{ID: 1}},
			expectedComments: []model.Comment{
				{Model: gorm.Model{ID: 1}, Body: "comment 1", UserID: 1, ArticleID: 1},
				{Model: gorm.Model{ID: 2}, Body: "comment 2", UserID: 1, ArticleID: 1},
			},
			expectedError: nil,
		},
		{
			name: "Scenario 2: Article with No Comments",
			setupMock: func() {
				rows := sqlmock.NewRows([]string{"id", "body", "user_id", "article_id"})
				mock.ExpectQuery("^SELECT \\* FROM `comments` WHERE article_id = ?").
					WillReturnRows(rows)
			},
			article:          &model.Article{Model: gorm.Model{ID: 2}},
			expectedComments: []model.Comment{},
			expectedError:    nil,
		},
		{
			name: "Scenario 3: Article Not Found in Database",
			setupMock: func() {
				mock.ExpectQuery("^SELECT \\* FROM `comments` WHERE article_id = ?").
					WillReturnError(gorm.ErrRecordNotFound)
			},
			article:          &model.Article{Model: gorm.Model{ID: 9999}},
			expectedComments: []model.Comment{},
			expectedError:    gorm.ErrRecordNotFound,
		},
		{
			name: "Scenario 4: Database Error Encountered",
			setupMock: func() {
				mock.ExpectQuery("^SELECT \\* FROM `comments` WHERE article_id = ?").
					WillReturnError(gorm.ErrInvalidSQL)
			},
			article:          &model.Article{Model: gorm.Model{ID: 1}},
			expectedComments: nil,
			expectedError:    gorm.ErrInvalidSQL,
		},
		{
			name: "Scenario 5: Preload Author Functionality Works",
			setupMock: func() {
				rows := sqlmock.NewRows([]string{"id", "body", "user_id", "article_id"}).
					AddRow(1, "comment with author", 1, 1)
				mock.ExpectQuery("^SELECT \\* FROM `comments` WHERE article_id = ?").
					WillReturnRows(rows)
			},
			article: &model.Article{Model: gorm.Model{ID: 1}},
			expectedComments: []model.Comment{
				{Model: gorm.Model{ID: 1}, Body: "comment with author", UserID: 1, ArticleID: 1},
			},
			expectedError: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			tc.setupMock()

			comments, err := store.GetComments(tc.article)

			assert.Equal(t, tc.expectedComments, comments, "Expected comments do not match the actual comments")
			assert.Equal(t, tc.expectedError, err, "Expected error does not match the actual error")

			if err != nil {
				t.Logf("Expected error: %v, Got: %v", tc.expectedError, err)
			} else {
				t.Logf("Success: Retrieved comments match expected results")
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("Expectations not met: %v", err)
			}
		})
	}
}


/*
ROOST_METHOD_HASH=IsFavorited_7ef7d3ed9e
ROOST_METHOD_SIG_HASH=IsFavorited_f34d52378f


 */
func TestArticleStoreIsFavorited(t *testing.T) {
	type testCase struct {
		description   string
		article       *model.Article
		user          *model.User
		mockBehaviour func(sqlmock.Sqlmock)
		expectedBool  bool
		expectedErr   error
	}

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	gormDB, err := gorm.Open("sqlite3", db)
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening gorm database connection", err)
	}

	store := &ArticleStore{db: gormDB}

	testCases := []testCase{
		{
			description: "Scenario 1: Article and User Are Not Nil",
			article: &model.Article{
				Model: gorm.Model{ID: 1},
			},
			user: &model.User{
				Model: gorm.Model{ID: 1},
			},
			mockBehaviour: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT count\\(.*\\) FROM favorite_articles").
					WithArgs(1, 1).
					WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
			},
			expectedBool: true,
			expectedErr:  nil,
		},
		{
			description: "Scenario 2: Article and User Are Not Favorited",
			article: &model.Article{
				Model: gorm.Model{ID: 1},
			},
			user: &model.User{
				Model: gorm.Model{ID: 1},
			},
			mockBehaviour: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT count\\(.*\\) FROM favorite_articles").
					WithArgs(1, 1).
					WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
			},
			expectedBool: false,
			expectedErr:  nil,
		},
		{
			description: "Scenario 3: Nil Article Parameter",
			article:     nil,
			user: &model.User{
				Model: gorm.Model{ID: 1},
			},
			mockBehaviour: nil,
			expectedBool:  false,
			expectedErr:   nil,
		},
		{
			description: "Scenario 4: Nil User Parameter",
			article: &model.Article{
				Model: gorm.Model{ID: 1},
			},
			user:          nil,
			mockBehaviour: nil,
			expectedBool:  false,
			expectedErr:   nil,
		},
		{
			description: "Scenario 5: Database Error",
			article: &model.Article{
				Model: gorm.Model{ID: 1},
			},
			user: &model.User{
				Model: gorm.Model{ID: 1},
			},
			mockBehaviour: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT count\\(.*\\) FROM favorite_articles").
					WithArgs(1, 1).
					WillReturnError(gorm.ErrInvalidSQL)
			},
			expectedBool: false,
			expectedErr:  gorm.ErrInvalidSQL,
		},
		{
			description: "Scenario 6: No Favorited Articles Exist",
			article: &model.Article{
				Model: gorm.Model{ID: 1},
			},
			user: &model.User{
				Model: gorm.Model{ID: 1},
			},
			mockBehaviour: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT count\\(.*\\) FROM favorite_articles").
					WithArgs(1, 1).
					WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
			},
			expectedBool: false,
			expectedErr:  nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			if tc.mockBehaviour != nil {
				tc.mockBehaviour(mock)
			}

			result, err := store.IsFavorited(tc.article, tc.user)
			if result != tc.expectedBool {
				t.Errorf("expected %t, got %t", tc.expectedBool, result)
			}
			if err != tc.expectedErr {
				t.Errorf("expected error %v, got %v", tc.expectedErr, err)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}

			t.Logf("Test case '%s' executed successfully", tc.description)
		})
	}
}


/*
ROOST_METHOD_HASH=GetFeedArticles_9c4f57afe4
ROOST_METHOD_SIG_HASH=GetFeedArticles_cadca0e51b


 */
func TestArticleStoreGetFeedArticles(t *testing.T) {
	tests := []struct {
		name          string
		userIDs       []uint
		limit         int64
		offset        int64
		mockSetup     func(mock sqlmock.Sqlmock)
		expectedError error
		expectedCount int
	}{
		{
			name:    "Retrieve Articles for Specific User IDs",
			userIDs: []uint{1, 2},
			limit:   10,
			offset:  0,
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "title", "user_id"}).
					AddRow(1, "Article 1", 1).
					AddRow(2, "Article 2", 2)
				mock.ExpectQuery(`SELECT * FROM "articles" WHERE user_id in (\?,\?)`).
					WithArgs(1, 2).
					WillReturnRows(rows)
			},
			expectedError: nil,
			expectedCount: 2,
		},
		{
			name:    "Limit Number of Articles Retrieved",
			userIDs: []uint{1},
			limit:   1,
			offset:  0,
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "title", "user_id"}).
					AddRow(1, "Article 1", 1)
				mock.ExpectQuery(`SELECT * FROM "articles" WHERE user_id in (\?)`).
					WithArgs(1).
					WillReturnRows(rows)
			},
			expectedError: nil,
			expectedCount: 1,
		},
		{
			name:    "Handle Article Retrieval Offset",
			userIDs: []uint{1, 2, 3},
			limit:   10,
			offset:  2,
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "title", "user_id"}).
					AddRow(3, "Article 3", 1)
				mock.ExpectQuery(`SELECT * FROM "articles" WHERE user_id in (\?,\?,\?)`).
					WithArgs(1, 2, 3).
					WillReturnRows(rows)
			},
			expectedError: nil,
			expectedCount: 1,
		},
		{
			name:    "No Articles Found for Given User IDs",
			userIDs: []uint{99},
			limit:   10,
			offset:  0,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT * FROM "articles" WHERE user_id in (\?)`).
					WithArgs(99).
					WillReturnRows(sqlmock.NewRows([]string{"id", "title", "user_id"}))
			},
			expectedError: nil,
			expectedCount: 0,
		},
		{
			name:    "Ensure Error Handling for Database Issues",
			userIDs: []uint{1},
			limit:   10,
			offset:  0,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT * FROM "articles" WHERE user_id in (\?)`).
					WillReturnError(gorm.ErrInvalidSQL)
			},
			expectedError: gorm.ErrInvalidSQL,
			expectedCount: 0,
		},
		{
			name:    "Validate Preload of Associated Data",
			userIDs: []uint{1},
			limit:   10,
			offset:  0,
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "title", "user_id"}).
					AddRow(1, "Article 1", 1)
				mock.ExpectQuery(`SELECT * FROM "articles" WHERE user_id in (\?)`).
					WithArgs(1).
					WillReturnRows(rows)
				mock.ExpectQuery(`SELECT * FROM "users" WHERE "users"."id" = \?`).
					WithArgs(1).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
			},
			expectedError: nil,
			expectedCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			assert.NoError(t, err)
			defer db.Close()

			mockDb, err := gorm.Open("postgres", db)
			assert.NoError(t, err)

			tt.mockSetup(mock)

			store := &ArticleStore{mockDb}

			articles, err := store.GetFeedArticles(tt.userIDs, tt.limit, tt.offset)
			assert.Equal(t, tt.expectedError, err)
			assert.Equal(t, tt.expectedCount, len(articles))
		})
	}
}


/*
ROOST_METHOD_HASH=AddFavorite_2b0cb9d894
ROOST_METHOD_SIG_HASH=AddFavorite_c4dea0ee90


 */
func TestArticleStoreAddFavorite(t *testing.T) {
	tests := []struct {
		name             string
		articleSetup     func(mock sqlmock.Sqlmock, article *model.Article, user *model.User)
		expectedError    error
		expectedFavCount int32
		concurrentUsers  int
	}{
		{
			name: "Scenario 1: Successfully Add a Favorite to an Article",
			articleSetup: func(mock sqlmock.Sqlmock, article *model.Article, user *model.User) {
				mock.ExpectBegin()
				mock.ExpectExec(`INSERT INTO favorite_articles`).WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg()).WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectExec(`UPDATE articles SET favorites_count = favorites_count + ? WHERE id = ?`).WithArgs(1, article.ID).WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
			expectedError:    nil,
			expectedFavCount: 1,
			concurrentUsers:  1,
		},
		{
			name: "Scenario 2: Add Favorite When Article Already Favorited by the User",
			articleSetup: func(mock sqlmock.Sqlmock, article *model.Article, user *model.User) {
				mock.ExpectBegin()
				mock.ExpectExec(`INSERT INTO favorite_articles`).WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg()).WillReturnResult(sqlmock.NewResult(0, 0))
				mock.ExpectRollback()
			},
			expectedError:    nil,
			expectedFavCount: 0,
			concurrentUsers:  1,
		},
		{
			name: "Scenario 3: Add Favorite with Nonexistent Article",
			articleSetup: func(mock sqlmock.Sqlmock, article *model.Article, user *model.User) {
				mock.ExpectBegin()
				mock.ExpectExec(`INSERT INTO favorite_articles`).WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg()).WillReturnError(gorm.ErrRecordNotFound)
				mock.ExpectRollback()
			},
			expectedError:    gorm.ErrRecordNotFound,
			expectedFavCount: 0,
			concurrentUsers:  1,
		},
		{
			name: "Scenario 4: Database Transaction Rollback on Append Error",
			articleSetup: func(mock sqlmock.Sqlmock, article *model.Article, user *model.User) {
				mock.ExpectBegin()
				mock.ExpectExec(`INSERT INTO favorite_articles`).WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg()).WillReturnError(gorm.ErrInvalidTransaction)
				mock.ExpectRollback()
			},
			expectedError:    gorm.ErrInvalidTransaction,
			expectedFavCount: 0,
			concurrentUsers:  1,
		},
		{
			name: "Scenario 5: Database Transaction Rollback on Favorites Count Update Error",
			articleSetup: func(mock sqlmock.Sqlmock, article *model.Article, user *model.User) {
				mock.ExpectBegin()
				mock.ExpectExec(`INSERT INTO favorite_articles`).WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg()).WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectExec(`UPDATE articles SET favorites_count = favorites_count + ? WHERE id = ?`).WithArgs(1, article.ID).WillReturnError(gorm.ErrInvalidTransaction)
				mock.ExpectRollback()
			},
			expectedError:    gorm.ErrInvalidTransaction,
			expectedFavCount: 0,
			concurrentUsers:  1,
		},
		{
			name: "Scenario 6: Concurrent Favoriting by Multiple Users",
			articleSetup: func(mock sqlmock.Sqlmock, article *model.Article, user *model.User) {
				for i := 0; i < 3; i++ {
					mock.ExpectBegin()
					mock.ExpectExec(`INSERT INTO favorite_articles`).WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg()).WillReturnResult(sqlmock.NewResult(1, 1))
					mock.ExpectExec(`UPDATE articles SET favorites_count = favorites_count + ? WHERE id = ?`).WithArgs(1, article.ID).WillReturnResult(sqlmock.NewResult(1, 1))
					mock.ExpectCommit()
				}
			},
			expectedError:    nil,
			expectedFavCount: 3,
			concurrentUsers:  3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("failed to open sqlmock database: %s", err)
			}
			defer db.Close()

			gormDB, err := gorm.Open("mysql", db)
			if err != nil {
				t.Fatalf("failed to open gorm DB: %s", err)
			}
			defer gormDB.Close()

			store := &ArticleStore{db: gormDB}

			mockArticle := &model.Article{Model: gorm.Model{ID: 1}}
			mockUser := &model.User{Model: gorm.Model{ID: 1}}

			tt.articleSetup(mock, mockArticle, mockUser)

			ch := make(chan error, tt.concurrentUsers)
			for i := 0; i < tt.concurrentUsers; i++ {
				go func() {
					ch <- store.AddFavorite(mockArticle, mockUser)
				}()
			}

			var actualError error
			for i := 0; i < tt.concurrentUsers; i++ {
				err := <-ch
				if err != nil {
					actualError = err
				}
			}

			if actualError != tt.expectedError {
				t.Errorf("unexpected error, got: %v, want: %v", actualError, tt.expectedError)
			}

			if mockArticle.FavoritesCount != tt.expectedFavCount {
				t.Errorf("unexpected favorites count, got: %d, want: %d", mockArticle.FavoritesCount, tt.expectedFavCount)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}


/*
ROOST_METHOD_HASH=DeleteFavorite_a856bcbb70
ROOST_METHOD_SIG_HASH=DeleteFavorite_f7e5c0626f


 */
func TestArticleStoreDeleteFavorite(t *testing.T) {

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("An error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	gormDB, err := gorm.Open("postgres", db)
	if err != nil {
		t.Fatalf("An error '%s' was not expected when opening a gorm database", err)
	}

	aStore := &ArticleStore{db: gormDB}

	type testCase struct {
		name           string
		article        *model.Article
		user           *model.User
		mockSetup      func()
		expectedCount  int32
		expectError    bool
		finalFavorites []model.User
	}

	tests := []testCase{
		{
			name: "Scenario 1: Successfully Remove Favorite User from Article",
			article: &model.Article{
				FavoritesCount: 1,
				FavoritedUsers: []model.User{{Model: gorm.Model{ID: 1}}},
			},
			user: &model.User{Model: gorm.Model{ID: 1}},
			mockSetup: func() {
				mock.ExpectBegin()
				mock.ExpectExec("DELETE FROM \"favorite_articles\" WHERE").WithArgs(1, 1).WillReturnResult(sqlmock.NewResult(0, 1))
				mock.ExpectExec("UPDATE \"articles\" SET \"favorites_count\" = \"favorites_count\" - ? WHERE \"articles\".\"id\" = ?").
					WithArgs(1, 1).WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
			expectedCount:  0,
			expectError:    false,
			finalFavorites: []model.User{},
		},
		{
			name: "Scenario 2: Removing Non-existing Favorite User",
			article: &model.Article{
				FavoritesCount: 0,
				FavoritedUsers: []model.User{},
			},
			user: &model.User{Model: gorm.Model{ID: 1}},
			mockSetup: func() {
				mock.ExpectBegin()
				mock.ExpectExec("DELETE FROM \"favorite_articles\" WHERE").WithArgs(1, 1).WillReturnResult(sqlmock.NewResult(0, 0))
				mock.ExpectExec("UPDATE \"articles\" SET \"favorites_count\" = \"favorites_count\" - ? WHERE \"articles\".\"id\" = ?").
					WithArgs(1, 1).WillReturnError(errors.New("no rows were updated"))
				mock.ExpectRollback()
			},
			expectedCount:  0,
			expectError:    true,
			finalFavorites: []model.User{},
		},
		{
			name: "Scenario 3: Database Error During User Removal",
			article: &model.Article{
				FavoritesCount: 1,
				FavoritedUsers: []model.User{{Model: gorm.Model{ID: 1}}},
			},
			user: &model.User{Model: gorm.Model{ID: 1}},
			mockSetup: func() {
				mock.ExpectBegin()
				mock.ExpectExec("DELETE FROM \"favorite_articles\" WHERE").WithArgs(1, 1).WillReturnError(errors.New("db error"))
				mock.ExpectRollback()
			},
			expectedCount:  1,
			expectError:    true,
			finalFavorites: []model.User{{Model: gorm.Model{ID: 1}}},
		},
		{
			name: "Scenario 4: Database Error During Count Update",
			article: &model.Article{
				FavoritesCount: 1,
				FavoritedUsers: []model.User{{Model: gorm.Model{ID: 1}}},
			},
			user: &model.User{Model: gorm.Model{ID: 1}},
			mockSetup: func() {
				mock.ExpectBegin()
				mock.ExpectExec("DELETE FROM \"favorite_articles\" WHERE").WithArgs(1, 1).WillReturnResult(sqlmock.NewResult(0, 1))
				mock.ExpectExec("UPDATE \"articles\" SET \"favorites_count\" = \"favorites_count\" - ? WHERE \"articles\".\"id\" = ?").
					WithArgs(1, 1).WillReturnError(errors.New("count update error"))
				mock.ExpectRollback()
			},
			expectedCount:  1,
			expectError:    true,
			finalFavorites: []model.User{{Model: gorm.Model{ID: 1}}},
		},
		{
			name: "Scenario 5: User and Article Interaction With No Prior Relation",
			article: &model.Article{
				FavoritesCount: 0,
				FavoritedUsers: []model.User{},
			},
			user: &model.User{Model: gorm.Model{ID: 2}},
			mockSetup: func() {
				mock.ExpectBegin()
				mock.ExpectExec("DELETE FROM \"favorite_articles\" WHERE").WithArgs(1, 2).WillReturnResult(sqlmock.NewResult(0, 0))
				mock.ExpectExec("UPDATE \"articles\" SET \"favorites_count\" = \"favorites_count\" - ? WHERE \"articles\".\"id\" = ?").
					WithArgs(1, 1).WillReturnError(errors.New("no rows were updated"))
				mock.ExpectRollback()
			},
			expectedCount:  0,
			expectError:    true,
			finalFavorites: []model.User{},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tc.mockSetup()

			err := aStore.DeleteFavorite(tc.article, tc.user)
			if tc.expectError && err == nil {
				t.Errorf("Expected error but got nil")
			}
			if !tc.expectError && err != nil {
				t.Errorf("Did not expect error but got: %v", err)
			}

			if tc.article.FavoritesCount != tc.expectedCount {
				t.Errorf("Expected FavoritesCount to be %d but got %d", tc.expectedCount, tc.article.FavoritesCount)
			}

			if len(tc.article.FavoritedUsers) != len(tc.finalFavorites) {
				t.Errorf("Expected FavoritedUsers length to be %d but got %d", len(tc.finalFavorites), len(tc.article.FavoritedUsers))
			}

			err = mock.ExpectationsWereMet()
			if err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}


/*
ROOST_METHOD_HASH=GetArticles_6382a4fe7a
ROOST_METHOD_SIG_HASH=GetArticles_1a0b3b0e8b


 */
func TestArticleStoreGetArticles(t *testing.T) {

	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("error creating mock database: %s", err)
	}
	defer mockDB.Close()

	gormDB, err := gorm.Open("mysql", mockDB)
	if err != nil {
		t.Fatalf("error opening gorm db: %s", err)
	}
	defer gormDB.Close()

	articleStore := &ArticleStore{db: gormDB}

	tests := []struct {
		name          string
		tagName       string
		username      string
		favoritedBy   *model.User
		limit         int64
		offset        int64
		mockSetup     func()
		expectedError bool
		expectedCount int
	}{
		{
			name:    "Scenario 1: Retrieve Articles by Tag Name",
			tagName: "go",
			mockSetup: func() {
				mock.ExpectQuery(`SELECT .* FROM "articles" .* JOIN article_tags`).
					WithArgs("go").
					WillReturnRows(sqlmock.NewRows([]string{"id", "title"}).AddRow(1, "Go Concurrency").AddRow(2, "Golang Testing"))
			},
			expectedError: false,
			expectedCount: 2,
		},
		{
			name:     "Scenario 2: Retrieve Articles by Author Username",
			username: "johndoe",
			mockSetup: func() {
				mock.ExpectQuery(`SELECT .* FROM "articles" .* JOIN users`).
					WithArgs("johndoe").
					WillReturnRows(sqlmock.NewRows([]string{"id", "title"}).AddRow(1, "Understanding SQL Join"))
			},
			expectedError: false,
			expectedCount: 1,
		},
		{
			name:        "Scenario 3: Retrieve Favorited Articles by User",
			favoritedBy: &model.User{Model: gorm.Model{ID: 1}},
			limit:       10,
			offset:      0,
			mockSetup: func() {
				mock.ExpectQuery(`SELECT article_id FROM "favorite_articles" WHERE .+`).
					WillReturnRows(sqlmock.NewRows([]string{"article_id"}).AddRow(1))
				mock.ExpectQuery(`SELECT .* FROM "articles" WHERE id IN (.+)`).
					WillReturnRows(sqlmock.NewRows([]string{"id", "title"}).AddRow(1, "Favorited Article"))
			},
			expectedError: false,
			expectedCount: 1,
		},
		{
			name:   "Scenario 4: Retrieve Articles with Pagination",
			limit:  1,
			offset: 1,
			mockSetup: func() {
				mock.ExpectQuery(`SELECT .* FROM "articles"`).
					WillReturnRows(sqlmock.NewRows([]string{"id", "title"}).AddRow(2, "Paged Article"))
			},
			expectedError: false,
			expectedCount: 1,
		},
		{
			name:    "Scenario 5: Handle Empty Result Set Gracefully",
			tagName: "nonexistent",
			mockSetup: func() {
				mock.ExpectQuery(`SELECT .* FROM "articles"`).
					WithArgs("nonexistent").
					WillReturnRows(sqlmock.NewRows([]string{"id", "title"}))
			},
			expectedError: false,
			expectedCount: 0,
		},
		{
			name: "Scenario 6: Error Handling for Database Issues",
			mockSetup: func() {
				mock.ExpectQuery(`SELECT .* FROM "articles"`).
					WillReturnError(gorm.ErrCantStartTransaction)
			},
			expectedError: true,
			expectedCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			articles, err := articleStore.GetArticles(tt.tagName, tt.username, tt.favoritedBy, tt.limit, tt.offset)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedCount, len(articles), "expected articles count does not match")
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}

}

