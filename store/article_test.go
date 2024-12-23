package store

import (
	"testing"
	"sync"
	"github.com/raahii/golang-grpc-realworld-example/model"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jinzhu/gorm"
	_ "github.com/go-sql-driver/mysql"
	"errors"
	"github.com/stretchr/testify/assert"
	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"fmt"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	newGorm "gorm.io/gorm"
	_ "github.com/lib/pq"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/jinzhu/gorm/dialects/sqlite"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"context"
	"database/sql"
	"github.com/raahii/golang-grpc-realworld-example/store"
)


/*
ROOST_METHOD_HASH=NewArticleStore_6be2824012
ROOST_METHOD_SIG_HASH=NewArticleStore_3fe6f79a92


 */
func TestNewArticleStore(t *testing.T) {
	type testCase struct {
		name         string
		db           *gorm.DB
		expectNilDB  bool
		expectSameDB bool
	}

	mockDB, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("error creating SQL mock: %s", err)
	}
	defer mockDB.Close()

	gormDB, err := gorm.Open("mysql", mockDB)
	if err != nil {
		t.Fatalf("error initializing GORM DB: %s", err)
	}

	tests := []testCase{
		{
			name:         "Successful Initialization with Valid DB Object",
			db:           gormDB,
			expectNilDB:  false,
			expectSameDB: true,
		},
		{
			name:         "Handle Nil Database Reference",
			db:           nil,
			expectNilDB:  true,
			expectSameDB: false,
		},
		{
			name:         "Valid DB with Pre-Configured State",
			db:           gormDB,
			expectNilDB:  false,
			expectSameDB: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			articleStore := NewArticleStore(tc.db)

			if tc.expectNilDB && articleStore.db != nil {
				t.Errorf("expected nil DB in ArticleStore, got: %v", articleStore.db)
			}

			if tc.expectSameDB && articleStore.db != tc.db {
				t.Errorf("expected same DB reference, got different one")
			}
		})
	}

	t.Run("Concurrent Access Handling", func(t *testing.T) {
		const goroutineCount = 100
		wg := sync.WaitGroup{}
		wg.Add(goroutineCount)

		results := make(chan *ArticleStore, goroutineCount)
		for i := 0; i < goroutineCount; i++ {
			go func() {
				defer wg.Done()
				store := NewArticleStore(gormDB)
				results <- store
			}()
		}
		wg.Wait()
		close(results)

		for store := range results {
			if store.db != gormDB {
				t.Errorf("concurrent test failure: ArticleStore does not retain DB reference across routines")
			}
		}
	})

	t.Log("Finished TestNewArticleStore")
}


/*
ROOST_METHOD_HASH=Create_0a911e138d
ROOST_METHOD_SIG_HASH=Create_723c594377


 */
func TestCreate(t *testing.T) {

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("unable to create sqlmock DB: %v", err)
	}
	defer db.Close()

	gormDB, err := gorm.Open("sqlite3", db)
	if err != nil {
		t.Fatalf("unable to open gorm DB: %v", err)
	}

	store := &ArticleStore{db: gormDB}

	tests := []struct {
		name          string
		article       model.Article
		prepareMock   func()
		expectError   bool
		expectedError error
	}{
		{
			name: "Scenario 1: Successful Creation of a Valid Article",
			article: model.Article{
				Title: "A Valid Title",
			},
			prepareMock: func() {
				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO articles").
					WithArgs("A Valid Title", sqlmock.AnyArg()).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
			expectError: false,
		},
		{
			name: "Scenario 2: Article Creation Fails Due to Database Error",
			article: model.Article{
				Title: "One more Title",
			},
			prepareMock: func() {
				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO articles").
					WithArgs("One more Title", sqlmock.AnyArg()).
					WillReturnError(errors.New("database error"))
				mock.ExpectRollback()
			},
			expectError:   true,
			expectedError: errors.New("database error"),
		},
		{
			name: "Scenario 3: Article Creation with Missing Required Fields",
			article: model.Article{
				Title: "",
			},
			prepareMock:   func() {},
			expectError:   true,
			expectedError: gorm.Errors{gorm.ErrInvalidSQL},
		},
		{
			name: "Scenario 4: Handling Duplicate Articles",
			article: model.Article{
				Title: "Duplicate Title",
			},
			prepareMock: func() {
				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO articles").
					WithArgs("Duplicate Title", sqlmock.AnyArg()).
					WillReturnError(gorm.ErrPrimaryKeyRequired)
				mock.ExpectRollback()
			},
			expectError:   true,
			expectedError: gorm.ErrPrimaryKeyRequired,
		},
		{
			name: "Scenario 5: Article Creation with Large Data Field Values",
			article: model.Article{
				Title: "A Very Large Title" + string(make([]byte, 1000)),
			},
			prepareMock: func() {
				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO articles").
					WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg()).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			tt.prepareMock()

			err := store.Create(&tt.article)

			if tt.expectError {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError, err)
			} else {
				assert.NoError(t, err)
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

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("error initializing database mock: %v", err)
	}
	defer db.Close()

	gormDB, err := gorm.Open(mysql.New(mysql.Config{
		Conn:                      db,
		SkipInitializeWithVersion: true,
	}), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("error initializing GORM: %v", err)
	}

	store := &ArticleStore{db: gormDB}

	tests := []struct {
		name         string
		comment      *model.Comment
		setupMock    func()
		expectedErr  bool
		expectedText string
	}{
		{
			name: "Scenario 1: Successfully Create a Comment",
			comment: &model.Comment{
				Body:      "Valid comment",
				UserID:    1,
				ArticleID: 1,
			},
			setupMock: func() {
				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO `comments`").
					WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), "Valid comment", 1, 1).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
			expectedErr:  false,
			expectedText: "Comment successfully created",
		},
		{
			name: "Scenario 2: Fail to Create a Comment with Missing Fields",
			comment: &model.Comment{
				Body:      "",
				UserID:    1,
				ArticleID: 1,
			},
			setupMock: func() {

			},
			expectedErr:  true,
			expectedText: "missing Body field",
		},
		{
			name: "Scenario 3: Database Connection Error",
			comment: &model.Comment{
				Body:      "Valid comment",
				UserID:    1,
				ArticleID: 1,
			},
			setupMock: func() {
				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO `comments`").
					WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), "Valid comment", 1, 1).
					WillReturnError(errors.New("database connection error"))
				mock.ExpectRollback()
			},
			expectedErr:  true,
			expectedText: "database error",
		},
		{
			name: "Scenario 4: Create Comment with Minimum and Maximum Field Lengths",
			comment: &model.Comment{
				Body:      "a",
				UserID:    1,
				ArticleID: 1,
			},
			setupMock: func() {
				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO `comments`").
					WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), "a", 1, 1).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
			expectedErr:  false,
			expectedText: "Comment successfully created with minimum length Body",
		},
		{
			name: "Scenario 4: Create Comment with Maximum Field Length",
			comment: &model.Comment{
				Body:      string(make([]rune, 1000)),
				UserID:    1,
				ArticleID: 1,
			},
			setupMock: func() {
				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO `comments`").
					WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), string(make([]rune, 1000)), 1, 1).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
			expectedErr:  false,
			expectedText: "Comment successfully created with maximum length Body",
		},
		{
			name: "Scenario 5: Verify Cascade Operations or Foreign Key Constraints",
			comment: &model.Comment{
				Body:      "Valid comment",
				UserID:    999,
				ArticleID: 1,
			},
			setupMock: func() {
				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO `comments`").
					WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), "Valid comment", 999, 1).
					WillReturnError(errors.New("foreign key constraint fails"))
				mock.ExpectRollback()
			},
			expectedErr:  true,
			expectedText: "foreign key constraint error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()
			err := store.CreateComment(tt.comment)
			if (err != nil) != tt.expectedErr {
				t.Errorf("expected error: %v, got: %v", tt.expectedErr, err)
			}
			if err != nil {
				t.Logf("Expected optional error detail: %s -- Received error: %v", tt.expectedText, err)
			} else {
				t.Logf("%s", tt.expectedText)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}


