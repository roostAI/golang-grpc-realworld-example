package github

import (
	"testing"
	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/require"
	"time"
	"github.com/raahii/golang-grpc-realworld-example/model"
	"github.com/stretchr/testify/assert"
	"errors"
	_ "github.com/jinzhu/gorm/dialects/postgres"
)









/*
ROOST_METHOD_HASH=NewUserStore_6a331dd890
ROOST_METHOD_SIG_HASH=NewUserStore_4f0c2dfca9

FUNCTION_DEF=func NewUserStore(db *gorm.DB) *UserStore 

 */
func TestNewUserStore(t *testing.T) {
	tests := []struct {
		name        string
		setup       func() *gorm.DB
		expectNilDB bool
		expectType  bool
	}{
		{
			name: "Creating a UserStore with a Valid DB Connection",
			setup: func() *gorm.DB {

				db, mock, err := sqlmock.New()
				require.NoError(t, err)
				mock.ExpectQuery("SELECT 1").WillReturnRows(sqlmock.NewRows([]string{"1"}).AddRow(1))

				gormDB, err := gorm.Open("_", db)
				require.NoError(t, err)
				return gormDB
			},
			expectNilDB: false,
			expectType:  true,
		},
		{
			name: "Creating a UserStore with a Nil DB Connection",
			setup: func() *gorm.DB {
				return nil
			},
			expectNilDB: true,
			expectType:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := tt.setup()
			defer func() {
				if db != nil {
					_ = db.Close()
				}
			}()

			userStore := NewUserStore(db)

			if tt.expectNilDB {
				if userStore.db != nil {
					t.Errorf("expected nil db, got %v", userStore.db)
					t.Log("Failure Reason: UserStore.db should be nil when initialized with a nil DB connection")
				} else {
					t.Log("Success: UserStore.db is nil as expected when initialized with a nil DB connection")
				}
			} else {
				if userStore.db == nil {
					t.Errorf("expected a valid db, got nil")
					t.Log("Failure Reason: UserStore.db should not be nil when initialized with a valid DB connection")
				} else {
					t.Log("Success: UserStore.db is valid as expected when initialized with a valid DB connection")
				}
			}

			_, ok := interface{}(userStore).(*UserStore)
			if ok != tt.expectType {
				t.Errorf("expected type *UserStore, got different type")
				t.Log("Failure Reason: Returned object type is not *UserStore as expected")
			} else {
				t.Log("Success: Returned object type is *UserStore as expected")
			}
		})
	}
}


/*
ROOST_METHOD_HASH=GetByID_bbf946112e
ROOST_METHOD_SIG_HASH=GetByID_728dd55ed1

FUNCTION_DEF=func (s *UserStore) GetByID(id uint) (*model.User, error) 

 */
func TestUserStoreGetById(t *testing.T) {

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock database: %v", err)
	}
	defer db.Close()

	gormDB, err := gorm.Open("sqlite3", db)
	if err != nil {
		t.Fatalf("failed to open gorm DB: %v", err)
	}
	store := &UserStore{db: gormDB}

	type testCase struct {
		name         string
		setupMock    func()
		userID       uint
		expectedErr  error
		expectedUser *model.User
	}

	testCases := []testCase{
		{
			name: "Successfully Retrieve a User by ID",
			setupMock: func() {
				rows := sqlmock.NewRows([]string{"id", "username", "email", "password", "bio", "image", "created_at", "updated_at", "deleted_at"}).
					AddRow(1, "exampleUser", "user@example.com", "hashedpassword", "bio", "image", time.Now(), time.Now(), nil)

				mock.ExpectQuery("^SELECT \\* FROM `users` WHERE").
					WithArgs(1).
					WillReturnRows(rows)
			},
			userID:       1,
			expectedErr:  nil,
			expectedUser: &model.User{Model: gorm.Model{ID: 1}, Username: "exampleUser", Email: "user@example.com"},
		},
		{
			name: "User Not Found for Given ID",
			setupMock: func() {
				mock.ExpectQuery("^SELECT \\* FROM `users` WHERE").
					WithArgs(2).
					WillReturnError(gorm.ErrRecordNotFound)
			},
			userID:       2,
			expectedErr:  gorm.ErrRecordNotFound,
			expectedUser: nil,
		},
		{
			name: "Database Connection Error During User Retrieval",
			setupMock: func() {
				mock.ExpectQuery("^SELECT \\* FROM `users` WHERE").
					WithArgs(3).
					WillReturnError(sqlmock.ErrCancelled)
			},
			userID:       3,
			expectedErr:  sqlmock.ErrCancelled,
			expectedUser: nil,
		},
		{
			name: "Handle Invalid ID Parameter (Zero Value)",
			setupMock: func() {
				mock.ExpectQuery("^SELECT \\* FROM `users` WHERE").
					WithArgs(0).
					WillReturnError(gorm.ErrInvalidSQL)
			},
			userID:       0,
			expectedErr:  gorm.ErrInvalidSQL,
			expectedUser: nil,
		},
		{
			name: "Valid ID but Deleted User Record",
			setupMock: func() {
				deletedAt := time.Now()
				rows := sqlmock.NewRows([]string{"id", "username", "email", "password", "bio", "image", "created_at", "updated_at", "deleted_at"}).
					AddRow(4, "deletedUser", "deleted@example.com", "hashedpassword", "bio", "image", time.Now(), time.Now(), &deletedAt)

				mock.ExpectQuery("^SELECT \\* FROM `users` WHERE").
					WithArgs(4).
					WillReturnRows(rows)
			},
			userID:       4,
			expectedErr:  gorm.ErrRecordNotFound,
			expectedUser: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.setupMock()
			user, err := store.GetByID(tc.userID)

			if tc.expectedErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tc.expectedErr, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedUser, user)
			}

			err = mock.ExpectationsWereMet()
			assert.NoError(t, err)
		})
	}
}


