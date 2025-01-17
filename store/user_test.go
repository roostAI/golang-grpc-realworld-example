package store

import (
	"testing"
	"github.com/jinzhu/gorm"
	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/raahii/golang-grpc-realworld-example/model"
	"errors"
	"database/sql/driver"
	"time"
	"fmt"
	"log"
)






type TestData struct {
	name      string
	user      *model.User
	mockSetup func(mock sqlmock.Sqlmock)
	verify    func(err error, t *testing.T)
}


/*
ROOST_METHOD_HASH=NewUserStore_6a331dd890
ROOST_METHOD_SIG_HASH=NewUserStore_4f0c2dfca9

FUNCTION_DEF=func NewUserStore(db *gorm.DB) *UserStore 

 */
func TestNewUserStore(t *testing.T) {

	tests := []struct {
		name          string
		setupDB       func() *gorm.DB
		expectedDBNil bool
	}{
		{
			name: "Scenario 1: Successful UserStore Creation with a Valid DB",
			setupDB: func() *gorm.DB {
				db, _, _ := sqlmock.New()
				return &gorm.DB{Value: db}
			},
			expectedDBNil: false,
		},
		{
			name: "Scenario 2: Creation of UserStore with Nil DB",
			setupDB: func() *gorm.DB {
				return nil
			},
			expectedDBNil: true,
		},
		{
			name: "Scenario 3: Ensure UserStore Maintains Reference Integrity",
			setupDB: func() *gorm.DB {
				db, _, _ := sqlmock.New()
				return &gorm.DB{Value: db}
			},
			expectedDBNil: false,
		},
		{
			name: "Scenario 4: Behavior with Closed Database Connection",
			setupDB: func() *gorm.DB {
				db, mock, _ := sqlmock.New()
				mock.ExpectClose()
				gdb := &gorm.DB{Value: db}
				gdb.Close()
				return gdb
			},
			expectedDBNil: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := tt.setupDB()
			userStore := NewUserStore(db)

			t.Logf("Running %s", tt.name)

			if tt.expectedDBNil {
				assert.Nil(t, userStore.db, "expected db field to be nil")
			} else {
				assert.NotNil(t, userStore, "expected not nil UserStore")
				assert.Same(t, db, userStore.db, "expected the db reference to remain same")
			}
		})
	}
}


/*
ROOST_METHOD_HASH=Create_889fc0fc45
ROOST_METHOD_SIG_HASH=Create_4c48ec3920

FUNCTION_DEF=func (s *UserStore) Create(m *model.User) error 

 */
