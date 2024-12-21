package store

import (
	"testing"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/jinzhu/gorm"
	"github.com/raahii/golang-grpc-realworld-example/model"
	"time"
	"fmt"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"log"
	"sync"
	"errors"
	"github.com/stretchr/testify/require"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	sqlmock "github.com/DATA-DOG/go-sqlmock"
	_ "github.com/mattn/go-sqlite3"
	"bytes"
	"os"
	"database/sql"
)


/*
ROOST_METHOD_HASH=Create_0a911e138d
ROOST_METHOD_SIG_HASH=Create_723c594377


 */
func TestArticleStoreCreate(t *testing.T) {
	tests := []struct {
		name           string
		setup          func() (*ArticleStore, *model.Article)
		expectedError  bool
		expectedCreate bool
	}{
		{
			name: "Successful Article Creation",
			setup: func() (*ArticleStore, *model.Article) {

				db, mock, _ := sqlmock.New()
				sqlDB, _ := gorm.Open("sqlite3", db)
				articleStore := &ArticleStore{db: sqlDB}

				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO articles").WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()

				article := &model.Article{
					Title:       "Unique Title",
					Description: "A valid description.",
					Body:        "Interesting article body.",
					UserID:      1,
				}

				return articleStore, article
			},
			expectedError:  false,
			expectedCreate: true,
		},
		{
			name: "Article Creation with Missing Required Fields",
			setup: func() (*ArticleStore, *model.Article) {
				db, mock, _ := sqlmock.New()
				sqlDB, _ := gorm.Open("sqlite3", db)
				articleStore := &ArticleStore{db: sqlDB}

				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO articles").WillReturnError(gorm.ErrInvalidSQL)
				mock.ExpectRollback()

				article := &model.Article{

					UserID: 1,
				}

				return articleStore, article
			},
			expectedError:  true,
			expectedCreate: false,
		},
		{
			name: "Article Creation with Duplicate Title",
			setup: func() (*ArticleStore, *model.Article) {
				db, mock, _ := sqlmock.New()
				sqlDB, _ := gorm.Open("sqlite3", db)
				articleStore := &ArticleStore{db: sqlDB}

				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO articles").
					WillReturnError(gorm.ErrInvalidSQL)
				mock.ExpectRollback()

				article := &model.Article{
					Title:       "Duplicate Title",
					Description: "Another valid description.",
					Body:        "More interesting content.",
					UserID:      1,
				}

				return articleStore, article
			},
			expectedError:  true,
			expectedCreate: false,
		},
		{
			name: "Article Creation with Null Database Connection",
			setup: func() (*ArticleStore, *model.Article) {
				articleStore := &ArticleStore{db: nil}
				article := &model.Article{
					Title:       "Null DB Title",
					Description: "Description with null DB.",
					Body:        "Body text here.",
					UserID:      1,
				}

				return articleStore, article
			},
			expectedError:  true,
			expectedCreate: false,
		},
		{
			name: "Article Creation with Exceedingly Large Textual Data",
			setup: func() (*ArticleStore, *model.Article) {
				db, mock, _ := sqlmock.New()
				sqlDB, _ := gorm.Open("sqlite3", db)
				articleStore := &ArticleStore{db: sqlDB}

				largeText := "Lorem ipsum dolor sit amet, consectetur adipiscing elit."

				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO articles").
					WithArgs(sqlmock.AnyArg(), largeText, largeText, largeText, sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()

				article := &model.Article{
					Title:       largeText,
					Description: largeText,
					Body:        largeText,
					UserID:      1,
				}

				return articleStore, article
			},
			expectedError:  false,
			expectedCreate: true,
		},
		{
			name: "Article Creation with Non-Existent Author Reference",
			setup: func() (*ArticleStore, *model.Article) {
				db, mock, _ := sqlmock.New()
				sqlDB, _ := gorm.Open("sqlite3", db)
				articleStore := &ArticleStore{db: sqlDB}

				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO articles").
					WillReturnError(gorm.ErrInvalidTransaction)
				mock.ExpectRollback()

				article := &model.Article{
					Title:       "Valid Title",
					Description: "Checking author reference.",
					Body:        "Content goes here.",
					UserID:      9999,
				}

				return articleStore, article
			},
			expectedError:  true,
			expectedCreate: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			articleStore, article := tt.setup()
			err := articleStore.Create(article)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}


/*
ROOST_METHOD_HASH=CreateComment_58d394e2c6
ROOST_METHOD_SIG_HASH=CreateComment_28b95f60a6


 */
func TestCreateComment(t *testing.T) {
	type test struct {
		name    string
		comment model.Comment
		setup   func(mock sqlmock.Sqlmock)
		wantErr bool
	}

	var (
		errInvalidData    = errors.New("invalid data")
		errRecordNotFound = gorm.ErrRecordNotFound
	)

	tests := []test{
		{
			name: "Successfully Create a Valid Comment",
			comment: model.Comment{
				Model:     gorm.Model{ID: 1, CreatedAt: time.Now()},
				Body:      "Great article!",
				UserID:    1,
				ArticleID: 1,
			},
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec("^INSERT INTO \"comments\" (.+) VALUES (.+)$").
					WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), "Great article!", 1, 1).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
			wantErr: false,
		},
		{
			name: "Handle Comment Creation with Missing Body",
			comment: model.Comment{
				Model:     gorm.Model{ID: 2, CreatedAt: time.Now()},
				UserID:    1,
				ArticleID: 1,
			},
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec("^INSERT INTO \"comments\" (.+) VALUES (.+)$").
					WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), "", 1, 1).
					WillReturnError(errInvalidData)
				mock.ExpectRollback()
			},
			wantErr: true,
		},
		{
			name: "Handle Comment Creation with Nonexistent ArticleID",
			comment: model.Comment{
				Model:     gorm.Model{ID: 3, CreatedAt: time.Now()},
				Body:      "Nice write-up!",
				UserID:    1,
				ArticleID: 999,
			},
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec("^INSERT INTO \"comments\" (.+) VALUES (.+)$").
					WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), "Nice write-up!", 1, 999).
					WillReturnError(errRecordNotFound)
				mock.ExpectRollback()
			},
			wantErr: true,
		},
		{
			name: "Handle Comment Creation with Invalid UserID",
			comment: model.Comment{
				Model:     gorm.Model{ID: 4, CreatedAt: time.Now()},
				Body:      "Advice appreciated!",
				UserID:    999,
				ArticleID: 1,
			},
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec("^INSERT INTO \"comments\" (.+) VALUES (.+)$").
					WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), "Advice appreciated!", 999, 1).
					WillReturnError(errRecordNotFound)
				mock.ExpectRollback()
			},
			wantErr: true,
		},
		{
			name: "Ensure Duplicate Comment Creation Handles Properly",
			comment: model.Comment{
				Model:     gorm.Model{ID: 5, CreatedAt: time.Now()},
				Body:      "Great!",
				UserID:    1,
				ArticleID: 1,
			},
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec("^INSERT INTO \"comments\" (.+) VALUES (.+)$").
					WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), "Great!", 1, 1).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()

				mock.ExpectBegin()
				mock.ExpectExec("^INSERT INTO \"comments\" (.+) VALUES (.+)$").
					WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), "Great!", 1, 1).
					WillReturnError(errInvalidData)
				mock.ExpectRollback()
			},
			wantErr: true,
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
				t.Fatalf("Failed to initialize gorm db: %s", err)
			}
			defer gormDB.Close()

			tt.setup(mock)

			store := &ArticleStore{db: gormDB}

			err = store.CreateComment(&tt.comment)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected an error but got none")
				} else {
					t.Logf("Successfully caught expected error: %v", err)
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				t.Log("Comment created successfully without errors")
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
	t.Run("Scenario 1: Successfully Delete an Existing Article", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("error setting up mock database: %v", err)
		}
		defer db.Close()

		gormDB, _ := gorm.Open("postgres", db)
		articleStore := &ArticleStore{db: gormDB}

		article := &model.Article{
			Model: gorm.Model{ID: 1},
			Title: "Test Article",
		}

		mock.ExpectBegin()
		mock.ExpectExec("DELETE FROM").WithArgs(1).WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		err = articleStore.Delete(article)
		if err != nil {
			t.Errorf("expected no error, but got %v", err)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %v", err)
		}
		t.Log("Successfully deleted existing article.")
	})

	t.Run("Scenario 2: Attempt Deletion of a Non-Existent Article", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("error setting up mock database: %v", err)
		}
		defer db.Close()

		gormDB, _ := gorm.Open("postgres", db)
		articleStore := &ArticleStore{db: gormDB}

		article := &model.Article{
			Model: gorm.Model{ID: 2},
			Title: "Non-Existent Article",
		}

		mock.ExpectBegin()
		mock.ExpectExec("DELETE FROM").WithArgs(2).WillReturnResult(sqlmock.NewResult(2, 0))
		mock.ExpectCommit()

		err = articleStore.Delete(article)
		if err == nil {
			t.Errorf("expected error for non-existent article, got nil")
		} else {
			t.Logf("Received expected error: %v", err)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %v", err)
		}
		t.Log("Handled non-existent article deletion gracefully.")
	})

	t.Run("Scenario 3: Handle Database Error During Deletion", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("error setting up mock database: %v", err)
		}
		defer db.Close()

		gormDB, _ := gorm.Open("postgres", db)
		articleStore := &ArticleStore{db: gormDB}

		article := &model.Article{
			Model: gorm.Model{ID: 3},
			Title: "Article with DB Error",
		}

		mock.ExpectBegin()
		mock.ExpectExec("DELETE FROM").WithArgs(3).WillReturnError(fmt.Errorf("database error"))
		mock.ExpectRollback()

		err = articleStore.Delete(article)
		if err == nil {
			t.Errorf("expected database error, got nil")
		} else {
			t.Logf("Properly handled database error: %v", err)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %v", err)
		}
		t.Log("Database error during article deletion handled appropriately.")
	})

}


