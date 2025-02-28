package store

import (
	errors "errors"
	testing "testing"
	gorm "github.com/jinzhu/gorm"
	model "github.com/raahii/golang-grpc-realworld-example/model"
)





type mockDB struct {
	updateFunc func(interface{}) *gorm.DB
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
		mockDB  func(interface{}) *gorm.DB
		wantErr bool
	}{
		{
			name: "Successfully Update an Existing Article",
			article: &model.Article{
				Model: gorm.Model{ID: 1},
				Title: "Updated Title",
				Body:  "Updated Body",
			},
			mockDB: func(value interface{}) *gorm.DB {
				return &gorm.DB{Error: nil}
			},
			wantErr: false,
		},
		{
			name: "Attempt to Update a Non-existent Article",
			article: &model.Article{
				Model: gorm.Model{ID: 999},
				Title: "Non-existent Article",
			},
			mockDB: func(value interface{}) *gorm.DB {
				return &gorm.DB{Error: gorm.ErrRecordNotFound}
			},
			wantErr: true,
		},
		{
			name: "Update with Invalid Data",
			article: &model.Article{
				Model: gorm.Model{ID: 1},
				Title: "",
			},
			mockDB: func(value interface{}) *gorm.DB {
				return &gorm.DB{Error: errors.New("invalid data")}
			},
			wantErr: true,
		},
		{
			name: "Update with No Changes",
			article: &model.Article{
				Model: gorm.Model{ID: 1},
				Title: "Unchanged Title",
			},
			mockDB: func(value interface{}) *gorm.DB {
				return &gorm.DB{Error: nil, RowsAffected: 0}
			},
			wantErr: false,
		},
		{
			name: "Update Article Tags",
			article: &model.Article{
				Model: gorm.Model{ID: 1},
				Title: "Article with Tags",
				Tags:  []model.Tag{{Name: "new tag"}},
			},
			mockDB: func(value interface{}) *gorm.DB {
				return &gorm.DB{Error: nil}
			},
			wantErr: false,
		},
		{
			name: "Update with Database Connection Error",
			article: &model.Article{
				Model: gorm.Model{ID: 1},
				Title: "Connection Error",
			},
			mockDB: func(value interface{}) *gorm.DB {
				return &gorm.DB{Error: errors.New("connection error")}
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := &ArticleStore{
				db: &mockDB{
					updateFunc: tt.mockDB,
				},
			}

			err := store.Update(tt.article)

			if (err != nil) != tt.wantErr {
				t.Errorf("ArticleStore.Update() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func (m *mockDB) Model(value interface{}) *gorm.DB {
	return &gorm.DB{Error: nil, Value: value}
}

func (m *mockDB) Update(value interface{}) *gorm.DB {
	return m.updateFunc(value)
}