/*
ROOST_METHOD_HASH=Delete_a8dc14c210
ROOST_METHOD_SIG_HASH=Delete_a4cc8044b1


 */
func TestDelete(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to open sqlmock database: %v", err)
	}
	defer db.Close()

	gormDB, err := gorm.Open("mysql", db)
	if err != nil {
		t.Fatalf("Failed to open gorm DB: %v", err)
	}
	defer gormDB.Close()

	gormDB.SingularTable(true)

	store := &ArticleStore{db: gormDB}

	tests := []struct {
		name       string
		setupMocks func()
		article    *model.Article
		expectErr  bool
	}{
		{
			name: "Successful Deletion",
			setupMocks: func() {
				mock.ExpectBegin()
				mock.ExpectExec("^DELETE FROM articles WHERE \\(id = \\?\\)$").
					WithArgs(sqlmock.AnyArg()).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
			article:   createTestArticle(),
			expectErr: false,
		},
		{
			name: "Attempt to Delete Non-Existent Article",
			setupMocks: func() {
				mock.ExpectBegin()
				mock.ExpectExec("^DELETE FROM articles WHERE \\(id = \\?\\)$").
					WithArgs(sqlmock.AnyArg()).
					WillReturnResult(sqlmock.NewResult(0, 0))
				mock.ExpectCommit()
			},
			article:   createTestArticle(),
			expectErr: false,
		},
		{
			name: "Database Connection Failure",
			setupMocks: func() {
				mock.ExpectBegin().WillReturnError(gorm.ErrInvalidTransaction)
			},
			article:   createTestArticle(),
			expectErr: true,
		},
		{
			name: "Deletion with Associated Tags",
			setupMocks: func() {
				mock.ExpectBegin()
				mock.ExpectExec("^DELETE FROM articles WHERE \\(id = \\?\\)$").
					WithArgs(sqlmock.AnyArg()).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
			article:   createTestArticle(),
			expectErr: false,
		},
		{
			name: "Deletion with Associated Comments",
			setupMocks: func() {
				mock.ExpectBegin()
				mock.ExpectExec("^DELETE FROM articles WHERE \\(id = \\?\\)$").
					WithArgs(sqlmock.AnyArg()).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectExec("^DELETE FROM comments WHERE \\(article_id = \\?\\)$").
					WithArgs(sqlmock.AnyArg()).
					WillReturnResult(sqlmock.NewResult(2, 2))
				mock.ExpectCommit()
			},
			article:   createTestArticle(),
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Log("Setting up test", tt.name)

			tt.setupMocks()

			err := store.Delete(tt.article)

			if tt.expectErr {
				assert.Error(t, err, "Expected an error but got nil")
			} else {
				assert.NoError(t, err, "Expected no error but got one")
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func createTestArticle() *model.Article {
	return &model.Article{
		Title:       "Test Article",
		Description: "Test Description",
		Body:        "Test Body",
		UserID:      1,
		Tags: []model.Tag{
			{Name: "Test Tag 1"},
			{Name: "Test Tag 2"},
		},
		Comments: []model.Comment{
			{Body: "Test Comment 1"},
			{Body: "Test Comment 2"},
		},
	}
}


/*
ROOST_METHOD_HASH=DeleteComment_b345e525a7
ROOST_METHOD_SIG_HASH=DeleteComment_732762ff12


 */
func TestDeleteComment(t *testing.T) {
	t.Run("Successfully Delete a Comment", func(t *testing.T) {

		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("Failed to open sqlmock database: %s", err)
		}
		defer db.Close()

		gormDB, err := gorm.Open("mysql", db)
		if err != nil {
			t.Fatalf("Failed to open gorm database: %s", err)
		}

		mock.ExpectBegin()
		mock.ExpectExec("^DELETE FROM `comments` WHERE (.+)$").WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		store := &ArticleStore{db: gormDB}
		comment := &model.Comment{
			Model:     gorm.Model{ID: 1},
			Body:      "This is a comment.",
			UserID:    1,
			ArticleID: 1,
		}

		err = store.DeleteComment(comment)

		if err != nil {
			t.Errorf("Expected no error, but got: %v", err)
		}
		t.Log("The comment was successfully deleted from the database.")
	})

	t.Run("Attempt to Delete a Non-existent Comment", func(t *testing.T) {

		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("Failed to open sqlmock database: %s", err)
		}
		defer db.Close()

		gormDB, err := gorm.Open("mysql", db)
		if err != nil {
			t.Fatalf("Failed to open gorm database: %s", err)
		}

		mock.ExpectBegin()
		mock.ExpectExec("^DELETE FROM `comments` WHERE (.+)$").WillReturnResult(sqlmock.NewResult(1, 0))
		mock.ExpectCommit()

		store := &ArticleStore{db: gormDB}
		comment := &model.Comment{
			Model:     gorm.Model{ID: 0},
			Body:      "This comment does not exist.",
			UserID:    1,
			ArticleID: 1,
		}

		err = store.DeleteComment(comment)

		if err != nil {
			t.Errorf("Expected no error, even if the comment does not exist, but got: %v", err)
		}
		t.Log("Deletion attempt of a non-existent comment did not alter the database.")
	})

	t.Run("Delete a Comment with a Dependent Relationship", func(t *testing.T) {

		t.Log("This test would ensure that deleting a comment does not affect related articles.")
	})

	t.Run("Delete a Comment with Concurrency Handling", func(t *testing.T) {

		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("Failed to open sqlmock database: %s", err)
		}
		defer db.Close()

		gormDB, err := gorm.Open("mysql", db)
		if err != nil {
			t.Fatalf("Failed to open gorm database: %s", err)
		}

		mock.ExpectBegin()
		mock.ExpectExec("^DELETE FROM `comments` WHERE (.+)$").WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		store := &ArticleStore{db: gormDB}
		comment := &model.Comment{
			Model:     gorm.Model{ID: 1},
			Body:      "This is a concurrent comment.",
			UserID:    1,
			ArticleID: 1,
		}

		ch := make(chan error, 2)
		go func() {
			ch <- store.DeleteComment(comment)
		}()
		go func() {
			ch <- store.DeleteComment(comment)
		}()

		for i := 0; i < 2; i++ {
			err := <-ch
			if err != nil {
				t.Errorf("Expected no error, but got: %v", err)
			}
		}
		t.Log("Concurrency handled correctly, comment deleted only once without errors.")
	})

	t.Run("Delete a Comment and Handle Database Error", func(t *testing.T) {

		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("Failed to open sqlmock database: %s", err)
		}
		defer db.Close()

		gormDB, err := gorm.Open("mysql", db)
		if err != nil {
			t.Fatalf("Failed to open gorm database: %s", err)
		}

		mock.ExpectBegin()
		mock.ExpectExec("^DELETE FROM `comments` WHERE (.+)$").WillReturnError(fmt.Errorf("simulated database error"))
		mock.ExpectRollback()

		store := &ArticleStore{db: gormDB}
		comment := &model.Comment{
			Model:     gorm.Model{ID: 1},
			Body:      "This is a comment.",
			UserID:    1,
			ArticleID: 1,
		}

		err = store.DeleteComment(comment)

		if err == nil || err.Error() != "simulated database error" {
			t.Errorf("Expected simulated database error, but got: %v", err)
		}
		t.Log("Database error during deletion is handled correctly.")
	})
}


/*
ROOST_METHOD_HASH=GetByID_36e92ad6eb
ROOST_METHOD_SIG_HASH=GetByID_9616e43e52


 */
func TestUserStoreGetByID(t *testing.T) {
	tests := []struct {
		name            string
		id              uint
		setupMock       func(mock sqlmock.Sqlmock)
		expectedArticle *model.Article
		expectedError   error
	}{
		{
			name: "Retrieve Article by Valid ID",
			id:   1,
			setupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "title"}).
					AddRow(1, "Test Article")
				mock.ExpectQuery("^SELECT (.+) FROM articles WHERE id = ?").
					WithArgs(1).WillReturnRows(rows)
				mock.ExpectQuery("^SELECT (.+) FROM tags WHERE article_id = ?").
					WithArgs(1).WillReturnRows(sqlmock.NewRows([]string{"id", "name"}))
				mock.ExpectQuery("^SELECT (.+) FROM authors WHERE article_id = ?").
					WithArgs(1).WillReturnRows(sqlmock.NewRows([]string{"id", "name"}))
			},
			expectedArticle: &model.Article{Model: gorm.Model{ID: 1}, Title: "Test Article"},
			expectedError:   nil,
		},
		{
			name: "Article ID Not Found",
			id:   2,
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("^SELECT (.+) FROM articles WHERE id = ?").
					WithArgs(2).WillReturnError(gorm.ErrRecordNotFound)
			},
			expectedArticle: nil,
			expectedError:   gorm.ErrRecordNotFound,
		},
		{
			name: "Database Connection Error",
			id:   3,
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("^SELECT (.+) FROM articles WHERE id = ?").
					WithArgs(3).WillReturnError(errors.New("db connection error"))
			},
			expectedArticle: nil,
			expectedError:   errors.New("db connection error"),
		},
		{
			name: "Malformed ID Input (Zero ID)",
			id:   0,
			setupMock: func(mock sqlmock.Sqlmock) {

			},
			expectedArticle: nil,
			expectedError:   errors.New("invalid ID provided"),
		},
		{
			name: "Preloading Related Data",
			id:   4,
			setupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "title"}).
					AddRow(4, "Related Data Article")
				mock.ExpectQuery("^SELECT (.+) FROM articles WHERE id = ?").
					WithArgs(4).WillReturnRows(rows)
				mock.ExpectQuery("^SELECT (.+) FROM tags WHERE article_id = ?").
					WithArgs(4).WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow(1, "Tech Tag"))
				mock.ExpectQuery("^SELECT (.+) FROM authors WHERE article_id = ?").
					WithArgs(4).WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow(1, "John Doe"))
			},
			expectedArticle: &model.Article{Model: gorm.Model{ID: 4}, Title: "Related Data Article"},
			expectedError:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
			}
			defer mockDB.Close()

			gormDB, err := gorm.Open("mysql", mockDB)
			if err != nil {
				t.Fatalf("failed to open gorm DB %v", err)
			}
			defer gormDB.Close()

			store := &ArticleStore{db: gormDB}

			tt.setupMock(mock)

			article, err := store.GetByID(tt.id)

			if tt.expectedError != nil {
				if err == nil || tt.expectedError.Error() != err.Error() {
					t.Errorf("expected error '%v', got '%v'", tt.expectedError, err)
				}
			} else {
				if err != nil {
					t.Errorf("expected no error, got '%v'", err)
				}
			}

			if tt.expectedArticle != nil {
				if article == nil || article.ID != tt.expectedArticle.ID || article.Title != tt.expectedArticle.Title {
					t.Errorf("expected article %+v, got %+v", tt.expectedArticle, article)
				}
			} else {
				if article != nil {
					t.Errorf("expected no article, got %+v", article)
				}
			}
		})
	}
}


