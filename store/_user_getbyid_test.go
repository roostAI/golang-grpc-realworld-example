package store

import (
	"errors"
	"testing"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/assert"
	"github.com/raahii/golang-grpc-realworld-example/model"
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
func TestUserStoreGetByID(t *testing.T) {
	type testCase struct {
		name          string
		setupMock     func(sqlmock.Sqlmock)
		id            uint
		expectedUser  *model.User
		expectedError error
	}

	t.Log("Initializing sqlmock for database mocking.")
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("An error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	gormDB, err := gorm.Open("postgres", db)
	if err != nil {
		t.Fatalf("Failed to open gorm DB with sqlmock, error: %s", err)
	}
	defer gormDB.Close()

	store := &UserStore{db: gormDB}

	tests := []testCase{
		{
			name: "Retrieve User Successfully by Valid ID",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("^SELECT (.+) FROM \"users\" WHERE (.+)$").
					WithArgs(1).
					WillReturnRows(sqlmock.NewRows([]string{"id", "username", "email", "password", "bio", "image"}).
						AddRow(1, "user1", "user1@example.com", "password", "bio", "image"))
			},
			id: 1,
			expectedUser: &model.User{
				Model:    gorm.Model{ID: 1},
				Username: "user1",
				Email:    "user1@example.com",
				Password: "password",
				Bio:      "bio",
				Image:    "image",
			},
			expectedError: nil,
		},
		{
			name: "Fail to Retrieve User for Non-Existent ID",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("^SELECT (.+) FROM \"users\" WHERE (.+)$").
					WithArgs(9999).
					WillReturnError(gorm.ErrRecordNotFound)
			},
			id:            9999,
			expectedUser:  nil,
			expectedError: gorm.ErrRecordNotFound,
		},
		{
			name: "Return Error Due to Database Connection Issue",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("^SELECT (.+) FROM \"users\" WHERE (.+)$").
					WithArgs(1).
					WillReturnError(errors.New("connection kill"))
			},
			id:            1,
			expectedUser:  nil,
			expectedError: errors.New("connection kill"),
		},
		{
			name: "Handle Large Numeric User ID Inputs",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("^SELECT (.+) FROM \"users\" WHERE (.+)$").
					WithArgs(9223372036854775807).
					WillReturnError(gorm.ErrRecordNotFound)
			},
			id:            9223372036854775807,
			expectedUser:  nil,
			expectedError: gorm.ErrRecordNotFound,
		},
		{
			name: "Validate Function's Response to Zero as User ID",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("^SELECT (.+) FROM \"users\" WHERE (.+)$").
					WithArgs(0).
					WillReturnError(gorm.ErrRecordNotFound)
			},
			id:            0,
			expectedUser:  nil,
			expectedError: gorm.ErrRecordNotFound,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Logf("Running test case: %s", tc.name)
			tc.setupMock(mock)

			actualUser, actualError := store.GetByID(tc.id)

			assert.Equal(t, tc.expectedUser, actualUser, "The user returned does not match the expected value.")
			assert.Equal(t, tc.expectedError, actualError, "The error returned does not match the expected value.")

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}
