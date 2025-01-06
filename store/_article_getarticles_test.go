package store

import (
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




type T struct {
	common
	isParallel bool
	isEnvSet   bool
	context    *testContext // For running tests and subtests.
}
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
