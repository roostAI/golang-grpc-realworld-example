package store

import (
	"testing"
	"github.com/DATA-DOG/go-sqlmock"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
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

type Rows struct {
	converter driver.ValueConverter
	cols      []string
	def       []*Column
	rows      [][]driver.Value
	pos       int
	nextErr   map[int]error
	closeErr  error
}




type T struct {
	common
	isParallel bool
	isEnvSet   bool
	context    *testContext // For running tests and subtests.
}
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
