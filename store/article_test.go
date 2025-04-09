package store

import (
	debug "runtime/debug"
	testing "testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	gorm "github.com/jinzhu/gorm"
)

func TestNewArticleStore(t *testing.T) {

	type testCase struct {
		name     string
		db       *gorm.DB
		wantNil  bool
		validate func(*testing.T, *ArticleStore)
	}

	mockDB, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock DB: %v", err)
	}
	defer mockDB.Close()

	gormDB, err := gorm.Open("mysql", mockDB)
	if err != nil {
		t.Fatalf("Failed to create GORM DB: %v", err)
	}
	defer gormDB.Close()

	tests := []testCase{
		{
			name: "Scenario 1: Successfully Create New ArticleStore with Valid DB Connection",
			db:   gormDB,
			validate: func(t *testing.T, store *ArticleStore) {
				if store == nil {
					t.Error("Expected non-nil ArticleStore")
				}
				if store.db != gormDB {
					t.Error("DB connection not properly set")
				}
			},
		},
		{
			name: "Scenario 2: Create ArticleStore with Nil DB Connection",
			db:   nil,
			validate: func(t *testing.T, store *ArticleStore) {
				if store == nil {
					t.Error("Expected non-nil ArticleStore even with nil DB")
				}
				if store.db != nil {
					t.Error("Expected nil DB in ArticleStore")
				}
			},
		},
		{
			name: "Scenario 3: Verify ArticleStore Instance Independence",
			db:   gormDB,
			validate: func(t *testing.T, store *ArticleStore) {

				store2 := NewArticleStore(gormDB)
				if store == store2 {
					t.Error("Expected different instances of ArticleStore")
				}
				if store.db != store2.db {
					t.Error("Expected same DB connection in both instances")
				}
			},
		},
		{
			name: "Scenario 4: Verify DB Connection Persistence",
			db:   gormDB,
			validate: func(t *testing.T, store *ArticleStore) {
				if store.db != gormDB {
					t.Error("DB connection not persisted correctly")
				}

				if store.db.Error != nil {
					t.Errorf("DB connection invalid: %v", store.db.Error)
				}
			},
		},
		{
			name: "Scenario 5: Memory Resource Management",
			db:   gormDB,
			validate: func(t *testing.T, store *ArticleStore) {

				var stores []*ArticleStore
				for i := 0; i < 100; i++ {
					stores = append(stores, NewArticleStore(gormDB))
				}

				for i, s := range stores {
					if s == nil {
						t.Errorf("Instance %d is nil", i)
					}
					if s.db != gormDB {
						t.Errorf("Instance %d has incorrect DB reference", i)
					}
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			defer func() {
				if r := recover(); r != nil {
					t.Logf("Panic encountered so failing test. %v\n%s", r, string(debug.Stack()))
					t.Fail()
				}
			}()

			t.Log("Starting test:", tt.name)

			got := NewArticleStore(tt.db)

			if tt.validate != nil {
				tt.validate(t, got)
			}

			t.Log("Successfully completed test:", tt.name)
		})
	}
}
