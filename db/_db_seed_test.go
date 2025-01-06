package db

import (
	"errors"
	"io/ioutil"
	"os"
	"testing"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/BurntSushi/toml"
	"github.com/jinzhu/gorm"
	"github.com/raahii/golang-grpc-realworld-example/model"
)

var origReadFile = ioutil.ReadFiletype ExpectedBegin struct {
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

type ExpectedRollback struct {
	commonExpectation
}




type T struct {
	common
	isParallel bool
	isEnvSet   bool
	context    *testContext // For running tests and subtests.
}
func TestSeed(t *testing.T) {
	tests := []struct {
		name          string
		setup         func() *gorm.DB
		expectedError error
	}{
		{
			name: "Scenario 1: Successful Seeding of Users from TOML File",
			setup: func() *gorm.DB {
				db, mock := mockDB(t)

				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO users").WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()

				data := `users = [{name = "John Doe", email = "john@example.com"}]`

				ioutil.ReadFile = func(filename string) ([]byte, error) { return []byte(data), nil }

				return db
			},
			expectedError: nil,
		},
		{
			name: "Scenario 2: Missing TOML File",
			setup: func() *gorm.DB {
				db, _ := mockDB(t)

				ioutil.ReadFile = func(filename string) ([]byte, error) { return nil, os.ErrNotExist }

				return db
			},
			expectedError: os.ErrNotExist,
		},
		{
			name: "Scenario 3: Malformed TOML File",
			setup: func() *gorm.DB {
				db, _ := mockDB(t)

				data := `users = [name = "John Doe"`
				ioutil.ReadFile = func(filename string) ([]byte, error) { return []byte(data), nil }

				return db
			},
			expectedError: errors.New("TOML parsing issue expected"),
		},
		{
			name: "Scenario 4: Database Create Operation Error",
			setup: func() *gorm.DB {
				db, mock := mockDB(t)

				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO users").WillReturnError(errors.New("database error"))
				mock.ExpectRollback()

				data := `users = [{name = "John Doe", email = "john@example.com"}]`
				ioutil.ReadFile = func(filename string) ([]byte, error) { return []byte(data), nil }

				return db
			},
			expectedError: errors.New("database error"),
		},
		{
			name: "Scenario 5: Empty User List in TOML File",
			setup: func() *gorm.DB {
				db, mock := mockDB(t)

				mock.ExpectBegin()
				mock.ExpectCommit()

				data := `users = []`
				ioutil.ReadFile = func(filename string) ([]byte, error) { return []byte(data), nil }

				return db
			},
			expectedError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {

				ioutil.ReadFile = origReadFile
			}()

			db := tt.setup()

			err := Seed(db)
			if tt.expectedError == nil && err != nil {
				t.Errorf("expected no error, but got %v", err)
			} else if tt.expectedError != nil && err == nil {
				t.Errorf("expected an error, but got none")
			} else if tt.expectedError != nil && err.Error() != tt.expectedError.Error() {
				t.Errorf("expected error %v, but got %v", tt.expectedError, err)
			}

			db.Close()
		})
	}
}
func mockDB(t *testing.T) (*gorm.DB, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock database: %s", err)
	}

	gdb, err := gorm.Open("postgres", db)
	if err != nil {
		t.Fatalf("failed to initialize gorm DB: %s", err)
	}
	return gdb, mock
}
