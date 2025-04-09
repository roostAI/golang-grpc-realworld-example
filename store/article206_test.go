package store

import (
	errors "errors"
	testing "testing"

	"github.com/DATA-DOG/go-sqlmock"
	gorm "github.com/jinzhu/gorm"
)

func NewMockDB() *gorm.DB {
	mockDB, _, _ := sqlmock.New()
	mockGorm, _ := gorm.Open("postgres", mockDB)
	return mockGorm
}

func TestNewArticleStoreScenarios(t *testing.T) {
	tests := []struct {
		name   string
		db     *gorm.DB
		expect func(t *testing.T, store *ArticleStore, err error)
	}{
		{
			name: "Successfully Creating a New Article Store with a Valid Database",
			db:   NewMockDB(),
			expect: func(t *testing.T, store *ArticleStore, err error) {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
					return
				}
				if store == nil {
					t.Errorf("expected an ArticleStore, got nil")
				}
			},
		},
		{
			name: "Error when Null Database is Provided",
			db:   nil,
			expect: func(t *testing.T, store *ArticleStore, err error) {
				if err != errors.New("unable to create ArticleStore with nil DB") {
					t.Errorf("expected error, got nil")
				}
			},
		},
		{
			name: "Successful Reinitialization of Article Store",
			db:   NewMockDB(),
			expect: func(t *testing.T, store *ArticleStore, err error) {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
					return
				}
				if store == nil {
					t.Errorf("expected an ArticleStore, got nil")
				}

				newStore := NewArticleStore(NewMockDB())

				if newStore == nil {
					t.Errorf("expected an ArticleStore, got nil")
				}
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("unexpected panic: %v", r)
				}
			}()

			store := NewArticleStore(tc.db)
			var err error
			if store == nil {
				err = errors.New("returned store is nil")
			}

			tc.expect(t, store, err)
		})
	}
}
