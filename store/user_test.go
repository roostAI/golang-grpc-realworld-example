package store

import (
	"errors"
	"testing"
	"time"
	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/jinzhu/gorm"
	"github.com/raahii/golang-grpc-realworld-example/model"
	"github.com/DATA-DOG/go-sqlmock"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/raahii/golang-grpc-realworld-example/store"
	"github.com/stretchr/testify/assert"
	"log"
	"reflect"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
)


/*
ROOST_METHOD_HASH=Create_889fc0fc45
ROOST_METHOD_SIG_HASH=Create_4c48ec3920


 */
func TestCreate(t *testing.T) {
	tests := []struct {
		name        string
		setupMock   func(sqlmock.Sqlmock)
		userInput   model.User
		expectedErr error
	}{
		{
			name: "CreateSuccessfully",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO \"users\"").
					WithArgs(sqlmock.AnyArg(), "uniqueusername", "uniqueemail@example.com", sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
			userInput: model.User{
				Username: "uniqueusername",
				Email:    "uniqueemail@example.com",
				Password: "strongpassword123",
				Bio:      "This is a bio",
				Image:    "image.jpg",
			},
			expectedErr: nil,
		},
		{
			name: "UsernameAlreadyExists",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO \"users\"").
					WithArgs(sqlmock.AnyArg(), "existingusername", sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
					WillReturnError(errors.New("unique constraint violated"))
				mock.ExpectRollback()
			},
			userInput: model.User{
				Username: "existingusername",
				Email:    "newemail@example.com",
				Password: "anotherpassword123",
				Bio:      "Another bio",
				Image:    "anotherimage.jpg",
			},
			expectedErr: errors.New("unique constraint violated"),
		},
		{
			name: "MissingRequiredFields",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO \"users\"").
					WithArgs(sqlmock.AnyArg(), "", sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
					WillReturnError(errors.New("missing required field"))
				mock.ExpectRollback()
			},
			userInput: model.User{
				Username: "",
				Email:    "anothermissingemail@example.com",
			},
			expectedErr: errors.New("missing required field"),
		},
		{
			name: "DatabaseConnectionFails",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin().WillReturnError(errors.New("connection failed"))
			},
			userInput: model.User{
				Username: "anotherusername",
				Email:    "validemail@example.com",
				Password: "validpassword",
				Bio:      "Yet another bio",
				Image:    "image.png",
			},
			expectedErr: errors.New("connection failed"),
		},
		{
			name: "VerifyTimestampAssignment",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				currentTime := time.Now()
				mock.ExpectExec("INSERT INTO \"users\"").
					WithArgs(sqlmock.AnyArg(), "timestampuser", sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), currentTime, currentTime).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
			userInput: model.User{
				Username: "timestampuser",
				Email:    "timestampemail@example.com",
				Password: "timestamp123",
				Bio:      "Timestamp bio",
				Image:    "timestamp.jpg",
			},
			expectedErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("An error '%s' was not expected when opening a stub database connection", err)
			}
			defer db.Close()

			gormDB, err := gorm.Open("postgres", db)
			if err != nil {
				t.Fatalf("An error '%s' was not expected when initializing gorm database", err)
			}
			defer gormDB.Close()

			userStore := UserStore{db: gormDB}
			tt.setupMock(mock)

			err = userStore.Create(&tt.userInput)

			if err != nil && err.Error() != tt.expectedErr.Error() {
				t.Errorf("expected error: %v, got: %v", tt.expectedErr, err)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}


/*
ROOST_METHOD_HASH=Follow_48fdf1257b
ROOST_METHOD_SIG_HASH=Follow_8217e61c06


 */
