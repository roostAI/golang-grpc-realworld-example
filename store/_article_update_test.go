package store

import (
	"testing"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jinzhu/gorm"
	"github.com/raahii/golang-grpc-realworld-example/model"
	_ "github.com/jinzhu/gorm/dialects/postgres"
)

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




type T struct {
	common
	isParallel bool
	isEnvSet   bool
	context    *testContext // For running tests and subtests.
}
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
