package store

import (
	"database/sql/driver"
	"testing"
	"github.com/DATA-DOG/go-sqlmock"
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