func TestUserStoreCreate(t *testing.T) {
	type testCase struct {
		name      string
		setup     func(sqlmock.Sqlmock)
		inputUser *model.User
		expectErr bool
	}

	tests := []testCase{
		{
			name: "Create a User Successfully",
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec(`INSERT INTO "users"`).
					WithArgs(sqlmock.AnyArg(), "testuser", "test@example.com", sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
			inputUser: &model.User{
				Username: "testuser",
				Email:    "test@example.com",
				Password: "securepassword",
				Bio:      "Test Bio",
				Image:    "image_url",
			},
			expectErr: false,
		},
		{
			name:  "Handle Invalid User Input",
			setup: func(mock sqlmock.Sqlmock) {},
			inputUser: &model.User{
				Username: "invaliduser",
				Password: "noemail",
			},
			expectErr: true,
		},
		{
			name: "Database Error Handling",
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin().WillReturnError(gorm.ErrInvalidTransaction)
			},
			inputUser: &model.User{
				Username: "testuser",
				Email:    "test@example.com",
				Password: "securepassword",
				Bio:      "Test Bio",
				Image:    "image_url",
			},
			expectErr: true,
		},
		{
			name: "Duplicate User Registration",
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec(`INSERT INTO "users"`).
					WithArgs(sqlmock.AnyArg(), "duplicateuser", "duplicate@example.com", sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
					WillReturnError(&gorm.Errors{gorm.ErrInvalidTransaction})
				mock.ExpectRollback()
			},
			inputUser: &model.User{
				Username: "duplicateuser",
				Email:    "duplicate@example.com",
				Password: "securepassword",
				Bio:      "Test Bio",
				Image:    "image_url",
			},
			expectErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {

			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
			}
			defer db.Close()

			gormDB, err := gorm.Open("postgres", db)
			if err != nil {
				t.Fatalf("failed to open GORM DB: %v", err)
			}
			defer gormDB.Close()

			userStore := &UserStore{db: gormDB}

			tc.setup(mock)

			err = userStore.Create(tc.inputUser)

			if tc.expectErr && err == nil {
				t.Errorf("expected error but got none")
			} else if !tc.expectErr && err != nil {
				t.Errorf("expected no error but got: %v", err)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
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
		t.Fatalf("An error occurred when opening a stub database connection: %s", err)
	}
	defer db.Close()

	gormDB, err := gorm.Open("mysql", db)
	if err != nil {
		t.Fatalf("An error occurred when setting up a GORM DB: %s", err)
	}

	userStore := &UserStore{db: gormDB}

	tests := []struct {
		name        string
		setupMock   func()
		id          uint
		expected    *model.User
		expectError bool
	}{
		{
			name: "Scenario 1: Successfully Retrieve User by ID",
			setupMock: func() {
				rows := sqlmock.NewRows([]string{"id", "username", "email"}).
					AddRow(1, "testuser", "test@example.com")
				mock.ExpectQuery("^SELECT (.+) FROM `users` WHERE (.+)$").WillReturnRows(rows)
			},
			id:          1,
			expected:    &model.User{Model: gorm.Model{ID: 1}, Username: "testuser", Email: "test@example.com"},
			expectError: false,
		},
		{
			name: "Scenario 2: User Not Found",
			setupMock: func() {
				mock.ExpectQuery("^SELECT (.+) FROM `users` WHERE (.+)$").WillReturnError(gorm.ErrRecordNotFound)
			},
			id:          2,
			expected:    nil,
			expectError: true,
		},
		{
			name: "Scenario 3: Database Error Encountered",
			setupMock: func() {
				mock.ExpectQuery("^SELECT (.+) FROM `users` WHERE (.+)$").WillReturnError(errors.New("database error"))
			},
			id:          3,
			expected:    nil,
			expectError: true,
		},
		{
			name: "Scenario 4: Retrieve User with Maximum Possible ID",
			setupMock: func() {
				rows := sqlmock.NewRows([]string{"id", "username", "email"}).
					AddRow(uint(^uint(0)>>1), "maxiduser", "maxiduser@example.com")
				mock.ExpectQuery("^SELECT (.+) FROM `users` WHERE (.+)$").WillReturnRows(rows)
			},
			id:          uint(^uint(0) >> 1),
			expected:    &model.User{Model: gorm.Model{ID: uint(^uint(0) >> 1)}, Username: "maxiduser", Email: "maxiduser@example.com"},
			expectError: false,
		},
		{
			name: "Scenario 5: Retrieve User with ID of '0'",
			setupMock: func() {
				mock.ExpectQuery("^SELECT (.+) FROM `users` WHERE (.+)$").WillReturnError(gorm.ErrRecordNotFound)
			},
			id:          0,
			expected:    nil,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			user, err := userStore.GetByID(tt.id)

			if (err != nil) != tt.expectError {
				t.Errorf("Got error = %v, expected error = %v", err != nil, tt.expectError)
			}

			if user == nil && tt.expected != nil || user != nil && tt.expected == nil ||
				(user != nil && tt.expected != nil && !compareUsers(user, tt.expected)) {
				t.Errorf("Got user = %+v, expected = %+v", user, tt.expected)
			}

			t.Logf("Executed test case: %s", tt.name)
		})
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("There were unfulfilled expectations: %s", err)
	}
}

func compareUsers(u1, u2 *model.User) bool {
	return u1.ID == u2.ID && u1.Username == u2.Username && u1.Email == u2.Email
}


/*
ROOST_METHOD_HASH=GetByEmail_3574af40e5
ROOST_METHOD_SIG_HASH=GetByEmail_5731b833c1

FUNCTION_DEF=func (s *UserStore) GetByEmail(email string) (*model.User, error) 

 */
func TestUserStoreGetByEmail(t *testing.T) {

	tests := []struct {
		name         string
		setupMock    func(sqlmock.Sqlmock)
		email        string
		expectedUser *model.User
		expectedErr  string
	}{
		{
			name: "Scenario 1: Retrieve User Successfully by Email",
			setupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "username", "email", "password", "bio", "image"}).
					AddRow(1, "testuser", "user@example.com", "hashed_password", "user bio", "user image")
				mock.ExpectQuery("^SELECT (.+) FROM `users` WHERE (.+)").WithArgs("user@example.com").WillReturnRows(rows)
			},
			email: "user@example.com",
			expectedUser: &model.User{
				Model:    gorm.Model{ID: 1},
				Username: "testuser",
				Email:    "user@example.com",
				Password: "hashed_password",
				Bio:      "user bio",
				Image:    "user image",
			},
			expectedErr: "",
		},
		{
			name: "Scenario 2: User Not Found by Email",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("^SELECT (.+) FROM `users` WHERE (.+)").WithArgs("missing@example.com").
					WillReturnError(gorm.ErrRecordNotFound)
			},
			email:        "missing@example.com",
			expectedUser: nil,
			expectedErr:  gorm.ErrRecordNotFound.Error(),
		},
		{
			name: "Scenario 3: Database Connection Error",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("^SELECT (.+) FROM `users` WHERE (.+)").WithArgs("user@example.com").WillReturnError(errors.New("connection error"))
			},
			email:        "user@example.com",
			expectedUser: nil,
			expectedErr:  "connection error",
		},
		{
			name: "Scenario 4: Invalid Email Format Handling",
			setupMock: func(mock sqlmock.Sqlmock) {

			},
			email:        "invalid-email",
			expectedUser: nil,
			expectedErr:  gorm.ErrRecordNotFound.Error(),
		},
		{
			name: "Scenario 5: Retrieval of User with Special Characters in Email",
			setupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "username", "email", "password", "bio", "image"}).
					AddRow(1, "testuser", "user+test@example.com", "hashed_password", "user bio", "user image")
				mock.ExpectQuery("^SELECT (.+) FROM `users` WHERE (.+)").WithArgs("user+test@example.com").WillReturnRows(rows)
			},
			email: "user+test@example.com",
			expectedUser: &model.User{
				Model:    gorm.Model{ID: 1},
				Username: "testuser",
				Email:    "user+test@example.com",
				Password: "hashed_password",
				Bio:      "user bio",
				Image:    "user image",
			},
			expectedErr: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
			}
			defer db.Close()

			gormDB, err := gorm.Open("mysql", db)
			if err != nil {
				t.Fatalf("an error '%s' was not expected when initializing gorm", err)
			}

			userStore := &UserStore{db: gormDB}

			tt.setupMock(mock)

			user, err := userStore.GetByEmail(tt.email)

			if tt.expectedErr != "" {
				if err == nil || err.Error() != tt.expectedErr {
					t.Errorf("expected error: %v, got: %v", tt.expectedErr, err)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}

			if tt.expectedUser != nil && user != nil {
				if user.Email != tt.expectedUser.Email || user.Username != tt.expectedUser.Username {
					t.Errorf("unexpected user returned. Expected: %v, got: %v", tt.expectedUser, user)
				}
			} else if tt.expectedUser != user {
				t.Errorf("unexpected user returned. Expected: %v, got: %v", tt.expectedUser, user)
			}

			err = mock.ExpectationsWereMet()
			if err != nil {
				t.Errorf("unfulfilled expectations: %s", err)
			}

			t.Log("Success for", tt.name)
		})
	}
}


