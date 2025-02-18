package store

import (
	"errors"
	"sync"
	"testing"
	"github.com/jinzhu/gorm"
	"github.com/raahii/golang-grpc-realworld-example/model"
	"github.com/stretchr/testify/assert"
)





type MockArticleStore struct {
	db *MockDB
}
type MockDB struct {
	CreateFunc func(value interface{}) *gorm.DB
}
type mockDB struct {
	*gorm.DB
}


/*
ROOST_METHOD_HASH=ArticleStore_Create_1273475ade
ROOST_METHOD_SIG_HASH=ArticleStore_Create_a27282cad5

FUNCTION_DEF=func (s *ArticleStore) Create(m *model.Article) error // Create creates an article


*/
func (s *MockArticleStore) Create(m *model.Article) error {
	return s.db.Create(m).Error
}

func (m *MockDB) Create(value interface{}) *gorm.DB {
	return m.CreateFunc(value)
}

func (m *MockDB) NewScope(value interface{}) *gorm.Scope {
	return &gorm.Scope{}
}

func TestArticleStoreArticleStoreCreate(t *testing.T) {
	tests := []struct {
		name    string
		article *model.Article
		mockDB  *MockDB
		wantErr bool
	}{
		{
			name: "Successfully Create a New Article",
			article: &model.Article{
				Title:       "Test Article",
				Description: "Test Description",
				Body:        "Test Body",
				UserID:      1,
			},
			mockDB: &MockDB{
				CreateFunc: func(value interface{}) *gorm.DB {
					return &gorm.DB{Error: nil}
				},
			},
			wantErr: false,
		},
		{
			name: "Attempt to Create an Article with Invalid Data",
			article: &model.Article{
				Title: "",
				Body:  "Test Body",
			},
			mockDB: &MockDB{
				CreateFunc: func(value interface{}) *gorm.DB {
					return &gorm.DB{Error: errors.New("invalid data")}
				},
			},
			wantErr: true,
		},
		{
			name: "Create Article with Duplicate Unique Fields",
			article: &model.Article{
				Title: "Existing Title",
				Body:  "Test Body",
			},
			mockDB: &MockDB{
				CreateFunc: func(value interface{}) *gorm.DB {
					return &gorm.DB{Error: errors.New("duplicate entry")}
				},
			},
			wantErr: true,
		},
		{
			name: "Create Article with Associated Tags",
			article: &model.Article{
				Title: "Article with Tags",
				Body:  "Test Body",
				Tags:  []model.Tag{{Name: "tag1"}, {Name: "tag2"}},
			},
			mockDB: &MockDB{
				CreateFunc: func(value interface{}) *gorm.DB {
					return &gorm.DB{Error: nil}
				},
			},
			wantErr: false,
		},
		{
			name: "Create Article with Database Connection Error",
			article: &model.Article{
				Title: "Test Article",
				Body:  "Test Body",
			},
			mockDB: &MockDB{
				CreateFunc: func(value interface{}) *gorm.DB {
					return &gorm.DB{Error: errors.New("connection error")}
				},
			},
			wantErr: true,
		},
		{
			name: "Create Article with Very Large Content",
			article: &model.Article{
				Title: "Large Article",
				Body:  string(make([]byte, 1<<20)),
			},
			mockDB: &MockDB{
				CreateFunc: func(value interface{}) *gorm.DB {
					return &gorm.DB{Error: nil}
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := &MockArticleStore{
				db: tt.mockDB,
			}

			err := store.Create(tt.article)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}

	t.Run("Create Multiple Articles Concurrently", func(t *testing.T) {
		numArticles := 10
		var wg sync.WaitGroup
		mockDB := &MockDB{
			CreateFunc: func(value interface{}) *gorm.DB {
				return &gorm.DB{Error: nil}
			},
		}
		store := &MockArticleStore{
			db: mockDB,
		}

		for i := 0; i < numArticles; i++ {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				article := &model.Article{
					Title: "Concurrent Article",
					Body:  "Test Body",
				}
				err := store.Create(article)
				assert.NoError(t, err)
			}(i)
		}

		wg.Wait()
	})
}


/*
ROOST_METHOD_HASH=ArticleStore_GetArticles_101b7250e8
ROOST_METHOD_SIG_HASH=ArticleStore_GetArticles_91bc0a6760

FUNCTION_DEF=func (s *ArticleStore) GetArticles(tagName, username string, favoritedBy *model.User, limit, offset int64) ([ // GetArticles get global articles
]model.Article, error) 

*/
func TestArticleStoreArticleStoreGetArticles(t *testing.T) {
	tests := []struct {
		name        string
		tagName     string
		username    string
		favoritedBy *model.User
		limit       int64
		offset      int64
		mockSetup   func(*mockDB)
		want        []model.Article
		wantErr     bool
	}{
		{
			name:        "Get Articles Without Filters",
			tagName:     "",
			username:    "",
			favoritedBy: nil,
			limit:       10,
			offset:      0,
			mockSetup: func(m *mockDB) {
				m.DB.Error = nil

				m.DB.Error = nil
				m.DB.Value = []model.Article{{Title: "Test Article"}}
			},
			want:    []model.Article{{Title: "Test Article"}},
			wantErr: false,
		},
		{
			name:        "Get Articles by Tag Name",
			tagName:     "test-tag",
			username:    "",
			favoritedBy: nil,
			limit:       10,
			offset:      0,
			mockSetup: func(m *mockDB) {
				m.DB.Error = nil

				m.DB.Value = []model.Article{{Title: "Tagged Article", Tags: []model.Tag{{Name: "test-tag"}}}}
			},
			want:    []model.Article{{Title: "Tagged Article", Tags: []model.Tag{{Name: "test-tag"}}}},
			wantErr: false,
		},
		{
			name:        "Get Articles by Author Username",
			tagName:     "",
			username:    "test-user",
			favoritedBy: nil,
			limit:       10,
			offset:      0,
			mockSetup: func(m *mockDB) {
				m.DB.Error = nil

				m.DB.Value = []model.Article{{Title: "Author's Article", Author: model.User{Username: "test-user"}}}
			},
			want:    []model.Article{{Title: "Author's Article", Author: model.User{Username: "test-user"}}},
			wantErr: false,
		},
		{
			name:        "Get Favorited Articles",
			tagName:     "",
			username:    "",
			favoritedBy: &model.User{Model: gorm.Model{ID: 1}},
			limit:       10,
			offset:      0,
			mockSetup: func(m *mockDB) {
				m.DB.Error = nil

				m.DB.Value = []model.Article{{Title: "Favorited Article", FavoritesCount: 1}}
			},
			want:    []model.Article{{Title: "Favorited Article", FavoritesCount: 1}},
			wantErr: false,
		},
		{
			name:        "Database Error",
			tagName:     "",
			username:    "",
			favoritedBy: nil,
			limit:       10,
			offset:      0,
			mockSetup: func(m *mockDB) {
				m.DB.Error = gorm.ErrRecordNotFound
			},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := &mockDB{&gorm.DB{}}
			tt.mockSetup(mockDB)

			s := &ArticleStore{
				db: mockDB.DB,
			}

			got, err := s.GetArticles(tt.tagName, tt.username, tt.favoritedBy, tt.limit, tt.offset)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tt.want, got)

			if tt.name == "Verify Preloading of Author" {
				for _, article := range got {
					assert.NotNil(t, article.Author)
				}
			}
		})
	}
}

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

func (m *mockDB) Where(query interface{}, args ...interface{}) *gorm.DB {
	return m.DB
}

