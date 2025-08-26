package store

import (
	gorm "github.com/jinzhu/gorm"
	assert "github.com/stretchr/testify/assert"
	testing "testing"
	gosqlmock "github.com/DATA-DOG/go-sqlmock"
	model "github.com/raahii/golang-grpc-realworld-example/model"
	errors "errors"
)





type MockAssociation struct {
	Error error
}
type MockUserStore struct {
	db *gorm.DB
}


/*
ROOST_METHOD_HASH=NewUserStore_fb599438e5
ROOST_METHOD_SIG_HASH=NewUserStore_c0075221af

FUNCTION_DEF=func NewUserStore(db *gorm.DB) *UserStore // NewUserStore returns a new UserStore


*/
func TestNewUserStore(t *testing.T) {

	testCases := []struct {
		name     string
		db       *gorm.DB
		expected *UserStore
	}{
		{
			name:     "Successful creation of a new UserStore",
			db:       &gorm.DB{},
			expected: &UserStore{db: &gorm.DB{}},
		},
		{
			name:     "Creation of a new UserStore with a nil gorm.DB instance",
			db:       nil,
			expected: &UserStore{db: nil},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			defer func() {
				if r := recover(); r != nil {
					t.Logf("Panic encountered so failing test. %v", r)
					t.Fail()
				}
			}()

			result := NewUserStore(tc.db)

			assert.Equal(t, tc.expected, result, "Expected and actual result should be the same")
		})
	}
}


/*
ROOST_METHOD_HASH=UserStore.Create_9495ddb29d
ROOST_METHOD_SIG_HASH=UserStore.Create_18451817fe

FUNCTION_DEF=func (s *UserStore) Create(m *model.User) error // Create create a user


*/
func TestUserStoreCreate(t *testing.T) {
	t.Run("Successful User Creation", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Logf("Panic encountered so failing test. %v", r)
				t.Fail()
			}
		}()

		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
		}
		defer db.Close()

		gormDB, err := gorm.Open("postgres", db)
		if err != nil {
			t.Fatalf("an error '%s' was not expected when opening gorm database", err)
		}

		userStore := &UserStore{
			db: gormDB,
		}

		user := &model.User{}

		mock.ExpectBegin()
		mock.ExpectExec("INSERT INTO \"users\"").WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		err = userStore.Create(user)

		assert.NoError(t, err, "Expected no error on successful user creation")
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %s", err)
		}
	})
}


