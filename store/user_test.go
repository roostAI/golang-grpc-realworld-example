package store

import (
	"errors"
	"fmt"
	"testing"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jinzhu/gorm"
	"github.com/raahii/golang-grpc-realworld-example/model"
	"github.com/stretchr/testify/assert"
	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"strings"
	"log"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/raahii/golang-grpc-realworld-example/store"
	"reflect"
	"github.com/stretchr/testify/require"
)

/*
ROOST_METHOD_HASH=Create_889fc0fc45
ROOST_METHOD_SIG_HASH=Create_4c48ec3920


 */
func TestUserStoreCreate(t *testing.T) {

	tests := []struct {
		name      string
		setupMock func(sqlmock.Sqlmock)
		user      model.User
		wantErr   bool
		errorMsg  string
	}{
		{
			name: "Successful User Creation",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO \"users\" (.+) VALUES (.+)").WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
			user: model.User{
				Username: "uniqueUser",
				Email:    "unique@example.com",
				Password: "password",
				Bio:      "Bio",
				Image:    "image.png",
			},
			wantErr:  false,
			errorMsg: "Expected user to be created successfully without error",
		},
		{
			name: "Duplicate Username Error Handling",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO \"users\" (.+) VALUES (.+)").WillReturnError(fmt.Errorf("unique constraint violation"))
				mock.ExpectRollback()
			},
			user: model.User{
				Username: "duplicateUser",
				Email:    "unique@example.com",
				Password: "password",
				Bio:      "Bio",
				Image:    "image.png",
			},
			wantErr:  true,
			errorMsg: "Expected error due to duplicate username",
		},
		{
			name: "Duplicate Email Error Handling",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO \"users\" (.+) VALUES (.+)").WillReturnError(fmt.Errorf("unique constraint violation"))
				mock.ExpectRollback()
			},
			user: model.User{
				Username: "uniqueUser",
				Email:    "duplicate@example.com",
				Password: "password",
				Bio:      "Bio",
				Image:    "image.png",
			},
			wantErr:  true,
			errorMsg: "Expected error due to duplicate email",
		},
		{
			name: "Invalid Field Handling (Validation Error)",
			setupMock: func(mock sqlmock.Sqlmock) {

			},
			user: model.User{
				Username: "",
				Email:    "unique@example.com",
				Password: "password",
				Bio:      "Bio",
				Image:    "image.png",
			},
			wantErr:  true,
			errorMsg: "Expected error due to validation of missing username field",
		},
		{
			name: "Database Connection Error Handling",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin().WillReturnError(errors.New("cannot connect to database"))
			},
			user: model.User{
				Username: "uniqueUser",
				Email:    "unique@example.com",
				Password: "password",
				Bio:      "Bio",
				Image:    "image.png",
			},
			wantErr:  true,
			errorMsg: "Expected error due to database connection issues",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("failed to open sqlmock database: %s", err)
			}
			defer db.Close()

			gormDB, err := gorm.Open("sqlite3", db)
			if err != nil {
				t.Fatalf("failed to open gorm DB: %s", err)
			}
			defer gormDB.Close()

			tt.setupMock(mock)

			userStore := &UserStore{db: gormDB}
			err = userStore.Create(&tt.user)

			if tt.wantErr {
				assert.Error(t, err, tt.errorMsg)
			} else {
				assert.NoError(t, err, tt.errorMsg)
			}

			t.Log(tt.name, ": ", tt.errorMsg)
		})
	}
}

/*
ROOST_METHOD_HASH=Follow_48fdf1257b
ROOST_METHOD_SIG_HASH=Follow_8217e61c06


 */
