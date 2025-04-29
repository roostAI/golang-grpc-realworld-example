package store

import (
	sql "database/sql"
	errors "errors"
	debug "runtime/debug"
	testing "testing"

	"github.com/DATA-DOG/go-sqlmock"
	gorm "github.com/jinzhu/gorm"
	model "github.com/raahii/golang-grpc-realworld-example/model"
	assert "github.com/stretchr/testify/assert"
	require "github.com/stretchr/testify/require"
)

type testCase struct {
	name      string
	userIDs   []uint
	limit     int64
	offset    int64
	expected  []model.Article
	expectErr bool
	setupMock func(mockSql sqlmock.Sqlmock)
}

/*
ROOST_METHOD_HASH=ArticleStore_IsFavorited_799826fee5
ROOST_METHOD_SIG_HASH=ArticleStore_IsFavorited_f6d5e67492

FUNCTION_DEF=func (s *ArticleStore) IsFavorited(a *model.Article, u *model.User) (bool, error) // IsFavorited returns whether the article is favorited by the user
*/
func TestArticleStoreIsFavorited(t *testing.T) {

	tests := []struct {
		name        string
		article     *model.Article
		user        *model.User
		mockCount   int
		mockError   error
		expected    bool
		expectError bool
	}{
		{
			name:        "Both Article and User are nil",
			article:     nil,
			user:        nil,
			mockCount:   0,
			mockError:   nil,
			expected:    false,
			expectError: false,
		},
		{
			name:        "Article is nil, User is not nil",
			article:     nil,
			user:        &model.User{Model: gorm.Model{ID: 1}},
			mockCount:   0,
			mockError:   nil,
			expected:    false,
			expectError: false,
		},
		{
			name:        "Article is not nil, User is nil",
			article:     &model.Article{Model: gorm.Model{ID: 1}},
			user:        nil,
			mockCount:   0,
			mockError:   nil,
			expected:    false,
			expectError: false,
		},
		{
			name:        "Both Article and User are valid, Article is favorited by the User",
			article:     &model.Article{Model: gorm.Model{ID: 1}},
			user:        &model.User{Model: gorm.Model{ID: 1}},
			mockCount:   1,
			mockError:   nil,
			expected:    true,
			expectError: false,
		},
		{
			name:        "Both Article and User are valid, Article is not favorited by the User",
			article:     &model.Article{Model: gorm.Model{ID: 1}},
			user:        &model.User{Model: gorm.Model{ID: 1}},
			mockCount:   0,
			mockError:   nil,
			expected:    false,
			expectError: false,
		},
		{
			name:        "Database error occurs while checking favorited status",
			article:     &model.Article{Model: gorm.Model{ID: 1}},
			user:        &model.User{Model: gorm.Model{ID: 1}},
			mockCount:   0,
			mockError:   sql.ErrConnDone,
			expected:    false,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Logf("Panic encountered so failing test. %v\n%s", r, string(debug.Stack()))
					t.Fail()
				}
			}()

			db, mock, err := sqlmock.New()
			require.NoError(t, err)
			defer db.Close()

			if tt.mockError != nil {
				mock.ExpectQuery("SELECT count").WithArgs(tt.article.ID, tt.user.ID).WillReturnError(tt.mockError)
			} else {
				mock.ExpectQuery("SELECT count").WithArgs(tt.article.ID, tt.user.ID).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(tt.mockCount))
			}

			gormDB, err := gorm.Open("sqlmock", db)
			require.NoError(t, err)

			store := &ArticleStore{db: gormDB}

			favorited, err := store.IsFavorited(tt.article, tt.user)

			if tt.expectError {
				require.Error(t, err)
				t.Logf("Expected error: %v, got error: %v", tt.mockError, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.expected, favorited)
				t.Logf("Expected: %v, got: %v", tt.expected, favorited)
			}

			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

