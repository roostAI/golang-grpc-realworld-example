package store

import (
	testing "testing"
	reflect "reflect"
	gosqlmock "github.com/DATA-DOG/go-sqlmock"
	gorm "github.com/jinzhu/gorm"
	model "github.com/raahii/golang-grpc-realworld-example/model"
	require "github.com/stretchr/testify/require"
	fmt "fmt"
	debug "runtime/debug"
	assert "github.com/stretchr/testify/assert"
	regexp "regexp"
	errors "errors"
	_ "github.com/jinzhu/gorm/dialects/mysql"
)



var mockGormDB, _ = gorm.Open("mysql", mockDB)
var mockDB, _, _  = gosqlmock.New()
var tests = []test{
	{
		name: "Successful Unfollow Operation",
		user: &UserStore{db: mockGormDB},
		args: arg{
			a: &model.User{Username: "UserA"},
			b: &model.User{Username: "UserB"},
		},
		wantErr: false,
	},
	{
		name: "Unfollow a Non-following User",
		user: &UserStore{db: mockGormDB},
		args: arg{
			a: &model.User{Username: "UserA"},
			b: &model.User{Username: "UserC"},
		},
		wantErr: true,
	},
	{
		name: "Invalid User",
		user: &UserStore{db: mockGormDB},
		args: arg{
			a: &model.User{Username: "UserA"},
			b: &model.User{Username: "UserD"},
		},
		wantErr: true,
	},
	{
		name: "Database Connection Failure",
		user: &UserStore{},
		args: arg{
			a: &model.User{Username: "UserA"},
			b: &model.User{Username: "UserB"},
		},
		wantErr: true,
	},
}

type testdata struct {
	name     string
	db       *gorm.DB
	expected *UserStore
}
type MockDatabase struct {
	DB *gorm.DB
}
type MockUserStore struct {
	db *gorm.DB
	Fn func(a *model.User, b *model.User) error
}
type arg struct {
	a *model.User
	b *model.User
}
type test struct {
	name    string
	user    *UserStore
	args    arg
	wantErr bool
}


/*
ROOST_METHOD_HASH=NewUserStore_fb599438e5
ROOST_METHOD_SIG_HASH=NewUserStore_c0075221af

FUNCTION_DEF=func NewUserStore(db *gorm.DB) *UserStore // NewUserStore returns a new UserStore


*/
func TestNewUserStore(t *testing.T) {

	tests := []testdata{

		{
			"Observable DB",
			func() *gorm.DB {

				sqlDB, mock, _ := sqlmock.New()
				gormDB, _ := gorm.Open("mysql", sqlDB)
				mock.ExpectQuery("^SELECT (.+)").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

				return gormDB
			}(),
			&UserStore{db: func() *gorm.DB {

				sqlDB, mock, _ := sqlmock.New()
				gormDB, _ := gorm.Open("mysql", sqlDB)
				mock.ExpectQuery("^SELECT (.+)").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

				return gormDB
			}()},
		},
		{
			"Nil DB",
			nil,
			&UserStore{db: nil},
		},
		{
			"Shared DB",
			func() *gorm.DB {

				sqlDB, mock, _ := sqlmock.New()
				gormDB, _ := gorm.Open("mysql", sqlDB)
				mock.ExpectQuery("^SELECT (.+)").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

				return gormDB
			}(),
			&UserStore{db: func() *gorm.DB {

				sqlDB, mock, _ := sqlmock.New()
				gormDB, _ := gorm.Open("mysql", sqlDB)
				mock.ExpectQuery("^SELECT (.+)").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

				return gormDB
			}()},
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

			actual := NewUserStore(tt.db)

			if !reflect.DeepEqual(actual, tt.expected) {
				t.Errorf("TestNewUserStore(%s): expected %v, got %v", tt.name, tt.expected, actual)
			}
		})
	}

}


/*
ROOST_METHOD_HASH=UserStore_Create_9495ddb29d
ROOST_METHOD_SIG_HASH=UserStore_Create_18451817fe

FUNCTION_DEF=func (s *UserStore) Create(m *model.User) error // Create create a user


*/
func (mockDB *MockDatabase) Create(m *model.User) error {
	return mockDB.DB.Create(m).Error
}