/*
ROOST_METHOD_HASH=DeleteComment_b345e525a7
ROOST_METHOD_SIG_HASH=DeleteComment_732762ff12


 */
func TestDeleteComment(t *testing.T) {
	type testCase struct {
		name           string
		comment        model.Comment
		mockSetup      func(mock sqlmock.Sqlmock)
		expectedError  error
		expectedDelete bool
	}

	tests := []testCase{
		{
			name: "Successfully Delete an Existing Comment",
			comment: model.Comment{
				Model: gorm.Model{ID: 1},
				Body:  "Test comment",
			},
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec("DELETE FROM `comments` WHERE").WithArgs(1).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
			expectedError:  nil,
			expectedDelete: true,
		},
		{
			name: "Attempt to Delete a Non-Existent Comment",
			comment: model.Comment{
				Model: gorm.Model{ID: 2},
				Body:  "Non-existent comment",
			},
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec("DELETE FROM `comments` WHERE").WithArgs(2).
					WillReturnResult(sqlmock.NewResult(0, 0))
				mock.ExpectCommit()
			},
			expectedError:  gorm.ErrRecordNotFound,
			expectedDelete: false,
		},
		{
			name: "Delete Comment with Foreign Key Constraints",
			comment: model.Comment{
				Model: gorm.Model{ID: 3},
				Body:  "Foreign key constrained comment",
			},
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec("DELETE FROM `comments` WHERE").WithArgs(3).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
			expectedError:  nil,
			expectedDelete: true,
		},
		{
			name: "Fail to Delete Due to Database Connection Error",
			comment: model.Comment{
				Model: gorm.Model{ID: 4},
				Body:  "Comment with DB error",
			},
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("DELETE FROM `comments` WHERE").WithArgs(4).
					WillReturnError(gorm.ErrInvalidTransaction)
			},
			expectedError:  gorm.ErrInvalidTransaction,
			expectedDelete: false,
		},
		{
			name: "Handle Soft Delete and Verify Comment is Soft-Deleted",
			comment: model.Comment{
				Model: gorm.Model{ID: 5, DeletedAt: nil},
				Body:  "Soft-delete comment",
			},
			mockSetup: func(mock sqlmock.Sqlmock) {
				currentTime := time.Now()
				mock.ExpectBegin()
				mock.ExpectExec("UPDATE `comments` SET `deleted_at`=? WHERE").WithArgs(currentTime, 5).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
			expectedError:  nil,
			expectedDelete: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			assert.NoError(t, err)
			defer db.Close()

			gormDB, err := gorm.Open("postgres", db)
			assert.NoError(t, err)

			tc.mockSetup(mock)

			store := &ArticleStore{db: gormDB}
			err = store.DeleteComment(&tc.comment)

			if tc.expectedError != nil {
				assert.EqualError(t, err, tc.expectedError.Error())
			} else {
				assert.NoError(t, err)
			}

			if tc.expectedDelete {
				assert.NotNil(t, tc.comment.DeletedAt)
			} else {
				assert.Nil(t, tc.comment.DeletedAt)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}


/*
ROOST_METHOD_HASH=GetByID_36e92ad6eb
ROOST_METHOD_SIG_HASH=GetByID_9616e43e52


 */
func TestArticleStoreGetByID(t *testing.T) {

	db, mock, err := sqlmock.New()
	if err != nil {
		log.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	gdb, err := gorm.Open("postgres", db)
	if err != nil {
		log.Fatalf("an error '%s' was not expected when initializing gorm", err)
	}

	articleStore := &ArticleStore{db: gdb}

	article := model.Article{

		Model: gorm.Model{ID: 1},
		Title: "Test Article",
		Body:  "This is a test article body.",
		Tags: []model.Tag{
			{Name: "Go"},
			{Name: "Golang"},
		},
		Author: model.User{
			Model: gorm.Model{ID: 1},
			Name:  "Author Name",
		},
	}

	testCases := []struct {
		name          string
		id            uint
		mockSetup     func()
		expectedError error
		expectedData  *model.Article
	}{
		{
			name: "Successfully Retrieve an Article by ID",
			id:   article.ID,
			mockSetup: func() {
				rows := sqlmock.NewRows([]string{"id", "title", "body"}).
					AddRow(article.ID, article.Title, article.Body)
				mock.ExpectQuery("SELECT \\* FROM \"articles\" WHERE \"articles\".\"id\" = \\$1").
					WithArgs(article.ID).
					WillReturnRows(rows)

			},
			expectedError: nil,
			expectedData:  &article,
		},
		{
			name: "Article Not Found",
			id:   2,
			mockSetup: func() {
				rows := sqlmock.NewRows([]string{"id", "title", "body"})
				mock.ExpectQuery("SELECT \\* FROM \"articles\" WHERE \"articles\".\"id\" = \\$1").
					WithArgs(2).
					WillReturnRows(rows)
			},
			expectedError: gorm.ErrRecordNotFound,
			expectedData:  nil,
		},
		{
			name: "Database Error Handling",
			id:   1,
			mockSetup: func() {
				mock.ExpectQuery("SELECT \\* FROM \"articles\" WHERE \"articles\".\"id\" = \\$1").
					WithArgs(1).
					WillReturnError(gorm.ErrInvalidTransaction)
			},
			expectedError: gorm.ErrInvalidTransaction,
			expectedData:  nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.mockSetup()

			article, err := articleStore.GetByID(tc.id)

			assert.Equal(t, tc.expectedError, err)
			assert.Equal(t, tc.expectedData, article)

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}

	t.Run("Handling of Concurrent Access", func(t *testing.T) {
		mock.ExpectQuery("SELECT \\* FROM \"articles\" WHERE \"articles\".\"id\" = \\$1").
			WithArgs(1).
			WillReturnRows(sqlmock.NewRows([]string{"id", "title", "body"}).
				AddRow(article.ID, article.Title, article.Body))

		var wg sync.WaitGroup
		concurrency := 5
		results := make([]*model.Article, concurrency)
		errors := make([]error, concurrency)

		for i := 0; i < concurrency; i++ {
			wg.Add(1)
			go func(index int) {
				defer wg.Done()
				results[index], errors[index] = articleStore.GetByID(article.ID)
			}(i)
		}
		wg.Wait()

		for i := 0; i < concurrency; i++ {
			assert.NoError(t, errors[i])
			assert.Equal(t, &article, results[i])
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %s", err)
		}
	})

}


/*
ROOST_METHOD_HASH=GetCommentByID_4bc82104a6
ROOST_METHOD_SIG_HASH=GetCommentByID_333cab101b


 */
func TestGetCommentByID(t *testing.T) {
	type testCase struct {
		name            string
		commentID       uint
		setupMock       func(mock sqlmock.Sqlmock)
		expectedComment *model.Comment
		expectedError   string
	}

	gormDB, mock := setupMockDB(t)
	defer gormDB.Close()

	store := &ArticleStore{db: gormDB}

	testCases := []testCase{
		{
			name:      "Scenario 1: Successfully retrieve a comment by ID",
			commentID: 1,
			setupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "body", "user_id", "article_id"}).
					AddRow(1, "Sample comment", 1, 1)
				mock.ExpectQuery("^SELECT \\* FROM `comments` WHERE `comments`.`id` = \\?").
					WithArgs(1).
					WillReturnRows(rows)
			},
			expectedComment: &model.Comment{
				Model:     gorm.Model{ID: 1},
				Body:      "Sample comment",
				UserID:    1,
				ArticleID: 1,
			},
			expectedError: "",
		},
		{
			name:      "Scenario 2: Retrieve a comment with an invalid ID",
			commentID: 99,
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("^SELECT \\* FROM `comments` WHERE `comments`.`id` = \\?").
					WithArgs(99).
					WillReturnError(gorm.ErrRecordNotFound)
			},
			expectedComment: nil,
			expectedError:   gorm.ErrRecordNotFound.Error(),
		},
		{
			name:      "Scenario 3: Database error during comment retrieval",
			commentID: 1,
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("^SELECT \\* FROM `comments` WHERE `comments`.`id` = \\?").
					WithArgs(1).
					WillReturnError(errors.New("database error"))
			},
			expectedComment: nil,
			expectedError:   "database error",
		},
		{
			name:      "Scenario 5: Retrieve a comment from an empty database",
			commentID: 1,
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("^SELECT \\* FROM `comments` WHERE `comments`.`id` = \\?").
					WithArgs(1).
					WillReturnError(gorm.ErrRecordNotFound)
			},
			expectedComment: nil,
			expectedError:   gorm.ErrRecordNotFound.Error(),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.setupMock(mock)

			comment, err := store.GetCommentByID(tc.commentID)

			if tc.expectedError == "" {
				require.NoError(t, err, "Expected no error, got %v", err)
				assert.Equal(t, tc.expectedComment, comment, "Comments should match")
			} else {
				require.Error(t, err, "Expected an error but got none")
				assert.EqualError(t, err, tc.expectedError, "Error messages should match")
			}

			require.NoError(t, mock.ExpectationsWereMet(), "All expectations should be met")
		})
	}
}