func TestUserStoreFollow(t *testing.T) {
	type args struct {
		a *model.User
		b *model.User
	}

	tests := []struct {
		name         string
		args         args
		setupMock    func(sqlmock.Sqlmock)
		expectError  bool
		successCheck func(t *testing.T, a *model.User)
	}{
		{
			name: "Successfully Follow Another User",
			args: args{
				a: &model.User{Model: gorm.Model{ID: 1}, Username: "A"},
				b: &model.User{Model: gorm.Model{ID: 2}, Username: "B"},
			},
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO follows").WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
			expectError: false,
			successCheck: func(t *testing.T, a *model.User) {
				assert.Contains(t, a.Follows, model.User{Model: gorm.Model{ID: 2}, Username: "B"}, "User B should be in User A's follows list")
			},
		},
		{
			name: "Attempt to Follow User with Database Error",
			args: args{
				a: &model.User{Username: "A"},
				b: &model.User{Username: "B"},
			},
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO follows").WillReturnError(gorm.ErrInvalidTransaction)
				mock.ExpectRollback()
			},
			expectError: true,
		},
		{
			name: "Following Self",
			args: args{
				a: &model.User{Username: "A"},
				b: &model.User{Username: "A"},
			},
			setupMock:   func(mock sqlmock.Sqlmock) {},
			expectError: true,
		},
		{
			name: "Attempt to Follow Non-existent User",
			args: args{
				a: &model.User{Model: gorm.Model{ID: 1}, Username: "A"},
				b: nil,
			},
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO follows").WillReturnError(gorm.ErrRecordNotFound)
				mock.ExpectRollback()
			},
			expectError: true,
		},
		{
			name: "Ensure No Duplicate Follows",
			args: args{
				a: &model.User{Model: gorm.Model{ID: 1}, Username: "A", Follows: []model.User{{Model: gorm.Model{ID: 2}, Username: "B"}}},
				b: &model.User{Model: gorm.Model{ID: 2}, Username: "B"},
			},
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()

				mock.ExpectExec("INSERT INTO follows").WillReturnError(gorm.ErrUniqueConstraint)
				mock.ExpectRollback()
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
			}
			defer db.Close()

			mockGormDB, err := gorm.Open("sqlite3", db)
			if err != nil {
				t.Fatalf("could not create gorm db: %v", err)
			}

			userStore := &UserStore{db: mockGormDB}

			tt.setupMock(mock)

			err = userStore.Follow(tt.args.a, tt.args.b)

			if tt.expectError {
				assert.Error(t, err, "expected an error but didn't get one")
			} else {
				assert.NoError(t, err, "did not expect an error but got one")
				if tt.successCheck != nil {
					tt.successCheck(t, tt.args.a)
				}
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

/*
ROOST_METHOD_HASH=GetByEmail_3574af40e5
ROOST_METHOD_SIG_HASH=GetByEmail_5731b833c1


 */
func TestGetByEmail(t *testing.T) {
	tests := []struct {
		name          string
		email         string
		setupMock     func(sqlmock.Sqlmock)
		expectedUser  *model.User
		expectError   bool
		errorContains string
	}{
		{
			name:  "Retrieve existing user by email",
			email: "existing@example.com",
			setupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "email", "username"}).
					AddRow(uint(1), "existing@example.com", "existing_user")
				mock.ExpectQuery("SELECT \\* FROM \"users\" WHERE email = ?").
					WithArgs("existing@example.com").WillReturnRows(rows)
			},
			expectedUser: &model.User{ID: uint(1), Email: "existing@example.com", Username: "existing_user"},
			expectError:  false,
		},
		{
			name:  "Handle non-existent user by email",
			email: "nonexistent@example.com",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT \\* FROM \"users\" WHERE email = ?").
					WithArgs("nonexistent@example.com").WillReturnError(gorm.ErrRecordNotFound)
			},
			expectedUser:  nil,
			expectError:   true,
			errorContains: "record not found",
		},
		{
			name:  "Handle database error scenario",
			email: "error@example.com",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT \\* FROM \"users\" WHERE email = ?").
					WithArgs("error@example.com").WillReturnError(errors.New("db error"))
			},
			expectedUser:  nil,
			expectError:   true,
			errorContains: "db error",
		},
		{
			name:  "Retrieve user using an email containing special characters",
			email: "special+char@example.com",
			setupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "email", "username"}).
					AddRow(uint(2), "special+char@example.com", "special_user")
				mock.ExpectQuery("SELECT \\* FROM \"users\" WHERE email = ?").
					WithArgs("special+char@example.com").WillReturnRows(rows)
			},
			expectedUser: &model.User{ID: uint(2), Email: "special+char@example.com", Username: "special_user"},
			expectError:  false,
		},
		{
			name:  "Handle case sensitivity in email retrieval",
			email: "CASE@example.com",
			setupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "email", "username"}).
					AddRow(uint(3), "case@example.com", "case_user")
				mock.ExpectQuery("SELECT \\* FROM \"users\" WHERE email = ?").
					WithArgs("CASE@example.com").WillReturnRows(rows)
			},
			expectedUser: &model.User{ID: uint(3), Email: "case@example.com", Username: "case_user"},
			expectError:  false,
		},
		{
			name:  "Ensure safe handling of email input types",
			email: "",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT \\* FROM \"users\" WHERE email = ?").
					WithArgs("").WillReturnError(gorm.ErrRecordNotFound)
			},
			expectedUser:  nil,
			expectError:   true,
			errorContains: "record not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
			}
			defer db.Close()

			gormDB, err := gorm.Open("postgres", db)
			if err != nil {
				t.Fatalf("failed to initialize gorm with mock: %v", err)
			}

			userStore := &UserStore{db: gormDB}

			tt.setupMock(mock)

			user, err := userStore.GetByEmail(tt.email)

			if tt.expectError {
				if err == nil || (tt.errorContains != "" && !contains(err.Error(), tt.errorContains)) {
					t.Errorf("expected error containing '%s', got '%v'", tt.errorContains, err)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if user == nil || user.ID != tt.expectedUser.ID || user.Email != tt.expectedUser.Email || user.Username != tt.expectedUser.Username {
					t.Errorf("expected user %+v, got %+v", tt.expectedUser, user)
				}
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}

/*
ROOST_METHOD_HASH=GetByID_bbf946112e
ROOST_METHOD_SIG_HASH=GetByID_728dd55ed1


 */
func TestUserStore_GetByID(t *testing.T) {
	type scenario struct {
		name          string
		setupMockDB   func(sqlmock.Sqlmock)
		input         uint
		expectedUser  *model.User
		expectedError error
	}

	scenarios := []scenario{
		{
			name: "Retrieve User by Valid ID",
			setupMockDB: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("^SELECT \\* FROM \"users\" WHERE (\"users\".\"id\" = \\$1 LIMIT 1)").
					WithArgs(1).
					WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow(1, "John Doe"))
			},
			input:         1,
			expectedUser:  &model.User{ID: 1, Name: "John Doe"},
			expectedError: nil,
		},
		{
			name: "Handle Non-existent User ID",
			setupMockDB: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("^SELECT \\* FROM \"users\" WHERE (\"users\".\"id\" = \\$1 LIMIT 1)").
					WithArgs(999).
					WillReturnError(gorm.ErrRecordNotFound)
			},
			input:         999,
			expectedUser:  nil,
			expectedError: gorm.ErrRecordNotFound,
		},
		{
			name: "Database Access Error",
			setupMockDB: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("^SELECT \\* FROM \"users\" WHERE (\"users\".\"id\" = \\$1 LIMIT 1)").
					WithArgs(1).
					WillReturnError(errors.New("db access error"))
			},
			input:         1,
			expectedUser:  nil,
			expectedError: errors.New("db access error"),
		},
		{
			name: "Retrieve User with Minimum ID Value",
			setupMockDB: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("^SELECT \\* FROM \"users\" WHERE (\"users\".\"id\" = \\$1 LIMIT 1)").
					WithArgs(1).
					WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow(1, "John Doe"))
			},
			input:         1,
			expectedUser:  &model.User{ID: 1, Name: "John Doe"},
			expectedError: nil,
		},
	}

	for _, s := range scenarios {
		t.Run(s.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("failed to create sqlmock: %v", err)
			}
			defer db.Close()

			gormDB, err := gorm.Open("postgres", db)
			if err != nil {
				t.Fatalf("failed to open gorm db: %v", err)
			}
			defer gormDB.Close()

			s.setupMockDB(mock)

			userStore := &UserStore{db: gormDB}

			t.Logf("Executing scenario: %s", s.name)

			actualUser, actualError := userStore.GetByID(s.input)

			if !errors.Is(actualError, s.expectedError) {
				t.Logf("Expected error: %v, got: %v", s.expectedError, actualError)
				t.Fail()
			}

			if s.expectedUser != nil && actualUser != nil {
				if actualUser.ID != s.expectedUser.ID {
					t.Logf("Expected user ID: %v, got: %v", s.expectedUser.ID, actualUser.ID)
					t.Fail()
				}
				if actualUser.Name != s.expectedUser.Name {
					t.Logf("Expected user Name: %v, got: %v", s.expectedUser.Name, actualUser.Name)
					t.Fail()
				}
			} else if s.expectedUser == nil && actualUser != nil {
				t.Logf("Expected nil user, got: %v", actualUser)
				t.Fail()
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

/*
ROOST_METHOD_HASH=GetByUsername_f11f114df2
ROOST_METHOD_SIG_HASH=GetByUsername_954d096e24


 */
func TestUserStoreGetByUsername(t *testing.T) {
	database, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock SQL database connection: %v", err)
	}
	defer database.Close()

	gormDB, err := gorm.Open("postgres", database)
	if err != nil {
		t.Fatalf("Failed to open gorm DB: %v", err)
	}
	defer gormDB.Close()

	userStore := &UserStore{db: gormDB}

	tests := []struct {
		name          string
		username      string
		setupMock     func()
		expectedUser  *model.User
		expectedError error
	}{
		{
			name:     "Valid Username Returns User",
			username: "validUser",
			setupMock: func() {
				rows := sqlmock.NewRows([]string{"username", "email"}).AddRow("validUser", "user@example.com")
				mock.ExpectQuery("^SELECT \\* FROM \"users\" WHERE \\(username = \\$1\\) LIMIT 1$").WithArgs("validUser").WillReturnRows(rows)
			},
			expectedUser:  &model.User{Username: "validUser", Email: "user@example.com"},
			expectedError: nil,
		},
		{
			name:     "Non-Existent Username Returns Error",
			username: "nonExistentUser",
			setupMock: func() {
				mock.ExpectQuery("^SELECT \\* FROM \"users\" WHERE \\(username = \\$1\\) LIMIT 1$").WithArgs("nonExistentUser").WillReturnError(gorm.ErrRecordNotFound)
			},
			expectedUser:  nil,
			expectedError: gorm.ErrRecordNotFound,
		},
		{
			name:     "Database Query Error Handling",
			username: "anyUser",
			setupMock: func() {
				mock.ExpectQuery("^SELECT \\* FROM \"users\" WHERE \\(username = \\$1\\) LIMIT 1$").WithArgs("anyUser").WillReturnError(errors.New("connection error"))
			},
			expectedUser:  nil,
			expectedError: errors.New("connection error"),
		},
		{
			name:     "Handling of Special Characters in Username",
			username: "special!User",
			setupMock: func() {
				rows := sqlmock.NewRows([]string{"username", "email"}).AddRow("special!User", "special@example.com")
				mock.ExpectQuery("^SELECT \\* FROM \"users\" WHERE \\(username = \\$1\\) LIMIT 1$").WithArgs("special!User").WillReturnRows(rows)
			},
			expectedUser:  &model.User{Username: "special!User", Email: "special@example.com"},
			expectedError: nil,
		},
		{
			name:     "Empty or Null Username Input",
			username: "",
			setupMock: func() {
				mock.ExpectQuery("^SELECT \\* FROM \"users\" WHERE \\(username = \\$1\\) LIMIT 1$").WithArgs("").WillReturnError(gorm.ErrRecordNotFound)
			},
			expectedUser:  nil,
			expectedError: gorm.ErrRecordNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			user, err := userStore.GetByUsername(tt.username)

			if tt.expectedError != nil && !errors.Is(err, tt.expectedError) {
				t.Fatalf("Expected error %v, got %v", tt.expectedError, err)
			}

			if tt.expectedUser != nil && (user == nil || user.Username != tt.expectedUser.Username || user.Email != tt.expectedUser.Email) {
				t.Fatalf("Expected user %v, got %v", tt.expectedUser, user)
			}

			if err == nil && tt.expectedError == nil {
				t.Log("Success: no error and expected user object returned.")
			} else {
				t.Log("Success: error correctly handled.")
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatalf("There were unfulfilled expectations: %v", err)
			}
		})
	}
}