func TestUserStoreCreate(t *testing.T) {
	cases := []struct {
		name    string
		user    *model.User
		mockDb  *MockDatabase
		wantErr bool
	}{
		{
			name: "UserStore.Create successfully creates a user",
			user: &model.User{},
			mockDb: func() *MockDatabase {
				sqlDb, mock, _ := sqlmock.New()
				db, _ := gorm.Open("postgres", sqlDb)
				db.LogMode(false)
				mock.ExpectExec("INSERT INTO").WillReturnResult(sqlmock.NewResult(1, 1))

				return &MockDatabase{
					DB: db,
				}
			}(),
			wantErr: false,
		},
		{
			name: "UserStore.Create fails when fields are not properly set",
			user: &model.User{},
			mockDb: func() *MockDatabase {
				sqlDb, mock, _ := sqlmock.New()
				db, _ := gorm.Open("postgres", sqlDb)
				db.LogMode(false)
				mock.ExpectExec("INSERT INTO").WillReturnError(gorm.ErrRecordNotFound)

				return &MockDatabase{
					DB: db,
				}
			}(),
			wantErr: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			userStore := &UserStore{
				db: tc.mockDb.DB,
			}
			err := userStore.Create(tc.user)
			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}


/*
ROOST_METHOD_HASH=UserStore_GetByID_1f5f06165b
ROOST_METHOD_SIG_HASH=UserStore_GetByID_2a864916bb

FUNCTION_DEF=func (s *UserStore) GetByID(id uint) (*model.User, error) // GetByID finds a user from id


*/
func TestUserStoreGetById(t *testing.T) {
	var testCases = []struct {
		name            string
		userId          uint
		expectedUser    *model.User
		expectedError   error
		mockExpectation func(mock sqlmock.Sqlmock, user *model.User)
	}{
		{
			name:          "Existing User Data Retrieval Test",
			userId:        1,
			expectedUser:  &model.User{Model: gorm.Model{ID: uint(1)}, Username: "Test User"},
			expectedError: nil,
			mockExpectation: func(mock sqlmock.Sqlmock, user *model.User) {
				rows := sqlmock.NewRows([]string{"id", "username"}).
					AddRow(user.Model.ID, user.Username)
				mock.ExpectQuery("^SELECT").WillReturnRows(rows)
			},
		},
		{
			name:          "Non-Existent User Data Retrieval Test",
			userId:        2,
			expectedUser:  nil,
			expectedError: gorm.ErrRecordNotFound,
			mockExpectation: func(mock sqlmock.Sqlmock, user *model.User) {
				mock.ExpectQuery("^SELECT").WillReturnError(gorm.ErrRecordNotFound)
			},
		},
		{
			name:          "Edge Case for User Data Retrieval with Lowest Possible ID",
			userId:        0,
			expectedUser:  nil,
			expectedError: gorm.ErrRecordNotFound,
			mockExpectation: func(mock sqlmock.Sqlmock, user *model.User) {
				mock.ExpectQuery("^SELECT").WillReturnError(gorm.ErrRecordNotFound)
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Logf("Panic encountered so failing test. %v\n%s", r, string(debug.Stack()))
					t.Fail()
				}
			}()

			db, mock, _ := sqlmock.New()
			defer db.Close()
			gormDb, _ := gorm.Open("postgres", db)
			testCase.mockExpectation(mock, testCase.expectedUser)
			userStore := &UserStore{db: gormDb}

			user, err := userStore.GetByID(testCase.userId)
			if testCase.expectedError != nil {
				assert.Equal(t, testCase.expectedError, err)
				t.Log(fmt.Sprintf("%s: Error expected: %v, Error received: %v", testCase.name, testCase.expectedError, err))
			} else {
				assert.Equal(t, testCase.expectedUser, user)
				t.Log(fmt.Sprintf("%s: Data expected: %v, Data received: %v", testCase.name, testCase.expectedUser, user))
			}
		})
	}
}


