package store

import (
	"testing"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jinzhu/gorm"
	"github.com/raahii/golang-grpc-realworld-example/model"
	"github.com/stretchr/testify/assert"
	_ "github.com/jinzhu/gorm/dialects/postgres"
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
func TestUserStoreUpdate(t *testing.T) {

	type testCase struct {
		name         string
		setup        func(sqlmock.Sqlmock)
		inputUser    model.User
		expectedErr  bool
		expectedRows int64
	}

	validUser := model.User{
		Username: "validUser",
		Email:    "valid@example.com",
		Password: "Password123",
		Bio:      "This is a bio",
		Image:    "http://example.com/image.jpg",
	}

	invalidUser := model.User{
		Username: "",
		Email:    "",
		Password: "Password123",
		Bio:      "This is a bio",
		Image:    "http://example.com/image.jpg",
	}

	testCases := []testCase{
		{
			name: "Update with Valid User Data",
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec(`^UPDATE "users" SET (.+)$`).
					WithArgs("newEmail@example.com", "Password123", "This is a bio", "http://example.com/image.jpg", "validUser").
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
			inputUser: model.User{
				Username: "validUser",
				Email:    "newEmail@example.com",
				Password: "Password123",
				Bio:      "This is a bio",
				Image:    "http://example.com/image.jpg",
			},
			expectedErr:  false,
			expectedRows: 1,
		},
		{
			name: "Update a Non-Existing User",
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec(`^UPDATE "users" SET (.+)$`).
					WithArgs("nonExistingUser@example.com", "Password123", "This is a bio", "http://example.com/image.jpg", "nonExistingUser").
					WillReturnResult(sqlmock.NewResult(0, 0))
				mock.ExpectCommit()
			},
			inputUser: model.User{
				Username: "nonExistingUser",
				Email:    "nonExistingUser@example.com",
				Password: "Password123",
				Bio:      "This is a bio",
				Image:    "http://example.com/image.jpg",
			},
			expectedErr:  true,
			expectedRows: 0,
		},
		{
			name: "Update with Invalid User Data",
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec(`^UPDATE "users" SET (.+)$`).
					WithArgs(invalidUser.Email, invalidUser.Password, invalidUser.Bio, invalidUser.Image, invalidUser.Username).
					WillReturnResult(sqlmock.NewResult(0, 0))
				mock.ExpectCommit()
			},
			inputUser:    invalidUser,
			expectedErr:  true,
			expectedRows: 0,
		},
		{
			name: "Update When DB Connection Fails",
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec(`^UPDATE "users" SET (.+)$`).
					WithArgs(validUser.Email, validUser.Password, validUser.Bio, validUser.Image, validUser.Username).
					WillReturnError(gorm.ErrInvalidSQL)
				mock.ExpectRollback()
			},
			inputUser:   validUser,
			expectedErr: true,
		},
		{
			name: "No Changes in Update Operation",
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec(`^UPDATE "users" SET (.+)$`).
					WithArgs(validUser.Email, validUser.Password, validUser.Bio, validUser.Image, validUser.Username).
					WillReturnResult(sqlmock.NewResult(1, 0))
				mock.ExpectCommit()
			},
			inputUser:    validUser,
			expectedErr:  false,
			expectedRows: 0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			assert.NoError(t, err)
			defer db.Close()

			gormDB, err := gorm.Open("postgres", db)
			assert.NoError(t, err)

			tc.setup(mock)

			store := &UserStore{db: gormDB}

			err = store.Update(&tc.inputUser)

			assert.Equal(t, tc.expectedErr, err != nil)

			if !tc.expectedErr {
				assert.Equal(t, tc.expectedRows, store.db.RowsAffected)
			}

			assert.Nil(t, mock.ExpectationsWereMet(), "there were unfulfilled expectations")
		})
	}
}
