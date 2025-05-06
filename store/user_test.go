package store

import (
	errors "errors"
	reflect "reflect"
	debug "runtime/debug"
	sync "sync"
	testing "testing"

	"github.com/DATA-DOG/go-sqlmock"
	gorm "github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	model "github.com/raahii/golang-grpc-realworld-example/model"
	assert "github.com/stretchr/testify/assert"
)

type User struct {
	ID       uint   `gorm:"primaryKey"`
	Username string `gorm:"not null;unique"`
	Email    string `gorm:"not null;unique"`
	Password string `gorm:"not null"`
}

/*
ROOST_METHOD_HASH=UserStore_Create_9495ddb29d
ROOST_METHOD_SIG_HASH=UserStore_Create_18451817fe

FUNCTION_DEF=func (s *UserStore) Create(m *model.User) error // Create create a user
*/
func TestUserStoreCreate(t *testing.T) {
	tests := []struct {
		name         string
		user         *model.User
		mockBehavior func(mock sqlmock.Sqlmock)
		wantErr      bool
	}{
		{
			name: "Successful User Creation",
			user: &model.User{
				Username: "testuser",
				Email:    "test@example.com",
				Password: "securepassword",
			},
			mockBehavior: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO \"users\"").
					WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
			wantErr: false,
		},
		{
			name: "Database Error During User Creation",
			user: &model.User{
				Username: "testuser",
				Email:    "test@example.com",
				Password: "securepassword",
			},
			mockBehavior: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO \"users\"").
					WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
					WillReturnError(errors.New("database error"))
				mock.ExpectRollback()
			},
			wantErr: true,
		},
		{
			name: "Invalid User Model",
			user: &model.User{
				Username: "",
				Email:    "",
				Password: "",
			},
			mockBehavior: func(mock sqlmock.Sqlmock) {},
			wantErr:      true,
		},
		{
			name:         "Empty User Model",
			user:         &model.User{},
			mockBehavior: func(mock sqlmock.Sqlmock) {},
			wantErr:      true,
		},
		{
			name: "User Creation with Existing Email",
			user: &model.User{
				Username: "testuser",
				Email:    "existing@example.com",
				Password: "securepassword",
			},
			mockBehavior: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO \"users\"").
					WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
					WillReturnError(errors.New("duplicate email"))
				mock.ExpectRollback()
			},
			wantErr: true,
		},
		{
			name: "User Creation with Invalid Email Format",
			user: &model.User{
				Username: "testuser",
				Email:    "invalid-email",
				Password: "securepassword",
			},
			mockBehavior: func(mock sqlmock.Sqlmock) {},
			wantErr:      true,
		},
		{
			name: "User Creation with Long Password",
			user: &model.User{
				Username: "testuser",
				Email:    "test@example.com",
				Password: "thisisaverylongpasswordthatisevenlongerthanthemaximumallowedlength",
			},
			mockBehavior: func(mock sqlmock.Sqlmock) {},
			wantErr:      true,
		},
		{
			name: "User Creation with Special Characters in Username",
			user: &model.User{
				Username: "test@user",
				Email:    "test@example.com",
				Password: "securepassword",
			},
			mockBehavior: func(mock sqlmock.Sqlmock) {},
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Logf("Panic encountered so failing test. %v\n%s", r, string(debug.Stack()))
					t.Fail()
				}
			}()

			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
			}
			defer db.Close()

			gormDB, err := gorm.Open("sqlmock", db)
			if err != nil {
				t.Fatalf("an error '%s' was not expected when opening a gorm database connection", err)
			}
			defer gormDB.Close()

			tt.mockBehavior(mock)

			userStore := &UserStore{db: gormDB}

			err = userStore.Create(tt.user)

			if (err != nil) != tt.wantErr {
				t.Errorf("UserStore.Create() error = %v, wantErr %v", err, tt.wantErr)
			} else {
				t.Logf("Test passed for scenario: %s", tt.name)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

/*
ROOST_METHOD_HASH=UserStore_Update_4fd6d3d1c1
ROOST_METHOD_SIG_HASH=UserStore_Update_ddd5c151cf

FUNCTION_DEF=func (s *UserStore) Update(m *model.User) error // Update update all of user fields
*/
func TestUserStoreUpdate(t *testing.T) {

	type testCase struct {
		name        string
		setup       func() (*UserStore, *model.User, func())
		expectedErr error
		validate    func(*testing.T, *UserStore, *model.User)
	}

	newMockedUserStore := func() (*UserStore, sqlmock.Sqlmock, func()) {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
		}
		gormDB, err := gorm.Open("_", db)
		if err != nil {
			t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
		}
		store := &UserStore{db: gormDB}
		return store, mock, func() {
			db.Close()
		}
	}

	tests := []testCase{
		{
			name: "Successful User Update",
			setup: func() (*UserStore, *model.User, func()) {
				store, mock, teardown := newMockedUserStore()
				mock.ExpectBegin()
				mock.ExpectExec("UPDATE \"users\"").WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
				user := &model.User{Model: gorm.Model{ID: 1}, Username: "updatedUser", Email: "updated@example.com"}
				return store, user, teardown
			},
			expectedErr: nil,
			validate: func(t *testing.T, store *UserStore, user *model.User) {
				var updatedUser model.User
				err := store.db.First(&updatedUser, user.ID).Error
				if err != nil {
					t.Errorf("error fetching updated user: %v", err)
				}
				if !reflect.DeepEqual(updatedUser, *user) {
					t.Errorf("user not updated correctly, got %v, want %v", updatedUser, *user)
				}
			},
		},
		{
			name: "Update with Non-Existent User",
			setup: func() (*UserStore, *model.User, func()) {
				store, mock, teardown := newMockedUserStore()
				mock.ExpectBegin()
				mock.ExpectExec("UPDATE \"users\"").WillReturnError(gorm.ErrRecordNotFound)
				mock.ExpectCommit()
				user := &model.User{Model: gorm.Model{ID: 999}, Username: "nonExistentUser", Email: "nonexistent@example.com"}
				return store, user, teardown
			},
			expectedErr: gorm.ErrRecordNotFound,
			validate: func(t *testing.T, store *UserStore, user *model.User) {

			},
		},
		{
			name: "Update with Invalid User Model",
			setup: func() (*UserStore, *model.User, func()) {
				store, mock, teardown := newMockedUserStore()
				mock.ExpectBegin()
				mock.ExpectExec("UPDATE \"users\"").WillReturnError(gorm.ErrInvalidTransaction)
				mock.ExpectCommit()
				user := &model.User{Model: gorm.Model{ID: 1}, Username: "", Email: "invalid@example.com"}
				return store, user, teardown
			},
			expectedErr: gorm.ErrInvalidTransaction,
			validate: func(t *testing.T, store *UserStore, user *model.User) {

			},
		},
		{
			name: "Database Connection Error",
			setup: func() (*UserStore, *model.User, func()) {
				store, mock, teardown := newMockedUserStore()
				mock.ExpectBegin()
				mock.ExpectExec("UPDATE \"users\"").WillReturnError(gorm.ErrInvalidTransaction)
				mock.ExpectCommit()
				user := &model.User{Model: gorm.Model{ID: 1}, Username: "dbErrorUser", Email: "dberror@example.com"}
				return store, user, teardown
			},
			expectedErr: gorm.ErrInvalidTransaction,
			validate: func(t *testing.T, store *UserStore, user *model.User) {

			},
		},
		{
			name: "Concurrent Updates",
			setup: func() (*UserStore, *model.User, func()) {
				store, mock, teardown := newMockedUserStore()
				mock.ExpectBegin()
				mock.ExpectExec("UPDATE \"users\"").WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
				user := &model.User{Model: gorm.Model{ID: 1}, Username: "concurrentUser", Email: "concurrent@example.com"}
				return store, user, teardown
			},
			expectedErr: nil,
			validate: func(t *testing.T, store *UserStore, user *model.User) {
				var wg sync.WaitGroup
				numUpdates := 5
				wg.Add(numUpdates)
				for i := 0; i < numUpdates; i++ {
					go func() {
						defer wg.Done()
						err := store.Update(user)
						if err != nil {
							t.Errorf("error updating user concurrently: %v", err)
						}
					}()
				}
				wg.Wait()
				var finalUser model.User
				err := store.db.First(&finalUser, user.ID).Error
				if err != nil {
					t.Errorf("error fetching final user: %v", err)
				}
				if !reflect.DeepEqual(finalUser, *user) {
					t.Errorf("final user state is inconsistent, got %v, want %v", finalUser, *user)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Logf("Panic encountered so failing test. %v\n%s", r, string(debug.Stack()))
					t.Fail()
				}
			}()

			store, user, teardown := tt.setup()
			defer teardown()

			err := store.Update(user)
			if err != tt.expectedErr {
				t.Errorf("expected error %v, but got %v", tt.expectedErr, err)
			}

			tt.validate(t, store, user)
		})
	}
}