func setupMockDB(t *testing.T) (*gorm.DB, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err, "Failed to create SQL mock")
	gormDB, err := gorm.Open("mysql", db)
	require.NoError(t, err, "Failed to open gorm DB")
	return gormDB, mock
}


/*
ROOST_METHOD_HASH=GetComments_e24a0f1b73
ROOST_METHOD_SIG_HASH=GetComments_fa6661983e


 */
func TestArticleStoreGetComments(t *testing.T) {
	tests := []struct {
		name             string
		setup            func(mock sqlmock.Sqlmock)
		article          model.Article
		expectedError    error
		expectedComments []model.Comment
	}{
		{
			name: "Fetch Comments Successfully",
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("^SELECT \\* FROM \"comments\" WHERE").
					WillReturnRows(
						sqlmock.NewRows([]string{"id", "body", "user_id", "article_id"}).
							AddRow(1, "Comment 1", 1, 1).
							AddRow(2, "Comment 2", 2, 1),
					)
				mock.ExpectQuery("^SELECT \\* FROM \"users\" WHERE").
					WillReturnRows(
						sqlmock.NewRows([]string{"id", "username"}).
							AddRow(1, "author1").
							AddRow(2, "author2"),
					)
			},
			article: model.Article{Model: gorm.Model{ID: 1}},
			expectedComments: []model.Comment{
				{Model: gorm.Model{ID: 1}, Body: "Comment 1", UserID: 1, Author: model.User{Model: gorm.Model{ID: 1}, Username: "author1"}},
				{Model: gorm.Model{ID: 2}, Body: "Comment 2", UserID: 2, Author: model.User{Model: gorm.Model{ID: 2}, Username: "author2"}},
			},
			expectedError: nil,
		},
		{
			name: "No Comments Available",
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("^SELECT \\* FROM \"comments\" WHERE").
					WillReturnRows(sqlmock.NewRows([]string{}))
			},
			article:          model.Article{Model: gorm.Model{ID: 2}},
			expectedComments: []model.Comment{},
			expectedError:    nil,
		},
		{
			name: "Article Not Found in Database",
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("^SELECT \\* FROM \"comments\" WHERE").
					WillReturnError(errors.New("record not found"))
			},
			article:          model.Article{Model: gorm.Model{ID: 999}},
			expectedComments: nil,
			expectedError:    errors.New("record not found"),
		},
		{
			name: "Database Query Error",
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("^SELECT \\* FROM \"comments\" WHERE").
					WillReturnError(errors.New("query error"))
			},
			article:          model.Article{Model: gorm.Model{ID: 1}},
			expectedComments: nil,
			expectedError:    errors.New("query error"),
		},
		{
			name: "Retrieve Comments with Preload Author",
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("^SELECT \\* FROM \"comments\" WHERE").
					WillReturnRows(
						sqlmock.NewRows([]string{"id", "body", "user_id", "article_id"}).
							AddRow(1, "Comment 1", 1, 1),
					)
				mock.ExpectQuery("^SELECT \\* FROM \"users\" WHERE").
					WillReturnRows(
						sqlmock.NewRows([]string{"id", "username"}).
							AddRow(1, "author1"),
					)
			},
			article: model.Article{Model: gorm.Model{ID: 1}},
			expectedComments: []model.Comment{
				{Model: gorm.Model{ID: 1}, Body: "Comment 1", UserID: 1, Author: model.User{Model: gorm.Model{ID: 1}, Username: "author1"}},
			},
			expectedError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			db, mock, err := sqlmock.New()
			assert.NoError(t, err)
			defer db.Close()

			gdb, err := gorm.Open("postgres", db)
			assert.NoError(t, err)

			tt.setup(mock)

			articleStore := &ArticleStore{db: gdb}
			comments, err := articleStore.GetComments(&tt.article)

			assert.Equal(t, tt.expectedComments, comments)
			assert.Equal(t, tt.expectedError, err)

			err = mock.ExpectationsWereMet()
			if err != nil {
				t.Errorf("Test scenario '%s', expectation error: %s", tt.name, err)
			}
		})
	}

	for _, tt := range tests {
		fmt.Printf("Scenario: %s\n", tt.name)
		if tt.expectedError != nil {
			fmt.Printf("Expected error: %s\n", tt.expectedError.Error())
		} else {
			fmt.Printf("Expected comments: %v\n", tt.expectedComments)
		}
	}
}