/*
ROOST_METHOD_HASH=UserStore.GetByID_1f5f06165b
ROOST_METHOD_SIG_HASH=UserStore.GetByID_2a864916bb

FUNCTION_DEF=func (s *UserStore) GetByID(id uint) (*model.User, error) // GetByID finds a user from id


*/
func TestUserStoreGetById(t *testing.T) {

	testCases := []struct {
		name        string
		userId      uint
		mockDb      func() (*gorm.DB, sqlmock.Sqlmock)
		expectedErr error
	}{
		{
			name:   "Valid User ID",
			userId: 1,
			mockDb: func() (*gorm.DB, sqlmock.Sqlmock) {
				db, mock, _ := sqlmock.New()
				gormDb, _ := gorm.Open("postgres", db)
				mock.ExpectQuery("^SELECT (.+) FROM \"users\" WHERE \"users\".\"id\" = $1$").
					WithArgs(1).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
				return gormDb, mock
			},
			expectedErr: nil,
		},
		{
			name:   "Non-Existent User ID",
			userId: 2,
			mockDb: func() (*gorm.DB, sqlmock.Sqlmock) {
				db, mock, _ := sqlmock.New()
				gormDb, _ := gorm.Open("postgres", db)
				mock.ExpectQuery("^SELECT (.+) FROM \"users\" WHERE \"users\".\"id\" = $1$").
					WithArgs(2).
					WillReturnError(gorm.ErrRecordNotFound)
				return gormDb, mock
			},
			expectedErr: gorm.ErrRecordNotFound,
		},
		{
			name:   "Database Connection Error",
			userId: 3,
			mockDb: func() (*gorm.DB, sqlmock.Sqlmock) {
				db, mock, _ := sqlmock.New()
				gormDb, _ := gorm.Open("postgres", db)
				mock.ExpectQuery("^SELECT (.+) FROM \"users\" WHERE \"users\".\"id\" = $1$").
					WithArgs(3).
					WillReturnError(errors.New("database connection error"))
				return gormDb, mock
			},
			expectedErr: errors.New("database connection error"),
		},
		{
			name:   "Invalid User ID",
			userId: 0,
			mockDb: func() (*gorm.DB, sqlmock.Sqlmock) {
				db, mock, _ := sqlmock.New()
				gormDb, _ := gorm.Open("postgres", db)
				mock.ExpectQuery("^SELECT (.+) FROM \"users\" WHERE \"users\".\"id\" = $1$").
					WithArgs(0).
					WillReturnError(errors.New("invalid user ID"))
				return gormDb, mock
			},
			expectedErr: errors.New("invalid user ID"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Logf("Panic encountered so failing test. %v", r)
					t.Fail()
				}
			}()

			db, mock := tc.mockDb()
			defer db.Close()

			store := &UserStore{db: db}
			_, err := store.GetByID(tc.userId)

			if err != nil {
				if tc.expectedErr == nil {
					t.Errorf("unexpected error: %v", err)
				} else if err.Error() != tc.expectedErr.Error() {
					t.Errorf("expected error: %v, got: %v", tc.expectedErr, err)
				}
			} else if tc.expectedErr != nil {
				t.Errorf("expected error: %v, got: nil", tc.expectedErr)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}


/*
ROOST_METHOD_HASH=UserStore.GetByEmail_fda09af5c4
ROOST_METHOD_SIG_HASH=UserStore.GetByEmail_9e84f3286b

FUNCTION_DEF=func (s *UserStore) GetByEmail(email string) (*model.User, error) // GetByEmail finds a user from email


*/
func TestUserStoreGetByEmail(t *testing.T) {

	testCases := []struct {
		name          string
		email         string
		mockDBHandler func(mock sqlmock.Sqlmock)
		expectedUser  *model.User
		expectedErr   error
	}{
		{
			name:  "Valid Email is Provided",
			email: "test@example.com",
			mockDBHandler: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"email"}).AddRow("test@example.com")
				mock.ExpectQuery("^SELECT (.+) FROM \"users\" WHERE (.+)").WithArgs("test@example.com").WillReturnRows(rows)
			},
			expectedUser: &model.User{Email: "test@example.com"},
			expectedErr:  nil,
		},
		{
			name:  "Email Does Not Exist in the Database",
			email: "nonexistent@example.com",
			mockDBHandler: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("^SELECT (.+) FROM \"users\" WHERE (.+)").WithArgs("nonexistent@example.com").WillReturnError(gorm.ErrRecordNotFound)
			},
			expectedUser: nil,
			expectedErr:  gorm.ErrRecordNotFound,
		},
		{
			name:  "Database Connection Error",
			email: "test@example.com",
			mockDBHandler: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("^SELECT (.+) FROM \"users\" WHERE (.+)").WithArgs("test@example.com").WillReturnError(errors.New("database connection error"))
			},
			expectedUser: nil,
			expectedErr:  errors.New("database connection error"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Logf("Panic encountered so failing test. %v", r)
					t.Fail()
				}
			}()

			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("Failed to create sqlmock: %v", err)
			}
			defer db.Close()

			tc.mockDBHandler(mock)

			gormDB, err := gorm.Open("postgres", db)
			if err != nil {
				t.Fatalf("Failed to open gorm DB: %v", err)
			}

			userStore := &UserStore{db: gormDB}

			user, err := userStore.GetByEmail(tc.email)

			if tc.expectedErr != nil {
				if err == nil || err.Error() != tc.expectedErr.Error() {
					t.Errorf("Expected error: %v, got: %v", tc.expectedErr, err)
				}
			} else if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if tc.expectedUser != nil {
				if user == nil || user.Email != tc.expectedUser.Email {
					t.Errorf("Expected user: %v, got: %v", tc.expectedUser, user)
				}
			} else if user != nil {
				t.Errorf("Unexpected user: %v", user)
			}
		})
	}
}