/*
ROOST_METHOD_HASH=ArticleStore_GetFeedArticles_a37e1934b6
ROOST_METHOD_SIG_HASH=ArticleStore_GetFeedArticles_f5f09c020e

FUNCTION_DEF=func (s *ArticleStore) GetFeedArticles(userIDs [ // GetFeedArticles returns following users' articles
]uint, limit, offset int64) ([]model.Article, error)
*/
func TestArticleStoreGetFeedArticles(t *testing.T) {
	tests := []testCase{
		{
			name:     "Normal Operation with Valid Parameters",
			userIDs:  []uint{1, 2, 3},
			limit:    5,
			offset:   0,
			expected: []model.Article{{Model: gorm.Model{ID: 1}, UserID: 1}, {Model: gorm.Model{ID: 2}, UserID: 2}, {Model: gorm.Model{ID: 3}, UserID: 3}},
			setupMock: func(mockSql sqlmock.Sqlmock) {
				mockSql.ExpectQuery("SELECT").WillReturnRows(sqlmock.NewRows([]string{"id", "user_id"}).AddRow(1, 1).AddRow(2, 2).AddRow(3, 3))
			},
		},
		{
			name:     "Empty User IDs List",
			userIDs:  []uint{},
			limit:    5,
			offset:   0,
			expected: []model.Article{},
			setupMock: func(mockSql sqlmock.Sqlmock) {
				mockSql.ExpectQuery("SELECT").WillReturnRows(sqlmock.NewRows([]string{"id", "user_id"}))
			},
		},
		{
			name:     "Limit Exceeds Available Articles",
			userIDs:  []uint{1, 2, 3},
			limit:    10,
			offset:   0,
			expected: []model.Article{{Model: gorm.Model{ID: 1}, UserID: 1}, {Model: gorm.Model{ID: 2}, UserID: 2}, {Model: gorm.Model{ID: 3}, UserID: 3}},
			setupMock: func(mockSql sqlmock.Sqlmock) {
				mockSql.ExpectQuery("SELECT").WillReturnRows(sqlmock.NewRows([]string{"id", "user_id"}).AddRow(1, 1).AddRow(2, 2).AddRow(3, 3))
			},
		},
		{
			name:     "Offset Exceeds Available Articles",
			userIDs:  []uint{1, 2, 3},
			limit:    5,
			offset:   10,
			expected: []model.Article{},
			setupMock: func(mockSql sqlmock.Sqlmock) {
				mockSql.ExpectQuery("SELECT").WillReturnRows(sqlmock.NewRows([]string{"id", "user_id"}))
			},
		},
		{
			name:      "Negative Limit",
			userIDs:   []uint{1, 2, 3},
			limit:     -5,
			offset:    0,
			expected:  []model.Article{},
			expectErr: true,
		},
		{
			name:      "Negative Offset",
			userIDs:   []uint{1, 2, 3},
			limit:     5,
			offset:    -1,
			expected:  []model.Article{},
			expectErr: true,
		},
		{
			name:      "Database Error",
			userIDs:   []uint{1, 2, 3},
			limit:     5,
			offset:    0,
			expected:  []model.Article{},
			expectErr: true,
			setupMock: func(mockSql sqlmock.Sqlmock) {
				mockSql.ExpectQuery("SELECT").WillReturnError(errors.New("database error"))
			},
		},
		{
			name:     "No Articles Available",
			userIDs:  []uint{1, 2, 3},
			limit:    5,
			offset:   0,
			expected: []model.Article{},
			setupMock: func(mockSql sqlmock.Sqlmock) {
				mockSql.ExpectQuery("SELECT").WillReturnRows(sqlmock.NewRows([]string{"id", "user_id"}))
			},
		},
		{
			name:     "Preload Author Association",
			userIDs:  []uint{1, 2, 3},
			limit:    5,
			offset:   0,
			expected: []model.Article{{Model: gorm.Model{ID: 1}, UserID: 1, Author: model.User{Model: gorm.Model{ID: 1}}}, {Model: gorm.Model{ID: 2}, UserID: 2, Author: model.User{Model: gorm.Model{ID: 2}}}, {Model: gorm.Model{ID: 3}, UserID: 3, Author: model.User{Model: gorm.Model{ID: 3}}}},
			setupMock: func(mockSql sqlmock.Sqlmock) {
				mockSql.ExpectQuery("SELECT").WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "author_id"}).AddRow(1, 1, 1).AddRow(2, 2, 2).AddRow(3, 3, 3))
			},
		},
		{
			name:     "Large Number of User IDs",
			userIDs:  []uint{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
			limit:    5,
			offset:   0,
			expected: []model.Article{{Model: gorm.Model{ID: 1}, UserID: 1}, {Model: gorm.Model{ID: 2}, UserID: 2}, {Model: gorm.Model{ID: 3}, UserID: 3}, {Model: gorm.Model{ID: 4}, UserID: 4}, {Model: gorm.Model{ID: 5}, UserID: 5}},
			setupMock: func(mockSql sqlmock.Sqlmock) {
				mockSql.ExpectQuery("SELECT").WillReturnRows(sqlmock.NewRows([]string{"id", "user_id"}).AddRow(1, 1).AddRow(2, 2).AddRow(3, 3).AddRow(4, 4).AddRow(5, 5))
			},
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

			db, mockSql, err := sqlmock.New()
			assert.NoError(t, err)
			defer db.Close()

			gormDB, err := gorm.Open("mysql", db)
			assert.NoError(t, err)

			if tc.setupMock != nil {
				tc.setupMock(mockSql)
			}

			store := &ArticleStore{
				db: gormDB,
			}

			articles, err := store.GetFeedArticles(tc.userIDs, tc.limit, tc.offset)

			if tc.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expected, articles)
			}

			err = mockSql.ExpectationsWereMet()
			assert.NoError(t, err)
		})
	}
}