func TestFollow(t *testing.T) {
	testCases := []struct {
		name                    string
		follower                *model.User
		followee                *model.User
		mockSetup               func(sqlmock.Sqlmock)
		expectedError           error
		expectedFollowerFollows int
	}{
		{
			name:     "Successful Follow Between Two Users",
			follower: &model.User{Model: gorm.Model{ID: 1}, Username: "userA"},
			followee: &model.User{Model: gorm.Model{ID: 2}, Username: "userB"},
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec(`INSERT INTO "follows"`).
					WithArgs(sqlmock.AnyArg(), 1, 2).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
			expectedError:           nil,
			expectedFollowerFollows: 1,
		},
		{
			name:     "Follow User Already Followed",
			follower: &model.User{Model: gorm.Model{ID: 1}, Username: "userA"},
			followee: &model.User{Model: gorm.Model{ID: 2}, Username: "userB"},
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec(`INSERT INTO "follows"`).
					WithArgs(sqlmock.AnyArg(), 1, 2).
					WillReturnError(gorm.ErrRecordNotFound)
				mock.ExpectRollback()
			},
			expectedError:           gorm.ErrRecordNotFound,
			expectedFollowerFollows: 1,
		},
		{
			name:     "Attempt to Follow a Nonexistent User",
			follower: &model.User{Model: gorm.Model{ID: 1}, Username: "userA"},
			followee: &model.User{Model: gorm.Model{ID: 999}, Username: "userX"},
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec(`INSERT INTO "follows"`).
					WithArgs(sqlmock.AnyArg(), 1, 999).
					WillReturnError(gorm.ErrRecordNotFound)
				mock.ExpectRollback()
			},
			expectedError:           gorm.ErrRecordNotFound,
			expectedFollowerFollows: 0,
		},
		{
			name:     "Database Error During Follow Operation",
			follower: &model.User{Model: gorm.Model{ID: 1}, Username: "userA"},
			followee: &model.User{Model: gorm.Model{ID: 2}, Username: "userB"},
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec(`INSERT INTO "follows"`).
					WithArgs(sqlmock.AnyArg(), 1, 2).
					WillReturnError(errors.New("database error"))
				mock.ExpectRollback()
			},
			expectedError:           errors.New("database error"),
			expectedFollowerFollows: 0,
		},
		{
			name:     "Null User Reference as Follower or Followee",
			follower: nil,
			followee: &model.User{Model: gorm.Model{ID: 2}, Username: "userB"},
			mockSetup: func(mock sqlmock.Sqlmock) {

			},
			expectedError:           errors.New("nil user reference"),
			expectedFollowerFollows: 0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("error initializing mock: %v", err)
			}
			defer db.Close()

			gormDB, err := gorm.Open("postgres", db)
			if err != nil {
				t.Fatalf("error opening gorm db: %v", err)
			}
			defer gormDB.Close()

			store := UserStore{db: gormDB}

			if tc.follower == nil || tc.followee == nil {
				if err := store.Follow(tc.follower, tc.followee); err == nil || err.Error() != tc.expectedError.Error() {
					t.Fatalf("expected error: %v, got: %v", tc.expectedError, err)
				}
				return
			}

			tc.mockSetup(mock)

			err = store.Follow(tc.follower, tc.followee)
			if tc.expectedError != nil && err == nil {
				t.Fatalf("expected error '%v', but got nil", tc.expectedError)
			} else if tc.expectedError == nil && err != nil {
				t.Fatalf("expected no error, but got one: %v", err)
			} else if tc.expectedError != nil && err != nil && tc.expectedError.Error() != err.Error() {
				t.Fatalf("expected error '%v', but got '%v'", tc.expectedError, err)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unmet expectations: %v", err)
			}

			t.Logf("Test '%s' passed.", tc.name)
		})
	}
}


/*
ROOST_METHOD_HASH=GetByEmail_3574af40e5
ROOST_METHOD_SIG_HASH=GetByEmail_5731b833c1


 */
