package store

import (
	"errors"
	"math"
	"testing"

	"github.com/jinzhu/gorm"
	"github.com/raahii/golang-grpc-realworld-example/model"
	"github.com/stretchr/testify/assert"
)

/*
ROOST_METHOD_HASH=GetCommentByID_7ecaa81f20
ROOST_METHOD_SIG_HASH=GetCommentByID_f6f8a51973

FUNCTION_DEF=func (s *ArticleStore) GetCommentByID(id uint) (*model.Comment, error) // GetCommentByID finds an comment from id
*/
func TestArticleStoreGetCommentById(t *testing.T) {
	tests := []struct {
		name    string
		id      uint
		want    *model.Comment
		wantErr error
		mockDB  func() *gorm.DB
	}{
		{
			name: "Successfully retrieve an existing comment",
			id:   1,
			want: &model.Comment{
				Model: gorm.Model{ID: 1},
				Body:  "Test comment",
			},
			wantErr: nil,
			mockDB: func() *gorm.DB {
				db := &gorm.DB{}
				db.AddError(nil)
				return db
			},
		},
		{
			name:    "Attempt to retrieve a non-existent comment",
			id:      999,
			want:    nil,
			wantErr: gorm.ErrRecordNotFound,
			mockDB: func() *gorm.DB {
				db := &gorm.DB{}
				db.AddError(gorm.ErrRecordNotFound)
				return db
			},
		},
		{
			name:    "Handle database error during retrieval",
			id:      2,
			want:    nil,
			wantErr: errors.New("database error"),
			mockDB: func() *gorm.DB {
				db := &gorm.DB{}
				db.AddError(errors.New("database error"))
				return db
			},
		},
		{
			name:    "Retrieve a comment with ID 0",
			id:      0,
			want:    nil,
			wantErr: gorm.ErrRecordNotFound,
			mockDB: func() *gorm.DB {
				db := &gorm.DB{}
				db.AddError(gorm.ErrRecordNotFound)
				return db
			},
		},
		{
			name:    "Retrieve a comment with maximum uint value",
			id:      math.MaxUint32,
			want:    nil,
			wantErr: gorm.ErrRecordNotFound,
			mockDB: func() *gorm.DB {
				db := &gorm.DB{}
				db.AddError(gorm.ErrRecordNotFound)
				return db
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := tt.mockDB()
			s := &ArticleStore{
				db: mockDB,
			}

			got, err := s.GetCommentByID(tt.id)

			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.wantErr, err)
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tt.want, got)
		})
	}
}