/*
ROOST_METHOD_HASH=IsFollowing_f53a5d9cef
ROOST_METHOD_SIG_HASH=IsFollowing_9eba5a0e9c


 */
func TestIsFollowing(t *testing.T) {

	tests := []struct {
		name           string
		userA          *model.User
		userB          *model.User
		mockQuery      func(mock sqlmock.Sqlmock)
		expectedResult bool
		expectedError  error
	}{
		{
			name:  "Scenario 1: Check Following Status Between Two Valid Users",
			userA: &model.User{Model: gorm.Model{ID: 1}},
			userB: &model.User{Model: gorm.Model{ID: 2}},
			mockQuery: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT count(.+) FROM follows WHERE from_user_id = \\? AND to_user_id = \\?").
					WithArgs(1, 2).
					WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
			},
			expectedResult: true,
			expectedError:  nil,
		},
		{
			name:  "Scenario 2: Check Non-Following Status Between Two Valid Users",
			userA: &model.User{Model: gorm.Model{ID: 1}},
			userB: &model.User{Model: gorm.Model{ID: 3}},
			mockQuery: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT count(.+) FROM follows WHERE from_user_id = \\? AND to_user_id = \\?").
					WithArgs(1, 3).
					WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
			},
			expectedResult: false,
			expectedError:  nil,
		},
		{
			name:           "Scenario 3: Handle Nil User A",
			userA:          nil,
			userB:          &model.User{Model: gorm.Model{ID: 2}},
			mockQuery:      nil,
			expectedResult: false,
			expectedError:  nil,
		},
		{
			name:           "Scenario 4: Handle Nil User B",
			userA:          &model.User{Model: gorm.Model{ID: 1}},
			userB:          nil,
			mockQuery:      nil,
			expectedResult: false,
			expectedError:  nil,
		},
		{
			name:  "Scenario 5: Handle Database Error",
			userA: &model.User{Model: gorm.Model{ID: 1}},
			userB: &model.User{Model: gorm.Model{ID: 2}},
			mockQuery: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT count(.+) FROM follows WHERE from_user_id = \\? AND to_user_id = \\?").
					WithArgs(1, 2).
					WillReturnError(errors.New("database error"))
			},
			expectedResult: false,
			expectedError:  errors.New("database error"),
		},
		{
			name:  "Scenario 6: Handle Empty Follows Table",
			userA: &model.User{Model: gorm.Model{ID: 1}},
			userB: &model.User{Model: gorm.Model{ID: 2}},
			mockQuery: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT count(.+) FROM follows WHERE from_user_id = \\? AND to_user_id = \\?").
					WithArgs(1, 2).
					WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
			},
			expectedResult: false,
			expectedError:  nil,
		},
		{
			name:  "Scenario 7: Handle Self-Following",
			userA: &model.User{Model: gorm.Model{ID: 1}},
			userB: &model.User{Model: gorm.Model{ID: 1}},
			mockQuery: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT count(.+) FROM follows WHERE from_user_id = \\? AND to_user_id = \\?").
					WithArgs(1, 1).
					WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
			},
			expectedResult: false,
			expectedError:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			db, mock, err := sqlmock.New()
			assert.NoError(t, err, "failed to create mock SQL DB")

			gormDB, err := gorm.Open("sqlite3", db)
			assert.NoError(t, err, "failed to open Gorm DB connection")

			store := &store.UserStore{DB: gormDB}

			if tt.mockQuery != nil {
				tt.mockQuery(mock)
			}

			result, err := store.IsFollowing(tt.userA, tt.userB)

			assert.Equal(t, tt.expectedResult, result)
			if tt.expectedError != nil {
				assert.EqualError(t, err, tt.expectedError.Error())
			} else {
				assert.NoError(t, err)
			}

			assert.NoError(t, mock.ExpectationsWereMet(), "not all expectations fulfilled")
		})
	}
}

