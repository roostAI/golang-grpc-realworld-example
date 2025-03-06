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
		dbError error
		wantErr bool
	}{
		{
			name: "Successfully Update an Existing Article",
			article: &model.Article{
				Model: gorm.Model{ID: 1},
				Title: "Updated Title",
				Body:  "Updated Body",
			},
			dbError: nil,
			wantErr: false,
		},
		{
			name: "Attempt to Update a Non-existent Article",
			article: &model.Article{
				Model: gorm.Model{ID: 999},
				Title: "Non-existent Article",
			},
			dbError: gorm.ErrRecordNotFound,
			wantErr: true,
		},
		{
			name: "Handle Database Connection Error",
			article: &model.Article{
				Model: gorm.Model{ID: 2},
				Title: "Connection Error Article",
			},
			dbError: errors.New("database connection error"),
			wantErr: true,
		},
		{
			name: "Update Article with Empty Fields",
			article: &model.Article{
				Model: gorm.Model{ID: 3},
				Title: "",
				Body:  "",
			},
			dbError: nil,
			wantErr: false,
		},
		{
			name: "Update Article with Very Large Content",
			article: &model.Article{
				Model: gorm.Model{ID: 4},
				Title: "Large Content Article",
				Body:  string(make([]byte, 1000000)),
			},
			dbError: nil,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := new(mockDB)
			mockDB.On("Model", mock.Anything).Return(mockDB)
			mockDB.On("Update", mock.Anything).Return(&gorm.DB{Error: tt.dbError})

			store := &ArticleStore{db: mockDB}

			err := store.Update(tt.article)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.dbError == gorm.ErrRecordNotFound {
					assert.Equal(t, gorm.ErrRecordNotFound, err)
				}
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

		article1 := &model.Article{Model: gorm.Model{ID: 5}, Title: "Concurrent Update 1"}
		article2 := &model.Article{Model: gorm.Model{ID: 5}, Title: "Concurrent Update 2"}

		var wg sync.WaitGroup
		wg.Add(2)

		go func() {
			defer wg.Done()
			err := store.Update(article1)
			assert.NoError(t, err)
		}()

		go func() {
			defer wg.Done()
			err := store.Update(article2)
			assert.NoError(t, err)
		}()

		wg.Wait()

		mockDB.AssertNumberOfCalls(t, "Update", 2)
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