/*
ROOST_METHOD_HASH=UserStore_GetByEmail_fda09af5c4
ROOST_METHOD_SIG_HASH=UserStore_GetByEmail_9e84f3286b

FUNCTION_DEF=func (s *UserStore) GetByEmail(email string) (*model.User, error) // GetByEmail finds a user from email


*/
func TestUserStoreGetByEmail(t *testing.T) {

	tests := []struct {
		name      string
		isError   bool
		setupMock func(mock sqlmock.Sqlmock, email string)
	}{
		{
			"Valid User Account Retrieval",
			false,
			func(mock sqlmock.Sqlmock, email string) {
				rows := sqlmock.NewRows([]string{"id", "email"}).
					AddRow("1", "test@example.com")
				mock.ExpectQuery("SELECT \\* FROM").WithArgs(email).WillReturnRows(rows)
			},
		},
		{
			"Non-existent User Account Retrieval",
			true,
			func(mock sqlmock.Sqlmock, email string) {
				mock.ExpectQuery("SELECT \\* FROM").WithArgs(email).WillReturnError(gorm.ErrRecordNotFound)
			},
		},
		{
			"Invalid email",
			true,
			func(mock sqlmock.Sqlmock, email string) {
				mock.ExpectQuery("SELECT \\* FROM").WithArgs(email).WillReturnError(gorm.ErrInvalidSQL)
			},
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

			db, mock, _ := sqlmock.New()
			dbInstance, _ := gorm.Open("mysql", db)
			defer dbInstance.Close()
			userStore := UserStore{db: dbInstance}

			tt.setupMock(mock, "test@example.com")

			res, err := userStore.GetByEmail("test@example.com")

			if tt.isError {
				assert.Error(t, err)
				assert.Nil(t, res)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, res)
				assert.Equal(t, "test@example.com", res.Email)
			}

			err = mock.ExpectationsWereMet()
			if err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}


/*
ROOST_METHOD_HASH=UserStore_GetByUsername_622b1b9e41
ROOST_METHOD_SIG_HASH=UserStore_GetByUsername_992f00baec

FUNCTION_DEF=func (s *UserStore) GetByUsername(username string) (*model.User, error) // GetByUsername finds a user from username


*/
func TestUserStoreGetByUsername(t *testing.T) {

	var (
		existingUsername     = "existingUser"
		nonExistingUsername  = "nonExistingUser"
		noDataError          = gorm.ErrRecordNotFound
		dbError              = fmt.Errorf("db error")
		emptyUsername        = ""
		invalidUsernameError = fmt.Errorf("Username must be provided")
	)

	tests := map[string]struct {
		username         string
		mockReturnsError error
		expectedUser     *model.User
		expectedError    error
	}{
		"user exists in database": {
			username:         existingUsername,
			mockReturnsError: nil,
			expectedUser:     &model.User{Username: existingUsername},
			expectedError:    nil,
		},
		"user does not exist in database": {
			username:         nonExistingUsername,
			mockReturnsError: noDataError,
			expectedUser:     nil,
			expectedError:    noDataError,
		},
		"database retrieval error": {
			username:         existingUsername,
			mockReturnsError: dbError,
			expectedUser:     nil,
			expectedError:    dbError,
		},
		"empty username": {
			username:         emptyUsername,
			mockReturnsError: nil,
			expectedUser:     nil,
			expectedError:    invalidUsernameError,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Logf("Panic encountered so failing test. %v\n%s", r, string(debug.Stack()))
					t.Fail()
				}
			}()

			db, mock, _ := gosqlmock.New()
			defer db.Close()
			gormDB, _ := gorm.Open("postgres", db)

			userStore := &UserStore{db: gormDB}

			mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM \"users\" WHERE (username = ?) ORDER BY \"users\".\"id\" ASC LIMIT 1")).
				WithArgs(test.username).
				WillReturnError(test.mockReturnsError)

			user, err := userStore.GetByUsername(test.username)

			if err != nil && err != test.expectedError {
				t.Errorf("error was not as expected: got %v, want %v", err, test.expectedError)
			}
			if test.expectedUser != nil &&
				(user == nil || user.Username != test.expectedUser.Username) {
				t.Errorf("user was not as expected: got %v, want %v", user, test.expectedUser)
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

	type Scenario struct {
		name string
		user *model.User
		db   *gorm.DB
		err  error
	}

	scenarios := []Scenario{
		{
			name: "Scenario 1: Successful Update operation",
			user: &model.User{Username: "TestUser", Email: "TestUser@gmail.com"},
			db:   mockSuccessDB(),
			err:  nil,
		},
		{
			name: "Scenario 2: Failed Update operation due to nil User model",
			user: nil,
			db:   mockSuccessDB(),
			err:  errors.New("user model cannot be null"),
		},
		{
			name: "Scenario 3: Failed Update operation due to a faulty DB connection",
			user: &model.User{Username: "TestUser", Email: "TestUser@gmail.com"},
			db:   mockFailureDB(),
			err:  errors.New("db connection error"),
		},
		{
			name: "Scenario 4: Successful Update operation for a non-existing User",
			user: &model.User{Username: "NotExistingUser", Email: "NotExistingUser@gmail.com"},
			db:   mockSuccessDB(),
			err:  gorm.ErrRecordNotFound,
		},
	}

	for _, s := range scenarios {
		t.Run(s.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Fatalf("Test %s panicked: %v\n", s.name, r)
				}
			}()

			store := &UserStore{s.db}
			err := store.Update(s.user)

			if (err != nil && s.err == nil) ||
				(err == nil && s.err != nil) ||
				(err != nil && s.err != nil && err.Error() != s.err.Error()) {
				t.Errorf("Test %s failed, expected error: %v, got: %v\n", s.name, s.err, err)
			}
		})
	}
}