func TestGetByEmail(t *testing.T) {

	type testScenario struct {
		name         string
		setupMocks   func(mock sqlmock.Sqlmock)
		email        string
		expectedUser *model.User
		expectedErr  error
	}

	const validEmail = "test@example.com"
	const nonExistentEmail = "notfound@example.com"
	const invalidEmail = "invalid"

	user := &model.User{

		UserID: 1,
		Email:  validEmail,
	}

	scenarios := []testScenario{
		{
			name: "Valid Email with Existing User",
			setupMocks: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("^SELECT (.+) FROM `users` WHERE (.+)$").
					WithArgs(validEmail).
					WillReturnRows(sqlmock.NewRows([]string{"user_id", "email"}).AddRow(user.UserID, user.Email))
			},
			email:        validEmail,
			expectedUser: user,
			expectedErr:  nil,
		},
		{
			name: "Non-Existent Email",
			setupMocks: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("^SELECT (.+) FROM `users` WHERE (.+)$").
					WithArgs(nonExistentEmail).
					WillReturnError(gorm.ErrRecordNotFound)
			},
			email:        nonExistentEmail,
			expectedUser: nil,
			expectedErr:  gorm.ErrRecordNotFound,
		},
		{
			name: "Invalid Email Format",
			setupMocks: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("^SELECT (.+) FROM `users` WHERE (.+)$").
					WithArgs(invalidEmail).
					WillReturnError(gorm.ErrRecordNotFound)
			},
			email:        invalidEmail,
			expectedUser: nil,
			expectedErr:  gorm.ErrRecordNotFound,
		},
		{
			name: "Database Connection Failure",
			setupMocks: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("^SELECT (.+) FROM `users` WHERE (.+)$").
					WillReturnError(errors.New("database connection error"))
			},
			email:        validEmail,
			expectedUser: nil,
			expectedErr:  errors.New("database connection error"),
		},
	}

	for _, scenario := range scenarios {
		scenario := scenario
		t.Run(scenario.name, func(t *testing.T) {

			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("failed to open sqlmock database connection: %v", err)
			}
			defer db.Close()

			gormDB, err := gorm.Open("mysql", db)
			if err != nil {
				t.Fatalf("failed to initialize gorm DB: %v", err)
			}
			defer gormDB.Close()

			userStore := &UserStore{db: gormDB}
			scenario.setupMocks(mock)

			result, err := userStore.GetByEmail(scenario.email)

			if (err != nil) && (err.Error() != scenario.expectedErr.Error()) {
				t.Errorf("expected error %v, but got %v", scenario.expectedErr, err)
			}

			if result != nil {
				if result.Email != scenario.expectedUser.Email {
					t.Errorf("expected user email %v, but got %v", scenario.expectedUser.Email, result.Email)
				}
			} else if scenario.expectedUser != nil {
				t.Errorf("expected user %v, but got nil", scenario.expectedUser)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %v", err)
			}
		})
	}
}


/*
ROOST_METHOD_HASH=GetByID_bbf946112e
ROOST_METHOD_SIG_HASH=GetByID_728dd55ed1


 */
func TestGetByID_DBConnectionError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open mock sql db, %v", err)
	}
	defer db.Close()

	gormDB, err := gorm.Open("postgres", db)
	if err != nil {
		t.Fatalf("failed to initialize gorm db, %v", err)
	}

	store := &UserStore{db: gormDB}

	testID := 1

	mock.ExpectQuery("SELECT * FROM users WHERE id = ?").WithArgs(testID).WillReturnError(errors.New("connection failed"))

	user, err := store.GetByID(uint(testID))
	if err == nil || err.Error() != "connection failed" {
		t.Errorf("expected connection error, got %v", err)
	}

	if user != nil {
		t.Errorf("expected user to be nil, got %v", user)
	}

	t.Log("Scenario 3: Database connection error handling - Passed")
}

func TestGetByID_InvalidID(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open mock sql db, %v", err)
	}
	defer db.Close()

	gormDB, err := gorm.Open("postgres", db)
	if err != nil {
		t.Fatalf("failed to initialize gorm db, %v", err)
	}

	store := &UserStore{db: gormDB}

	testID := 0

	user, err := store.GetByID(uint(testID))
	if err == nil {
		t.Errorf("expected an error for zero ID, got nil")
	}

	if user != nil {
		t.Errorf("expected user to be nil for zero ID, got %v", user)
	}

	t.Log("Scenario 4: Invalid ID parameter (Zero ID) - Passed")
}

func TestGetByID_SuccessfulRetrieval(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open mock sql db, %v", err)
	}
	defer db.Close()

	gormDB, err := gorm.Open("postgres", db)
	if err != nil {
		t.Fatalf("failed to initialize gorm db, %v", err)
	}

	store := &UserStore{db: gormDB}

	testID := 1
	expectedUser := model.User{ID: uint(testID), Username: "John Doe"}

	rows := sqlmock.NewRows([]string{"id", "username"}).AddRow(expectedUser.ID, expectedUser.Username)
	mock.ExpectQuery("SELECT * FROM users WHERE id = ?").WithArgs(testID).WillReturnRows(rows)

	user, err := store.GetByID(uint(testID))
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if user == nil || user.ID != expectedUser.ID || user.Username != expectedUser.Username {
		t.Errorf("expected user %v, got %v", expectedUser, user)
	}

	t.Log("Scenario 1: Successfully retrieve user by existing ID - Passed")
}