/*
ROOST_METHOD_HASH=GetTags_ac049ebded
ROOST_METHOD_SIG_HASH=GetTags_25034b82b0


 */
func TestArticleStoreGetTags(t *testing.T) {
	tests := []struct {
		name          string
		setupMock     func(sqlmock.Sqlmock)
		expectedTags  []model.Tag
		expectedError bool
	}{
		{
			name: "Successfully Retrieve Tags",
			setupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "name"}).
					AddRow(1, "Tag1").
					AddRow(2, "Tag2")

				mock.ExpectQuery("^SELECT (.+) FROM `tags`").
					WillReturnRows(rows)
			},
			expectedTags: []model.Tag{
				{Model: gorm.Model{ID: 1}, Name: "Tag1"},
				{Model: gorm.Model{ID: 2}, Name: "Tag2"},
			},
			expectedError: false,
		},
		{
			name: "Retrieve Tags Returns Empty List",
			setupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "name"})
				mock.ExpectQuery("^SELECT (.+) FROM `tags`").
					WillReturnRows(rows)
			},
			expectedTags:  []model.Tag{},
			expectedError: false,
		},
		{
			name: "Database Error Occurs",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("^SELECT (.+) FROM `tags`").
					WillReturnError(gorm.ErrInvalidTransaction)
			},
			expectedTags:  nil,
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("error creating mock db: %v", err)
			}
			defer db.Close()

			gormDB, err := gorm.Open("sqlite3", db)
			if err != nil {
				t.Fatalf("error initializing gorm db: %v", err)
			}

			tt.setupMock(mock)

			store := ArticleStore{db: gormDB}

			tags, err := store.GetTags()

			if tt.expectedError && err == nil {
				t.Errorf("expected error but got none")
			}
			if !tt.expectedError && err != nil {
				t.Errorf("did not expect error, but got: %v", err)
			}

			if len(tags) != len(tt.expectedTags) {
				t.Errorf("expected %d tags, got %d", len(tt.expectedTags), len(tags))
			}

			for i, tag := range tags {
				if tag.ID != tt.expectedTags[i].ID || tag.Name != tt.expectedTags[i].Name {
					t.Errorf("expected tag %v, got %v", tt.expectedTags[i], tag)
				}
			}

			t.Logf("Test '%s' completed with returned tags: %#v", tt.name, tags)
		})
	}
}


/*
ROOST_METHOD_HASH=IsFavorited_7ef7d3ed9e
ROOST_METHOD_SIG_HASH=IsFavorited_f34d52378f


 */