/*
ROOST_METHOD_HASH=GetByEmail_3574af40e5
ROOST_METHOD_SIG_HASH=GetByEmail_5731b833c1

FUNCTION_DEF=func (s *UserStore) GetByEmail(email string) (*model.User, error) 

 */
func TestUserStoreGetByEmail(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	gormDB, err := gorm.Open("sqlmock", db)
	if err != nil {
		t.Fatalf("an error '%s' was not expected when initializing GORM with sqlmock", err)
	}

	userStore := UserStore{db: gormDB}

	tableTests := []struct {
		name          string
		email         string
		setupMock     func()
		expectedUser  *model.User
		expectedError error
	}{
		{
			name:  "Scenario 1: Retrieve Existing User by Email",
			email: "test@example.com",
			setupMock: func() {
				rows := sqlmock.NewRows([]string{"id", "username", "email", "password", "bio", "image"}).
					AddRow(1, "testuser", "test@example.com", "hashedpassword", "Bio", "ImageURL")

				mock.ExpectQuery(`SELECT * FROM "users" WHERE (email = .+ .+) AND "users"."deleted_at" IS NULL ORDER BY "users"."id" ASC LIMIT 1`).
					WithArgs("test@example.com").
					WillReturnRows(rows)
			},
			expectedUser: &model.User{
				Model:    gorm.Model{ID: 1},
				Username: "testuser",
				Email:    "test@example.com",
				Password: "hashedpassword",
				Bio:      "Bio",
				Image:    "ImageURL",
			},
			expectedError: nil,
		},
		{
			name:  "Scenario 2: Email Not Found",
			email: "nonexistent@example.com",
			setupMock: func() {
				mock.ExpectQuery(`SELECT * FROM "users" WHERE (email = .+ .+) AND "users"."deleted_at" IS NULL ORDER BY "users"."id" ASC LIMIT 1`).
					WithArgs("nonexistent@example.com").
					WillReturnRows(sqlmock.NewRows(nil))
			},
			expectedUser:  nil,
			expectedError: gorm.ErrRecordNotFound,
		},
		{
			name:  "Scenario 3: Invalid Email Format",
			email: "invalid-email-format",
			setupMock: func() {

			},
			expectedUser:  nil,
			expectedError: gorm.ErrRecordNotFound,
		},
		{
			name:  "Scenario 4: Database Error Handling",
			email: "valid@example.com",
			setupMock: func() {
				mock.ExpectQuery(`SELECT * FROM "users" WHERE (email = .+ .+) AND "users"."deleted_at" IS NULL ORDER BY "users"."id" ASC LIMIT 1`).
					WithArgs("valid@example.com").
					WillReturnError(errors.New("DB error"))
			},
			expectedUser:  nil,
			expectedError: errors.New("DB error"),
		},
	}

	for _, tt := range tableTests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			result, err := userStore.GetByEmail(tt.email)
			if err != nil {
				if tt.expectedError == nil || err.Error() != tt.expectedError.Error() {
					t.Fatalf("unexpected error: got %v, want %v", err, tt.expectedError)
				}
			}

			if result != nil && tt.expectedUser != nil {
				if result.ID != tt.expectedUser.ID || result.Email != tt.expectedUser.Email || result.Username != tt.expectedUser.Username {
					t.Errorf("unexpected user: got %+v, want %+v", result, tt.expectedUser)
				}
			} else if result != tt.expectedUser {
				t.Errorf("unexpected user: got %+v, want nil", result)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}


/*
ROOST_METHOD_HASH=Update_68f27dd78a
ROOST_METHOD_SIG_HASH=Update_87150d6435

FUNCTION_DEF=func (s *UserStore) Update(m *model.User) error 

 */
func TestUserStoreUpdate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		prep         func(sqlmock.Sqlmock, *model.User)
		user         *model.User
		expectedErr  error
		expectedRows int64
	}{
		{
			name: "Update User Successfully",
			prep: func(mock sqlmock.Sqlmock, user *model.User) {
				mock.ExpectBegin()
				mock.ExpectExec("UPDATE \"users\"").
					WithArgs(user.Username, user.Email, user.Password, user.Bio, user.Image, user.ID).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
			user: &model.User{
				Model:    gorm.Model{ID: 1},
				Username: "updated_user",
				Email:    "updated_user@example.com",
				Password: "updatedpass",
				Bio:      "Updated bio",
				Image:    "updated_image",
			},
			expectedErr:  nil,
			expectedRows: 1,
		},
		{
			name: "Handle Database Error During Update",
			prep: func(mock sqlmock.Sqlmock, user *model.User) {
				mock.ExpectBegin()
				mock.ExpectExec("UPDATE \"users\"").
					WithArgs(user.Username, user.Email, user.Password, user.Bio, user.Image, user.ID).
					WillReturnError(errors.New("update failed"))
				mock.ExpectRollback()
			},
			user: &model.User{
				Model:    gorm.Model{ID: 2},
				Username: "error_user",
				Email:    "error_user@example.com",
				Password: "errorpass",
				Bio:      "Error bio",
				Image:    "error_image",
			},
			expectedErr: errors.New("update failed"),
		},
		{
			name: "Update User with No Changes",
			prep: func(mock sqlmock.Sqlmock, user *model.User) {
				mock.ExpectBegin()
				mock.ExpectExec("UPDATE \"users\"").
					WithArgs(user.Username, user.Email, user.Password, user.Bio, user.Image, user.ID).
					WillReturnResult(sqlmock.NewResult(1, 0))
				mock.ExpectCommit()
			},
			user: &model.User{
				Model:    gorm.Model{ID: 3},
				Username: "no_change_user",
				Email:    "no_change_user@example.com",
				Password: "nopass",
				Bio:      "No change bio",
				Image:    "no_image",
			},
			expectedErr:  nil,
			expectedRows: 0,
		},
		{
			name: "Update User with Constraints Violations",
			prep: func(mock sqlmock.Sqlmock, user *model.User) {
				mock.ExpectBegin()
				mock.ExpectExec("UPDATE \"users\"").
					WithArgs(user.Username, user.Email, user.Password, user.Bio, user.Image, user.ID).
					WillReturnError(errors.New("unique constraint violated"))
				mock.ExpectRollback()
			},
			user: &model.User{
				Model:    gorm.Model{ID: 4},
				Username: "duplicate_user",
				Email:    "duplicate_user@example.com",
				Password: "duplicatepass",
				Bio:      "Duplicate bio",
				Image:    "duplicate_image",
			},
			expectedErr: errors.New("unique constraint violated"),
		},
		{
			name: "Attempt to Update Non-existent User",
			prep: func(mock sqlmock.Sqlmock, user *model.User) {
				mock.ExpectBegin()
				mock.ExpectExec("UPDATE \"users\"").
					WithArgs(user.Username, user.Email, user.Password, user.Bio, user.Image, user.ID).
					WillReturnResult(sqlmock.NewResult(0, 0))
				mock.ExpectCommit()
			},
			user: &model.User{
				Model:    gorm.Model{ID: 404},
				Username: "ghost_user",
				Email:    "ghost_user@example.com",
				Password: "ghostpass",
				Bio:      "Ghost bio",
				Image:    "ghost_image",
			},
			expectedErr:  nil,
			expectedRows: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("Failed to open sqlmock database: %s", err)
			}
			defer db.Close()

			mockDB, err := gorm.Open("_", db)
			if err != nil {
				t.Fatalf("Failed to open gorm DB: %s", err)
			}

			store := &UserStore{db: mockDB}

			tt.prep(mock, tt.user)

			err = store.Update(tt.user)
			if err != nil && tt.expectedErr == nil {
				t.Errorf("Expected no error, got %v", err)
			}
			if err == nil && tt.expectedErr != nil {
				t.Errorf("Expected error %v, got none", tt.expectedErr)
			}

			if tt.expectedRows != store.db.RowsAffected {
				t.Errorf("Expected affected rows %v, got %v", tt.expectedRows, store.db.RowsAffected)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("There were unfulfilled expectations: %s", err)
			}
		})
	}
}


