package store

import (
	"errors"
	"testing"
	"github.com/DATA-DOG/go-sqlmock"
	_ "github.com/jinzhu/gorm/dialects/postgres"
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
func TestUserStoreGetByEmail(t *testing.T) {

	tests := []struct {
		name          string
		email         string
		setupMock     func(sqlmock.Sqlmock)
		expectedUser  *model.User
		expectedError error
	}{
		{
			name:  "Retrieve Existing User by Email",
			email: "existinguser@example.com",
			setupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "username", "email", "password", "bio", "image"}).
					AddRow(1, "ExistingUser", "existinguser@example.com", "hashedpassword", "Bio", "ImageURL")
				mock.ExpectQuery("^SELECT \\* FROM \"users\" WHERE \\(email = \\$1\\) ORDER BY \"users\".\"id\" ASC LIMIT 1$").
					WithArgs("existinguser@example.com").WillReturnRows(rows)
			},
			expectedUser: &model.User{
				Model:    gorm.Model{ID: 1},
				Username: "ExistingUser",
				Email:    "existinguser@example.com",
				Password: "hashedpassword",
				Bio:      "Bio",
				Image:    "ImageURL",
			},
			expectedError: nil,
		},
		{
			name:  "Email Does Not Exist in Database",
			email: "nonexistent@example.com",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("^SELECT \\* FROM \"users\" WHERE \\(email = \\$1\\) ORDER BY \"users\".\"id\" ASC LIMIT 1$").
					WithArgs("nonexistent@example.com").WillReturnError(gorm.ErrRecordNotFound)
			},
			expectedUser:  nil,
			expectedError: gorm.ErrRecordNotFound,
		},
		{
			name:  "Invalid Email Format",
			email: "invalid-email",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("^SELECT \\* FROM \"users\" WHERE \\(email = \\$1\\) ORDER BY \"users\".\"id\" ASC LIMIT 1$").
					WithArgs("invalid-email").WillReturnError(gorm.ErrRecordNotFound)
			},
			expectedUser:  nil,
			expectedError: gorm.ErrRecordNotFound,
		},
		{
			name:  "Database Connection Error",
			email: "dberror@example.com",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("^SELECT \\* FROM \"users\" WHERE \\(email = \\$1\\) ORDER BY \"users\".\"id\" ASC LIMIT 1$").
					WithArgs("dberror@example.com").WillReturnError(errors.New("connection error"))
			},
			expectedUser:  nil,
			expectedError: errors.New("connection error"),
		},
		{
			name:  "Multiple Users with the Same Email",
			email: "duplicateemail@example.com",
			setupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "username", "email", "password", "bio", "image"}).
					AddRow(1, "User1", "duplicateemail@example.com", "password1", "Bio1", "Image1").
					AddRow(2, "User2", "duplicateemail@example.com", "password2", "Bio2", "Image2")
				mock.ExpectQuery("^SELECT \\* FROM \"users\" WHERE \\(email = \\$1\\) ORDER BY \"users\".\"id\" ASC LIMIT 1$").
					WithArgs("duplicateemail@example.com").WillReturnRows(rows)
			},
			expectedUser: &model.User{
				Model:    gorm.Model{ID: 1},
				Username: "User1",
				Email:    "duplicateemail@example.com",
				Password: "password1",
				Bio:      "Bio1",
				Image:    "Image1",
			},
			expectedError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("error while opening a stub database connection: %s", err)
			}
			defer db.Close()

			mockSQL, gormErr := gorm.Open("postgres", db)
			if gormErr != nil {
				t.Fatalf("error initializing gorm db mock: %s", gormErr)
			}

			tt.setupMock(mock)

			store := &UserStore{db: mockSQL}
			user, err := store.GetByEmail(tt.email)

			if tt.expectedError != nil && err != nil {
				if tt.expectedError.Error() != err.Error() {
					t.Errorf("unexpected error. expected: %v, got: %v", tt.expectedError, err)
				}
			} else if (tt.expectedUser == nil) != (user == nil) {
				t.Fatalf("unexpected user result. expected: %v, got: %v", tt.expectedUser, user)
			} else if tt.expectedUser != nil && user != nil {

				if tt.expectedUser.Model.ID != user.Model.ID ||
					tt.expectedUser.Username != user.Username ||
					tt.expectedUser.Email != user.Email ||
					tt.expectedUser.Password != user.Password ||
					tt.expectedUser.Bio != user.Bio ||
					tt.expectedUser.Image != user.Image {
					t.Errorf("mismatched user. expected: %v, got: %v", tt.expectedUser, user)
				}
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatalf("there were unfulfilled expectations: %s", err)
			}

			t.Log("Test case executed:", tt.name)
		})
	}
}
