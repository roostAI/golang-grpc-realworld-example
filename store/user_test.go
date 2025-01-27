package github

import (
	"reflect"
	"testing"
	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/jinzhu/gorm"
	"errors"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/raahii/golang-grpc-realworld-example/model"
	"fmt"
	"regexp"
	"gopkg.in/DATA-DOG/go-sqlmock.v1"
	"bytes"
	"log"
	"sync"
)









/*
ROOST_METHOD_HASH=NewUserStore_6a331dd890
ROOST_METHOD_SIG_HASH=NewUserStore_4f0c2dfca9

FUNCTION_DEF=func NewUserStore(db *gorm.DB) *UserStore 

 */
func TestNewUserStore(t *testing.T) {
	tests := []struct {
		name           string
		db             *gorm.DB
		expectNilDB    bool
		expectSameType bool
	}{
		{
			name:           "Non-nil Database",
			db:             validDB(t),
			expectNilDB:    false,
			expectSameType: true,
		},
		{
			name:           "Nil Database",
			db:             nil,
			expectNilDB:    true,
			expectSameType: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userStore := NewUserStore(tt.db)

			if !tt.expectNilDB && userStore.db != tt.db {
				t.Errorf("expected db attribute to be set correctly, got %v want %v", userStore.db, tt.db)
			} else if tt.expectNilDB && userStore.db != nil {

				t.Errorf("expected db attribute to be nil, got %v", userStore.db)
			}

			if _, ok := interface{}(userStore).(*UserStore); !ok {
				t.Errorf("expected return type *UserStore, got %T", userStore)
			}

			if tt.db != nil {
				originalDB := *tt.db
				modifiedDB := *userStore.db
				if !reflect.DeepEqual(originalDB, modifiedDB) {
					t.Errorf("expected userStore's DB to remain unaffected by changes in original db")
				}
			}

			t.Logf("Scenario '%s' passed", tt.name)
		})
	}
}

func validDB(t *testing.T) *gorm.DB {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("could not create mock db: %v", err)
	}

	gormDB, _ := gorm.Open("mysql", db)
	return gormDB
}


/*
ROOST_METHOD_HASH=Create_889fc0fc45
ROOST_METHOD_SIG_HASH=Create_4c48ec3920

FUNCTION_DEF=func (s *UserStore) Create(m *model.User) error 

 */
