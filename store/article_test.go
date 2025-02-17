package store

import (
	"database/sql"
	"errors"
	"testing"

	"github.com/jinzhu/gorm"
	"github.com/raahii/golang-grpc-realworld-example/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockArticleStore struct {
	mock.Mock
}

type MockDB struct {
	mock.Mock
}

type mockDB struct {
	*gorm.DB
}

/*
ROOST_METHOD_HASH=Create_1273475ade
ROOST_METHOD_SIG_HASH=Create_a27282cad5

FUNCTION_DEF=func (s *ArticleStore) Create(m *model.Article) error // Create creates an article
*/
func (m *MockArticleStore) Create(article *model.Article) error {
	args := m.Called(article)
	return args.Error(0)
}

func (m *MockDB) Create(value interface{}) *gorm.DB {
	args := m.Called(value)
	return args.Get(0).(*gorm.DB)
}

func (m *MockDB) NewScope(value interface{}) *gorm.Scope {
	args := m.Called(value)
	return args.Get(0).(*gorm.Scope)
}

func TestArticleStoreCreate(t *testing.T) {
	tests := []struct {
		name    string
		article *model.Article
		dbError error
		wantErr bool
	}{
		{
			name: "Successfully Create a New Article",
			article: &model.Article{
				Title:       "Test Article",
				Description: "This is a test article",
				Body:        "Test body content",
				UserID:      1,
			},
			dbError: nil,
			wantErr: false,
		},
		{
			name: "Attempt to Create an Article with Invalid Data",
			article: &model.Article{
				Title:       "",
				Description: "This is a test article",
				Body:        "Test body content",
				UserID:      1,
			},
			dbError: errors.New("validation error"),
			wantErr: true,
		},
		{
			name: "Database Error During Article Creation",
			article: &model.Article{
				Title:       "Test Article",
				Description: "This is a test article",
				Body:        "Test body content",
				UserID:      1,
			},
			dbError: errors.New("database error"),
			wantErr: true,
		},
		{
			name: "Create Article with Associated Tags",
			article: &model.Article{
				Title:       "Test Article with Tags",
				Description: "This is a test article with tags",
				Body:        "Test body content",
				UserID:      1,
				Tags: []model.Tag{
					{Name: "tag1"},
					{Name: "tag2"},
				},
			},
			dbError: nil,
			wantErr: false,
		},
		{
			name: "Create Article with Maximum Allowed Content Length",
			article: &model.Article{
				Title:       string(make([]byte, 255)),
				Description: string(make([]byte, 1000)),
				Body:        string(make([]byte, 10000)),
				UserID:      1,
			},
			dbError: nil,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStore := new(MockArticleStore)

			mockStore.On("Create", tt.article).Return(tt.dbError)

			err := mockStore.Create(tt.article)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Equal(t, tt.dbError, err)
			} else {
				assert.NoError(t, err)
			}

			mockStore.AssertCalled(t, "Create", tt.article)
		})
	}
}

/*
ROOST_METHOD_HASH=GetArticles_101b7250e8
ROOST_METHOD_SIG_HASH=GetArticles_91bc0a6760

FUNCTION_DEF=func (s *ArticleStore) GetArticles(tagName, username string, favoritedBy *model.User, limit, offset int64) ([]model.Article, error) // GetArticles get global articles
*/
func (m *mockDB) Find(out interface{}, where ...interface{}) *gorm.DB {
	return m.DB
}

func (m *mockDB) Joins(query string, args ...interface{}) *gorm.DB {
	return m.DB
}

func (m *mockDB) Limit(limit interface{}) *gorm.DB {
	return m.DB
}

func (m *mockDB) Offset(offset interface{}) *gorm.DB {
	return m.DB
}

func (m *mockDB) Preload(column string, conditions ...interface{}) *gorm.DB {
	return m.DB
}

func (m *mockDB) Rows() (*sql.Rows, error) {
	return nil, nil
}

func (m *mockDB) Select(query interface{}, args ...interface{}) *gorm.DB {
	return m.DB
}

func (m *mockDB) Table(name string) *gorm.DB {
	return m.DB
}

func TestArticleStoreGetArticles(t *testing.T) {
	tests := []struct {
		name        string
		tagName     string
		username    string
		favoritedBy *model.User
		limit       int64
		offset      int64
		mockSetup   func(*mockDB)
		expected    []model.Article
		expectedErr error
	}{
		{
			name:        "Get Articles with No Filters",
			tagName:     "",
			username:    "",
			favoritedBy: nil,
			limit:       10,
			offset:      0,
			mockSetup: func(m *mockDB) {
			},
			expected:    []model.Article{},
			expectedErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := &mockDB{DB: &gorm.DB{}}
			tt.mockSetup(mockDB)

			store := &ArticleStore{db: mockDB.DB}

			articles, err := store.GetArticles(tt.tagName, tt.username, tt.favoritedBy, tt.limit, tt.offset)

			assert.Equal(t, tt.expectedErr, err)
			assert.Equal(t, tt.expected, articles)
		})
	}
}

func (m *mockDB) Where(query interface{}, args ...interface{}) *gorm.DB {
	return m.DB
}