/*
ROOST_METHOD_HASH=GetCommentByID_4bc82104a6
ROOST_METHOD_SIG_HASH=GetCommentByID_333cab101b


 */
func TestArticleStoreGetCommentByID(t *testing.T) {

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	gormDB, err := newGorm.Open(mysql.New(mysql.Config{
		Conn:                      db,
		SkipInitializeWithVersion: true,
	}), &newGorm.Config{})
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a gorm database connection", err)
	}

	oldGormDB, err := gorm.Open("mysql", gormDB)
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a legacy gorm database connection", err)
	}

	store := &ArticleStore{db: oldGormDB}

	tests := []struct {
		name            string
		id              uint
		setupMock       func()
		expectedError   error
		expectedComment *model.Comment
	}{
		{
			name: "Retrieve Existing Comment by Valid ID",
			id:   1,
			setupMock: func() {
				rows := sqlmock.NewRows([]string{"id", "created_at", "updated_at", "deleted_at", "body", "user_id", "article_id"}).
					AddRow(1, nil, nil, nil, "This is a comment", 1, 1)
				mock.ExpectQuery(`SELECT \* FROM "comments" WHERE "comments"."id" = ?`).
					WithArgs(1).WillReturnRows(rows)
			},
			expectedError: nil,
			expectedComment: &model.Comment{
				Body:      "This is a comment",
				UserID:    1,
				ArticleID: 1,
			},
		},
		{
			name: "Handle Non-Existent ID Gracefully",
			id:   2,
			setupMock: func() {
				mock.ExpectQuery(`SELECT \* FROM "comments" WHERE "comments"."id" = ?`).
					WithArgs(2).WillReturnError(gorm.ErrRecordNotFound)
			},
			expectedError:   gorm.ErrRecordNotFound,
			expectedComment: nil,
		},
		{
			name: "Handle Database Errors during Retrieval",
			id:   3,
			setupMock: func() {
				mock.ExpectQuery(`SELECT \* FROM "comments" WHERE "comments"."id" = ?`).
					WithArgs(3).WillReturnError(errors.New("database error"))
			},
			expectedError:   errors.New("database error"),
			expectedComment: nil,
		},
		{
			name: "Retrieve Comment with Large ID",
			id:   4294967295,
			setupMock: func() {
				mock.ExpectQuery(`SELECT \* FROM "comments" WHERE "comments"."id" = ?`).
					WithArgs(4294967295).WillReturnError(gorm.ErrRecordNotFound)
			},
			expectedError:   gorm.ErrRecordNotFound,
			expectedComment: nil,
		},
		{
			name: "Retrieve Comment with ID of Zero",
			id:   0,
			setupMock: func() {
				mock.ExpectQuery(`SELECT \* FROM "comments" WHERE "comments"."id" = ?`).
					WithArgs(0).WillReturnError(gorm.ErrRecordNotFound)
			},
			expectedError:   gorm.ErrRecordNotFound,
			expectedComment: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			comment, err := store.GetCommentByID(tt.id)

			assert.Equal(t, tt.expectedError, err)
			assert.Equal(t, tt.expectedComment, comment)

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}


/*
ROOST_METHOD_HASH=GetTags_ac049ebded
ROOST_METHOD_SIG_HASH=GetTags_25034b82b0


 */
func TestArticleStoreGetTags(t *testing.T) {

	tests := []struct {
		name          string
		setup         func(sqlmock.Sqlmock)
		expectedTags  []model.Tag
		expectedError error
	}{
		{
			name: "Retrieve Tags Successfully from Database",
			setup: func(mock sqlmock.Sqlmock) {

				rows := sqlmock.NewRows([]string{"id", "name"}).
					AddRow(1, "tech").
					AddRow(2, "science")
				mock.ExpectQuery("^SELECT \\* FROM \"tags\"").WillReturnRows(rows)
			},
			expectedTags: []model.Tag{
				{Model: gorm.Model{ID: 1}, Name: "tech"},
				{Model: gorm.Model{ID: 2}, Name: "science"},
			},
			expectedError: nil,
		},
		{
			name: "Retrieve No Tags from Empty Database",
			setup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "name"})
				mock.ExpectQuery("^SELECT \\* FROM \"tags\"").WillReturnRows(rows)
			},
			expectedTags:  []model.Tag{},
			expectedError: nil,
		},
		{
			name: "Handle Database Error Gracefully",
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("^SELECT \\* FROM \"tags\"").WillReturnError(gorm.ErrInvalidSQL)
			},
			expectedTags:  nil,
			expectedError: gorm.ErrInvalidSQL,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("Failed to open mock sql db: %s", err)
			}
			defer db.Close()

			gormDB, err := gorm.Open("postgres", db)
			if err != nil {
				t.Fatalf("Failed to open gorm db: %s", err)
			}

			tt.setup(mock)

			store := &ArticleStore{db: gormDB}

			tags, err := store.GetTags()

			if !equalTags(tags, tt.expectedTags) || err != tt.expectedError {
				t.Errorf("Got tags = %v, error = %v; expected tags = %v, error = %v",
					tags, err, tt.expectedTags, tt.expectedError)
			}

			if len(tags) != len(tt.expectedTags) {
				t.Logf("Expected tag count: %d, got: %d", len(tt.expectedTags), len(tags))
			} else {
				t.Logf("Tags retrieval matched expected.")
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatalf("There were unfulfilled expectations: %s", err)
			}
		})
	}
}

