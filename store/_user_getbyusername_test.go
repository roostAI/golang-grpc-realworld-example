package store

import (
	"errors"
	"testing"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jinzhu/gorm"
	"github.com/raahii/golang-grpc-realworld-example/model"
)


type ExpectedQuery struct {
	queryBasedExpectation
	rows             driver.Rows
	delay            time.Duration
	rowsMustBeClosed bool
	rowsWereClosed   bool
}

type Rows struct {
	converter driver.ValueConverter
	cols      []string
	def       []*Column
	rows      [][]driver.Value
	pos       int
	nextErr   map[int]error
	closeErr  error
}





type T struct {
	common
	isParallel bool
	isEnvSet   bool
	context    *testContext // For running tests and subtests.
}

func TestUserStoreGetByUsername(t *testing.T) {
	type testCase struct {
		description  string
		username     string
		mockSetup    func(sqlmock.Sqlmock)
		expectedUser *model.User
		expectedErr  error
	}

	tests := []testCase{
		{
			description: "Retrieve Existing User by Username",
			username:    "existing_user",
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "username", "email", "password", "bio", "image"}).AddRow(1, "existing_user", "user@example.com", "password123", "bio data", "image.png")
				mock.ExpectQuery(`SELECT \* FROM "users" WHERE \(username = \?\) ORDER BY "users"\."id" ASC LIMIT 1`).WithArgs("existing_user").WillReturnRows(rows)
			},
			expectedUser: &model.User{
				Model:    gorm.Model{ID: 1},
				Username: "existing_user",
				Email:    "user@example.com",
				Password: "password123",
				Bio:      "bio data",
				Image:    "image.png",
			},
			expectedErr: nil,
		},
		{
			description: "Username Not Found",
			username:    "non_existent_user",
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT \* FROM "users" WHERE \(username = \?\) ORDER BY "users"\."id" ASC LIMIT 1`).WithArgs("non_existent_user").WillReturnError(gorm.ErrRecordNotFound)
			},
			expectedUser: nil,
			expectedErr:  gorm.ErrRecordNotFound,
		},
		{
			description: "Database Connectivity Issues",
			username:    "any_user",
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT \* FROM "users" WHERE \(username = \?\) ORDER BY "users"\."id" ASC LIMIT 1`).WithArgs("any_user").WillReturnError(errors.New("db connection error"))
			},
			expectedUser: nil,
			expectedErr:  errors.New("db connection error"),
		},
		{
			description: "Invalid Username Input",
			username:    "",
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT \* FROM "users" WHERE \(username = \?\) ORDER BY "users"\."id" ASC LIMIT 1`).WithArgs("").WillReturnError(gorm.ErrRecordNotFound)
			},
			expectedUser: nil,
			expectedErr:  gorm.ErrRecordNotFound,
		},
		{
			description: "Multiple Users with the Same Username",
			username:    "duplicate_user",
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "username", "email", "password", "bio", "image"}).
					AddRow(1, "duplicate_user", "user1@example.com", "password1", "bio1", "image1.png").
					AddRow(2, "duplicate_user", "user2@example.com", "password2", "bio2", "image2.png")
				mock.ExpectQuery(`SELECT \* FROM "users" WHERE \(username = \?\) ORDER BY "users"\."id" ASC LIMIT 1`).WithArgs("duplicate_user").WillReturnRows(rows)
			},
			expectedUser: &model.User{
				Model:    gorm.Model{ID: 1},
				Username: "duplicate_user",
				Email:    "user1@example.com",
				Password: "password1",
				Bio:      "bio1",
				Image:    "image1.png",
			},
			expectedErr: nil,
		},
	}

	for _, tc := range tests {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("error opening stub database connection: %v", err)
		}

		defer db.Close()

		if tc.mockSetup != nil {
			tc.mockSetup(mock)
		}

		gormDB, err := gorm.Open("postgres", db)
		if err != nil {
			t.Fatalf("failed to create GORM DB from sqlmock: %v", err)
		}

		store := &UserStore{db: gormDB}

		user, err := store.GetByUsername(tc.username)

		if tc.expectedErr != nil && err != nil && errors.Is(err, tc.expectedErr) {
			t.Logf("%s: Expected error matches actual error: %v", tc.description, err)
		} else if tc.expectedErr == nil && err == nil && equalUsers(tc.expectedUser, user) {
			t.Logf("%s: User data matches expectation", tc.description)
		} else {
			t.Errorf("%s: unexpected result, got (user: %v, err: %v)", tc.description, user, err)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("There were unmet expectations: %s", err)
		}
	}
}
func equalUsers(expected, actual *model.User) bool {
	if expected == nil || actual == nil {
		return expected == actual
	}
	return expected.ID == actual.ID &&
		expected.Username == actual.Username &&
		expected.Email == actual.Email &&
		expected.Password == actual.Password &&
		expected.Bio == actual.Bio &&
		expected.Image == actual.Image
}