/*
ROOST_METHOD_HASH=UserStore_Follow_fe0976e4eb
ROOST_METHOD_SIG_HASH=UserStore_Follow_0e703b23f8

FUNCTION_DEF=func (s *UserStore) Follow(a *model.User, b *model.User) error // Follow create follow relashionship to User B from user A
*/
func TestUserStoreFollow(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	gormDB, err := gorm.Open("sqlite3", db)
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}

	store := &UserStore{db: gormDB}

	tests := []struct {
		name    string
		setup   func()
		userA   *model.User
		userB   *model.User
		wantErr bool
		errMsg  string
	}{
		{
			name: "Successful Follow Relationship Creation",
			setup: func() {
				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO \"users\"").WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectExec("INSERT INTO \"user_follows\"").WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
			userA:   &model.User{Username: "userA"},
			userB:   &model.User{Username: "userB"},
			wantErr: false,
		},
		{
			name: "Follow Relationship Already Exists",
			setup: func() {
				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO \"users\"").WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectExec("INSERT INTO \"user_follows\"").WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
			userA:   &model.User{Username: "userA"},
			userB:   &model.User{Username: "userB"},
			wantErr: false,
		},
		{
			name: "Follow Relationship with Non-Existent User",
			setup: func() {
				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO \"users\"").WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
			userA:   &model.User{Username: "userA"},
			userB:   &model.User{Username: "userB"},
			wantErr: true,
			errMsg:  "record not found",
		},
		{
			name: "Follow Relationship with Nil User",
			setup: func() {
				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO \"users\"").WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
			userA:   &model.User{Username: "userA"},
			userB:   nil,
			wantErr: true,
			errMsg:  "record not found",
		},
		{
			name: "Follow Relationship with Self",
			setup: func() {
				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO \"users\"").WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
			userA:   &model.User{Username: "userA"},
			userB:   &model.User{Username: "userA"},
			wantErr: true,
			errMsg:  "invalid association",
		},
		{
			name: "Database Error During Follow Relationship Creation",
			setup: func() {
				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO \"users\"").WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectExec("INSERT INTO \"user_follows\"").WillReturnError(errors.New("database error"))
				mock.ExpectRollback()
			},
			userA:   &model.User{Username: "userA"},
			userB:   &model.User{Username: "userB"},
			wantErr: true,
			errMsg:  "database error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Logf("Panic encountered so failing test. %v", r)
					t.Fail()
				}
			}()

			tt.setup()

			err := store.Follow(tt.userA, tt.userB)
			if (err != nil) != tt.wantErr {
				t.Errorf("UserStore.Follow() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil && tt.errMsg != "" && !errors.Is(err, errors.New(tt.errMsg)) {
				t.Errorf("UserStore.Follow() error message = %v, wantErrMsg %v", err, tt.errMsg)
			}
			t.Logf("Test %s passed successfully", tt.name)
		})
	}
}