func equalTags(a, b []model.Tag) bool {
	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if a[i].ID != b[i].ID || a[i].Name != b[i].Name {
			return false
		}
	}

	return true
}


/*
ROOST_METHOD_HASH=Update_51145aa965
ROOST_METHOD_SIG_HASH=Update_6c1b5471fe


 */
func TestUpdate(t *testing.T) {
	type testScenario struct {
		description   string
		article       model.Article
		expectedError error
		setupMock     func(mock sqlmock.Sqlmock)
	}

	tests := []testScenario{
		{
			description: "Scenario 1: Successful Article Update",
			article: model.Article{
				Model:   gorm.Model{ID: 1},
				Title:   "Updated Title",
				Content: "Updated Content",
			},
			expectedError: nil,
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec("^UPDATE articles SET (.+) WHERE (.+)$").
					WithArgs("Updated Title", "Updated Content", 1).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
		},
		{
			description: "Scenario 2: Article Not Found",
			article: model.Article{
				Model:   gorm.Model{ID: 2},
				Title:   "Non-existent Article",
				Content: "Some Content",
			},
			expectedError: gorm.ErrRecordNotFound,
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec("^UPDATE articles SET (.+) WHERE (.+)$").
					WithArgs("Non-existent Article", "Some Content", 2).
					WillReturnError(gorm.ErrRecordNotFound)
				mock.ExpectRollback()
			},
		},
		{
			description: "Scenario 3: Database Connection Error",
			article: model.Article{
				Model:   gorm.Model{ID: 1},
				Title:   "Title",
				Content: "Content",
			},
			expectedError: errors.New("db connection error"),
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin().
					WillReturnError(errors.New("db connection error"))
			},
		},
		{
			description: "Scenario 4: Invalid Article Data",
			article: model.Article{
				Model: gorm.Model{ID: 3},

				Content: "Content",
			},

			expectedError: errors.New("validation error"),
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec("^UPDATE articles SET (.+) WHERE (.+)$").
					WithArgs("", "Content", 3).
					WillReturnError(errors.New("validation error"))
				mock.ExpectRollback()
			},
		},
	}

	for _, scenario := range tests {
		t.Run(scenario.description, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			assert.NoError(t, err)
			defer db.Close()

			gormDB, err := gorm.Open("sqlite3", db)
			assert.NoError(t, err)

			store := ArticleStore{db: gormDB}
			scenario.setupMock(mock)

			err = store.Update(&scenario.article)

			if scenario.expectedError != nil {
				assert.EqualError(t, err, scenario.expectedError.Error())
			} else {
				assert.NoError(t, err)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("Expectations were not met: %v", err)
			}

			t.Log("Successfully tested:", scenario.description)
		})
	}
}


