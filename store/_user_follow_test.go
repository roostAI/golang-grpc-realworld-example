package store

import (
	"errors"
	"testing"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
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

type ExpectedRollback struct {
	commonExpectation
}




type T struct {
	common
	isParallel bool
	isEnvSet   bool
	context    *testContext // For running tests and subtests.
}
func TestUserStoreFollow(t *testing.T) {

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open mock sql db, %v", err)
	}
	defer db.Close()

	gormDB, err := gorm.Open("sqlite3", db)
	if err != nil {
		t.Fatalf("failed to open gorm db, %v", err)
	}

	userStore := &UserStore{db: gormDB}

	tests := []struct {
		name          string
		setupMock     func()
		userA         *model.User
		userB         *model.User
		expectedError error
	}{
		{
			name: "Successfully Follow Another User",
			setupMock: func() {
				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO follows").WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
			userA:         &model.User{Username: "userA"},
			userB:         &model.User{Username: "userB"},
			expectedError: nil,
		},
		{
			name: "Fail to Follow a User Due to Database Error",
			setupMock: func() {
				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO follows").WillReturnError(errors.New("db error"))
				mock.ExpectRollback()
			},
			userA:         &model.User{Username: "userA"},
			userB:         &model.User{Username: "userB"},
			expectedError: errors.New("db error"),
		},
		{
			name: "Follow Yourself Operation",
			setupMock: func() {

			},
			userA:         &model.User{Username: "userA"},
			userB:         &model.User{Username: "userA"},
			expectedError: errors.New("cannot follow yourself"),
		},
		{
			name: "User Does Not Exist in Database",
			setupMock: func() {
				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO follows").WillReturnError(gorm.ErrRecordNotFound)
				mock.ExpectRollback()
			},
			userA:         &model.User{Username: "existentUser"},
			userB:         &model.User{Username: "nonExistentUser"},
			expectedError: gorm.ErrRecordNotFound,
		},
		{
			name: "Follow Operation Produces a Cycle in Following Graph",
			setupMock: func() {

				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO follows").WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
			userA:         &model.User{Username: "userA"},
			userB:         &model.User{Username: "userC"},
			expectedError: nil,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			err := userStore.Follow(tt.userA, tt.userB)
			if tt.expectedError != nil {
				if err == nil || err.Error() != tt.expectedError.Error() {
					t.Errorf("got error: %v, want: %v", err, tt.expectedError)
				}
			} else if err != nil {
				t.Errorf("did not expect error, but got: %v", err)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}