/*
ROOST_METHOD_HASH=Follow_48fdf1257b
ROOST_METHOD_SIG_HASH=Follow_8217e61c06

FUNCTION_DEF=func (s *UserStore) Follow(a *model.User, b *model.User) error 

 */
func TestUserStoreFollow(t *testing.T) {

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("error opening a stub database connection: %v", err)
	}
	defer db.Close()

	gormDB, err := gorm.Open("_", db)
	if err != nil {
		t.Fatalf("error opening a gorm database: %v", err)
	}

	userStore := &UserStore{db: gormDB}

	t.Run("Successful Follow Operation", func(t *testing.T) {
		userA := &model.User{Model: gorm.Model{ID: 1}, Username: "UserA"}
		userB := &model.User{Model: gorm.Model{ID: 2}, Username: "UserB"}

		mock.ExpectBegin()
		mock.ExpectExec(`INSERT INTO follows`).
			WithArgs(userA.ID, userB.ID).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		err := userStore.Follow(userA, userB)
		if err != nil {
			t.Errorf("Did not expect error, got %v", err)
		}
	})

	t.Run("Prevent Duplicates in Follows", func(t *testing.T) {
		userA := &model.User{Model: gorm.Model{ID: 1}, Username: "UserA"}
		userB := &model.User{Model: gorm.Model{ID: 2}, Username: "UserB"}

		mock.ExpectBegin()
		mock.ExpectExec(`INSERT INTO follows`).
			WithArgs(userA.ID, userB.ID).
			WillReturnError(errors.New("duplicate key value"))
		mock.ExpectRollback()

		err := userStore.Follow(userA, userB)
		if err == nil {
			t.Error("Expected an error, because a duplicate follow should not be allowed")
		}
	})

	t.Run("Invalid User Reference", func(t *testing.T) {
		userA := &model.User{Model: gorm.Model{ID: 1}, Username: "UserA"}

		err := userStore.Follow(userA, nil)
		if err == nil {
			t.Error("Expected an error when trying to follow with a nil user")
		}
	})

	t.Run("Database Error Handling", func(t *testing.T) {
		userA := &model.User{Model: gorm.Model{ID: 1}, Username: "UserA"}
		userB := &model.User{Model: gorm.Model{ID: 2}, Username: "UserB"}

		mock.ExpectBegin()
		mock.ExpectExec(`INSERT INTO follows`).
			WillReturnError(errors.New("database error"))
		mock.ExpectRollback()

		err := userStore.Follow(userA, userB)
		if err == nil {
			t.Error("Expected error due to database issue, but got none")
		}
	})

	t.Run("Self-Follow Attempt", func(t *testing.T) {
		userA := &model.User{Model: gorm.Model{ID: 1}, Username: "UserA"}

		err := userStore.Follow(userA, userA)
		if err == nil {
			t.Error("Expected an error or no action when user attempts to follow themselves")
		}
	})

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %v", err)
	}
}