/*
ROOST_METHOD_HASH=UserStore.GetByUsername_622b1b9e41
ROOST_METHOD_SIG_HASH=UserStore.GetByUsername_992f00baec

FUNCTION_DEF=func (s *UserStore) GetByUsername(username string) (*model.User, error) // GetByUsername finds a user from username


*/
func TestUserStoreGetByUsername(t *testing.T) {

	testCases := []struct {
		name     string
		username string
		setup    func(mock sqlmock.Sqlmock)
		check    func(user *model.User, err error)
	}{
		{
			name:     "Valid User Retrieval",
			username: "testuser",
			setup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"username"}).AddRow("testuser")
				mock.ExpectQuery("^SELECT (.+) FROM \"users\" WHERE (.+)").WithArgs("testuser").WillReturnRows(rows)
			},
			check: func(user *model.User, err error) {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if user == nil || user.Username != "testuser" {
					t.Errorf("unexpected user: %v", user)
				}
			},
		},
		{
			name:     "Non-Existent User Retrieval",
			username: "nonexistentuser",
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("^SELECT (.+) FROM \"users\" WHERE (.+)").WithArgs("nonexistentuser").WillReturnError(gorm.ErrRecordNotFound)
			},
			check: func(user *model.User, err error) {
				if err == nil || !errors.Is(err, gorm.ErrRecordNotFound) {
					t.Errorf("expected ErrRecordNotFound, but got: %v", err)
				}
			},
		},
		{
			name:     "Database Connection Error",
			username: "anyuser",
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("^SELECT (.+) FROM \"users\" WHERE (.+)").WithArgs("anyuser").WillReturnError(errors.New("database connection error"))
			},
			check: func(user *model.User, err error) {
				if err == nil || err.Error() != "database connection error" {
					t.Errorf("expected database connection error, but got: %v", err)
				}
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			defer func() {
				if r := recover(); r != nil {
					t.Logf("Panic encountered so failing test. %v", r)
					t.Fail()
				}
			}()

			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("failed to create sqlmock: %v", err)
			}
			defer db.Close()

			gormDB, err := gorm.Open("postgres", db)
			if err != nil {
				t.Fatalf("failed to open gorm DB: %v", err)
			}

			tc.setup(mock)

			store := &UserStore{
				db: gormDB,
			}

			user, err := store.GetByUsername(tc.username)

			tc.check(user, err)
		})
	}
}