func TestUserStoreCreate(t *testing.T) {
	type args struct {
		user *model.User
	}
	tests := []struct {
		name    string
		args    args
		setup   func(sqlmock.Sqlmock)
		wantErr bool
		errMsg  string
	}{
		{
			name: "Scenario 1: Successfully Create a New User",
			args: args{
				user: &model.User{
					Username: "john_doe",
					Email:    "john@example.com",
					Password: "securepassword",
				},
			},
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO \"users\"").WithArgs().WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
			wantErr: false,
		},
		{
			name: "Scenario 2: Attempt to Create a User with Duplicate Email",
			args: args{
				user: &model.User{
					Username: "jane_doe",
					Email:    "john@example.com",
					Password: "anotherpassword",
				},
			},
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO \"users\"").WithArgs().WillReturnError(errors.New("UNIQUE constraint failed: users.email"))
				mock.ExpectRollback()
			},
			wantErr: true,
			errMsg:  "UNIQUE constraint failed: users.email",
		},
		{
			name: "Scenario 3: Attempt to Create a User with Missing Required Fields",
			args: args{
				user: &model.User{
					Email: "missing@username.com",
				},
			},
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO \"users\"").WithArgs().WillReturnError(errors.New("Field 'Username' is required"))
				mock.ExpectRollback()
			},
			wantErr: true,
			errMsg:  "Field 'Username' is required",
		},
		{
			name: "Scenario 4: Database Connection Error During User Creation",
			args: args{
				user: &model.User{
					Username: "charles_doe",
					Email:    "charles@example.com",
					Password: "charlespassword",
				},
			},
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin().WillReturnError(gorm.ErrCantStartTransaction)
			},
			wantErr: true,
			errMsg:  "transaction could not be started",
		},
		{
			name: "Scenario 5: Attempt to Create a User with Null Database",
			args: args{
				user: &model.User{
					Username: "null_db_user",
					Email:    "null@db.com",
					Password: "nullpassword",
				},
			},
			setup: func(sqlmock.Sqlmock) {

			},
			wantErr: true,
			errMsg:  "db pointer is nil",
		},
		{
			name: "Scenario 6: Create a User and Verify Model Auto-Population",
			args: args{
				user: &model.User{
					Username: "auto_populate",
					Email:    "auto@populate.com",
					Password: "autopassword",
				},
			},
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO \"users\"").WithArgs().WillReturnResult(sqlmock.NewResult(2, 1))
				mock.ExpectCommit()
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("Error setting up mock database: %v", err)
			}
			defer db.Close()

			gdb, err := gorm.Open("_", db)
			if err != nil {
				t.Fatalf("Error opening gorm connection: %v", err)
			}

			us := &UserStore{db: gdb}

			tt.setup(mock)

			err = us.Create(tt.args.user)

			if (err != nil) != tt.wantErr {
				t.Errorf("Create() error = %v, wantErr %v", err, tt.wantErr)
				t.Logf("Failed due to: %v", err)
			}

			if err != nil && tt.errMsg != "" && err.Error() != tt.errMsg {
				t.Errorf("Create() error = %v, expected error message = %v", err, tt.errMsg)
			}

			if !tt.wantErr && tt.name == "Scenario 6: Create a User and Verify Model Auto-Population" {
				if tt.args.user.CreatedAt.IsZero() || tt.args.user.ID == 0 {
					t.Error("Expected CreatedAt and ID to be auto-populated, but they weren't")
				}
			}

			err = mock.ExpectationsWereMet()
			if err != nil {
				t.Errorf("Expectations were not met: %v", err)
			}
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

	gormDB, err := gorm.Open("mysql", db)
	if err != nil {
		t.Fatalf("failed to open gorm DB, got error: %v", err)
	}
	defer gormDB.Close()

	userStore := &UserStore{db: gormDB}

	type testCase struct {
		email         string
		expectedUser  *model.User
		expectedError error
		mockSetup     func()
	}

	var testCases = []struct {
		description string
		tc          testCase
	}{
		{
			description: "Retrieve Existing User by Email",
			tc: testCase{
				email: "user@example.com",
				expectedUser: &model.User{
					Email:    "user@example.com",
					Username: "testuser",
					Bio:      "test bio",
					Image:    "http://example.com/image.jpg",
				},
				expectedError: nil,
				mockSetup: func() {
					rows := sqlmock.NewRows([]string{"email", "username", "bio", "image"}).
						AddRow("user@example.com", "testuser", "test bio", "http://example.com/image.jpg")
					mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `users` WHERE (email = ?) ORDER BY `users`.`id` ASC LIMIT 1")).
						WithArgs("user@example.com").WillReturnRows(rows)
				},
			},
		},
		{
			description: "Handle Non-Existent Email",
			tc: testCase{
				email:         "nonexist@example.com",
				expectedUser:  nil,
				expectedError: gorm.ErrRecordNotFound,
				mockSetup: func() {
					mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `users` WHERE (email = ?) ORDER BY `users`.`id` ASC LIMIT 1")).
						WithArgs("nonexist@example.com").WillReturnError(gorm.ErrRecordNotFound)
				},
			},
		},
		{
			description: "Handle Database Connection Error",
			tc: testCase{
				email:         "user@example.com",
				expectedUser:  nil,
				expectedError: fmt.Errorf("connection error"),
				mockSetup: func() {
					mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `users` WHERE (email = ?) ORDER BY `users`.`id` ASC LIMIT 1")).
						WithArgs("user@example.com").WillReturnError(fmt.Errorf("connection error"))
				},
			},
		},
		{
			description: "Handle Email with Special Characters",
			tc: testCase{
				email: "special+user@example.com",
				expectedUser: &model.User{
					Email:    "special+user@example.com",
					Username: "specialuser",
					Bio:      "special bio",
					Image:    "http://example.com/special_image.jpg",
				},
				expectedError: nil,
				mockSetup: func() {
					rows := sqlmock.NewRows([]string{"email", "username", "bio", "image"}).
						AddRow("special+user@example.com", "specialuser", "special bio", "http://example.com/special_image.jpg")
					mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `users` WHERE (email = ?) ORDER BY `users`.`id` ASC LIMIT 1")).
						WithArgs("special+user@example.com").WillReturnRows(rows)
				},
			},
		},
		{
			description: "Validate Case Insensitivity of Email Query",
			tc: testCase{
				email: "USER@EXAMPLE.COM",
				expectedUser: &model.User{
					Email:    "user@example.com",
					Username: "testuser",
					Bio:      "test bio",
					Image:    "http://example.com/image.jpg",
				},
				expectedError: nil,
				mockSetup: func() {
					rows := sqlmock.NewRows([]string{"email", "username", "bio", "image"}).
						AddRow("user@example.com", "testuser", "test bio", "http://example.com/image.jpg")
					mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `users` WHERE (email = ?) ORDER BY `users`.`id` ASC LIMIT 1")).
						WithArgs("USER@EXAMPLE.COM").WillReturnRows(rows)
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {

			tc.tc.mockSetup()

			user, err := userStore.GetByEmail(tc.tc.email)

			if tc.tc.expectedError != nil {
				if err == nil || tc.tc.expectedError.Error() != err.Error() {
					t.Errorf("expected error %v, got %v", tc.tc.expectedError, err)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				} else if user == nil || user.Email != tc.tc.expectedUser.Email || user.Username != tc.tc.expectedUser.Username {
					t.Errorf("expected user %v, got %v", tc.tc.expectedUser, user)
				}
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
	var mu sync.Mutex

	type args struct {
		user *model.User
	}
	tests := []struct {
		name     string
		prepare  func(mock sqlmock.Sqlmock)
		args     args
		wantErr  bool
		errorMsg string
	}{
		{
			name: "Scenario 1: Successful User Update",
			prepare: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec(`UPDATE "users"`).
					WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
			args: args{
				user: &model.User{
					Username: "test_user",
					Email:    "test@example.com",
					Bio:      "Just testing",
					Image:    "test.jpg",
				},
			},
			wantErr: false,
		},
		{
			name: "Scenario 2: Attempt to Update Non-Existent User",
			prepare: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec(`UPDATE "users"`).
					WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
					WillReturnError(gorm.ErrRecordNotFound)
				mock.ExpectCommit()
			},
			args: args{
				user: &model.User{
					Username: "non_existent_user",
				},
			},
			wantErr:  true,
			errorMsg: gorm.ErrRecordNotFound.Error(),
		},
		{
			name: "Scenario 3: Update with Invalid User Data",
			prepare: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec(`UPDATE "users"`).
					WillReturnError(fmt.Errorf("duplicate key value violates unique constraint"))
				mock.ExpectCommit()
			},
			args: args{
				user: &model.User{
					Username: "duplicate_user",
				},
			},
			wantErr:  true,
			errorMsg: "duplicate key value violates unique constraint",
		},
		{
			name: "Scenario 4: Database Connection Issue During Update",
			prepare: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec(`UPDATE "users"`).
					WillReturnError(fmt.Errorf("connection error"))
				mock.ExpectCommit()
			},
			args: args{
				user: &model.User{
					Username: "connection_error_user",
				},
			},
			wantErr:  true,
			errorMsg: "connection error",
		},
		{
			name: "Scenario 5: Partial Update Type",
			prepare: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec(`UPDATE "users"`).
					WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
			args: args{
				user: &model.User{
					Email: "updated_email@example.com",
					Bio:   "",
				},
			},
			wantErr: false,
		},
		{
			name: "Scenario 6: Concurrency Test for Simultaneous Updates",
			prepare: func(mock sqlmock.Sqlmock) {

				mock.ExpectBegin()
				mock.ExpectExec(`UPDATE "users"`).
					WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
			args: args{
				user: &model.User{
					Username: "concurrent_user",
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("Failed to create mock database: %v", err)
			}
			defer db.Close()

			gormDB, err := gorm.Open("_", db)
			if err != nil {
				t.Fatalf("Failed to open gorm db: %v", err)
			}

			mu.Lock()
			defer mu.Unlock()

			if tt.name == "Scenario 6: Concurrency Test for Simultaneous Updates" {
				go tt.prepare(mock)
				go tt.prepare(mock)
			} else {
				tt.prepare(mock)
			}

			store := &UserStore{db: gormDB}

			var buf bytes.Buffer
			log.SetOutput(&buf)

			err = store.Update(tt.args.user)

			if (err != nil) != tt.wantErr {
				t.Errorf("UserStore.Update() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantErr && err != nil && tt.errorMsg != "" && err.Error() != tt.errorMsg {
				t.Errorf("UserStore.Update() error = %v, expected error message %v", err, tt.errorMsg)
			}

			t.Logf("Scenario: %s - Log: %s", tt.name, buf.String())

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
		t.Fatalf("Error setting up mock DB: %v", err)
	}
	defer db.Close()
	gormDB, err := gorm.Open("_", db)
	if err != nil {
		t.Fatalf("Error opening gorm DB: %v", err)
	}
	store := UserStore{db: gormDB}

	userA := &model.User{Model: gorm.Model{ID: 1}, Username: "userA"}
	userB := &model.User{Model: gorm.Model{ID: 2}, Username: "userB"}

	testCases := []struct {
		name        string
		setupMock   func()
		expectError bool
	}{
		{
			name: "Successfully Unfollow a User",
			setupMock: func() {

				mock.ExpectBegin()
				mock.ExpectExec(`DELETE FROM "follows"`).
					WithArgs(userA.ID, userB.ID).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
			expectError: false,
		},
		{
			name: "Error When User Tries to Unfollow a Non-Followed User",
			setupMock: func() {

				mock.ExpectBegin()
				mock.ExpectExec(`DELETE FROM "follows"`).
					WithArgs(userA.ID, userB.ID).
					WillReturnError(errors.New("not following"))
				mock.ExpectRollback()
			},
			expectError: true,
		},
		{
			name: "Handles Database Connection Error During Unfollow",
			setupMock: func() {

				mock.ExpectBegin().WillReturnError(errors.New("connection error"))
			},
			expectError: true,
		},
		{
			name: "Edge Case - Unfollow for Non-Existent User",
			setupMock: func() {

				mock.ExpectBegin()
				mock.ExpectExec(`DELETE FROM "follows"`).
					WithArgs(userA.ID, 99).
					WillReturnError(errors.New("user not found"))
				mock.ExpectRollback()
			},
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.setupMock()
			err := store.Unfollow(userA, userB)
			if tc.expectError && err == nil {
				t.Errorf("Expected error, got nil")
			}
			if !tc.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("Unfulfilled expectations: %v", err)
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
		scenario    string
		mockSetup   func(mock sqlmock.Sqlmock)
		expectedIDs []uint
		expectedErr bool
	}{
		{
			scenario: "Successfully Retrieve Following User IDs",
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"to_user_id"}).AddRow(2).AddRow(3)
				mock.ExpectQuery("^SELECT (.+) FROM follows WHERE").WithArgs(uint(1)).WillReturnRows(rows)
			},
			expectedIDs: []uint{2, 3},
			expectedErr: false,
		},
		{
			scenario: "User is Not Following Anyone",
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"to_user_id"})
				mock.ExpectQuery("^SELECT (.+) FROM follows WHERE").WithArgs(uint(1)).WillReturnRows(rows)
			},
			expectedIDs: []uint{},
			expectedErr: false,
		},
		{
			scenario: "Database Error Occurs",
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("^SELECT (.+) FROM follows WHERE").WithArgs(uint(1)).WillReturnError(gorm.ErrInvalidSQL)
			},
			expectedIDs: nil,
			expectedErr: true,
		},
		{
			scenario: "Retrieve Following User IDs for User with Edge Case ID",
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"to_user_id"}).AddRow(2).AddRow(4)
				mock.ExpectQuery("^SELECT (.+) FROM follows WHERE").WithArgs(uint(0)).WillReturnRows(rows)
			},
			expectedIDs: []uint{2, 4},
			expectedErr: false,
		},
		{
			scenario: "Retrieving Following User IDs with Multiple Followers",
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"to_user_id"}).AddRow(2).AddRow(3).AddRow(4)
				mock.ExpectQuery("^SELECT (.+) FROM follows WHERE").WithArgs(uint(1)).WillReturnRows(rows)
			},
			expectedIDs: []uint{2, 3, 4},
			expectedErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.scenario, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("error creating DB: %s", err)
			}
			defer db.Close()

			gormDB, err := gorm.Open("sqlite3", db)
			if err != nil {
				t.Fatalf("failed to open gorm DB: %s", err)
			}

			us := &UserStore{db: gormDB}

			tt.mockSetup(mock)

			user := &model.User{Model: gorm.Model{ID: uint(1)}}

			result, err := us.GetFollowingUserIDs(user)
			if tt.expectedErr && err == nil {
				t.Errorf("expected error, got none")
			} else if !tt.expectedErr && err != nil {
				t.Errorf("did not expect error, got %v", err)
			}

			if tt.expectedErr {
				t.Logf("Scenario '%s' check: Correctly produced an error and an empty slice", tt.scenario)
				if result != nil {
					t.Errorf("Expected nil slice, got %v", result)
				}
			} else {
				if !equal(result, tt.expectedIDs) {
					t.Errorf("Expected IDs %v, got %v", tt.expectedIDs, result)
				} else {
					t.Logf("Scenario '%s' check: Retrieved IDs match expected list - %v", tt.scenario, result)
				}
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("ensured all expectations were met, got err: %s", err)
			}
		})
	}
}

func equal(a, b []uint) bool {
	if len(a) != len(b) {
		return false
	}

	for i, v := range a {
		if v != b[i] {
			return false
		}
	}

	return true
}