func TestGetByID_UserNotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open mock sql db, %v", err)
	}
	defer db.Close()

	gormDB, err := gorm.Open("postgres", db)
	if err != nil {
		t.Fatalf("failed to initialize gorm db, %v", err)
	}

	store := &UserStore{db: gormDB}

	testID := 100

	mock.ExpectQuery("SELECT * FROM users WHERE id = ?").WithArgs(testID).WillReturnError(gorm.ErrRecordNotFound)

	user, err := store.GetByID(uint(testID))
	if err == nil || err != gorm.ErrRecordNotFound {
		t.Errorf("expected error %v, got %v", gorm.ErrRecordNotFound, err)
	}

	if user != nil {
		t.Errorf("expected user to be nil, got %v", user)
	}

	t.Log("Scenario 2: User not found with a given ID - Passed")
}


/*
ROOST_METHOD_HASH=GetByUsername_f11f114df2
ROOST_METHOD_SIG_HASH=GetByUsername_954d096e24


 */
func TestGetByUsername(t *testing.T) {
	t.Run("Scenario 1: Successful retrieval of a user by a valid username", func(t *testing.T) {

		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
		}
		defer db.Close()

		username := "testuser"
		expectedUser := model.User{Username: username}

		mock.ExpectQuery("^SELECT (.+) FROM `users` WHERE username = \\?$").
			WithArgs(username).
			WillReturnRows(sqlmock.NewRows([]string{"username"}).AddRow(username))

		gormDB, err := gorm.Open("sqlite3", db)
		if err != nil {
			t.Fatalf("failed to open gorm db: %v", err)
		}
		defer gormDB.Close()

		store := UserStore{db: gormDB}

		user, err := store.GetByUsername(username)

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if user.Username != expectedUser.Username {
			t.Errorf("expected %v, got %v", expectedUser, *user)
		}

		t.Log("Successfully retrieved user by valid username")
	})

	t.Run("Scenario 2: User not found for a non-existent username", func(t *testing.T) {

		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
		}
		defer db.Close()

		username := "nonexistent"
		mock.ExpectQuery("^SELECT (.+) FROM `users` WHERE username = \\?$").
			WithArgs(username).
			WillReturnRows(sqlmock.NewRows([]string{"username"}))

		gormDB, err := gorm.Open("sqlite3", db)
		if err != nil {
			t.Fatalf("failed to open gorm db: %v", err)
		}
		defer gormDB.Close()

		store := UserStore{db: gormDB}

		user, err := store.GetByUsername(username)

		if err == nil {
			t.Errorf("expected error, got nil")
		}
		if user != nil {
			t.Errorf("expected nil user, got %v", user)
		}

		t.Log("Handled user not found by non-existent username")
	})

	t.Run("Scenario 3: Database error during query execution", func(t *testing.T) {

		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
		}
		defer db.Close()

		username := "dberror"
		mock.ExpectQuery("^SELECT (.+) FROM `users` WHERE username = \\?$").
			WithArgs(username).
			WillReturnError(errors.New("query error"))

		gormDB, err := gorm.Open("sqlite3", db)
		if err != nil {
			t.Fatalf("failed to open gorm db: %v", err)
		}
		defer gormDB.Close()

		store := UserStore{db: gormDB}

		user, err := store.GetByUsername(username)

		if err == nil {
			t.Errorf("expected error, got nil")
		}
		if user != nil {
			t.Errorf("expected nil user, got %v", user)
		}

		t.Log("Handled database error during query execution")
	})

	t.Run("Scenario 4: Handling of SQL injection attempt in username", func(t *testing.T) {

		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
		}
		defer db.Close()

		username := "'; DROP TABLE users; --"
		mock.ExpectQuery("^SELECT (.+) FROM `users` WHERE username = \\?$").
			WithArgs(username).
			WillReturnRows(sqlmock.NewRows([]string{"username"}))

		gormDB, err := gorm.Open("sqlite3", db)
		if err != nil {
			t.Fatalf("failed to open gorm db: %v", err)
		}
		defer gormDB.Close()

		store := UserStore{db: gormDB}

		user, err := store.GetByUsername(username)

		if err == nil || err != gorm.ErrRecordNotFound {
			t.Errorf("expected no error, or specific not found error, got %v", err)
		}
		if user != nil {
			t.Errorf("expected nil user, got %v", user)
		}

		t.Log("Safely handled SQL injection attempt in username")
	})

	t.Run("Scenario 5: Retrieval of a user when the username includes special characters", func(t *testing.T) {

		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
		}
		defer db.Close()

		username := "user+name_test"
		expectedUser := model.User{Username: username}
		mock.ExpectQuery("^SELECT (.+) FROM `users` WHERE username = \\?$").
			WithArgs(username).
			WillReturnRows(sqlmock.NewRows([]string{"username"}).AddRow(username))

		gormDB, err := gorm.Open("sqlite3", db)
		if err != nil {
			t.Fatalf("failed to open gorm db: %v", err)
		}
		defer gormDB.Close()

		store := UserStore{db: gormDB}

		user, err := store.GetByUsername(username)

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if user.Username != expectedUser.Username {
			t.Errorf("expected %v, got %v", expectedUser, *user)
		}

		t.Log("Successfully retrieved user with special character username")
	})

}


