package store

import (
	errors "errors"
	testing "testing"
	gorm "github.com/jinzhu/gorm"
	model "github.com/raahii/golang-grpc-realworld-example/model"
)





type mockDB struct {
	updateErr error
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
		dbErr   error
		wantErr bool
	}{
		{
			name: "Successfully Update an Existing Article",
			article: &model.Article{
				Model: gorm.Model{ID: 1},
				Title: "Updated Title",
				Body:  "Updated Body",
			},
			dbErr:   nil,
			wantErr: false,
		},
		{
			name: "Attempt to Update a Non-existent Article",
			article: &model.Article{
				Model: gorm.Model{ID: 999},
				Title: "Non-existent Article",
			},
			dbErr:   gorm.ErrRecordNotFound,
			wantErr: true,
		},
		{
			name: "Update with Invalid Data",
			article: &model.Article{
				Model: gorm.Model{ID: 1},
				Title: "",
			},
			dbErr:   errors.New("invalid data"),
			wantErr: true,
		},
		{
			name: "Update with No Changes",
			article: &model.Article{
				Model: gorm.Model{ID: 1},
				Title: "Existing Title",
				Body:  "Existing Body",
			},
			dbErr:   nil,
			wantErr: false,
		},
		{
			name: "Update Article Tags",
			article: &model.Article{
				Model: gorm.Model{ID: 1},
				Title: "Article with Tags",
				Tags:  []model.Tag{{Name: "new-tag"}},
			},
			dbErr:   nil,
			wantErr: false,
		},
		{
			name: "Update with Database Connection Error",
			article: &model.Article{
				Model: gorm.Model{ID: 1},
				Title: "Connection Error Article",
			},
			dbErr:   errors.New("database connection error"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := &mockDB{updateErr: tt.dbErr}
			s := &ArticleStore{db: mockDB}

			err := s.Update(tt.article)

			if (err != nil) != tt.wantErr {
				t.Errorf("ArticleStore.Update() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func (m *mockDB) Model(value interface{}) *gorm.DB {
	return &gorm.DB{}
}

func (m *mockDB) Update(attrs ...interface{}) *gorm.DB {
	return &gorm.DB{Error: m.updateErr}
}

