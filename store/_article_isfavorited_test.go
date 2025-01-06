package store

import (
	"testing"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jinzhu/gorm"
	"github.com/raahii/golang-grpc-realworld-example/model"
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
