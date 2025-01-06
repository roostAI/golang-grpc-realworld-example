package store

import (
	"errors"
	"testing"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/raahii/golang-grpc-realworld-example/model"
	"github.com/stretchr/testify/assert"
)

type ExpectedQuery struct {
	queryBasedExpectation
	rows             driver.Rows
	delay            time.Duration
	rowsMustBeClosed bool
	rowsWereClosed   bool
}




type T struct {
	common
	isParallel bool
	isEnvSet   bool
	context    *testContext // For running tests and subtests.
}
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