/*
ROOST_METHOD_HASH=IsFollowing_f53a5d9cef
ROOST_METHOD_SIG_HASH=IsFollowing_9eba5a0e9c


 */
func TestIsFollowing(t *testing.T) {

	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	gormDB, err := gorm.Open("postgres", db)
	assert.NoError(t, err)

	userStore := store.UserStore{DB: gormDB}

	t.Run("Scenario 1: Null user input", func(t *testing.T) {

		userA := &model.User{}
		userB := &model.User{Model: gorm.Model{ID: 2}}

		isFollowing, err := userStore.IsFollowing(nil, userB)
		assert.NoError(t, err)
		assert.False(t, isFollowing, "Expected false when userA is nil")

		isFollowing, err = userStore.IsFollowing(userA, nil)
		assert.NoError(t, err)
		assert.False(t, isFollowing, "Expected false when userB is nil")

		isFollowing, err = userStore.IsFollowing(nil, nil)
		assert.NoError(t, err)
		assert.False(t, isFollowing, "Expected false when both users are nil")
	})

	t.Run("Scenario 2: User `a` is following User `b`", func(t *testing.T) {
		userA := &model.User{Model: gorm.Model{ID: 1}}
		userB := &model.User{Model: gorm.Model{ID: 2}}

		mock.ExpectQuery("^SELECT COUNT\\(\\*\\) FROM follows WHERE from_user_id = \\? AND to_user_id = \\?$").
			WithArgs(userA.ID, userB.ID).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

		isFollowing, err := userStore.IsFollowing(userA, userB)
		assert.NoError(t, err)
		assert.True(t, isFollowing, "Expected true when userA is following userB")
	})

	t.Run("Scenario 3: User `a` is not following User `b`", func(t *testing.T) {
		userA := &model.User{Model: gorm.Model{ID: 1}}
		userB := &model.User{Model: gorm.Model{ID: 2}}

		mock.ExpectQuery("^SELECT COUNT\\(\\*\\) FROM follows WHERE from_user_id = \\? AND to_user_id = \\?$").
			WithArgs(userA.ID, userB.ID).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

		isFollowing, err := userStore.IsFollowing(userA, userB)
		assert.NoError(t, err)
		assert.False(t, isFollowing, "Expected false when userA is not following userB")
	})

	t.Run("Scenario 4: Database error during follow check", func(t *testing.T) {
		userA := &model.User{Model: gorm.Model{ID: 1}}
		userB := &model.User{Model: gorm.Model{ID: 2}}

		mock.ExpectQuery("^SELECT COUNT\\(\\*\\) FROM follows WHERE from_user_id = \\? AND to_user_id = \\?$").
			WithArgs(userA.ID, userB.ID).
			WillReturnError(errors.New("database error"))

		isFollowing, err := userStore.IsFollowing(userA, userB)
		assert.Error(t, err)
		assert.EqualError(t, err, "database error")
		assert.False(t, isFollowing, "Expected false when there is a database error")
	})

	t.Run("Scenario 5: User `a` has no follow records", func(t *testing.T) {
		userA := &model.User{Model: gorm.Model{ID: 1}}
		userB := &model.User{Model: gorm.Model{ID: 2}}

		mock.ExpectQuery("^SELECT COUNT\\(\\*\\) FROM follows WHERE from_user_id = \\? AND to_user_id = \\?$").
			WithArgs(userA.ID, userB.ID).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

		isFollowing, err := userStore.IsFollowing(userA, userB)
		assert.NoError(t, err)
		assert.False(t, isFollowing, "Expected false when userA has no follow records")
	})
}


/*
ROOST_METHOD_HASH=NewUserStore_6a331dd890
ROOST_METHOD_SIG_HASH=NewUserStore_4f0c2dfca9


 */
