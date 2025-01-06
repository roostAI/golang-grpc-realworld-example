package store

import (
	"testing"
	"github.com/jinzhu/gorm"
	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/raahii/golang-grpc-realworld-example/store"
)

type ExpectedQuery struct {
	queryBasedExpectation
	rows             driver.Rows
	delay            time.Duration
	rowsMustBeClosed bool
	rowsWereClosed   bool
}

type T struct {
	common
	isParallel bool
	isEnvSet   bool
	context    *testContext // For running tests and subtests.
}
func TestNewUserStore(t *testing.T) {

	t.Run("Scenario 1: Basic Initialization with a Valid gorm.DB Object", func(t *testing.T) {

		db, _, err := sqlmock.New()
		if err != nil {
			t.Fatalf("Failed to create sqlmock: %v", err)
		}
		defer db.Close()
		gormDB, err := gorm.Open("sqlmock", db)
		if err != nil {
			t.Fatalf("Failed to open gorm DB: %v", err)
		}

		userStore := store.NewUserStore(gormDB)

		if userStore.db != gormDB {
			t.Errorf("Expected gorm.DB instance %v, but got %v", gormDB, userStore.db)
		} else {
			t.Log("Successfully initialized UserStore with valid gorm.DB")
		}
	})

	t.Run("Scenario 2: Handling of Nil gorm.DB Object", func(t *testing.T) {

		var nilDB *gorm.DB

		userStore := store.NewUserStore(nilDB)

		if userStore.db != nil {
			t.Errorf("Expected nil gorm.DB, but got %v", userStore.db)
		} else {
			t.Log("Correctly handled nil gorm.DB without panic")
		}
	})

	t.Run("Scenario 3: Integration with Other Components Using a Mocked DB", func(t *testing.T) {

		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("Failed to create sqlmock: %v", err)
		}
		defer db.Close()
		gormDB, err := gorm.Open("sqlmock", db)
		if err != nil {
			t.Fatalf("Failed to open gorm DB: %v", err)
		}

		mock.ExpectQuery("SELECT \\* FROM users").WillReturnRows(sqlmock.NewRows(nil))

		userStore := store.NewUserStore(gormDB)

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("There were unfulfilled expectations: %s", err)
		} else {
			t.Log("UserStore can successfully integrate with mocked DB")
		}
	})

	t.Run("Scenario 4: Multiple Initializations of UserStore", func(t *testing.T) {

		db1, _, err := sqlmock.New()
		if err != nil {
			t.Fatalf("Failed to create sqlmock: %v", err)
		}
		defer db1.Close()
		gormDB1, err := gorm.Open("sqlmock", db1)
		if err != nil {
			t.Fatalf("Failed to open gorm DB: %v", err)
		}

		db2, _, err := sqlmock.New()
		if err != nil {
			t.Fatalf("Failed to create sqlmock: %v", err)
		}
		defer db2.Close()
		gormDB2, err := gorm.Open("sqlmock", db2)
		if err != nil {
			t.Fatalf("Failed to open gorm DB: %v", err)
		}

		userStore1 := store.NewUserStore(gormDB1)
		userStore2 := store.NewUserStore(gormDB2)

		if userStore1 == userStore2 || userStore1.db == userStore2.db {
			t.Errorf("Expected distinct UserStore instances, but got identical instances")
		} else {
			t.Log("Validation succeeded: Distinct UserStore instances with separate DBs")
		}
	})
}