func mockFailureDB() *gorm.DB {
	return nil
}

func mockSuccessDB() *gorm.DB {
	sqlDB, _, _ := gosqlmock.New()
	db, _ := gorm.Open("postgres", sqlDB)
	return db
}


/*
ROOST_METHOD_HASH=UserStore_Follow_fe0976e4eb
ROOST_METHOD_SIG_HASH=UserStore_Follow_0e703b23f8

FUNCTION_DEF=func (s *UserStore) Follow(a *model.User, b *model.User) error // Follow create follow relashionship to User B from user A


*/
func (s *MockUserStore) Follow(a *model.User, b *model.User) error {
	return s.Fn(a, b)
}

func TestUserStoreFollow(t *testing.T) {
	testCases := []struct {
		name         string
		UserA        *model.User
		UserB        *model.User
		mockBehavior func(us *MockUserStore, a *model.User, b *model.User) error
		expectedErr  error
	}{
		{
			name: "Valid Follow Relationship Creation",
			UserA: &model.User{
				Username: "User A",
				Email:    "usera@example.com",
				Bio:      "",
				Image:    "",
				Follows:  nil,
				Password: "",
			},
			UserB: &model.User{
				Username: "User B",
				Email:    "userb@example.com",
				Bio:      "",
				Image:    "",
				Follows:  nil,
				Password: "",
			},
			mockBehavior: func(us *MockUserStore, a *model.User, b *model.User) error {
				return nil
			},
			expectedErr: nil,
		},
		{
			name: "Error when trying to follow a non-existing user",
			UserA: &model.User{
				Username: "User A",
				Email:    "usera@example.com",
				Bio:      "",
				Image:    "",
				Follows:  nil,
				Password: "",
			},
			UserB: &model.User{
				Username: "User B",
				Email:    "userb@example.com",
				Bio:      "",
				Image:    "",
				Follows:  nil,
				Password: "",
			},
			mockBehavior: func(us *MockUserStore, a *model.User, b *model.User) error {
				return gorm.ErrRecordNotFound
			},
			expectedErr: gorm.ErrRecordNotFound,
		},
		{
			name: "Error when a user tries to follow themselves",
			UserA: &model.User{
				Username: "User A",
				Email:    "usera@example.com",
				Bio:      "",
				Image:    "",
				Follows:  nil,
				Password: "",
			},
			UserB: &model.User{
				Username: "User A",
				Email:    "usera@example.com",
				Bio:      "",
				Image:    "",
				Follows:  nil,
				Password: "",
			},
			mockBehavior: func(us *MockUserStore, a *model.User, b *model.User) error {
				return errors.New(fmt.Sprintf("cannot follow self: %s", a.Username))
			},
			expectedErr: fmt.Errorf("cannot follow self: User A"),
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {

			mus := &MockUserStore{}
			mus.Fn = func(a *model.User, b *model.User) error {
				return tt.mockBehavior(mus, a, b)
			}

			err := mus.Follow(tt.UserA, tt.UserB)

			if tt.expectedErr != nil {
				if err.Error() != tt.expectedErr.Error() {
					t.Errorf("expected error %v, but got %v", tt.expectedErr, err)
				}
			} else if err != nil {
				t.Errorf("expected no error, but got %v", err)
			}
		})
	}
}