func TestArticleStoreIsFavorited(t *testing.T) {

	db, mock, err := sqlmock.New()
	if err != nil {
		log.Fatalf("Error '%s' when opening a stub database connection", err)
	}
	defer db.Close()

	gormDB, err := gorm.Open("postgres", db)
	if err != nil {
		log.Fatalf("Error '%s' when opening gorm DB", err)
	}

	store := &ArticleStore{db: gormDB}

	tests := []struct {
		name      string
		article   *model.Article
		user      *model.User
		mockSetup func()
		expected  bool
		expectErr bool
	}{
		{
			name:    "User has favorited the article",
			article: &model.Article{Model: gorm.Model{ID: 1}},
			user:    &model.User{Model: gorm.Model{ID: 1}},
			mockSetup: func() {
				mock.ExpectQuery("^SELECT (.+) FROM \"favorite_articles\" WHERE (.+)$").
					WithArgs(1, 1).
					WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
			},
			expected:  true,
			expectErr: false,
		},
		{
			name:    "User has not favorited the article",
			article: &model.Article{Model: gorm.Model{ID: 1}},
			user:    &model.User{Model: gorm.Model{ID: 2}},
			mockSetup: func() {
				mock.ExpectQuery("^SELECT (.+) FROM \"favorite_articles\" WHERE (.+)$").
					WithArgs(1, 2).
					WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
			},
			expected:  false,
			expectErr: false,
		},
		{
			name:    "Nil article input",
			article: nil,
			user:    &model.User{Model: gorm.Model{ID: 1}},
			mockSetup: func() {

			},
			expected:  false,
			expectErr: false,
		},
		{
			name:    "Nil user input",
			article: &model.Article{Model: gorm.Model{ID: 1}},
			user:    nil,
			mockSetup: func() {

			},
			expected:  false,
			expectErr: false,
		},
		{
			name:    "Database error during query",
			article: &model.Article{Model: gorm.Model{ID: 1}},
			user:    &model.User{Model: gorm.Model{ID: 1}},
			mockSetup: func() {
				mock.ExpectQuery("^SELECT (.+) FROM \"favorite_articles\" WHERE (.+)$").
					WithArgs(1, 1).
					WillReturnError(gorm.ErrInvalidSQL)
			},
			expected:  false,
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			result, err := store.IsFavorited(tt.article, tt.user)
			if (err != nil) != tt.expectErr {
				t.Errorf("expected error: %v, got: %v", tt.expectErr, err)
			}
			if result != tt.expected {
				t.Errorf("expected result: %v, got: %v", tt.expected, result)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}


/*
ROOST_METHOD_HASH=NewArticleStore_6be2824012
ROOST_METHOD_SIG_HASH=NewArticleStore_3fe6f79a92


 */
func TestNewArticleStore(t *testing.T) {
	t.Run("Scenario 1: Valid Database Instance", func(t *testing.T) {

		db, _, err := sqlmock.New()
		if err != nil {
			t.Fatalf("Failed to open sqlmock database: %v", err)
		}
		defer db.Close()

		gormDB, err := gorm.Open("mysql", db)
		if err != nil {
			t.Fatalf("failed to open gorm DB: %v", err)
		}

		articleStore := NewArticleStore(gormDB)

		assert.NotNil(t, articleStore, "ArticleStore should be instantiated")
		assert.Equal(t, gormDB, articleStore.db, "The db field should be set to the provided gorm.DB instance")

		t.Log("Scenario 1: Success. Valid DB instance correctly initializes ArticleStore.")
	})

	t.Run("Scenario 2: Nil Database Instance", func(t *testing.T) {

		var nilDB *gorm.DB

		articleStore := NewArticleStore(nilDB)

		assert.NotNil(t, articleStore, "ArticleStore should be instantiated even with nil db")
		assert.Nil(t, articleStore.db, "The db field should be nil")

		t.Log("Scenario 2: Success. ArticleStore handles nil DB instance gracefully.")
	})

	t.Run("Scenario 3: Correct Initialization of ArticleStore Fields", func(t *testing.T) {

		db, _, err := sqlmock.New()
		if err != nil {
			t.Fatalf("Failed to open sqlmock database: %v", err)
		}
		defer db.Close()

		gormDB, err := gorm.Open("mysql", db)
		if err != nil {
			t.Fatalf("failed to open gorm DB: %v", err)
		}

		articleStore := NewArticleStore(gormDB)

		assert.NotNil(t, articleStore, "ArticleStore should be instantiated")

		t.Log("Scenario 3: Success. ArticleStore fields initialized correctly.")
	})

	t.Run("Scenario 4: Persistent Database Connection", func(t *testing.T) {

		db, _, err := sqlmock.New()
		if err != nil {
			t.Fatalf("Failed to open sqlmock database: %v", err)
		}
		defer db.Close()

		gormDB, err := gorm.Open("mysql", db)
		if err != nil {
			t.Fatalf("failed to open gorm DB: %v", err)
		}

		articleStore1 := NewArticleStore(gormDB)
		articleStore2 := NewArticleStore(gormDB)

		assert.Equal(t, articleStore1.db, articleStore2.db, "All ArticleStore instances should maintain the same DB instance")
		t.Log("Scenario 4: Success. Persistent connection ensured across different ArticleStore instances.")
	})
}


/*
ROOST_METHOD_HASH=Update_51145aa965
ROOST_METHOD_SIG_HASH=Update_6c1b5471fe


 */
func TestArticleStoreUpdate(t *testing.T) {
	type testCase struct {
		name          string
		prepareMocks  func(sqlmock.Sqlmock, *model.Article)
		run           func(*ArticleStore, *model.Article) error
		expectedError bool
	}

	var (
		articleID uint = 1
		userID    uint = 1
	)

	mockArticle := model.Article{
		Model:       gorm.Model{ID: articleID},
		Title:       "Initial Title",
		Description: "Initial Description",
		Body:        "Initial Body",
		UserID:      userID,
		Tags:        []model.Tag{{Name: "tech"}},
	}

	tests := []testCase{
		{
			name: "Scenario 1: Successful Update of an Existing Article",
			prepareMocks: func(mock sqlmock.Sqlmock, article *model.Article) {
				mock.ExpectBegin()
				mock.ExpectExec(`UPDATE "articles"`).
					WithArgs(article.Title, article.Description, article.Body, articleID).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
			run: func(store *ArticleStore, article *model.Article) error {
				article.Title = "Updated Title"
				article.Description = "Updated Description"
				article.Body = "Updated Body"
				return store.Update(article)
			},
			expectedError: false,
		},
		{
			name: "Scenario 2: Update Article with No Changes (No-op)",
			prepareMocks: func(mock sqlmock.Sqlmock, article *model.Article) {
				mock.ExpectBegin()
				mock.ExpectExec(`UPDATE "articles"`).
					WithArgs(article.Title, article.Description, article.Body, articleID).
					WillReturnResult(sqlmock.NewResult(1, 0))
				mock.ExpectCommit()
			},
			run: func(store *ArticleStore, article *model.Article) error {
				return store.Update(article)
			},
			expectedError: false,
		},
		{
			name: "Scenario 3: Attempt to Update a Non-Existent Article",
			prepareMocks: func(mock sqlmock.Sqlmock, article *model.Article) {
				mock.ExpectBegin()
				mock.ExpectExec(`UPDATE "articles"`).
					WithArgs(article.Title, article.Description, article.Body, article.ID).
					WillReturnResult(sqlmock.NewResult(0, 0))
				mock.ExpectCommit()
			},
			run: func(store *ArticleStore, article *model.Article) error {
				article.ID = 999
				return store.Update(article)
			},
			expectedError: true,
		},
		{
			name: "Scenario 4: Update Article to Remove Tags",
			prepareMocks: func(mock sqlmock.Sqlmock, article *model.Article) {
				mock.ExpectBegin()
				mock.ExpectExec(`UPDATE "articles"`).
					WithArgs(article.Title, article.Description, article.Body, articleID).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
			run: func(store *ArticleStore, article *model.Article) error {
				article.Tags = []model.Tag{}
				return store.Update(article)
			},
			expectedError: false,
		},
		{
			name: "Scenario 5: Concurrency: Simultaneous Updates to Same Article",
			prepareMocks: func(mock sqlmock.Sqlmock, article *model.Article) {
				mock.ExpectBegin()
				mock.ExpectExec(`UPDATE "articles"`).
					WithArgs(article.Title, article.Description, article.Body, articleID).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
			run: func(store *ArticleStore, article *model.Article) error {
				done := make(chan error, 2)

				go func() {
					article.Title = "Concurrent Title 1"
					done <- store.Update(article)
				}()
				go func() {
					article.Title = "Concurrent Title 2"
					done <- store.Update(article)
				}()
				err1 := <-done
				err2 := <-done
				if err1 != nil || err2 != nil {
					return err1
				}
				return nil
			},
			expectedError: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				log.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
			}
			defer db.Close()

			gormDB, err := gorm.Open("sqlite3", db)
			if err != nil {
				log.Fatalf("an error '%s' was not expected when creating Gorm DB connection", err)
			}

			store := &ArticleStore{db: gormDB}
			article := mockArticle

			tc.prepareMocks(mock, &article)

			err = tc.run(store, &article)

			if tc.expectedError && err == nil {
				t.Errorf("expected an error but got nil")
			}
			if !tc.expectedError && err != nil {
				t.Errorf("expected no error but got %v", err)
			}

			expectedErr := mock.ExpectationsWereMet()
			if expectedErr != nil {
				t.Errorf("there were unfulfilled expectations: %s", expectedErr)
			}

			t.Logf("Test %s passed", tc.name)
		})
	}
}


/*
ROOST_METHOD_HASH=GetFeedArticles_9c4f57afe4
ROOST_METHOD_SIG_HASH=GetFeedArticles_cadca0e51b


 */
func TestGetFeedArticles(t *testing.T) {
	type test struct {
		name           string
		userIDs        []uint
		limit          int64
		offset         int64
		mockSetup      func(mock sqlmock.Sqlmock)
		expectedError  bool
		expectedResult []model.Article
	}

	tests := []test{
		{
			name:    "Scenario 1: Fetch articles for a valid list of user IDs",
			userIDs: []uint{1, 2, 3},
			limit:   2,
			offset:  0,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("^SELECT \\* FROM \"articles\" WHERE \\(user_id in (\\$1,\\$2,\\$3)\\) LIMIT \\$4 OFFSET \\$5$").
					WithArgs(1, 2, 3, 2, 0).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1).AddRow(2))
			},
			expectedError:  false,
			expectedResult: []model.Article{{Model: gorm.Model{ID: 1}}, {Model: gorm.Model{ID: 2}}},
		},
		{
			name:    "Scenario 2: Fetch with no user IDs in the list",
			userIDs: []uint{},
			limit:   5,
			offset:  0,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("^SELECT \\* FROM \"articles\" WHERE \\(user_id in \\(\\)\\) LIMIT \\$1 OFFSET \\$2$").
					WithArgs(5, 0).
					WillReturnRows(sqlmock.NewRows([]string{"id"}))
			},
			expectedError:  false,
			expectedResult: []model.Article{},
		},
		{
			name:    "Scenario 3: Handling a database error",
			userIDs: []uint{1},
			limit:   5,
			offset:  0,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("^SELECT \\* FROM \"articles\" WHERE \\(user_id in (\\$1)\\) LIMIT \\$2 OFFSET \\$3$").
					WithArgs(1, 5, 0).
					WillReturnError(fmt.Errorf("database error"))
			},
			expectedError:  true,
			expectedResult: nil,
		},
		{
			name:    "Scenario 4: Limit exceeds available articles",
			userIDs: []uint{1},
			limit:   10,
			offset:  0,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("^SELECT \\* FROM \"articles\" WHERE \\(user_id in (\\$1)\\) LIMIT \\$2 OFFSET \\$3$").
					WithArgs(1, 10, 0).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1).AddRow(2).AddRow(3))
			},
			expectedError:  false,
			expectedResult: []model.Article{{Model: gorm.Model{ID: 1}}, {Model: gorm.Model{ID: 2}}, {Model: gorm.Model{ID: 3}}},
		},
		{
			name:    "Scenario 5: Articles fetched for a single user",
			userIDs: []uint{1},
			limit:   3,
			offset:  0,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("^SELECT \\* FROM \"articles\" WHERE \\(user_id in (\\$1)\\) LIMIT \\$2 OFFSET \\$3$").
					WithArgs(1, 3, 0).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1).AddRow(2))
			},
			expectedError:  false,
			expectedResult: []model.Article{{Model: gorm.Model{ID: 1}}, {Model: gorm.Model{ID: 2}}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			stdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
			}
			defer db.Close()

			gormDB, err := gorm.Open("postgres", db)
			if err != nil {
				t.Fatalf("failed to open gorm DB, %v", err)
			}

			tt.mockSetup(mock)

			articleStore := &ArticleStore{db: gormDB}

			result, err := articleStore.GetFeedArticles(tt.userIDs, tt.limit, tt.offset)

			w.Close()
			os.Stdout = stdout

			var buf bytes.Buffer
			fmt.Fscanf(r, "%v", &buf)

			if (err != nil) != tt.expectedError {
				t.Errorf("expected error: %v, got: %v", tt.expectedError, err)
			}
			if len(result) != len(tt.expectedResult) {
				t.Errorf("expected %d articles, got %d", len(tt.expectedResult), len(result))
			}

			for i, article := range result {
				if article.ID != tt.expectedResult[i].ID {
					t.Errorf("expected article ID %v, got %v", tt.expectedResult[i].ID, article.ID)
				}
			}

			t.Logf("Test %s completed with reason: %v", tt.name, buf.String())
		})
	}
}