/*
ROOST_METHOD_HASH=NewUserStore_6a331dd890
ROOST_METHOD_SIG_HASH=NewUserStore_4f0c2dfca9


 */
func TestNewUserStore(t *testing.T) {
	t.Run("Scenario 1: Initialization with a Valid gorm.DB Instance", func(t *testing.T) {

		db, _, err := sqlmock.New()
		assert.NoError(t, err, "Creating sqlmock should not error")
		gormDB, err := gorm.Open("postgres", db)
		assert.NoError(t, err, "Opening gorm DB should not error")
		defer gormDB.Close()

		userStore := NewUserStore(gormDB)

		assert.NotNil(t, userStore, "UserStore should not be nil")
		assert.True(t, reflect.DeepEqual(userStore.db, gormDB), "UserStore DB should match the provided DB instance")
		t.Log("Successfully validated initialization with a valid gorm.DB instance")
	})

	t.Run("Scenario 2: Handling a Nil gorm.DB Instance", func(t *testing.T) {

		var nilDB *gorm.DB

		userStore := NewUserStore(nilDB)

		assert.NotNil(t, userStore, "UserStore should not be nil even when nil DB is passed")
		assert.Nil(t, userStore.db, "UserStore DB should be nil when nil DB is passed")
		t.Log("Successfully validated function does not panic and DB field is nil")
	})

	t.Run("Scenario 3: Verification of Type Integrity", func(t *testing.T) {

		db, _, err := sqlmock.New()
		assert.NoError(t, err, "Creating sqlmock should not error")
		gormDB, err := gorm.Open("postgres", db)
		assert.NoError(t, err, "Opening gorm DB should not error")
		defer gormDB.Close()

		userStore := NewUserStore(gormDB)

		assert.IsType(t, &UserStore{}, userStore, "Result should be a pointer to UserStore type")
		t.Log("Successfully validated type integrity")
	})

	t.Run("Scenario 4: Repeated Calls with Different gorm.DB Instances", func(t *testing.T) {

		db1, _, err := sqlmock.New()
		assert.NoError(t, err, "Creating first sqlmock should not error")
		gormDB1, err := gorm.Open("postgres", db1)
		assert.NoError(t, err, "Opening first gorm DB should not error")
		defer gormDB1.Close()

		db2, _, err := sqlmock.New()
		assert.NoError(t, err, "Creating second sqlmock should not error")
		gormDB2, err := gorm.Open("postgres", db2)
		assert.NoError(t, err, "Opening second gorm DB should not error")
		defer gormDB2.Close()

		userStore1 := NewUserStore(gormDB1)
		userStore2 := NewUserStore(gormDB2)

		assert.NotEqual(t, userStore1, userStore2, "Each UserStore instance should be unique")
		assert.True(t, reflect.DeepEqual(userStore1.db, gormDB1), "First UserStore should have first DB")
		assert.True(t, reflect.DeepEqual(userStore2.db, gormDB2), "Second UserStore should have second DB")
		t.Log("Successfully validated repeated calls with different gorm.DB instances")
	})

}