/*
ROOST_METHOD_HASH=UserStore_Unfollow_29d3ef7f50
ROOST_METHOD_SIG_HASH=UserStore_Unfollow_31d9214353

FUNCTION_DEF=func (s *UserStore) Unfollow(a *model.User, b *model.User) error // Unfollow delete follow relashionship to User B from user A
*/
func TestUserStoreUnfollow(t *testing.T) {

	scenarios := []struct {
		name        string
		setup       func() (*UserStore, *model.User, *model.User, error)
		expectError bool
	}{
		{
			name: "Successful Unfollow",
			setup: func() (*UserStore, *model.User, *model.User, error) {
				db, mock, err := sqlmock.New()
				if err != nil {
					t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
				}
				mock.ExpectBegin()
				mock.ExpectExec("DELETE FROM.*").WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()

				gormDB, err := gorm.Open("postgres", db)
				if err != nil {
					t.Fatalf("failed to open gorm db: %v", err)
				}
				userStore := &UserStore{db: gormDB}
				userA := &model.User{Model: gorm.Model{ID: 1}}
				userB := &model.User{Model: gorm.Model{ID: 2}}

				userStore.db.Create(userA)
				userStore.db.Create(userB)
				userStore.db.Model(userA).Association("Follows").Append(userB)

				return userStore, userA, userB, nil
			},
			expectError: false,
		},
		{
			name: "Unfollow Non-Existent User",
			setup: func() (*UserStore, *model.User, *model.User, error) {
				db, mock, err := sqlmock.New()
				if err != nil {
					t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
				}
				mock.ExpectBegin()
				mock.ExpectExec("DELETE FROM.*").WillReturnError(errors.New("user does not exist"))
				mock.ExpectCommit()

				gormDB, err := gorm.Open("postgres", db)
				if err != nil {
					t.Fatalf("failed to open gorm db: %v", err)
				}
				userStore := &UserStore{db: gormDB}
				userA := &model.User{Model: gorm.Model{ID: 1}}
				userB := &model.User{Model: gorm.Model{ID: 2}}

				userStore.db.Create(userA)

				return userStore, userA, userB, nil
			},
			expectError: true,
		},
		{
			name: "Unfollow Self",
			setup: func() (*UserStore, *model.User, *model.User, error) {
				db, _, err := sqlmock.New()
				if err != nil {
					t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
				}

				gormDB, err := gorm.Open("postgres", db)
				if err != nil {
					t.Fatalf("failed to open gorm db: %v", err)
				}
				userStore := &UserStore{db: gormDB}
				userA := &model.User{Model: gorm.Model{ID: 1}}

				return userStore, userA, userA, nil
			},
			expectError: true,
		},
		{
			name: "Unfollow Already Unfollowed User",
			setup: func() (*UserStore, *model.User, *model.User, error) {
				db, mock, err := sqlmock.New()
				if err != nil {
					t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
				}
				mock.ExpectBegin()
				mock.ExpectExec("DELETE FROM.*").WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()

				gormDB, err := gorm.Open("postgres", db)
				if err != nil {
					t.Fatalf("failed to open gorm db: %v", err)
				}
				userStore := &UserStore{db: gormDB}
				userA := &model.User{Model: gorm.Model{ID: 1}}
				userB := &model.User{Model: gorm.Model{ID: 2}}

				userStore.db.Create(userA)
				userStore.db.Create(userB)

				return userStore, userA, userB, nil
			},
			expectError: false,
		},
		{
			name: "Database Error",
			setup: func() (*UserStore, *model.User, *model.User, error) {
				db, mock, err := sqlmock.New()
				if err != nil {
					t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
				}
				mock.ExpectBegin()
				mock.ExpectExec("DELETE FROM.*").WillReturnError(errors.New("database error"))
				mock.ExpectCommit()

				gormDB, err := gorm.Open("postgres", db)
				if err != nil {
					t.Fatalf("failed to open gorm db: %v", err)
				}
				userStore := &UserStore{db: gormDB}
				userA := &model.User{Model: gorm.Model{ID: 1}}
				userB := &model.User{Model: gorm.Model{ID: 2}}

				userStore.db.Create(userA)
				userStore.db.Create(userB)
				userStore.db.Model(userA).Association("Follows").Append(userB)

				return userStore, userA, userB, nil
			},
			expectError: true,
		},
	}

	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {

			defer func() {
				if r := recover(); r != nil {
					t.Logf("Panic encountered so failing test. %v\n%s", r, string(debug.Stack()))
					t.Fail()
				}
			}()

			userStore, userA, userB, setupErr := scenario.setup()
			if setupErr != nil {
				t.Fatalf("Setup failed: %v", setupErr)
			}

			err := userStore.Unfollow(userA, userB)
			if scenario.expectError {
				assert.Error(t, err, "Expected error but got nil")
				t.Logf("Successfully tested scenario: %s", scenario.name)
			} else {
				assert.NoError(t, err, "Expected no error but got: %v", err)
				t.Logf("Successfully tested scenario: %s", scenario.name)
			}
		})
	}
}