func TestNewUserStore(t *testing.T) {
	tests := []struct {
		name         string
		setup        func() *gorm.DB
		expectedNil  bool
		expectedType string
		verifyResult func(*UserStore) bool
	}{
		{
			name: "Valid gorm.DB instance",
			setup: func() *gorm.DB {

				db, _, err := sqlmock.New()
				if err != nil {
					log.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
				}
				return &gorm.DB{
					Config: &gorm.Config{},
				}

			},
			expectedNil:  false,
			expectedType: "*store.UserStore",
			verifyResult: func(us *UserStore) bool {
				return us.DB != nil
			},
		},
		{
			name: "Nil gorm.DB instance",
			setup: func() *gorm.DB {
				return nil
			},
			expectedNil:  true,
			expectedType: "*store.UserStore",
			verifyResult: func(us *UserStore) bool {
				return us.DB == nil
			},
		},
		{
			name: "Idempotency with a single gorm.DB instance",
			setup: func() *gorm.DB {
				db, _, err := sqlmock.New()
				if err != nil {
					log.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
				}
				return &gorm.DB{
					Config: &gorm.Config{},
				}
			},
			expectedNil:  false,
			expectedType: "*store.UserStore",
			verifyResult: func(us *UserStore) bool {

				mockDB := &gorm.DB{Config: us.DB.Config}
				first := store.NewUserStore(mockDB)
				second := store.NewUserStore(mockDB)
				return first != second
			},
		},
		{
			name: "Mocked DB with Error",
			setup: func() *gorm.DB {
				db, mock, err := sqlmock.New()
				if err != nil {
					log.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
				}
				mock.ExpectQuery("SELECT 1").WillReturnError(gorm.ErrInvalidSQL)
				return &gorm.DB{
					Config: &gorm.Config{},
					Error:  gorm.ErrInvalidSQL,
				}
			},
			expectedNil:  false,
			expectedType: "*store.UserStore",
			verifyResult: func(us *UserStore) bool {
				return us.DB.Error != nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := tt.setup()
			got := store.NewUserStore(db)

			if reflect.TypeOf(got).String() != tt.expectedType {
				t.Errorf("expected type %s, but got %s", tt.expectedType, reflect.TypeOf(got).String())
			}

			if tt.verifyResult(got) != tt.expectedNil {
				t.Errorf("unexpected UserStore initialization; nil check failed")
			}
		})
	}

	t.Log("All test cases passed successfully for the NewUserStore function")
}


/*
ROOST_METHOD_HASH=Unfollow_57959a8a53
ROOST_METHOD_SIG_HASH=Unfollow_8bd8e0bc55


 */