/*
ROOST_METHOD_HASH=GetByUsername_f11f114df2
ROOST_METHOD_SIG_HASH=GetByUsername_954d096e24

FUNCTION_DEF=func (s *UserStore) GetByUsername(username string) (*model.User, error) 

 */
func TestUserStoreGetByUsername(t *testing.T) {

	type testCase struct {
		name         string
		username     string
		mockFunc     func(mock sqlmock.Sqlmock)
		expectedUser *model.User
		expectError  bool
	}

	testCases := []testCase{
		{
			name:     "Successful Retrieval of Existing User by Username",
			username: "existingUser",
			mockFunc: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("^SELECT (.+) FROM `users` WHERE (.+)$").
					WithArgs("existingUser").
					WillReturnRows(sqlmock.NewRows([]string{"id", "username", "email"}).
						AddRow(1, "existingUser", "user@example.com"))
			},
			expectedUser: &model.User{
				Model:    gorm.Model{ID: 1},
				Username: "existingUser",
				Email:    "user@example.com",
			},
			expectError: false,
		},
		{
			name:     "Error Handling When Username Does Not Exist",
			username: "nonExistentUser",
			mockFunc: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("^SELECT (.+) FROM `users` WHERE (.+)$").
					WithArgs("nonExistentUser").
					WillReturnError(gorm.ErrRecordNotFound)
			},
			expectedUser: nil,
			expectError:  true,
		},
		{
			name:     "Database Connection Error Scenario",
			username: "anyUser",
			mockFunc: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("^SELECT (.+) FROM `users` WHERE (.+)$").
					WithArgs("anyUser").
					WillReturnError(gorm.ErrInvalidTransaction)
			},
			expectedUser: nil,
			expectError:  true,
		},
		{
			name:     "Handling of Unexpected Database Errors",
			username: "anyUser",
			mockFunc: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("^SELECT (.+) FROM `users` WHERE (.+)$").
					WithArgs("anyUser").
					WillReturnError(gorm.ErrInvalidTransaction)
			},
			expectedUser: nil,
			expectError:  true,
		},
		{
			name:     "Case Sensitivity Test",
			username: "Existinguser",
			mockFunc: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("^SELECT (.+) FROM `users` WHERE (.+)$").
					WithArgs("Existinguser").
					WillReturnRows(sqlmock.NewRows([]string{"id", "username", "email"}).
						AddRow(2, "Existinguser", "user@example.com"))
			},
			expectedUser: &model.User{
				Model:    gorm.Model{ID: 2},
				Username: "Existinguser",
				Email:    "user@example.com",
			},
			expectError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
			}
			defer db.Close()

			gormDB, err := gorm.Open("mysql", db)
			if err != nil {
				t.Fatalf("failed to open gorm DB: %v", err)
			}

			userStore := &UserStore{db: gormDB}

			tc.mockFunc(mock)

			user, err := userStore.GetByUsername(tc.username)

			if (err != nil) != tc.expectError {
				t.Errorf("expected error: %v, got: %v", tc.expectError, err)
			}

			if err == nil && user != nil {
				if user.ID != tc.expectedUser.ID ||
					user.Username != tc.expectedUser.Username ||
					user.Email != tc.expectedUser.Email {
					t.Errorf("unexpected user result: got %+v, want %+v", user, tc.expectedUser)
				}
			} else if err == nil && (user == nil || tc.expectedUser == nil) {
				t.Error("expected non-nil user, got nil")
			}

			if err != nil {
				t.Logf("Test '%s' failed with error: %v", tc.name, err)
			} else {
				t.Logf("Test '%s' succeeded, retrieved user: %v", tc.name, user.Username)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("unmet expectations: %s", err)
			}
		})
	}
}