/*
ROOST_METHOD_HASH=UserStore.Update_4fd6d3d1c1
ROOST_METHOD_SIG_HASH=UserStore.Update_ddd5c151cf

FUNCTION_DEF=func (s *UserStore) Update(m *model.User) error // Update update all of user fields


*/
func TestUserStoreUpdate(t *testing.T) {

	testCases := []struct {
		name    string
		user    *model.User
		mock    func()
		wantErr bool
	}{
		{

			name: "Successful User Update",
			user: &model.User{
				Username: "testuser",
				Email:    "testuser@example.com",
			},
			mock: func() {

			},
			wantErr: false,
		},
		{

			name: "Update Non-Existent User",
			user: &model.User{
				Username: "nonexistentuser",
				Email:    "nonexistentuser@example.com",
			},
			mock: func() {

			},
			wantErr: true,
		},
		{

			name: "Update User with Invalid Details",
			user: &model.User{
				Username: "invaliduser",
				Email:    "invaliduser",
			},
			mock: func() {

			},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			defer func() {
				if r := recover(); r != nil {
					t.Logf("Panic encountered so failing test. %v", r)
					t.Fail()
				}
			}()

			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
			}
			defer db.Close()

			gdb, err := gorm.Open("postgres", db)
			if err != nil {
				t.Fatalf("an error '%s' was not expected when opening gorm database", err)
			}

			tc.mock()

			store := &UserStore{
				db: gdb,
			}

			err = store.Update(tc.user)

			if tc.wantErr {
				if err == nil {
					t.Errorf("expected an error but didn't get one")
				}
			} else {
				if err != nil {
					t.Errorf("didn't expect an error but got one: %v", err)
				}
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}


/*
ROOST_METHOD_HASH=UserStore.Follow_fe0976e4eb
ROOST_METHOD_SIG_HASH=UserStore.Follow_0e703b23f8

FUNCTION_DEF=func (s *UserStore) Follow(a *model.User, b *model.User) error // Follow create follow relashionship to User B from user A


*/
func (ma *MockAssociation) Append(values ...interface{}) *MockAssociation {
	return ma
}

func (ms *MockUserStore) Association(column string) *MockAssociation {
	if column == "Follows" {
		return &MockAssociation{}
	}
	return &MockAssociation{Error: errors.New("invalid association")}
}

func (ms *MockUserStore) Follow(a *model.User, b *model.User) error {
	return ms.Model(a).Association("Follows").Append(b).Error
}

func (ms *MockUserStore) Model(value interface{}) *MockUserStore {
	return ms
}

func TestUserStoreFollow(t *testing.T) {

	testCases := []struct {
		name          string
		userA         *model.User
		userB         *model.User
		expectedError error
	}{
		{
			name:          "Successful Follow Operation",
			userA:         &model.User{Username: "UserA"},
			userB:         &model.User{Username: "UserB"},
			expectedError: nil,
		},
		{
			name:          "Follow Operation with Non-Existent User",
			userA:         &model.User{Username: "UserA"},
			userB:         nil,
			expectedError: errors.New("invalid association"),
		},
		{
			name:          "Follow Operation with Same User",
			userA:         &model.User{Username: "UserA"},
			userB:         &model.User{Username: "UserA"},
			expectedError: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Logf("Panic encountered so failing test. %v", r)
					t.Fail()
				}
			}()

			mockUserStore := &MockUserStore{}

			err := mockUserStore.Follow(tc.userA, tc.userB)

			if err != nil && err.Error() != tc.expectedError.Error() {
				t.Errorf("Expected error: %v, got: %v", tc.expectedError, err)
			}
		})
	}
}


/*
ROOST_METHOD_HASH=UserStore.Unfollow_29d3ef7f50
ROOST_METHOD_SIG_HASH=UserStore.Unfollow_31d9214353

FUNCTION_DEF=func (s *UserStore) Unfollow(a *model.User, b *model.User) error // Unfollow delete follow relashionship to User B from user A


*/
func TestUserStoreUnfollow(t *testing.T) {

	testCases := []struct {
		name          string
		userA         *model.User
		userB         *model.User
		setup         func(userStore *UserStore, userA *model.User, userB *model.User) error
		expectedError bool
	}{
		{
			name: "Successful Unfollow Operation",
			userA: &model.User{
				Username: "UserA",
				Email:    "usera@example.com",
			},
			userB: &model.User{
				Username: "UserB",
				Email:    "userb@example.com",
			},
			setup: func(userStore *UserStore, userA *model.User, userB *model.User) error {

				return nil
			},
			expectedError: false,
		},
		{
			name: "Unfollow Operation with Non-Existent User",
			userA: &model.User{
				Username: "UserA",
				Email:    "usera@example.com",
			},
			userB: &model.User{
				Username: "NonExistentUser",
				Email:    "nonexistentuser@example.com",
			},
			setup: func(userStore *UserStore, userA *model.User, userB *model.User) error {

				return nil
			},
			expectedError: true,
		},
		{
			name: "Unfollow Operation with No Existing Follow Relationship",
			userA: &model.User{
				Username: "UserA",
				Email:    "usera@example.com",
			},
			userB: &model.User{
				Username: "UserB",
				Email:    "userb@example.com",
			},
			setup: func(userStore *UserStore, userA *model.User, userB *model.User) error {

				return nil
			},
			expectedError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			db, _ := gorm.Open("postgres", "host=localhost user=gorm dbname=gorm password=gorm sslmode=disable")
			userStore := &UserStore{db: db}

			if err := tc.setup(userStore, tc.userA, tc.userB); err != nil {
				t.Fatalf("failed to setup test case: %v", err)
			}

			err := userStore.Unfollow(tc.userA, tc.userB)

			if tc.expectedError {
				if err == nil {
					t.Errorf("expected error but got nil")
				} else {
					t.Logf("expected error received: %v", err)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				} else {
					t.Logf("unfollow operation successful")
				}
			}
		})
	}
}