/*
ROOST_METHOD_HASH=IsFollowing_f53a5d9cef
ROOST_METHOD_SIG_HASH=IsFollowing_9eba5a0e9c

FUNCTION_DEF=func (s *UserStore) IsFollowing(a *model.User, b *model.User) (bool, error) 

 */
func TestUserStoreIsFollowing(t *testing.T) {

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	gormDB, err := gorm.Open("mysql", db)
	if err != nil {
		t.Fatalf("failed to open gorm DB connection: %v", err)
	}
	defer gormDB.Close()

	userStore := &UserStore{db: gormDB}

	tests := []struct {
		name           string
		userA          *model.User
		userB          *model.User
		mockQuery      func()
		expectedResult bool
		expectedError  error
	}{
		{
			name:  "Scenario 1: Both Users are Valid and Following Exists",
			userA: &model.User{Model: gorm.Model{ID: 1}},
			userB: &model.User{Model: gorm.Model{ID: 2}},
			mockQuery: func() {
				mock.ExpectQuery("SELECT count(.+) FROM follows WHERE").
					WithArgs(1, 2).
					WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
			},
			expectedResult: true,
			expectedError:  nil,
		},
		{
			name:  "Scenario 2: Both Users are Valid and No Following Exists",
			userA: &model.User{Model: gorm.Model{ID: 3}},
			userB: &model.User{Model: gorm.Model{ID: 4}},
			mockQuery: func() {
				mock.ExpectQuery("SELECT count(.+) FROM follows WHERE").
					WithArgs(3, 4).
					WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
			},
			expectedResult: false,
			expectedError:  nil,
		},
		{
			name:           "Scenario 3: User 'a' is Nil",
			userA:          nil,
			userB:          &model.User{Model: gorm.Model{ID: 5}},
			mockQuery:      func() {},
			expectedResult: false,
			expectedError:  nil,
		},
		{
			name:           "Scenario 4: User 'b' is Nil",
			userA:          &model.User{Model: gorm.Model{ID: 6}},
			userB:          nil,
			mockQuery:      func() {},
			expectedResult: false,
			expectedError:  nil,
		},
		{
			name:  "Scenario 5: Error Occurs During Database Query",
			userA: &model.User{Model: gorm.Model{ID: 7}},
			userB: &model.User{Model: gorm.Model{ID: 8}},
			mockQuery: func() {
				mock.ExpectQuery("SELECT count(.+) FROM follows WHERE").
					WithArgs(7, 8).
					WillReturnError(gorm.ErrInvalidSQL)
			},
			expectedResult: false,
			expectedError:  gorm.ErrInvalidSQL,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			tt.mockQuery()
			result, err := userStore.IsFollowing(tt.userA, tt.userB)
			if result != tt.expectedResult || err != tt.expectedError {
				t.Errorf("IsFollowing() = %v, err = %v; expected %v, err = %v", result, err, tt.expectedResult, tt.expectedError)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
			t.Logf("%s: Successfully completed", tt.name)
		})
	}

}