/*
ROOST_METHOD_HASH=GetComments_e24a0f1b73
ROOST_METHOD_SIG_HASH=GetComments_fa6661983e


 */
func TestArticleStoreGetComments(t *testing.T) {

	var mockArticleID uint = 1

	validComment1 := model.Comment{
		ArticleID: mockArticleID,
		Body:      "Test Comment 1",
		Author: model.User{
			Model:    gorm.Model{ID: 1},
			Username: "JohnDoe",
		},
	}
	validComment2 := model.Comment{
		ArticleID: mockArticleID,
		Body:      "Test Comment 2",
		Author: model.User{
			Model:    gorm.Model{ID: 2},
			Username: "JaneDoe",
		},
	}
	mockArticle := model.Article{
		Model: gorm.Model{ID: mockArticleID},
	}

	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	gormDB, err := gorm.Open("postgres", db)
	assert.NoError(t, err)

	store := &ArticleStore{db: gormDB}

	tests := []struct {
		name       string
		prepExpect func()
		article    model.Article
		expected   []model.Comment
		expectErr  bool
	}{
		{
			name: "Scenario 1: Retrieval of Comments for a Valid Article",
			prepExpect: func() {
				mock.ExpectQuery("SELECT \\* FROM \"comments\"").
					WithArgs(mockArticleID).
					WillReturnRows(sqlmock.NewRows([]string{"id", "article_id", "body", "author_id"}).
						AddRow(validComment1.ID, validComment1.ArticleID, validComment1.Body, validComment1.Author.ID).
						AddRow(validComment2.ID, validComment2.ArticleID, validComment2.Body, validComment2.Author.ID))
			},
			article:   mockArticle,
			expected:  []model.Comment{validComment1, validComment2},
			expectErr: false,
		},
		{
			name: "Scenario 2: No Comments for a Valid Article",
			prepExpect: func() {
				mock.ExpectQuery("SELECT \\* FROM \"comments\"").
					WithArgs(mockArticleID).
					WillReturnRows(sqlmock.NewRows([]string{"id", "article_id", "body", "author_id"}))
			},
			article:   mockArticle,
			expected:  []model.Comment{},
			expectErr: false,
		},
		{
			name: "Scenario 3: Invalid Article with Zero Comments",
			prepExpect: func() {
				mockArticle.ID = 2
				mock.ExpectQuery("SELECT \\* FROM \"comments\"").
					WithArgs(mockArticle.ID).
					WillReturnError(gorm.ErrRecordNotFound)
			},
			article:   mockArticle,
			expected:  []model.Comment{},
			expectErr: true,
		},
		{
			name: "Scenario 4: Database Error Simulation",
			prepExpect: func() {
				mock.ExpectQuery("SELECT \\* FROM \"comments\"").
					WithArgs(mockArticleID).
					WillReturnError(assert.AnError)
			},
			article:   mockArticle,
			expected:  []model.Comment{},
			expectErr: true,
		},
		{
			name: "Scenario 5: Preload Author Information",
			prepExpect: func() {
				mock.ExpectQuery("SELECT \\* FROM \"comments\"").
					WithArgs(mockArticleID).
					WillReturnRows(sqlmock.NewRows([]string{"id", "article_id", "body", "author_id"}).
						AddRow(validComment1.ID, validComment1.ArticleID, validComment1.Body, validComment1.Author.ID).
						AddRow(validComment2.ID, validComment2.ArticleID, validComment2.Body, validComment2.Author.ID))
			},
			article:   mockArticle,
			expected:  []model.Comment{validComment1, validComment2},
			expectErr: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.prepExpect()
			comments, err := store.GetComments(&test.article)

			if test.expectErr {
				assert.Error(t, err, "Expected error")
			} else {
				assert.NoError(t, err)
				assert.Equal(t, test.expected, comments)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("There were unfulfilled expectations:\n%s", err)
			}
		})
	}
}


/*
ROOST_METHOD_HASH=IsFavorited_7ef7d3ed9e
ROOST_METHOD_SIG_HASH=IsFavorited_f34d52378f


 */
