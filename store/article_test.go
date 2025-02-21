package store

import (
	"database/sql"
	"runtime/debug"
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/raahii/golang-grpc-realworld-example/model"
	"github.com/stretchr/testify/assert"
)

/*
ROOST_METHOD_HASH=ArticleStore_GetByID_6fe18728fc
ROOST_METHOD_SIG_HASH=ArticleStore_GetByID_bb488e542f

FUNCTION_DEF=func (s *ArticleStore) GetByID(id uint) (*model.Article, error) // GetByID finds an article from id
*/
func TestArticleStoreGetById(t *testing.T) {
	type testCase struct {
		name          string
		id            uint
		mockSetup     func(mock sqlmock.Sqlmock)
		expectedError error
		expectArticle bool
	}

	tests := []testCase{
		{
			name: "Successfully retrieve article with valid ID",
			id:   1,
			mockSetup: func(mock sqlmock.Sqlmock) {
				articleColumns := []string{"id", "created_at", "updated_at", "deleted_at", "title", "description", "body", "user_id", "favorites_count"}

				mock.ExpectQuery(`SELECT \* FROM "articles"`).
					WithArgs(1).
					WillReturnRows(sqlmock.NewRows(articleColumns).
						AddRow(1, time.Now(), time.Now(), nil, "Test Title", "Test Description", "Test Body", 1, 0))

				mock.ExpectQuery(`SELECT \* FROM "tags"`).
					WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).
						AddRow(1, "test-tag"))

				mock.ExpectQuery(`SELECT \* FROM "users"`).
					WillReturnRows(sqlmock.NewRows([]string{"id", "username", "email", "password", "bio", "image"}).
						AddRow(1, "testuser", "test@example.com", "password", "bio", "image"))
			},
			expectedError: nil,
			expectArticle: true,
		},
		{
			name: "Article not found",
			id:   999,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT \* FROM "articles"`).
					WithArgs(999).
					WillReturnError(gorm.ErrRecordNotFound)
			},
			expectedError: gorm.ErrRecordNotFound,
			expectArticle: false,
		},
		{
			name: "Database connection error",
			id:   1,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT \* FROM "articles"`).
					WithArgs(1).
					WillReturnError(sql.ErrConnDone)
			},
			expectedError: sql.ErrConnDone,
			expectArticle: false,
		},
		{
			name: "Zero ID value",
			id:   0,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT \* FROM "articles"`).
					WithArgs(0).
					WillReturnError(gorm.ErrRecordNotFound)
			},
			expectedError: gorm.ErrRecordNotFound,
			expectArticle: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Logf("Panic encountered so failing test. %v\n%s", r, string(debug.Stack()))
					t.Fail()
				}
			}()

			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("Failed to create mock database: %v", err)
			}
			defer db.Close()

			gormDB, err := gorm.Open("postgres", db)
			if err != nil {
				t.Fatalf("Failed to open gorm connection: %v", err)
			}
			defer gormDB.Close()

			tc.mockSetup(mock)

			store := &ArticleStore{
				db: gormDB,
			}

			article, err := store.GetByID(tc.id)

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("Unfulfilled expectations: %s", err)
			}

			if tc.expectedError != nil {
				assert.Equal(t, tc.expectedError, err)
			}

			if tc.expectArticle {
				assert.NotNil(t, article)
				assert.IsType(t, &model.Article{}, article)
			} else {
				assert.Nil(t, article)
			}

			t.Logf("Test case '%s' completed successfully", tc.name)
		})
	}
}
