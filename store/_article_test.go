package store

import (
	errors "errors"
	sync "sync"
	testing "testing"
	gorm "github.com/jinzhu/gorm"
	model "github.com/raahii/golang-grpc-realworld-example/model"
	assert "github.com/stretchr/testify/assert"
	mock "github.com/stretchr/testify/mock"
)





type mockDB struct {
	mock.Mock
}


/*
ROOST_METHOD_HASH=ArticleStore_Update_3cddacb803
ROOST_METHOD_SIG_HASH=ArticleStore_Update_e245edd177

FUNCTION_DEF=func (s *ArticleStore) Update(m *model.Article) error // Update updates an article


*/
func TestArticleStoreUpdate(t *testing.T) {
	tests := []struct {
		name    string
		article *model.Article
		mockDB  func() *mockDB
		wantErr bool
	}{
		{
			name: "Successfully Update an Existing Article",
			article: &model.Article{
				Model: gorm.Model{ID: 1},
				Title: "Updated Title",
				Body:  "Updated Body",
			},
			mockDB: func() *mockDB {
				db := new(mockDB)
				db.On("Model", mock.Anything).Return(db)
				db.On("Update", mock.Anything).Return(&gorm.DB{Error: nil})
				return db
			},
			wantErr: false,
		},
		{
			name: "Attempt to Update a Non-existent Article",
			article: &model.Article{
				Model: gorm.Model{ID: 999},
				Title: "Non-existent Article",
			},
			mockDB: func() *mockDB {
				db := new(mockDB)
				db.On("Model", mock.Anything).Return(db)
				db.On("Update", mock.Anything).Return(&gorm.DB{Error: gorm.ErrRecordNotFound})
				return db
			},
			wantErr: true,
		},
		{
			name: "Handle Database Connection Error",
			article: &model.Article{
				Model: gorm.Model{ID: 1},
				Title: "Connection Error",
			},
			mockDB: func() *mockDB {
				db := new(mockDB)
				db.On("Model", mock.Anything).Return(db)
				db.On("Update", mock.Anything).Return(&gorm.DB{Error: errors.New("connection error")})
				return db
			},
			wantErr: true,
		},
		{
			name: "Update Article with Empty Fields",
			article: &model.Article{
				Model: gorm.Model{ID: 1},
				Title: "",
				Body:  "",
			},
			mockDB: func() *mockDB {
				db := new(mockDB)
				db.On("Model", mock.Anything).Return(db)
				db.On("Update", mock.Anything).Return(&gorm.DB{Error: nil})
				return db
			},
			wantErr: false,
		},
		{
			name: "Update Article with Very Large Content",
			article: &model.Article{
				Model: gorm.Model{ID: 1},
				Title: string(make([]byte, 10000)),
				Body:  string(make([]byte, 100000)),
			},
			mockDB: func() *mockDB {
				db := new(mockDB)
				db.On("Model", mock.Anything).Return(db)
				db.On("Update", mock.Anything).Return(&gorm.DB{Error: nil})
				return db
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := tt.mockDB()
			store := &ArticleStore{db: mockDB}

			err := store.Update(tt.article)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockDB.AssertExpectations(t)
		})
	}

	t.Run("Concurrent Updates to the Same Article", func(t *testing.T) {
		mockDB := new(mockDB)
		mockDB.On("Model", mock.Anything).Return(mockDB)
		mockDB.On("Update", mock.Anything).Return(&gorm.DB{Error: nil})

		store := &ArticleStore{db: mockDB}
		article := &model.Article{Model: gorm.Model{ID: 1}}

		var wg sync.WaitGroup
		updateCount := 10

		for i := 0; i < updateCount; i++ {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				article.Title = "Concurrent Update " + string(i)
				err := store.Update(article)
				assert.NoError(t, err)
			}(i)
		}

		wg.Wait()
		mockDB.AssertNumberOfCalls(t, "Update", updateCount)
	})
}

func (m *mockDB) Model(value interface{}) *gorm.DB {
	args := m.Called(value)
	return args.Get(0).(*gorm.DB)
}

func (m *mockDB) Update(attrs ...interface{}) *gorm.DB {
	args := m.Called(attrs...)
	return args.Get(0).(*gorm.DB)
}