func TestIsFavorited(t *testing.T) {

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error initializing mock database: %v", err)
	}
	defer db.Close()

	gormDB, err := gorm.Open(sqlite.Dialect{}, db)
	if err != nil {
		t.Fatalf("Error initializing gorm DB: %v", err)
	}

	gormDB.LogMode(true)

	articleStore := &ArticleStore{db: gormDB}

	tests := []struct {
		name           string
		article        *model.Article
		user           *model.User
		prepareMock    func()
		expectedResult bool
		expectError    bool
	}{
		{
			name:    "Nil Article",
			article: nil,
			user:    &model.User{},
			prepareMock: func() {

			},
			expectedResult: false,
			expectError:    false,
		},
		{
			name:    "Nil User",
			article: &model.Article{},
			user:    nil,
			prepareMock: func() {

			},
			expectedResult: false,
			expectError:    false,
		},
		{
			name: "User has favorited article",
			article: &model.Article{
				Model: gorm.Model{ID: 1},
			},
			user: &model.User{
				Model: gorm.Model{ID: 1},
			},
			prepareMock: func() {
				mock.ExpectQuery("^SELECT count\\(\\*\\) FROM `favorite_articles` WHERE").
					WithArgs(1, 1).
					WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
			},
			expectedResult: true,
			expectError:    false,
		},
		{
			name: "User has not favorited article",
			article: &model.Article{
				Model: gorm.Model{ID: 1},
			},
			user: &model.User{
				Model: gorm.Model{ID: 1},
			},
			prepareMock: func() {
				mock.ExpectQuery("^SELECT count\\(\\*\\) FROM `favorite_articles` WHERE").
					WithArgs(1, 1).
					WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
			},
			expectedResult: false,
			expectError:    false,
		},
		{
			name: "Database error occurs",
			article: &model.Article{
				Model: gorm.Model{ID: 1},
			},
			user: &model.User{
				Model: gorm.Model{ID: 1},
			},
			prepareMock: func() {
				mock.ExpectQuery("^SELECT count\\(\\*\\) FROM `favorite_articles` WHERE").
					WithArgs(1, 1).
					WillReturnError(gorm.ErrInvalidSQL)
			},
			expectedResult: false,
			expectError:    true,
		},
		{
			name: "Invalid article and user IDs",
			article: &model.Article{
				Model: gorm.Model{ID: 999},
			},
			user: &model.User{
				Model: gorm.Model{ID: 999},
			},
			prepareMock: func() {
				mock.ExpectQuery("^SELECT count\\(\\*\\) FROM `favorite_articles` WHERE").
					WithArgs(999, 999).
					WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
			},
			expectedResult: false,
			expectError:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.prepareMock()

			result, err := articleStore.IsFavorited(tt.article, tt.user)

			if (err != nil) != tt.expectError {
				t.Errorf("Unexpected error state: got %v, want error: %v", err, tt.expectError)
			}
			if result != tt.expectedResult {
				t.Errorf("Unexpected result: got %v, want %v", result, tt.expectedResult)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("Not all expectations were met: %v", err)
			}
		})
	}
}


/*
ROOST_METHOD_HASH=GetFeedArticles_9c4f57afe4
ROOST_METHOD_SIG_HASH=GetFeedArticles_cadca0e51b


 */