/*
ROOST_METHOD_HASH=UserStore_Unfollow_29d3ef7f50
ROOST_METHOD_SIG_HASH=UserStore_Unfollow_31d9214353

FUNCTION_DEF=func (s *UserStore) Unfollow(a *model.User, b *model.User) error // Unfollow delete follow relashionship to User B from user A


*/
func TestUserStoreUnfollow(t *testing.T) {
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Logf("Panic encountered so failing test. %v", r)
					t.Fail()
				}
			}()

			err := tt.user.Unfollow(tt.args.a, tt.args.b)

			if (err != nil) != tt.wantErr {
				t.Errorf("UserStore.Unfollow() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if (err != nil) && tt.wantErr {
				assert.Equal(t,
					err,
					errors.New(fmt.Sprintf("unfollow operation failed for user %s towards user %s", tt.args.a.Username, tt.args.b.Username)),
					fmt.Sprintf("UserStore.Unfollow() test case failed for %s", tt.name),
				)
			}
		})
	}
}


/*
ROOST_METHOD_HASH=UserStore_GetFollowingUserIDs_ee9c1008ff
ROOST_METHOD_SIG_HASH=UserStore_GetFollowingUserIDs_d7746035ec

FUNCTION_DEF=func (s *UserStore) GetFollowingUserIDs(m *model.User) ([ // GetFollowingUserIDs returns user ids current user follows
]uint, error) 

*/
func TestUserStoreGetFollowingUserIDs(t *testing.T) {
	scenarios := []struct {
		name    string
		known   bool
		follows bool
		fail    bool
	}{
		{"Getting IDs of users being followed by an existing user from the system", true, true, false},
		{"No User IDs being followed by an existing user", true, false, false},
		{"Attempt to fetch list for Non-existing User", false, true, true},
		{"Database connection failure", true, true, true},
	}

	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Logf("Panic encountered. %v\n", r)
					t.Fail()
				}
			}()

			mockDB, mock, _ := sqlmock.New()
			defer mockDB.Close()
			gdb, _ := gorm.Open("postgres", mockDB)

			userStore := &UserStore{db: gdb}

			m := &model.User{Model: gorm.Model{ID: 1}}

			if scenario.known {
				if scenario.fail {
					mock.ExpectQuery("^SELECT (.+) FROM \"follows\" WHERE (.+)$").WithArgs(m.ID).WillReturnError(fmt.Errorf("db connection error"))
				} else {
					rows := sqlmock.NewRows([]string{"to_user_id"})
					if scenario.follows {
						rows.AddRow(1).AddRow(2).AddRow(3)
					}
					mock.ExpectQuery("^SELECT (.+) FROM \"follows\" WHERE (.+)$").WithArgs(m.ID).WillReturnRows(rows)
				}
			} else {
				mock.ExpectQuery("^SELECT (.+) FROM \"follows\" WHERE (.+)$").WithArgs(m.ID).WillReturnError(gorm.ErrRecordNotFound)
			}

			res, err := userStore.GetFollowingUserIDs(m)

			if scenario.fail {
				if err == nil {
					t.Errorf("Expected error. Got nil error.")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %s", err.Error())
					return
				}
				if scenario.follows {
					if len(res) != 3 {
						t.Errorf("Expected 3 user ids, got %d", len(res))
					}
				} else {
					if len(res) != 0 {
						t.Errorf("Expected 0 user ids, got %d", len(res))
					}
				}
			}
		})
	}
}

