package store

import (
	"errors"
	"testing"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jinzhu/gorm"
	"github.com/raahii/golang-grpc-realworld-example/model"
)

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