/*
ROOST_METHOD_HASH=Update_68f27dd78a
ROOST_METHOD_SIG_HASH=Update_87150d6435

FUNCTION_DEF=func (s *UserStore) Update(m *model.User) error 

 */
func (a AnyTime) Match(v driver.Value) bool {
	_, ok := v.(time.Time)
	return ok
}

func TestUserStoreUpdate(t *testing.T) {
	tests := []TestData{
		{
			name: "Successfully Update a User's Information",
			user: &model.User{
				Model:    gorm.Model{ID: 1, UpdatedAt: time.Now()},
				Username: "new_username",
				Email:    "new_email@example.com",
			},
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec(`UPDATE "users"`).
					WithArgs(AnyTime{}, 1, "new_username", "new_email@example.com", nil, nil, nil, 1).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
			verify: func(err error, t *testing.T) {
				if err != nil {
					t.Errorf("Expected no error, got %v", err)
				}
			},
		},
		{
			name: "Attempt to Update a Non-existent User",
			user: &model.User{
				Model: gorm.Model{ID: 999},
			},
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec(`UPDATE "users"`).
					WithArgs(AnyTime{}, 999).
					WillReturnError(gorm.ErrRecordNotFound)
				mock.ExpectRollback()
			},
			verify: func(err error, t *testing.T) {
				if err == nil || !errors.Is(err, gorm.ErrRecordNotFound) {
					t.Error("Expected record not found error")
				}
			},
		},
		{
			name: "Handle Database Connection Error During Update",
			user: &model.User{
				Model: gorm.Model{ID: 1},
			},
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec(`UPDATE "users"`).
					WithArgs(AnyTime{}, 1).
					WillReturnError(errors.New("connection error"))
				mock.ExpectRollback()
			},
			verify: func(err error, t *testing.T) {
				if err == nil || err.Error() != "connection error" {
					t.Error("Expected connection error")
				}
			},
		},
		{
			name: "Attempt to Update User with Invalid Data",
			user: &model.User{
				Model: gorm.Model{ID: 1},
				Email: "",
			},
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec(`UPDATE "users"`).
					WithArgs(AnyTime{}, 1, "", nil, nil, nil, nil, 1).
					WillReturnError(errors.New("validation error"))
				mock.ExpectRollback()
			},
			verify: func(err error, t *testing.T) {
				if err == nil || err.Error() != "validation error" {
					t.Error("Expected validation error")
				}
			},
		},
		{
			name: "Concurrent Update Request Handling",
			user: &model.User{
				Model: gorm.Model{ID: 1},
				Email: "concurrent@example.com",
			},
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec(`UPDATE "users"`).
					WithArgs(AnyTime{}, 1, "concurrent@example.com", nil, nil, nil, nil, 1).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
				mock.ExpectBegin()
				mock.ExpectExec(`UPDATE "users"`).
					WithArgs(AnyTime{}, 1, "concurrent@example.com", nil, nil, nil, nil, 1).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
			verify: func(err error, t *testing.T) {
				if err != nil {
					t.Errorf("Expected no error, got %v", err)
				}
			},
		},
		{
			name: "Update User with No Changes",
			user: &model.User{
				Model: gorm.Model{ID: 1},
				Email: "same@example.com",
			},
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec(`UPDATE "users"`).
					WithArgs(AnyTime{}, 1, "same@example.com", nil, nil, nil, nil, 1).
					WillReturnResult(sqlmock.NewResult(1, 0))
				mock.ExpectCommit()
			},
			verify: func(err error, t *testing.T) {
				if err != nil {
					t.Errorf("Expected no error, got %v", err)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
			}
			defer db.Close()

			gormDB, err := gorm.Open("sqlite3", db)
			if err != nil {
				t.Fatalf("Failed to open gorm DB: %v", err)
			}

			userStore := &UserStore{db: gormDB}

			tt.mockSetup(mock)

			err = userStore.Update(tt.user)

			tt.verify(err, t)
		})
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

	gormDB, err := gorm.Open("postgres", db)
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub gorm connection", err)
	}

	store := &UserStore{db: gormDB}

	userA := &model.User{Model: gorm.Model{ID: 1}}
	userB := &model.User{Model: gorm.Model{ID: 2}}

	tests := []struct {
		name      string
		userA     *model.User
		userB     *model.User
		setupMock func()
		want      bool
		wantErr   bool
	}{
		{
			name:  "Valid Following Relationship Exists",
			userA: userA,
			userB: userB,
			setupMock: func() {
				rows := sqlmock.NewRows([]string{"count"}).AddRow(1)
				mock.ExpectQuery("^SELECT count(.+) FROM follows").
					WithArgs(userA.ID, userB.ID).
					WillReturnRows(rows)
			},
			want:    true,
			wantErr: false,
		},
		{
			name:  "No Following Relationship Exists",
			userA: userA,
			userB: userB,
			setupMock: func() {
				rows := sqlmock.NewRows([]string{"count"}).AddRow(0)
				mock.ExpectQuery("^SELECT count(.+) FROM follows").
					WithArgs(userA.ID, userB.ID).
					WillReturnRows(rows)
			},
			want:    false,
			wantErr: false,
		},
		{
			name:      "Handling Null User Parameter (User B nil)",
			userA:     userA,
			userB:     nil,
			setupMock: func() {},
			want:      false,
			wantErr:   false,
		},
		{
			name:  "Database Error Occurrence",
			userA: userA,
			userB: userB,
			setupMock: func() {
				mock.ExpectQuery("^SELECT count(.+) FROM follows").
					WithArgs(userA.ID, userB.ID).
					WillReturnError(gorm.ErrInvalidSQL)
			},
			want:    false,
			wantErr: true,
		},
		{
			name:  "Multiple Follow Entries Between Users",
			userA: userA,
			userB: userB,
			setupMock: func() {
				rows := sqlmock.NewRows([]string{"count"}).AddRow(5)
				mock.ExpectQuery("^SELECT count(.+) FROM follows").
					WithArgs(userA.ID, userB.ID).
					WillReturnRows(rows)
			},
			want:    true,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			got, err := store.IsFollowing(tt.userA, tt.userB)
			if (err != nil) != tt.wantErr {
				t.Errorf("IsFollowing() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("IsFollowing() = %v, want %v", got, tt.want)
			}

			if err != nil && tt.wantErr {
				t.Logf("Test %s passed. Error: %v", tt.name, err)
			} else if got == tt.want {
				t.Logf("Test %s passed. Expected output: %v", tt.name, tt.want)
			} else {
				t.Logf("Test %s failed. Expected output: %v but got %v", tt.name, tt.want, got)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}


/*
ROOST_METHOD_HASH=Unfollow_57959a8a53
ROOST_METHOD_SIG_HASH=Unfollow_8bd8e0bc55

FUNCTION_DEF=func (s *UserStore) Unfollow(a *model.User, b *model.User) error 

 */
func TestUserStoreUnfollow(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	gormDB, err := gorm.Open("sqlite3", db)
	if err != nil {
		t.Fatalf("failed to open gorm db, got error: %v", err)
	}

	userStore := &UserStore{db: gormDB}

	tests := []struct {
		name      string
		userA     *model.User
		userB     *model.User
		setupMock func()
		wantErr   bool
	}{
		{
			name: "Scenario 1: Successfully Unfollowing a User",
			userA: &model.User{
				Model: gorm.Model{ID: 1},
				Follows: []model.User{
					{Model: gorm.Model{ID: 2}},
				},
			},
			userB: &model.User{Model: gorm.Model{ID: 2}},
			setupMock: func() {

				mock.ExpectBegin()
				mock.ExpectExec(`DELETE FROM "follows"`).
					WithArgs(1, 2).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
			wantErr: false,
		},
		{
			name: "Scenario 2: Attempt to Unfollow When Not Following",
			userA: &model.User{
				Model:   gorm.Model{ID: 3},
				Follows: []model.User{},
			},
			userB: &model.User{Model: gorm.Model{ID: 4}},
			setupMock: func() {

				mock.ExpectQuery(`SELECT COUNT(1) FROM "follows" WHERE "from_user_id" = ? AND "to_user_id" = ?`).
					WithArgs(3, 4).
					WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
			},
			wantErr: false,
		},
		{
			name:  "Scenario 3: Handling a Non-Existent User",
			userA: &model.User{Model: gorm.Model{ID: 5}},
			userB: &model.User{Model: gorm.Model{ID: 6}},
			setupMock: func() {

				mock.ExpectQuery(`SELECT COUNT(1) FROM "follows" WHERE "from_user_id" = ? AND "to_user_id" = ?`).
					WithArgs(5, 6).
					WillReturnError(gorm.ErrRecordNotFound)
			},
			wantErr: true,
		},
		{
			name: "Scenario 4: Database Error During Unfollow",
			userA: &model.User{
				Model: gorm.Model{ID: 7},
				Follows: []model.User{
					{Model: gorm.Model{ID: 8}},
				},
			},
			userB: &model.User{Model: gorm.Model{ID: 8}},
			setupMock: func() {

				mock.ExpectBegin()
				mock.ExpectExec(`DELETE FROM "follows"`).
					WithArgs(7, 8).
					WillReturnError(gorm.ErrInvalidSQL)
				mock.ExpectRollback()
			},
			wantErr: true,
		},
		{
			name:  "Scenario 5: Unfollow Self Attempt",
			userA: &model.User{Model: gorm.Model{ID: 9}},
			userB: &model.User{Model: gorm.Model{ID: 9}},
			setupMock: func() {

				mock.ExpectQuery(`SELECT COUNT(1) FROM "follows" WHERE "from_user_id" = ? AND "to_user_id" = ?`).
					WithArgs(9, 9).
					WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()
			err := userStore.Unfollow(tt.userA, tt.userB)
			if (err != nil) != tt.wantErr {
				t.Errorf("UserStore.Unfollow() error = %v, wantErr %v", err, tt.wantErr)
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
		name          string
		mockSetup     func(sqlmock.Sqlmock)
		inputUser     *model.User
		expectedIDs   []uint
		expectedError bool
	}{
		{
			name: "Retrieve Following User IDs Successfully",
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"to_user_id"}).
					AddRow(uint(101)).
					AddRow(uint(102))
				mock.ExpectQuery("^SELECT to_user_id FROM follows WHERE").
					WithArgs(1).
					WillReturnRows(rows)
			},
			inputUser:     &model.User{Model: gorm.Model{ID: 1}},
			expectedIDs:   []uint{101, 102},
			expectedError: false,
		},
		{
			name: "No Followings for the User",
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"to_user_id"})
				mock.ExpectQuery("^SELECT to_user_id FROM follows WHERE").
					WithArgs(2).
					WillReturnRows(rows)
			},
			inputUser:     &model.User{Model: gorm.Model{ID: 2}},
			expectedIDs:   []uint{},
			expectedError: false,
		},
		{
			name: "Database Error When Fetching User IDs",
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("^SELECT to_user_id FROM follows WHERE").
					WithArgs(3).
					WillReturnError(fmt.Errorf("db error"))
			},
			inputUser:     &model.User{Model: gorm.Model{ID: 3}},
			expectedIDs:   []uint{},
			expectedError: true,
		},
		{
			name: "Valid SQL Query Execution",
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"to_user_id"})
				mock.ExpectQuery("^SELECT to_user_id FROM follows WHERE").
					WithArgs(4).
					WillReturnRows(rows)
			},
			inputUser:     &model.User{Model: gorm.Model{ID: 4}},
			expectedIDs:   []uint{},
			expectedError: false,
		},
		{
			name: "Handling Large Number of Followings",
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"to_user_id"})
				for i := 1; i <= 1000; i++ {
					rows.AddRow(uint(i))
				}
				mock.ExpectQuery("^SELECT to_user_id FROM follows WHERE").
					WithArgs(5).
					WillReturnRows(rows)
			},
			inputUser:     &model.User{Model: gorm.Model{ID: 5}},
			expectedIDs:   generateIDs(1000),
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				log.Fatalf("An error '%s' was not expected when opening a stub database connection", err)
			}
			defer db.Close()

			tt.mockSetup(mock)

			gormDB, err := gorm.Open("mysql", db)
			if err != nil {
				log.Fatalf("An error '%s' was not expected when opening the gorm database", err)
			}

			store := UserStore{db: gormDB}

			ids, err := store.GetFollowingUserIDs(tt.inputUser)

			if (err != nil) != tt.expectedError {
				t.Errorf("expected error: %v, got: %v", tt.expectedError, err)
			}

			if !equalIDs(ids, tt.expectedIDs) {
				t.Errorf("expected IDs: %v, got: %v", tt.expectedIDs, ids)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("unfulfilled expectations: %s", err)
			}

			t.Logf("Test %s succeeded", tt.name)
		})
	}
}

func equalIDs(a, b []uint) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func generateIDs(count int) []uint {
	ids := make([]uint, count)
	for i := range ids {
		ids[i] = uint(i + 1)
	}
	return ids
}

