package store

import (
	errors "errors"
	testing "testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	gorm "github.com/jinzhu/gorm"
)

func TestNewArticleStore(t *testing.T) {

	tests := []struct {
		name   string
		db     *gorm.DB
		expect func(t *testing.T, store *ArticleStore, err error)
	}{
		{
			name: "Correctly Initialized ArticleStore",
			db: func() *gorm.DB {
				mockDB, _, _ := sqlmock.New()
				mockGorm, _ := gorm.Open("postgres", mockDB)
				return mockGorm
			}(),
			expect: func(t *testing.T, store *ArticleStore, err error) {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
					return
				}
				if store == nil {
					t.Errorf("expected an ArticleStore, got nil")
				}
				t.Log("pass Correctly Initialized ArticleStore")
			},
		},
		{
			name: "Error when Null DB provided",
			db:   nil,
			expect: func(t *testing.T, store *ArticleStore, err error) {
				if err != errors.New("unable to create ArticleStore with nil DB") {
					t.Errorf("expected error, got nil")
				}
				t.Log("pass Error when Null DB provided")
			},
		},
		{
			name: "ArticleStore Reinitialization",
			db: func() *gorm.DB {
				mockDB, _, _ := sqlmock.New()
				mockGorm, _ := gorm.Open("postgres", mockDB)
				return mockGorm
			}(),
			expect: func(t *testing.T, store *ArticleStore, err error) {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
					return
				}
				if store == nil {
					t.Errorf("expected an ArticleStore, got nil")
				}
				t.Log("pass ArticleStore Reinitialization")
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Logf("Panic encountered so failing test. %v \n", r)
					t.Fail()
				}
			}()
			result := NewArticleStore(tc.db)
			if result != nil {
				tc.expect(t, result, nil)
				return
			}

		})
	}
}
