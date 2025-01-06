package store

import (
	"errors"
	"testing"
	"time"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/raahii/golang-grpc-realworld-example/model"
)

type ExpectedBegin struct {
	commonExpectation
	delay time.Duration
}

type ExpectedCommit struct {
	commonExpectation
}

type ExpectedExec struct {
	queryBasedExpectation
	result driver.Result
	delay  time.Duration
}




type T struct {
	common
	isParallel bool
	isEnvSet   bool
	context    *testContext // For running tests and subtests.
}
func TestUserStoreCreate(t *testing.T) {
	t.Run("Successful User Creation", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
		}
		defer db.Close()

		mock.ExpectBegin()
		mock.ExpectExec("INSERT INTO \"users\"").
			WithArgs(sqlmock.AnyArg(), "unique_user", "unique_email@example.com", sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		gormDB, err := gorm.Open("postgres", db)
		if err != nil {
			t.Fatalf("failed to initialize gorm db: %v", err)
		}
		defer gormDB.Close()

		userStore := &UserStore{db: gormDB}
		user := &model.User{
			Username: "unique_user",
			Email:    "unique_email@example.com",
			Password: "securepassword",
			Bio:      "User Bio",
			Image:    "User Image",
		}

		err = userStore.Create(user)
		if err != nil {
			t.Errorf("expected no error, got '%v'", err)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %s", err)
		}
	})

	t.Run("Duplicate Email Error", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
		}
		defer db.Close()

		mock.ExpectBegin()
		mock.ExpectExec("INSERT INTO \"users\"").
			WithArgs(sqlmock.AnyArg(), "another_user", "duplicate_email@example.com", sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnError(errors.New("duplicate key value violates unique constraint"))

		gormDB, err := gorm.Open("postgres", db)
		if err != nil {
			t.Fatalf("failed to initialize gorm db: %v", err)
		}
		defer gormDB.Close()

		userStore := &UserStore{db: gormDB}
		user := &model.User{
			Username: "another_user",
			Email:    "duplicate_email@example.com",
			Password: "securepassword",
			Bio:      "User Bio",
			Image:    "User Image",
		}

		err = userStore.Create(user)
		if err == nil {
			t.Error("expected error, got none")
		}
		if err.Error() != "duplicate key value violates unique constraint" {
			t.Errorf("unexpected error message: %s", err)
		}
	})

	t.Run("Duplicate Username Error", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
		}
		defer db.Close()

		mock.ExpectBegin()
		mock.ExpectExec("INSERT INTO \"users\"").
			WithArgs(sqlmock.AnyArg(), "duplicate_user", "new_email@example.com", sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnError(errors.New("duplicate key value violates unique constraint"))

		gormDB, err := gorm.Open("postgres", db)
		if err != nil {
			t.Fatalf("failed to initialize gorm db: %v", err)
		}
		defer gormDB.Close()

		userStore := &UserStore{db: gormDB}
		user := &model.User{
			Username: "duplicate_user",
			Email:    "new_email@example.com",
			Password: "securepassword",
			Bio:      "User Bio",
			Image:    "User Image",
		}

		err = userStore.Create(user)
		if err == nil {
			t.Error("expected error, got none")
		}
		if err.Error() != "duplicate key value violates unique constraint" {
			t.Errorf("unexpected error message: %s", err)
		}
	})

	t.Run("Database Connection Error", func(t *testing.T) {
		userStore := &UserStore{db: nil}
		user := &model.User{
			Username: "new_user",
			Email:    "new_email@example.com",
			Password: "securepassword",
			Bio:      "User Bio",
			Image:    "User Image",
		}

		err := userStore.Create(user)
		if err == nil {
			t.Error("expected error, got none")
		}
	})

	t.Run("Invalid Model Fields", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
		}
		defer db.Close()

		mock.ExpectBegin()
		mock.ExpectExec("INSERT INTO \"users\"").
			WithArgs(sqlmock.AnyArg(), "", "", sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnError(errors.New("not null constraint failed"))

		gormDB, err := gorm.Open("postgres", db)
		if err != nil {
			t.Fatalf("failed to initialize gorm db: %v", err)
		}
		defer gormDB.Close()

		userStore := &UserStore{db: gormDB}
		user := &model.User{}

		err = userStore.Create(user)
		if err == nil {
			t.Error("expected error, got none")
		}
	})

	t.Run("Successful User Creation with Empty Follows and Favorites", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
		}
		defer db.Close()

		mock.ExpectBegin()
		mock.ExpectExec("INSERT INTO \"users\"").
			WithArgs(sqlmock.AnyArg(), "user_with_empty_lists", "email@example.com", sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		gormDB, err := gorm.Open("postgres", db)
		if err != nil {
			t.Fatalf("failed to initialize gorm db: %v", err)
		}
		defer gormDB.Close()

		userStore := &UserStore{db: gormDB}
		user := &model.User{
			Username:         "user_with_empty_lists",
			Email:            "email@example.com",
			Password:         "securepassword",
			Bio:              "User Bio",
			Image:            "User Image",
			Follows:          []model.User{},
			FavoriteArticles: []model.Article{},
		}

		err = userStore.Create(user)
		if err != nil {
			t.Errorf("expected no error, got '%v'", err)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %s", err)
		}
	})

	t.Run("Error Handling in Overridden Time Function", func(t *testing.T) {
		db, _, err := sqlmock.New()
		if err != nil {
			t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
		}
		defer db.Close()

		gormDB, err := gorm.Open("postgres", db)
		if err != nil {
			t.Fatalf("failed to initialize gorm db: %v", err)
		}
		defer gormDB.Close()

		originalNowFunc := gormDB.Set("gorm:now_func", func() time.Time {
			return time.Time{}
		})
		defer gormDB.Set("gorm:now_func", originalNowFunc.(func() time.Time))

		userStore := &UserStore{db: gormDB}
		user := &model.User{
			Username: "test_user",
			Email:    "test_email@example.com",
			Password: "securepassword",
			Bio:      "User Bio",
			Image:    "User Image",
		}

		err = userStore.Create(user)
		if err == nil {
			t.Error("expected error due to invalid timestamp, got none")
		}
	})

	t.Run("Concurrency with Global Locks", func(t *testing.T) {
		t.Log("Concurrency tests might be needed based on the actual implementation specifics")
	})
}
