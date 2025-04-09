package store

import (
	sql "database/sql"
	debug "runtime/debug"
	sync "sync"
	testing "testing"

	"github.com/DATA-DOG/go-sqlmock"
	gorm "github.com/jinzhu/gorm"
)

func TestNewArticleStoreConcurrentAndTransactions(t *testing.T) {
	type testCase struct {
		name     string
		setup    func() *gorm.DB
		validate func(*testing.T, *ArticleStore)
	}

	createMockDB := func() (*gorm.DB, sqlmock.Sqlmock) {
		mockDB, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("Failed to create mock DB: %v", err)
		}
		gormDB, err := gorm.Open("mysql", mockDB)
		if err != nil {
			t.Fatalf("Failed to create GORM DB: %v", err)
		}
		return gormDB, mock
	}

	tests := []testCase{
		{
			name: "Concurrent ArticleStore Creation",
			setup: func() *gorm.DB {
				gormDB, _ := createMockDB()
				return gormDB
			},
			validate: func(t *testing.T, _ *ArticleStore) {
				var wg sync.WaitGroup
				stores := make([]*ArticleStore, 10)
				gormDB, _ := createMockDB()

				for i := 0; i < 10; i++ {
					wg.Add(1)
					go func(index int) {
						defer wg.Done()
						stores[index] = NewArticleStore(gormDB)
					}(i)
				}
				wg.Wait()

				for i, store := range stores {
					if store == nil {
						t.Errorf("Store %d was not created", i)
					}
					if store.db != gormDB {
						t.Errorf("Store %d has incorrect DB reference", i)
					}
				}
			},
		},
		{
			name: "Create ArticleStore with Transaction",
			setup: func() *gorm.DB {
				gormDB, mock := createMockDB()
				mock.ExpectBegin()
				return gormDB
			},
			validate: func(t *testing.T, store *ArticleStore) {
				tx := store.db.Begin()
				if tx.Error != nil {
					t.Errorf("Failed to begin transaction: %v", tx.Error)
				}

				txStore := NewArticleStore(tx)
				if txStore == nil {
					t.Error("Failed to create ArticleStore with transaction")
				}
				if txStore.db != tx {
					t.Error("Transaction DB not properly set in ArticleStore")
				}
			},
		},
		{
			name: "Create ArticleStore with Closed Connection",
			setup: func() *gorm.DB {
				gormDB, _ := createMockDB()
				gormDB.Close()
				return gormDB
			},
			validate: func(t *testing.T, store *ArticleStore) {
				if store == nil {
					t.Error("Expected non-nil ArticleStore even with closed connection")
				}
				if store.db == nil {
					t.Error("Expected non-nil DB in ArticleStore")
				}

				if err := store.db.DB().Ping(); err != sql.ErrConnDone {
					t.Errorf("Expected error on closed connection, got: %v", err)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Logf("Panic recovered in %s: %v\n%s", tt.name, r, debug.Stack())
					t.Fail()
				}
			}()

			db := tt.setup()
			store := NewArticleStore(db)
			tt.validate(t, store)
		})
	}
}
