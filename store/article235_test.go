package store

import (
	driver "database/sql/driver"
	debug "runtime/debug"
	sync "sync"
	testing "testing"

	"github.com/DATA-DOG/go-sqlmock"
	gorm "github.com/jinzhu/gorm"
)

func TestNewArticleStore_Enhanced(t *testing.T) {
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
			name: "Scenario 6: Handle Closed DB Connection",
			db:   gormDB,
			validate: func(t *testing.T, store *ArticleStore) {
				store.db.Close()
				stats := store.db.DB().Stats()
				if stats.OpenConnections > 0 {
					t.Error("Expected DB connection to be closed")
				}
			},
		},
		{
			name: "Scenario 7: Concurrent Access Safety",
			db:   gormDB,
			validate: func(t *testing.T, store *ArticleStore) {
				var wg sync.WaitGroup
				stores := make([]*ArticleStore, 10)
				for i := 0; i < 10; i++ {
					wg.Add(1)
					go func(index int) {
						defer wg.Done()
						stores[index] = NewArticleStore(gormDB)
					}(i)
				}
				wg.Wait()

				for i, s := range stores {
					if s == nil {
						t.Errorf("Concurrent creation failed for store %d", i)
					}
					if s.db != gormDB {
						t.Errorf("Concurrent creation resulted in incorrect DB reference for store %d", i)
					}
				}
			},
		},
		{
			name: "Scenario 8: DB Connection Type Validation",
			db:   gormDB,
			validate: func(t *testing.T, store *ArticleStore) {
				sqlDB := store.db.DB()
				if sqlDB == nil {
					t.Error("Expected valid DB connection")
					return
				}

				if _, ok := sqlDB.Driver().(driver.Driver); !ok {
					t.Error("Expected DB driver to implement sql.Driver interface")
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
