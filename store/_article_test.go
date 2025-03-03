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
		dbSetup func(*mockDB)
		wantErr bool
	}{
		{
			name: "Successfully Update an Existing Article",
			article: &model.Article{
				Model: gorm.Model{ID: 1},
				Title: "Updated Title",
				Body:  "Updated Body",
			},
			dbSetup: func(db *mockDB) {
				db.On("Model", mock.Anything).Return(db)
				db.On("Update", mock.Anything).Return(&gorm.DB{Error: nil})
			},
			wantErr: false,
		},
		{
			name: "Attempt to Update a Non-existent Article",
			article: &model.Article{
				Model: gorm.Model{ID: 999},
				Title: "Non-existent Article",
			},
			dbSetup: func(db *mockDB) {
				db.On("Model", mock.Anything).Return(db)
				db.On("Update", mock.Anything).Return(&gorm.DB{Error: gorm.ErrRecordNotFound})
			},
			wantErr: true,
		},
		{
			name: "Handle Database Connection Error",
			article: &model.Article{
				Model: gorm.Model{ID: 1},
				Title: "Connection Error Test",
			},
			dbSetup: func(db *mockDB) {
				db.On("Model", mock.Anything).Return(db)
				db.On("Update", mock.Anything).Return(&gorm.DB{Error: errors.New("connection error")})
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
			dbSetup: func(db *mockDB) {
				db.On("Model", mock.Anything).Return(db)
				db.On("Update", mock.Anything).Return(&gorm.DB{Error: nil})
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := new(mockDB)
			tt.dbSetup(mockDB)

			store := &ArticleStore{
				db: mockDB,
			}

			err := store.Update(tt.article)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockDB.AssertExpectations(t)
		})
	}
}

func TestArticleStoreUpdateConcurrent(t *testing.T) {
	mockDB := new(mockDB)
	mockDB.On("Model", mock.Anything).Return(mockDB)
	mockDB.On("Update", mock.Anything).Return(&gorm.DB{Error: nil})

	store := &ArticleStore{
		db: mockDB,
	}

	article1 := &model.Article{Model: gorm.Model{ID: 1}, Title: "Concurrent Update 1"}
	article2 := &model.Article{Model: gorm.Model{ID: 1}, Title: "Concurrent Update 2"}

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

	mockDB.AssertExpectations(t)
}

func (m *mockDB) Model(value interface{}) *gorm.DB {
	args := m.Called(value)
	return args.Get(0).(*gorm.DB)
}

func (m *mockDB) Update(attrs ...interface{}) *gorm.DB {
	args := m.Called(attrs...)
	return args.Get(0).(*gorm.DB)
}