/*
ROOST_METHOD_HASH=Unfollow_57959a8a53
ROOST_METHOD_SIG_HASH=Unfollow_8bd8e0bc55


 */
func TestUserStoreUnfollow(t *testing.T) {

	tests := []struct {
		name        string
		setup       func() (*UserStore, *model.User, *model.User)
		expectedErr bool
	}{
		{
			name: "Successfully Unfollow a User",
			setup: func() (*UserStore, *model.User, *model.User) {
				db, mock, _ := sqlmock.New()
				mock.ExpectExec("DELETE FROM follows").
					WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg()).
					WillReturnResult(sqlmock.NewResult(1, 1))

				gormDB, _ := gorm.Open("sqlite3", db)
				userStore := &UserStore{db: gormDB}

				userA := &model.User{Username: "userA"}
				userB := &model.User{Username: "userB"}
				userA.Follows = []model.User{*userB}

				return userStore, userA, userB
			},
			expectedErr: false,
		},
		{
			name: "Unfollow a User Not Being Followed",
			setup: func() (*UserStore, *model.User, *model.User) {
				db, mock, _ := sqlmock.New()
				mock.ExpectExec("DELETE FROM follows").
					WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg()).
					WillReturnResult(sqlmock.NewResult(1, 0))

				gormDB, _ := gorm.Open("sqlite3", db)
				userStore := &UserStore{db: gormDB}

				userA := &model.User{Username: "userA"}
				userB := &model.User{Username: "userB"}

				return userStore, userA, userB
			},
			expectedErr: false,
		},
		{
			name: "Unfollow with Database Error",
			setup: func() (*UserStore, *model.User, *model.User) {
				db, mock, _ := sqlmock.New()
				mock.ExpectExec("DELETE FROM follows").
					WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg()).
					WillReturnError(errors.New("database error"))

				gormDB, _ := gorm.Open("sqlite3", db)
				userStore := &UserStore{db: gormDB}

				userA := &model.User{Username: "userA"}
				userB := &model.User{Username: "userB"}
				userA.Follows = []model.User{*userB}

				return userStore, userA, userB
			},
			expectedErr: true,
		},
		{
			name: "Unfollow Null User",
			setup: func() (*UserStore, *model.User, *model.User) {
				db, _, _ := sqlmock.New()
				gormDB, _ := gorm.Open("sqlite3", db)
				userStore := &UserStore{db: gormDB}

				userA := &model.User{Username: "userA"}
				userB := (*model.User)(nil)

				return userStore, userA, userB
			},
			expectedErr: true,
		},
		{
			name: "Unfollow with Null Store Database",
			setup: func() (*UserStore, *model.User, *model.User) {
				userStore := &UserStore{db: nil}

				userA := &model.User{Username: "userA"}
				userB := &model.User{Username: "userB"}

				return userStore, userA, userB
			},
			expectedErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store, userA, userB := tt.setup()
			err := store.Unfollow(userA, userB)

			if tt.expectedErr {
				assert.Error(t, err, "Expected an error but didn't get one.")
			} else {
				assert.NoError(t, err, "Expected no error but got one.")
				if userB != nil {
					for _, follow := range userA.Follows {
						assert.NotEqual(t, follow.Username, userB.Username, "userB should be removed from userA.Follows")
					}
				}
			}
		})
	}
}

