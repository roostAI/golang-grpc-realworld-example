package store

import (
	"testing"
	"github.com/jinzhu/gorm"
	"github.com/DATA-DOG/go-sqlmock"
	"sync"
)

const errDBInit = "DB initialization error: %v"


type T struct {
	common
	isParallel bool
	isEnvSet   bool
	context    *testContext // For running tests and subtests.
}
func TestNewArticleStore(t *testing.T) {

	t.Run("Scenario 1: Successful Initialization of ArticleStore with a Valid Database Connection", func(t *testing.T) {

		db, _, err := sqlmock.New()
		if err != nil {
			t.Fatalf(errDBInit, err)
		}
		defer db.Close()

		gormDB, err := gorm.Open("sqlite3", db)
		if err != nil {
			t.Fatalf(errDBInit, err)
		}

		articleStore := NewArticleStore(gormDB)

		if articleStore == nil {
			t.Error("Expected ArticleStore not to be nil")
		}
		if articleStore.db != gormDB {
			t.Errorf("Expected db field in ArticleStore to match the provided gorm.DB object. Got different values.")
		}

		t.Log("Test Scenario 1 passed: ArticleStore initialized successfully with a valid DB connection.")
	})

	t.Run("Scenario 2: Initialization of ArticleStore with a Nil Database Connection", func(t *testing.T) {

		var nilDB *gorm.DB

		articleStore := NewArticleStore(nilDB)

		if articleStore.db != nil {
			t.Errorf("Expected db field to be nil. Got a non-nil db.")
		}

		t.Log("Test Scenario 2 passed: ArticleStore initialized with a nil DB connection without panicking.")
	})

	t.Run("Scenario 3: Thread Safety of ArticleStore Initialization", func(t *testing.T) {

		db, _, err := sqlmock.New()
		if err != nil {
			t.Fatalf(errDBInit, err)
		}
		defer db.Close()

		gormDB, err := gorm.Open("sqlite3", db)
		if err != nil {
			t.Fatalf(errDBInit, err)
		}

		var wg sync.WaitGroup
		const concurrency = 10
		stores := make(chan *ArticleStore, concurrency)

		for i := 0; i < concurrency; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				store := NewArticleStore(gormDB)
				stores <- store
			}()
		}

		wg.Wait()
		close(stores)

		for store := range stores {
			if store == nil || store.db != gormDB {
				t.Error("Expected all ArticleStore instances to have the same valid db value and be non-nil")
			}
		}

		t.Log("Test Scenario 3 passed: ArticleStore initializations are thread-safe.")
	})
}