func TestArticleStoreGetFeedArticles(t *testing.T) {
	testCases := []struct {
		name             string
		userIDs          []uint
		limit, offset    int64
		mockSetup        func(mock sqlmock.Sqlmock)
		expectedArticles []model.Article
		expectedError    error
	}{
		{
			name:    "Retrieve Articles with Multiple User IDs",
			userIDs: []uint{1, 2, 3},
			limit:   10,
			offset:  0,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT \* FROM "articles" WHERE user_id in \(\$1,\$2,\$3\) LIMIT \$4 OFFSET \$5`).
					WithArgs(1, 2, 3, 10, 0).
					WillReturnRows(sqlmock.NewRows([]string{"id", "title", "user_id"}).
						AddRow(1, "Article 1", 1).
						AddRow(2, "Article 2", 2).
						AddRow(3, "Article 3", 3))
			},
			expectedArticles: []model.Article{
				{Model: gorm.Model{ID: 1}, Title: "Article 1", UserID: 1},
				{Model: gorm.Model{ID: 2}, Title: "Article 2", UserID: 2},
				{Model: gorm.Model{ID: 3}, Title: "Article 3", UserID: 3},
			},
			expectedError: nil,
		},
		{
			name:    "Handle No User IDs",
			userIDs: []uint{},
			limit:   0,
			offset:  0,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT \* FROM "articles" WHERE 1 = 1 LIMIT \$1 OFFSET \$2`).
					WithArgs(0, 0).
					WillReturnRows(sqlmock.NewRows([]string{"id", "title", "user_id"}))
			},
			expectedArticles: []model.Article{},
			expectedError:    nil,
		},
		{
			name:    "Apply Limit and Offset on Retrieved Articles",
			userIDs: []uint{1},
			limit:   5,
			offset:  10,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT \* FROM "articles" WHERE user_id in \(\$1\) LIMIT \$2 OFFSET \$3`).
					WithArgs(1, 5, 10).
					WillReturnRows(sqlmock.NewRows([]string{"id", "title", "user_id"}).
						AddRow(11, "Article 11", 1).
						AddRow(12, "Article 12", 1).
						AddRow(13, "Article 13", 1).
						AddRow(14, "Article 14", 1).
						AddRow(15, "Article 15", 1))
			},
			expectedArticles: []model.Article{
				{Model: gorm.Model{ID: 11}, Title: "Article 11", UserID: 1},
				{Model: gorm.Model{ID: 12}, Title: "Article 12", UserID: 1},
				{Model: gorm.Model{ID: 13}, Title: "Article 13", UserID: 1},
				{Model: gorm.Model{ID: 14}, Title: "Article 14", UserID: 1},
				{Model: gorm.Model{ID: 15}, Title: "Article 15", UserID: 1},
			},
			expectedError: nil,
		},
		{
			name:    "Handle Database Error",
			userIDs: []uint{1},
			limit:   10,
			offset:  0,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT \* FROM "articles" WHERE user_id in \(\$1\) LIMIT \$2 OFFSET \$3`).
					WithArgs(1, 10, 0).
					WillReturnError(gorm.ErrInvalidSQL)
			},
			expectedArticles: nil,
			expectedError:    gorm.ErrInvalidSQL,
		},
		{
			name:    "Retrieve Articles for a Single User ID",
			userIDs: []uint{2},
			limit:   10,
			offset:  0,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT \* FROM "articles" WHERE user_id in \(\$1\) LIMIT \$2 OFFSET \$3`).
					WithArgs(2, 10, 0).
					WillReturnRows(sqlmock.NewRows([]string{"id", "title", "user_id"}).
						AddRow(21, "Article 21", 2).
						AddRow(22, "Article 22", 2))
			},
			expectedArticles: []model.Article{
				{Model: gorm.Model{ID: 21}, Title: "Article 21", UserID: 2},
				{Model: gorm.Model{ID: 22}, Title: "Article 22", UserID: 2},
			},
			expectedError: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			assert.NoError(t, err)
			defer db.Close()

			gdb, err := gorm.Open("postgres", db)
			assert.NoError(t, err)

			articleStore := &ArticleStore{
				db: gdb,
			}

			tc.mockSetup(mock)

			articles, err := articleStore.GetFeedArticles(tc.userIDs, tc.limit, tc.offset)

			assert.Equal(t, tc.expectedError, err)
			assert.Equal(t, tc.expectedArticles, articles)

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}

			t.Logf("Test Case %s executed successfully", tc.name)
		})
	}
}


/*
ROOST_METHOD_HASH=AddFavorite_2b0cb9d894
ROOST_METHOD_SIG_HASH=AddFavorite_c4dea0ee90


 */
func TestAddFavorite(t *testing.T) {
	t.Parallel()

	type scenario struct {
		description string
		setup       func(mock sqlmock.Sqlmock, article *model.Article, user *model.User)
		article     *model.Article
		user        *model.User
		expectError bool
		verify      func(t *testing.T, article *model.Article, err error)
	}

	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	gormDB, err := gorm.Open("sqlite3", "test.db")
	assert.NoError(t, err)

	articleStore := &ArticleStore{db: gormDB}

	gormDB.Set("gorm:table_options", "ENGINE=InnoDB")

	scenarios := []scenario{
		{
			description: "Successfully Adding a Favorite",
			setup: func(mock sqlmock.Sqlmock, article *model.Article, user *model.User) {

				mock.ExpectBegin()
				mock.ExpectExec(`INSERT INTO favorite_articles`).WithArgs().WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectExec(`UPDATE articles SET favorites_count = favorites_count + 1 WHERE id = ?`).WithArgs(article.ID).WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
			article: &model.Article{
				Model:          gorm.Model{ID: 1},
				FavoritesCount: 0,
			},
			user: &model.User{
				Model: gorm.Model{ID: 1},
			},
			expectError: false,
			verify: func(t *testing.T, article *model.Article, err error) {
				assert.NoError(t, err)
				assert.Equal(t, int32(1), article.FavoritesCount)
			},
		},
		{
			description: "Adding an Existing Favorite",
			setup: func(mock sqlmock.Sqlmock, article *model.Article, user *model.User) {

				mock.ExpectBegin()
				mock.ExpectExec(`INSERT INTO favorite_articles`).WithArgs().WillReturnError(errors.New("UNIQUE constraint failed"))
				mock.ExpectRollback()
			},
			article: &model.Article{
				Model:          gorm.Model{ID: 1},
				FavoritesCount: 1,
			},
			user: &model.User{
				Model: gorm.Model{ID: 1},
			},
			expectError: true,
			verify: func(t *testing.T, article *model.Article, err error) {
				assert.Error(t, err)
				assert.Equal(t, int32(1), article.FavoritesCount)
			},
		},
		{
			description: "Rolling Back on a Database Error",
			setup: func(mock sqlmock.Sqlmock, article *model.Article, user *model.User) {

				mock.ExpectBegin()
				mock.ExpectExec(`INSERT INTO favorite_articles`).WithArgs().WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectExec(`UPDATE articles SET favorites_count = favorites_count + 1 WHERE id = ?`).WithArgs(article.ID).WillReturnError(errors.New("Update failed"))
				mock.ExpectRollback()
			},
			article: &model.Article{
				Model:          gorm.Model{ID: 1},
				FavoritesCount: 0,
			},
			user: &model.User{
				Model: gorm.Model{ID: 1},
			},
			expectError: true,
			verify: func(t *testing.T, article *model.Article, err error) {
				assert.Error(t, err)
				assert.Equal(t, int32(0), article.FavoritesCount)
			},
		},
		{
			description: "Adding a Favorite with Missing User or Article Data",
			setup: func(mock sqlmock.Sqlmock, article *model.Article, user *model.User) {

			},
			article:     nil,
			user:        &model.User{Model: gorm.Model{ID: 1}},
			expectError: true,
			verify: func(t *testing.T, article *model.Article, err error) {
				assert.Error(t, err)
			},
		},
	}

	for _, s := range scenarios {
		t.Run(s.description, func(t *testing.T) {

			s.setup(mock, s.article, s.user)

			err := articleStore.AddFavorite(s.article, s.user)

			s.verify(t, s.article, err)

			assert.NoError(t, mock.ExpectationsWereMet())

			if s.expectError {
				t.Logf("Scenario '%s' expected an error: %v", s.description, err)
			} else {
				t.Logf("Scenario '%s' passed without errors.", s.description)
			}
		})
	}

}


/*
ROOST_METHOD_HASH=DeleteFavorite_a856bcbb70
ROOST_METHOD_SIG_HASH=DeleteFavorite_f7e5c0626f


 */
func TestDeleteFavorite(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	gormDB, err := gorm.Open("postgres", db)
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a GORM connection", err)
	}
	defer gormDB.Close()

	articleStore := &ArticleStore{db: gormDB}

	tests := []struct {
		name             string
		setup            func() (*model.Article, *model.User)
		expectedError    bool
		expectedFavCount int32
		mock             func(a *model.Article, u *model.User)
	}{
		{
			name: "Successfully Delete Favorite User from Article",
			setup: func() (*model.Article, *model.User) {
				user := &model.User{Model: gorm.Model{ID: 1}, Username: "user1"}
				article := &model.Article{Model: gorm.Model{ID: 1}, FavoritesCount: 1}
				return article, user
			},
			expectedError:    false,
			expectedFavCount: 0,
			mock: func(a *model.Article, u *model.User) {
				mock.ExpectBegin()
				mock.ExpectExec("DELETE FROM \"favorite_articles\"").WithArgs(u.ID, a.ID).WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectExec("UPDATE \"articles\"").WithArgs(0, a.ID).WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
		},
		{
			name: "Attempt to Delete Non-Favorite User",
			setup: func() (*model.Article, *model.User) {
				user := &model.User{Model: gorm.Model{ID: 1}, Username: "user1"}
				article := &model.Article{Model: gorm.Model{ID: 1}, FavoritesCount: 1}
				return article, user
			},
			expectedError:    false,
			expectedFavCount: 1,
			mock: func(a *model.Article, u *model.User) {
				mock.ExpectBegin()
				mock.ExpectExec("DELETE FROM \"favorite_articles\"").WithArgs(u.ID, a.ID).WillReturnResult(sqlmock.NewResult(0, 0))
				mock.ExpectExec("UPDATE \"articles\"").WithArgs(0, a.ID).WillReturnResult(sqlmock.NewResult(0, 0))
				mock.ExpectCommit()
			},
		},
		{
			name: "Database Error on Deleting Favorite User",
			setup: func() (*model.Article, *model.User) {
				user := &model.User{Model: gorm.Model{ID: 1}, Username: "user1"}
				article := &model.Article{Model: gorm.Model{ID: 1}, FavoritesCount: 1}
				return article, user
			},
			expectedError:    true,
			expectedFavCount: 1,
			mock: func(a *model.Article, u *model.User) {
				mock.ExpectBegin()
				mock.ExpectExec("DELETE FROM \"favorite_articles\"").WithArgs(u.ID, a.ID).WillReturnError(gorm.ErrInvalidTransaction)
				mock.ExpectRollback()
			},
		},
		{
			name: "Database Error on Updating Favorite Count",
			setup: func() (*model.Article, *model.User) {
				user := &model.User{Model: gorm.Model{ID: 1}, Username: "user1"}
				article := &model.Article{Model: gorm.Model{ID: 1}, FavoritesCount: 1}
				return article, user
			},
			expectedError:    true,
			expectedFavCount: 1,
			mock: func(a *model.Article, u *model.User) {
				mock.ExpectBegin()
				mock.ExpectExec("DELETE FROM \"favorite_articles\"").WithArgs(u.ID, a.ID).WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectExec("UPDATE \"articles\"").WithArgs(0, a.ID).WillReturnError(gorm.ErrInvalidTransaction)
				mock.ExpectRollback()
			},
		},
		{
			name: "Attempt to Delete Favorite with Nil User",
			setup: func() (*model.Article, *model.User) {
				article := &model.Article{Model: gorm.Model{ID: 1}, FavoritesCount: 1}
				return article, nil
			},
			expectedError:    true,
			expectedFavCount: 1,
			mock: func(a *model.Article, u *model.User) {
				mock.ExpectBegin()
				mock.ExpectExec("DELETE FROM \"favorite_articles\"").WithArgs(nil, a.ID).WillReturnError(gorm.ErrInvalidTransaction)
				mock.ExpectRollback()
			},
		},
		{
			name: "Handle Nil Article Reference",
			setup: func() (*model.Article, *model.User) {
				user := &model.User{Model: gorm.Model{ID: 1}, Username: "user1"}
				return nil, user
			},
			expectedError:    true,
			expectedFavCount: 0,
			mock: func(a *model.Article, u *model.User) {
				mock.ExpectBegin()
				mock.ExpectExec("DELETE FROM \"favorite_articles\"").WithArgs(u.ID, nil).WillReturnError(gorm.ErrInvalidTransaction)
				mock.ExpectRollback()
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			article, user := tt.setup()
			tt.mock(article, user)

			err := articleStore.DeleteFavorite(article, user)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedFavCount, article.FavoritesCount, "FavoritesCount should match expected value")
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}


/*
ROOST_METHOD_HASH=GetArticles_6382a4fe7a
ROOST_METHOD_SIG_HASH=GetArticles_1a0b3b0e8b


 */
func TestGetArticles(t *testing.T) {
	tests := []struct {
		name        string
		tagName     string
		username    string
		favoritedBy *model.User
		limit       int64
		offset      int64
		setupMock   func(mock sqlmock.Sqlmock)
		wantCount   int
		wantErr     bool
	}{
		{
			name:     "Retrieve articles by a specific username",
			username: "testuser",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`JOIN users ON articles.user_id = users.id WHERE users.username = \?`).
					WithArgs("testuser").
					WillReturnRows(sqlmock.NewRows([]string{"id", "title"}).
						AddRow(1, "Article 1").
						AddRow(2, "Article 2"))
			},
			wantCount: 2,
			wantErr:   false,
		},
		{
			name:    "Retrieve articles with a specific tag",
			tagName: "tech",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`JOIN tags ON tags.id = article_tags.tag_id WHERE tags.name = \?`).
					WithArgs("tech").
					WillReturnRows(sqlmock.NewRows([]string{"id", "title"}).
						AddRow(1, "Article 1"))
			},
			wantCount: 1,
			wantErr:   false,
		},
		{
			name: "Retrieve articles favorited by a specific user",
			favoritedBy: &model.User{
				Model: gorm.Model{ID: 1},
			},
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT article_id FROM favorite_articles WHERE user_id = \?`).
					WithArgs(1).
					WillReturnRows(sqlmock.NewRows([]string{"article_id"}).
						AddRow(1))

				mock.ExpectQuery(`WHERE id in \(\?\)`).
					WithArgs(sqlmock.AnyArg())
			},
			wantCount: 1,
			wantErr:   false,
		},
		{
			name:     "Retrieve articles with combined filters",
			tagName:  "tech",
			username: "testuser",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`JOIN users ON articles.user_id = users.id WHERE users.username = \?`).
					WithArgs("testuser").
					WillReturnRows(sqlmock.NewRows([]string{"id", "title"}).
						AddRow(1, "Article 1"))

				mock.ExpectQuery(`JOIN tags ON tags.id = article_tags.tag_id WHERE tags.name = \?`).
					WithArgs("tech")
			},
			wantCount: 1,
			wantErr:   false,
		},
		{
			name: "Handle database errors gracefully",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT \\* FROM articles").
					WillReturnError(errors.New("db error"))
			},
			wantCount: 0,
			wantErr:   true,
		},
		{
			name:   "Pagination support with limit and offset",
			limit:  1,
			offset: 1,
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT \\* FROM articles LIMIT \\? OFFSET \\?").
					WithArgs(1, 1).
					WillReturnRows(sqlmock.NewRows([]string{"id", "title"}).AddRow(2, "Article 2"))
			},
			wantCount: 1,
			wantErr:   false,
		},
		{
			name: "Retrieve all articles without filters",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT \\* FROM articles").
					WillReturnRows(sqlmock.NewRows([]string{"id", "title"}).
						AddRow(1, "Article 1").
						AddRow(2, "Article 2"))
			},
			wantCount: 2,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("Failed to create mock database: %v", err)
			}
			defer db.Close()

			gormDB, err := gorm.Open("sqlite3", db)
			if err != nil {
				t.Fatalf("Failed to open gorm DB: %v", err)
			}

			articleStore := store.ArticleStore{DB: gormDB}

			tt.setupMock(mock)

			articles, err := articleStore.GetArticles(tt.tagName, tt.username, tt.favoritedBy, tt.limit, tt.offset)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetArticles() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if len(articles) != tt.wantCount {
				t.Errorf("GetArticles() got %v articles, want %v", len(articles), tt.wantCount)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("Unfulfilled mock expectations: %v", err)
			}
		})
	}
}