/*
ROOST_METHOD_HASH=AddFavorite_2b0cb9d894
ROOST_METHOD_SIG_HASH=AddFavorite_c4dea0ee90


 */
func TestArticleStoreAddFavorite(t *testing.T) {
	type scenarioData struct {
		desc            string
		article         *model.Article
		user            *model.User
		mockDBSetup     func(sqlmock.Sqlmock)
		expectedError   bool
		expectedCount   int32
		expectedFavList []model.User
	}

	tests := []scenarioData{
		{
			desc: "Successfully Add User to Article's Favorite List and Increment Favorites Count",
			article: &model.Article{
				Model:          gorm.Model{ID: 1},
				FavoritesCount: 0,
				FavoritedUsers: []model.User{},
			},
			user: &model.User{Model: gorm.Model{ID: 1}},
			mockDBSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO favorite_articles").WithArgs(1, 1).WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectExec("UPDATE articles SET favorites_count").WithArgs().WillReturnResult(sqlmock.NewResult(0, 1))
				mock.ExpectCommit()
			},
			expectedError:   false,
			expectedCount:   1,
			expectedFavList: []model.User{{Model: gorm.Model{ID: 1}}},
		},
		{
			desc: "Handle Error When Adding User to Favorite List",
			article: &model.Article{
				Model:          gorm.Model{ID: 2},
				FavoritesCount: 0,
				FavoritedUsers: []model.User{},
			},
			user: &model.User{Model: gorm.Model{ID: 2}},
			mockDBSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO favorite_articles").WithArgs(2, 2).WillReturnError(gorm.ErrInvalidTransaction)
				mock.ExpectRollback()
			},
			expectedError:   true,
			expectedCount:   0,
			expectedFavList: []model.User{},
		},
		{
			desc: "Handle Error When Incrementing Favorites Count",
			article: &model.Article{
				Model:          gorm.Model{ID: 3},
				FavoritesCount: 0,
				FavoritedUsers: []model.User{},
			},
			user: &model.User{Model: gorm.Model{ID: 3}},
			mockDBSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO favorite_articles").WithArgs(3, 3).WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectExec("UPDATE articles SET favorites_count").WillReturnError(gorm.ErrInvalidTransaction)
				mock.ExpectRollback()
			},
			expectedError:   true,
			expectedCount:   0,
			expectedFavList: []model.User{},
		},
		{
			desc: "Add Existing User to Article's Favorite List",
			article: &model.Article{
				Model:          gorm.Model{ID: 4},
				FavoritesCount: 1,
				FavoritedUsers: []model.User{{Model: gorm.Model{ID: 4}}},
			},
			user: &model.User{Model: gorm.Model{ID: 4}},
			mockDBSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO favorite_articles").WithArgs(4, 4).WillReturnResult(sqlmock.NewResult(0, 0))
				mock.ExpectCommit()
			},
			expectedError:   false,
			expectedCount:   1,
			expectedFavList: []model.User{{Model: gorm.Model{ID: 4}}},
		},
		{
			desc:            "Edge Case - Null Article Input",
			article:         nil,
			user:            &model.User{Model: gorm.Model{ID: 5}},
			mockDBSetup:     func(mock sqlmock.Sqlmock) {},
			expectedError:   true,
			expectedCount:   0,
			expectedFavList: nil,
		},
		{
			desc: "Edge Case - Null User Input",
			article: &model.Article{
				Model:          gorm.Model{ID: 6},
				FavoritesCount: 0,
				FavoritedUsers: []model.User{},
			},
			user:            nil,
			mockDBSetup:     func(mock sqlmock.Sqlmock) {},
			expectedError:   true,
			expectedCount:   0,
			expectedFavList: []model.User{},
		},
		{
			desc: "Successful Commit with Database Mock Verification",
			article: &model.Article{
				Model:          gorm.Model{ID: 7},
				FavoritesCount: 0,
				FavoritedUsers: []model.User{},
			},
			user: &model.User{Model: gorm.Model{ID: 7}},
			mockDBSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO favorite_articles").WithArgs(7, 7).WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectExec("UPDATE articles SET favorites_count").WillReturnResult(sqlmock.NewResult(0, 1))
				mock.ExpectCommit()
			},
			expectedError:   false,
			expectedCount:   1,
			expectedFavList: []model.User{{Model: gorm.Model{ID: 7}}},
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("error mocking database: %s", err)
			}
			defer db.Close()

			gormDB, err := gorm.Open("sqlite3", db)
			if err != nil {
				t.Fatalf("error opening gorm db: %s", err)
			}

			test.mockDBSetup(mock)

			articleStore := &ArticleStore{db: gormDB}

			err = articleStore.AddFavorite(test.article, test.user)

			if test.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, test.expectedCount, test.article.FavoritesCount)
				assert.Equal(t, test.expectedFavList, test.article.FavoritedUsers)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("expectations were not met: %s", err)
			}

			t.Log(test.desc, "completed")
		})
	}
}