func TestUserStoreUnfollow(t *testing.T) {

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock database connection: %v", err)
	}
	defer db.Close()

	gormDB, err := gorm.Open("sqlite3", db)
	if err != nil {
		t.Fatalf("failed to open gorm DB: %v", err)
	}

	store := &UserStore{db: gormDB}

	type testCase struct {
		Name          string
		Arrange       func()
		Act           func() error
		Assert        func(err error)
		ExpectedError error
	}

	tests := []testCase{
		{
			Name: "Successfully unfollow a user",
			Arrange: func() {
				userA := &model.User{Model: gorm.Model{ID: 1}, Username: "userA"}
				userB := &model.User{Model: gorm.Model{ID: 2}, Username: "userB"}

				mock.ExpectQuery("^SELECT (.+) FROM follows WHERE (.+)$").
					WithArgs(userA.ID, userB.ID).
					WillReturnRows(sqlmock.NewRows(nil))

				mock.ExpectExec("^DELETE FROM follows WHERE (.+)$").
					WithArgs(userA.ID, userB.ID).
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
			Act: func() error {
				userA := &model.User{Model: gorm.Model{ID: 1}}
				userB := &model.User{Model: gorm.Model{ID: 2}}
				return store.Unfollow(userA, userB)
			},
			Assert: func(err error) {
				assert.Nil(t, err)
			},
		},
		{
			Name: "Attempt to unfollow a user that is not being followed",
			Arrange: func() {
				userA := &model.User{Model: gorm.Model{ID: 1}, Username: "userA"}
				userB := &model.User{Model: gorm.Model{ID: 2}, Username: "userB"}

				mock.ExpectExec("^DELETE FROM follows WHERE (.+)$").
					WithArgs(userA.ID, userB.ID).
					WillReturnResult(sqlmock.NewResult(1, 0))
			},
			Act: func() error {
				userA := &model.User{Model: gorm.Model{ID: 1}}
				userB := &model.User{Model: gorm.Model{ID: 2}}
				return store.Unfollow(userA, userB)
			},
			Assert: func(err error) {
				assert.Nil(t, err)
			},
		},
		{
			Name: "Handle case where the database fails",
			Arrange: func() {
				userA := &model.User{Model: gorm.Model{ID: 1}, Username: "userA"}
				userB := &model.User{Model: gorm.Model{ID: 2}, Username: "userB"}

				mock.ExpectExec("^DELETE FROM follows WHERE (.+)$").
					WithArgs(userA.ID, userB.ID).
					WillReturnError(gorm.ErrInvalidSQL)
			},
			Act: func() error {
				userA := &model.User{Model: gorm.Model{ID: 1}}
				userB := &model.User{Model: gorm.Model{ID: 2}}
				return store.Unfollow(userA, userB)
			},
			Assert: func(err error) {
				assert.EqualError(t, err, gorm.ErrInvalidSQL.Error())
			},
		},
		{
			Name: "Unfollow when database operation affects multiple rows unexpectedly",
			Arrange: func() {
				userA := &model.User{Model: gorm.Model{ID: 1}, Username: "userA"}
				userB := &model.User{Model: gorm.Model{ID: 2}, Username: "userB"}

				mock.ExpectExec("^DELETE FROM follows WHERE (.+)$").
					WithArgs(userA.ID, userB.ID).
					WillReturnResult(sqlmock.NewResult(1, 2))
			},
			Act: func() error {
				userA := &model.User{Model: gorm.Model{ID: 1}}
				userB := &model.User{Model: gorm.Model{ID: 2}}
				return store.Unfollow(userA, userB)
			},
			Assert: func(err error) {
				assert.Error(t, err)

			},
		},
		{
			Name: "Attempt to unfollow a user with invalid input",
			Arrange: func() {

			},
			Act: func() error {
				userA := &model.User{Model: gorm.Model{ID: 1}}
				return store.Unfollow(userA, nil)
			},
			Assert: func(err error) {
				assert.Error(t, err)

			},
		},
		{
			Name: "Unfollow a user where userA and userB are the same",
			Arrange: func() {
				userA := &model.User{Model: gorm.Model{ID: 1}, Username: "userA"}

				mock.ExpectExec("^DELETE FROM follows WHERE (.+)$").
					WithArgs(userA.ID, userA.ID).
					WillReturnResult(sqlmock.NewResult(1, 0))
			},
			Act: func() error {
				userA := &model.User{Model: gorm.Model{ID: 1}}
				return store.Unfollow(userA, userA)
			},
			Assert: func(err error) {
				assert.Nil(t, err)

			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			tt.Arrange()
			err := tt.Act()
			tt.Assert(err)

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}


/*
ROOST_METHOD_HASH=Update_68f27dd78a
ROOST_METHOD_SIG_HASH=Update_87150d6435


 */
func TestUpdate(t *testing.T) {
	var (
		sqlDB *gorm.DB
		mock  sqlmock.Sqlmock
		err   error
	)

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open mock sql db, got error: %v", err)
	}
	defer db.Close()

	sqlDB, err = gorm.Open("mysql", db)
	if err != nil {
		t.Fatalf("failed to open gorm db, got error: %v", err)
	}
	defer sqlDB.Close()

	userStore := UserStore{db: sqlDB}

	tests := []struct {
		name      string
		setup     func() *model.User
		mock      func(user *model.User)
		wantError bool
	}{
		{
			name: "Successfully Update a User's Information",
			setup: func() *model.User {
				user := &model.User{
					Model:    gorm.Model{ID: 1},
					Username: "test_user",
					Email:    "test@example.com",
					Bio:      "This is a test user.",
					Image:    "image_url",
				}
				return user
			},
			mock: func(user *model.User) {
				mock.ExpectBegin()
				mock.ExpectExec("UPDATE `users`").
					WithArgs(user.Bio, user.Image, user.ID).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
			wantError: false,
		},
		{
			name: "Attempt to Update Non-Existent User",
			setup: func() *model.User {
				return &model.User{
					Model: gorm.Model{ID: 9999},
					Bio:   "Updated bio",
					Image: "updated_image_url",
				}
			},
			mock: func(user *model.User) {
				mock.ExpectBegin()
				mock.ExpectExec("UPDATE `users`").
					WithArgs(user.Bio, user.Image, user.ID).
					WillReturnResult(sqlmock.NewResult(1, 0))
				mock.ExpectCommit()
			},
			wantError: true,
		},
		{
			name: "Error During Database Operation",
			setup: func() *model.User {
				return &model.User{
					Model:    gorm.Model{ID: 1},
					Username: "test_user",
					Email:    "test@example.com",
					Bio:      "Bio with error simulation",
					Image:    "image_url_with_error",
				}
			},
			mock: func(user *model.User) {
				mock.ExpectBegin()
				mock.ExpectExec("UPDATE `users`").
					WithArgs(user.Bio, user.Image, user.ID).
					WillReturnError(errors.New("database operation error"))
				mock.ExpectRollback()
			},
			wantError: true,
		},
		{
			name: "Update User with Duplicate Constraints Violation",
			setup: func() *model.User {
				return &model.User{
					Model:    gorm.Model{ID: 2},
					Username: "user_two",
					Email:    "duplicate@example.com",
					Bio:      "Updated bio",
					Image:    "updated_image_url",
				}
			},
			mock: func(user *model.User) {
				mock.ExpectBegin()
				mock.ExpectExec("UPDATE `users`").
					WithArgs(user.Bio, user.Image, user.ID).
					WillReturnError(errors.New("duplicate email error"))
				mock.ExpectRollback()
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := tt.setup()
			tt.mock(user)

			err := userStore.Update(user)
			if (err != nil) != tt.wantError {
				t.Errorf("UserStore.Update() error = %v, wantError %v", err, tt.wantError)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}

			t.Log(tt.name + " executed")
		})
	}

}


/*
ROOST_METHOD_HASH=GetFollowingUserIDs_ba3670aa2c
ROOST_METHOD_SIG_HASH=GetFollowingUserIDs_55ccc2afd7


 */
func TestGetFollowingUserIDs(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock database connection: %s", err)
	}
	defer db.Close()

	gormDB, err := gorm.Open("sqlite3", db)
	if err != nil {
		t.Fatalf("failed to open gorm DB: %s", err)
	}
	defer gormDB.Close()

	userStore := &UserStore{db: gormDB}

	tests := []struct {
		name        string
		userID      uint
		setupMock   func()
		expectedIDs []uint
		expectErr   bool
	}{
		{
			name:   "Scenario 1: Retrieve IDs When User Follows Multiple Accounts",
			userID: 1,
			setupMock: func() {
				rows := sqlmock.NewRows([]string{"to_user_id"}).AddRow(2).AddRow(3).AddRow(4)
				mock.ExpectQuery("SELECT to_user_id FROM follows WHERE from_user_id = ?").
					WithArgs(1).
					WillReturnRows(rows)
			},
			expectedIDs: []uint{2, 3, 4},
			expectErr:   false,
		},
		{
			name:   "Scenario 2: No Following Accounts",
			userID: 2,
			setupMock: func() {
				rows := sqlmock.NewRows([]string{"to_user_id"})
				mock.ExpectQuery("SELECT to_user_id FROM follows WHERE from_user_id = ?").
					WithArgs(2).
					WillReturnRows(rows)
			},
			expectedIDs: []uint{},
			expectErr:   false,
		},
		{
			name:   "Scenario 3: Database Error on Query Execution",
			userID: 3,
			setupMock: func() {
				mock.ExpectQuery("SELECT to_user_id FROM follows WHERE from_user_id = ?").
					WithArgs(3).
					WillReturnError(gorm.ErrInvalidSQL)
			},
			expectedIDs: []uint{},
			expectErr:   true,
		},
		{
			name:   "Scenario 4: Retrieve IDs with Single Follow",
			userID: 4,
			setupMock: func() {
				rows := sqlmock.NewRows([]string{"to_user_id"}).AddRow(5)
				mock.ExpectQuery("SELECT to_user_id FROM follows WHERE from_user_id = ?").
					WithArgs(4).
					WillReturnRows(rows)
			},
			expectedIDs: []uint{5},
			expectErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			mUser := &model.User{Model: gorm.Model{ID: tt.userID}}
			ids, err := userStore.GetFollowingUserIDs(mUser)

			if tt.expectErr && err == nil {
				t.Errorf("expected an error but got none")
			}
			if !tt.expectErr && err != nil {
				t.Errorf("expected no error, but got: %v", err)
			}

			if len(ids) != len(tt.expectedIDs) {
				t.Errorf("expected %v, but got %v", tt.expectedIDs, ids)
			} else {
				for i, id := range ids {
					if id != tt.expectedIDs[i] {
						t.Errorf("expected ID %d, but got %d", tt.expectedIDs[i], id)
					}
				}
			}
		})
	}

	err = mock.ExpectationsWereMet()
	if err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