/*
ROOST_METHOD_HASH=Update_68f27dd78a
ROOST_METHOD_SIG_HASH=Update_87150d6435


 */
func TestUserStoreUpdate(t *testing.T) {

	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	gormDB, err := gorm.Open("mysql", db)
	require.NoError(t, err)
	defer gormDB.Close()

	userStore := UserStore{db: gormDB}

	tests := []struct {
		name          string
		setupDBMock   func()
		inputUser     *model.User
		expectedError error
	}{
		{
			name: "Scenario 1: Successful Update of a User",
			setupDBMock: func() {
				mock.ExpectBegin()
				mock.ExpectExec("UPDATE `users` SET (.+) WHERE (.+)").
					WithArgs("NewUsername", "newemail@example.com", sqlmock.AnyArg(), "NewBio", "NewImage", 1).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
			inputUser: &model.User{
				Model:    gorm.Model{ID: 1},
				Username: "NewUsername",
				Email:    "newemail@example.com",
				Bio:      "NewBio",
				Image:    "NewImage",
			},
			expectedError: nil,
		},
		{
			name: "Scenario 2: Update Fails Due to Nonexistent User",
			setupDBMock: func() {
				mock.ExpectBegin()
				mock.ExpectExec("UPDATE `users` SET (.+) WHERE (.+)").
					WithArgs("NewUsername", "newemail@example.com", sqlmock.AnyArg(), "NewBio", "NewImage", 2).
					WillReturnResult(sqlmock.NewResult(0, 0))
				mock.ExpectCommit()
			},
			inputUser: &model.User{
				Model:    gorm.Model{ID: 2},
				Username: "NewUsername",
				Email:    "newemail@example.com",
				Bio:      "NewBio",
				Image:    "NewImage",
			},
			expectedError: gorm.ErrRecordNotFound,
		},
		{
			name: "Scenario 3: Update Fails Due to Database Error",
			setupDBMock: func() {
				mock.ExpectBegin()
				mock.ExpectExec("UPDATE `users` SET (.+) WHERE (.+)").
					WithArgs("NewUsername", "newemail@example.com", sqlmock.AnyArg(), "NewBio", "NewImage", 3).
					WillReturnError(gorm.ErrInvalidTransaction)
				mock.ExpectRollback()
			},
			inputUser: &model.User{
				Model:    gorm.Model{ID: 3},
				Username: "NewUsername",
				Email:    "newemail@example.com",
				Bio:      "NewBio",
				Image:    "NewImage",
			},
			expectedError: gorm.ErrInvalidTransaction,
		},
		{
			name: "Scenario 4: Update Failures Due to Validation Constraints",
			setupDBMock: func() {
				mock.ExpectBegin()
				mock.ExpectExec("UPDATE `users` SET (.+) WHERE (.+)").
					WithArgs("NewUsername", "newemail@example.com", sqlmock.AnyArg(), "NewBio", "NewImage", 4).
					WillReturnError(gorm.ErrInvalidTransaction)
				mock.ExpectRollback()
			},
			inputUser: &model.User{
				Model:    gorm.Model{ID: 4},
				Username: "NewUsername",
				Email:    "newemail@example.com",
				Bio:      "NewBio",
				Image:    "NewImage",
			},
			expectedError: gorm.ErrInvalidTransaction,
		},
		{
			name: "Scenario 5: No Changes Made During Update",
			setupDBMock: func() {
				mock.ExpectBegin()
				mock.ExpectExec("UPDATE `users` SET (.+) WHERE (.+)").
					WithArgs("NewUsername", "newemail@example.com", sqlmock.AnyArg(), "NewBio", "NewImage", 5).
					WillReturnResult(sqlmock.NewResult(0, 0))
				mock.ExpectCommit()
			},
			inputUser: &model.User{
				Model:    gorm.Model{ID: 5},
				Username: "NewUsername",
				Email:    "newemail@example.com",
				Bio:      "NewBio",
				Image:    "NewImage",
			},
			expectedError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupDBMock()

			err := userStore.Update(tt.inputUser)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError, err)
			} else {
				assert.NoError(t, err)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("Expectations were not met: %v", err)
			}
		})
	}
}