/*
ROOST_METHOD_HASH=DeleteFavorite_a856bcbb70
ROOST_METHOD_SIG_HASH=DeleteFavorite_f7e5c0626f


 */
func TestArticleStoreDeleteFavorite(t *testing.T) {

	setupMockDB := func() (*gorm.DB, sqlmock.Sqlmock) {
		db, mock, err := sqlmock.New()
		if err != nil {
			log.Fatalf("An error '%s' was not expected when opening a stub database connection", err)
		}
		gormDb, err := gorm.Open("sqlite3", db)
		if err != nil {
			log.Fatalf("An error '%s' was not expected when opening a gorm database connection", err)
		}

		sqlDB := gormDb.DB()
		t.Cleanup(func() {
			sqlDB.Close()
		})

		return gormDb, mock
	}

	tests := []struct {
		name             string
		setup            func(mock sqlmock.Sqlmock, article *model.Article, user *model.User)
		article          *model.Article
		user             *model.User
		expectedFavCount int32
		expectError      bool
		detailedTestLog  string
	}{
		{
			name: "Scenario 1: Successfully Delete a User from Article's Favorited Users List",
			setup: func(mock sqlmock.Sqlmock, article *model.Article, user *model.User) {
				mock.ExpectBegin()
				mock.ExpectExec(`DELETE FROM "favorite_articles"`).
					WithArgs(user.ID, article.ID).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectExec(`UPDATE "articles"`).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
			article: &model.Article{
				Model:          gorm.Model{ID: 1},
				FavoritesCount: 5,
				FavoritedUsers: []model.User{
					{Model: gorm.Model{ID: 1}, Username: "testuser"},
				},
			},
			user:             &model.User{Model: gorm.Model{ID: 1}},
			expectedFavCount: 4,
			expectError:      false,
			detailedTestLog:  "User should be removed and favorites count should decrement",
		},
		{
			name: "Scenario 2: Attempt to Delete a Non-Existent User from Favorited Users List",
			setup: func(mock sqlmock.Sqlmock, article *model.Article, user *model.User) {
				mock.ExpectBegin()
				mock.ExpectExec(`DELETE FROM "favorite_articles"`).
					WithArgs(user.ID, article.ID).
					WillReturnResult(sqlmock.NewResult(0, 0))
				mock.ExpectCommit()
			},
			article: &model.Article{
				Model:          gorm.Model{ID: 1},
				FavoritesCount: 3,
				FavoritedUsers: []model.User{
					{Model: gorm.Model{ID: 2}, Username: "otheruser"},
				},
			},
			user:             &model.User{Model: gorm.Model{ID: 1}},
			expectedFavCount: 3,
			expectError:      false,
			detailedTestLog:  "No change in favorites count as user was not in the list",
		},
		{
			name: "Scenario 3: Error Handling During Database Transaction",
			setup: func(mock sqlmock.Sqlmock, article *model.Article, user *model.User) {
				mock.ExpectBegin()
				mock.ExpectExec(`DELETE FROM "favorite_articles"`).
					WithArgs(user.ID, article.ID).
					WillReturnError(errors.New("failed to delete"))
				mock.ExpectRollback()
			},
			article: &model.Article{
				Model:          gorm.Model{ID: 1},
				FavoritesCount: 3,
			},
			user:             &model.User{Model: gorm.Model{ID: 1}},
			expectedFavCount: 3,
			expectError:      true,
			detailedTestLog:  "Expect error due to failed delete operation, transaction to rollback",
		},
		{
			name: "Scenario 4: No Side Effects When Deleting from an Empty Favorited Users List",
			setup: func(mock sqlmock.Sqlmock, article *model.Article, user *model.User) {
				mock.ExpectBegin()
				mock.ExpectExec(`DELETE FROM "favorite_articles"`).
					WithArgs(user.ID, article.ID).
					WillReturnResult(sqlmock.NewResult(0, 0))
				mock.ExpectCommit()
			},
			article: &model.Article{
				Model:          gorm.Model{ID: 1},
				FavoritesCount: 0,
				FavoritedUsers: []model.User{},
			},
			user:             &model.User{Model: gorm.Model{ID: 1}},
			expectedFavCount: 0,
			expectError:      false,
			detailedTestLog:  "No changes expected when attempting to delete from empty list",
		},
		{
			name: "Scenario 5: Database Constraint Violation When Deleting User",
			setup: func(mock sqlmock.Sqlmock, article *model.Article, user *model.User) {
				mock.ExpectBegin()
				mock.ExpectExec(`DELETE FROM "favorite_articles"`).
					WithArgs(user.ID, article.ID).
					WillReturnError(errors.New("constraint violation"))
				mock.ExpectRollback()
			},
			article: &model.Article{
				Model:          gorm.Model{ID: 1},
				FavoritesCount: 3,
			},
			user:             &model.User{Model: gorm.Model{ID: 1}},
			expectedFavCount: 3,
			expectError:      true,
			detailedTestLog:  "Error due to constraint violation, expect rollback",
		},
		{
			name: "Scenario 6: Check if Article State Consistent After Deletion",
			setup: func(mock sqlmock.Sqlmock, article *model.Article, user *model.User) {
				mock.ExpectBegin()
				mock.ExpectExec(`DELETE FROM "favorite_articles"`).
					WithArgs(user.ID, article.ID).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectExec(`UPDATE "articles"`).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
			article: &model.Article{
				Model:          gorm.Model{ID: 1},
				Title:          "Test Title",
				Description:    "Test Description",
				Body:           "Test Body",
				Tags:           []model.Tag{{Name: "go"}, {Name: "testing"}},
				Author:         model.User{Model: gorm.Model{ID: 1}, Username: "authoruser"},
				FavoritesCount: 10,
				FavoritedUsers: []model.User{
					{Model: gorm.Model{ID: 1}, Username: "testuser"},
				},
				UserID: 1,
			},
			user:             &model.User{Model: gorm.Model{ID: 1}},
			expectedFavCount: 9,
			expectError:      false,
			detailedTestLog:  "All article properties except favorites_count should remain unchanged",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock := setupMockDB()
			store := &ArticleStore{db: db}

			tt.setup(mock, tt.article, tt.user)

			err := store.DeleteFavorite(tt.article, tt.user)
			if (err != nil) != tt.expectError {
				t.Errorf("Unexpected error state: %v, expected error: %v", err, tt.expectError)
			}

			if tt.article.FavoritesCount != tt.expectedFavCount {
				t.Errorf("FavoritesCount = %d, want %d", tt.article.FavoritesCount, tt.expectedFavCount)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("There were unfulfilled expectations: %s", err)
			}

			t.Log(tt.detailedTestLog)
		})
	}

}


/*
ROOST_METHOD_HASH=GetArticles_6382a4fe7a
ROOST_METHOD_SIG_HASH=GetArticles_1a0b3b0e8b


 */
func TestGetArticles(t *testing.T) {
	type args struct {
		tagName       string
		username      string
		favoritedBy   *model.User
		limit, offset int64
	}
	tests := []struct {
		name             string
		args             args
		setupMock        func(db *gorm.DB, mock sqlmock.Sqlmock)
		wantArticleCount int
		wantErr          bool
	}{
		{
			name: "Retrieve all articles with no filters",
			args: args{
				tagName:     "",
				username:    "",
				favoritedBy: nil,
				limit:       10,
				offset:      0,
			},
			setupMock: func(db *gorm.DB, mock sqlmock.Sqlmock) {
				mock.ExpectQuery("^SELECT (.+) FROM \"articles\"").
					WillReturnRows(sqlmock.NewRows([]string{"id", "title", "description", "body", "user_id"}).
						AddRow(1, "Test Title1", "Test Description1", "Test Body1", 1).
						AddRow(2, "Test Title2", "Test Description2", "Test Body2", 1))
			},
			wantArticleCount: 2,
			wantErr:          false,
		},
		{
			name: "Filter articles by username",
			args: args{
				tagName:     "",
				username:    "testuser",
				favoritedBy: nil,
				limit:       10,
				offset:      0,
			},
			setupMock: func(db *gorm.DB, mock sqlmock.Sqlmock) {
				mock.ExpectQuery("^SELECT (.+) FROM \"articles\" join users").
					WithArgs("testuser").
					WillReturnRows(sqlmock.NewRows([]string{"id", "title", "description", "body", "user_id"}).
						AddRow(1, "Filtered Title", "Filtered Description", "Filtered Body", 1))
			},
			wantArticleCount: 1,
			wantErr:          false,
		},
		{
			name: "Filter articles by tag name",
			args: args{
				tagName:     "testtag",
				username:    "",
				favoritedBy: nil,
				limit:       10,
				offset:      0,
			},
			setupMock: func(db *gorm.DB, mock sqlmock.Sqlmock) {
				mock.ExpectQuery("^SELECT (.+) FROM \"articles\" join article_tags").
					WithArgs("testtag").
					WillReturnRows(sqlmock.NewRows([]string{"id", "title", "description", "body", "user_id"}).
						AddRow(1, "Tagged Title", "Tagged Description", "Tagged Body", 1))
			},
			wantArticleCount: 1,
			wantErr:          false,
		},
		{
			name: "Retrieve articles favorited by a particular user",
			args: args{
				tagName:     "",
				username:    "",
				favoritedBy: &model.User{Model: gorm.Model{ID: 1}},
				limit:       10,
				offset:      0,
			},
			setupMock: func(db *gorm.DB, mock sqlmock.Sqlmock) {
				mock.ExpectQuery("^SELECT article_id FROM \"favorite_articles\"").
					WithArgs(1).
					WillReturnRows(sqlmock.NewRows([]string{"article_id"}).
						AddRow(1))
				mock.ExpectQuery("^SELECT (.+) FROM \"articles\"").
					WithArgs(sqlmock.AnyArg()).
					WillReturnRows(sqlmock.NewRows([]string{"id", "title", "description", "body", "user_id"}).
						AddRow(1, "Favorited Title", "Favorited Description", "Favorited Body", 1))
			},
			wantArticleCount: 1,
			wantErr:          false,
		},
		{
			name: "Limit and offset functionality",
			args: args{
				tagName:     "",
				username:    "",
				favoritedBy: nil,
				limit:       1,
				offset:      1,
			},
			setupMock: func(db *gorm.DB, mock sqlmock.Sqlmock) {
				mock.ExpectQuery("^SELECT (.+) FROM \"articles\"").
					WillReturnRows(sqlmock.NewRows([]string{"id", "title", "description", "body", "user_id"}).
						AddRow(1, "Offset Title", "Offset Description", "Offset Body", 1))
			},
			wantArticleCount: 1,
			wantErr:          false,
		},
		{
			name: "Error handling on database failure",
			args: args{
				tagName:     "",
				username:    "",
				favoritedBy: nil,
				limit:       10,
				offset:      0,
			},
			setupMock: func(db *gorm.DB, mock sqlmock.Sqlmock) {
				mock.ExpectQuery("^SELECT (.+) FROM \"articles\"").
					WillReturnError(fmt.Errorf("db error"))
			},
			wantArticleCount: 0,
			wantErr:          true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("failed to open sqlmock database: %s", err)
			}
			defer db.Close()

			gormDB, err := gorm.Open("postgres", db)
			if err != nil {
				t.Fatalf("failed to open gorm DB: %s", err)
			}

			tt.setupMock(gormDB, mock)

			store := &ArticleStore{db: gormDB}

			articles, err := store.GetArticles(tt.args.tagName, tt.args.username, tt.args.favoritedBy, tt.args.limit, tt.args.offset)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetArticles() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(articles) != tt.wantArticleCount {
				t.Errorf("GetArticles() got = %v articles, want %v articles", len(articles), tt.wantArticleCount)
			}

			t.Logf("Test case %s: retrieved %d articles", tt.name, len(articles))
		})
	}
}