/*
ROOST_METHOD_HASH=Unfollow_57959a8a53
ROOST_METHOD_SIG_HASH=Unfollow_8bd8e0bc55

FUNCTION_DEF=func (s *UserStore) Unfollow(a *model.User, b *model.User) error 

 */
func TestUserStoreUnfollow(t *testing.T) {

	createMockDB := func() (*gorm.DB, sqlmock.Sqlmock, error) {
		db, mock, err := sqlmock.New()
		if err != nil {
			return nil, nil, err
		}

		gdb, err := gorm.Open("_", db)
		if err != nil {
			return nil, nil, err
		}
		gdb.LogMode(true)

		return gdb, mock, nil
	}

	tests := []struct {
		name        string
		userA       *model.User
		userB       *model.User
		withMock    func(sqlmock sqlmock.Sqlmock)
		expectError bool
	}{
		{
			name: "Successful Unfollow Operation",
			userA: &model.User{
				Model:    gorm.Model{ID: 1},
				Username: "UserA",
			},
			userB: &model.User{
				Model:    gorm.Model{ID: 2},
				Username: "UserB",
			},
			withMock: func(mock sqlmock.Sqlmock) {

				mock.ExpectExec(`DELETE FROM "follows"`).
					WithArgs(2, 1).
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
			expectError: false,
		},
		{
			name: "Unfollow Non-Existent Association",
			userA: &model.User{
				Model:    gorm.Model{ID: 1},
				Username: "UserA",
			},
			userB: &model.User{
				Model:    gorm.Model{ID: 3},
				Username: "UserB",
			},
			withMock: func(mock sqlmock.Sqlmock) {

				mock.ExpectExec(`DELETE FROM "follows"`).
					WithArgs(3, 1).
					WillReturnResult(sqlmock.NewResult(1, 0))
			},
			expectError: false,
		},
		{
			name: "Unfollow with Null User",
			userA: &model.User{
				Model:    gorm.Model{ID: 1},
				Username: "UserA",
			},
			userB:       nil,
			withMock:    func(mock sqlmock.Sqlmock) {},
			expectError: true,
		},
		{
			name: "Unfollow When Database Returns an Error",
			userA: &model.User{
				Model:    gorm.Model{ID: 1},
				Username: "UserA",
			},
			userB: &model.User{
				Model:    gorm.Model{ID: 4},
				Username: "UserB",
			},
			withMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(`DELETE FROM "follows"`).
					WithArgs(4, 1).
					WillReturnError(errors.New("db error"))
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := createMockDB()
			if err != nil {
				t.Fatalf("failed to create mock db: %v", err)
			}
			defer db.Close()

			store := &UserStore{db: db}

			tt.withMock(mock)
			err = store.Unfollow(tt.userA, tt.userB)
			if (err != nil) != tt.expectError {
				t.Errorf("store.Unfollow() error = %v, expectError %v", err, tt.expectError)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}


/*
ROOST_METHOD_HASH=GetFollowingUserIDs_ba3670aa2c
ROOST_METHOD_SIG_HASH=GetFollowingUserIDs_55ccc2afd7

FUNCTION_DEF=func (s *UserStore) GetFollowingUserIDs(m *model.User) ([]uint, error) 

 */
func TestUserStoreGetFollowingUserIDs(t *testing.T) {

	tests := []struct {
		name        string
		setup       func(mock sqlmock.Sqlmock, userID uint)
		userID      uint
		expectedIDs []uint
		expectError bool
	}{
		{
			name: "Scenario 1: Retrieve IDs for a User Following Other Users",
			setup: func(mock sqlmock.Sqlmock, userID uint) {
				rows := sqlmock.NewRows([]string{"to_user_id"}).AddRow(2).AddRow(3)
				mock.ExpectQuery("^SELECT to_user_id FROM follows WHERE from_user_id = ?$").WithArgs(userID).WillReturnRows(rows)
			},
			userID:      1,
			expectedIDs: []uint{2, 3},
			expectError: false,
		},
		{
			name: "Scenario 2: No Following Relationships",
			setup: func(mock sqlmock.Sqlmock, userID uint) {
				rows := sqlmock.NewRows([]string{"to_user_id"})
				mock.ExpectQuery("^SELECT to_user_id FROM follows WHERE from_user_id = ?$").WithArgs(userID).WillReturnRows(rows)
			},
			userID:      1,
			expectedIDs: []uint{},
			expectError: false,
		},
		{
			name: "Scenario 3: Database Error Occurrence",
			setup: func(mock sqlmock.Sqlmock, userID uint) {
				mock.ExpectQuery("^SELECT to_user_id FROM follows WHERE from_user_id = ?$").WithArgs(userID).WillReturnError(gorm.ErrRecordNotFound)
			},
			userID:      1,
			expectedIDs: []uint{},
			expectError: true,
		},
		{
			name: "Scenario 4: A User Following a Large Number of Other Users",
			setup: func(mock sqlmock.Sqlmock, userID uint) {
				const numFollows = 1000
				rows := sqlmock.NewRows([]string{"to_user_id"})
				for i := 1; i <= numFollows; i++ {
					rows.AddRow(i)
				}
				mock.ExpectQuery("^SELECT to_user_id FROM follows WHERE from_user_id = ?$").WithArgs(userID).WillReturnRows(rows)
			},
			userID:      1,
			expectedIDs: makeRange(1, 1000),
			expectError: false,
		},
		{
			name: "Scenario 5: User Does Not Exist in Database",
			setup: func(mock sqlmock.Sqlmock, userID uint) {
				rows := sqlmock.NewRows([]string{"to_user_id"})
				mock.ExpectQuery("^SELECT to_user_id FROM follows WHERE from_user_id = ?$").WithArgs(userID).WillReturnRows(rows)
			},
			userID:      9999,
			expectedIDs: []uint{},
			expectError: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			db, mock := setupMockDB(t)
			defer db.Close()

			store := &UserStore{db: db}
			user := &model.User{Model: gorm.Model{ID: tc.userID}}

			tc.setup(mock, tc.userID)

			ids, err := store.GetFollowingUserIDs(user)
			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tc.expectedIDs, ids)

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func makeRange(min, max int) []uint {
	a := make([]uint, max-min+1)
	for i := range a {
		a[i] = uint(min + i)
	}
	return a
}

func setupMockDB(t *testing.T) (*gorm.DB, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("unexpected error while opening mock database: %s", err)
	}

	gormDB, err := gorm.Open("mysql", db)
	if err != nil {
		t.Fatalf("unexpected error while opening GORM DB: %s", err)
	}

	return gormDB, mock
}