/*
ROOST_METHOD_HASH=GetFollowingUserIDs_ba3670aa2c
ROOST_METHOD_SIG_HASH=GetFollowingUserIDs_55ccc2afd7


 */
func TestUserStoreGetFollowingUserIDs(t *testing.T) {
	tests := []struct {
		name        string
		setupMock   func(sqlmock.Sqlmock, *model.User)
		user        *model.User
		expectedIDs []uint
		expectError bool
	}{
		{
			name: "Retrieve Following User IDs Successfully",
			setupMock: func(mock sqlmock.Sqlmock, user *model.User) {
				rows := sqlmock.NewRows([]string{"to_user_id"}).
					AddRow(101).
					AddRow(102)
				mock.ExpectQuery("^SELECT (.+) FROM follows WHERE (.+)$").WithArgs(user.ID).WillReturnRows(rows)
			},
			user: &model.User{
				Model: gorm.Model{ID: 1},
			},
			expectedIDs: []uint{101, 102},
			expectError: false,
		},
		{
			name: "User with No Following",
			setupMock: func(mock sqlmock.Sqlmock, user *model.User) {
				rows := sqlmock.NewRows([]string{"to_user_id"})
				mock.ExpectQuery("^SELECT (.+) FROM follows WHERE (.+)$").WithArgs(user.ID).WillReturnRows(rows)
			},
			user: &model.User{
				Model: gorm.Model{ID: 2},
			},
			expectedIDs: []uint{},
			expectError: false,
		},
		{
			name: "Handle Database Error Gracefully",
			setupMock: func(mock sqlmock.Sqlmock, user *model.User) {
				mock.ExpectQuery("^SELECT (.+) FROM follows WHERE (.+)$").WithArgs(user.ID).WillReturnError(gorm.ErrInvalidTransaction)
			},
			user: &model.User{
				Model: gorm.Model{ID: 3},
			},
			expectedIDs: []uint{},
			expectError: true,
		},
		{
			name: "Verify SQL Injection Safety",
			setupMock: func(mock sqlmock.Sqlmock, user *model.User) {
				rows := sqlmock.NewRows([]string{"to_user_id"})
				mock.ExpectQuery("^SELECT (.+) FROM follows WHERE (.+)$").WithArgs(user.ID).WillReturnRows(rows)
			},
			user: &model.User{
				Model: gorm.Model{ID: 0xDEADBEEF},
			},
			expectedIDs: []uint{},
			expectError: false,
		},
		{
			name: "Large List of Following Users",
			setupMock: func(mock sqlmock.Sqlmock, user *model.User) {
				rows := sqlmock.NewRows([]string{"to_user_id"})
				for i := 0; i < 1000; i++ {
					rows.AddRow(uint(i))
				}
				mock.ExpectQuery("^SELECT (.+) FROM follows WHERE (.+)$").WithArgs(user.ID).WillReturnRows(rows)
			},
			user: &model.User{
				Model: gorm.Model{ID: 4},
			},
			expectedIDs: make([]uint, 1000),
			expectError: false,
		},
		{
			name: "Non-existent User ID",
			setupMock: func(mock sqlmock.Sqlmock, user *model.User) {
				rows := sqlmock.NewRows([]string{"to_user_id"})
				mock.ExpectQuery("^SELECT (.+) FROM follows WHERE (.+)$").WithArgs(user.ID).WillReturnRows(rows)
			},
			user: &model.User{
				Model: gorm.Model{ID: 5},
			},
			expectedIDs: []uint{},
			expectError: false,
		},
	}

	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {

			db, mock, err := sqlmock.New()
			if err != nil {
				log.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
			}
			defer db.Close()

			gormDB, err := gorm.Open("postgres", db)
			if err != nil {
				log.Fatalf("an error '%s' was not expected when creating gorm DB", err)
			}

			store := &UserStore{db: gormDB}

			tt.setupMock(mock, tt.user)

			ids, err := store.GetFollowingUserIDs(tt.user)

			if (err != nil) != tt.expectError {
				t.Errorf("expected error: %v, got: %v", tt.expectError, err)
			}

			if len(ids) != len(tt.expectedIDs) {
				t.Errorf("expected ids length: %d, got: %d", len(tt.expectedIDs), len(ids))
			}

			for i, id := range ids {
				if id != tt.expectedIDs[i] {
					t.Errorf("expected id: %d, got: %d", tt.expectedIDs[i], id)
				}
			}
		})
	}
}

