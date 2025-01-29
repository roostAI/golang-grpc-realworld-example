package store

import (
	"errors"
	"testing"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jinzhu/gorm"
	"github.com/raahii/golang-grpc-realworld-example/model"
	"fmt"
	_ "github.com/jinzhu/gorm/dialects/postgres"
)









/*
ROOST_METHOD_HASH=Create_889fc0fc45
ROOST_METHOD_SIG_HASH=Create_4c48ec3920

FUNCTION_DEF=func (s *UserStore) Create(m *model.User) error 

 */
func TestUserStoreCreate(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to open mock sql db, got error %v", err)
	}
	defer db.Close()

	gormDB, err := gorm.Open("_", db)
	if err != nil {
		t.Fatalf("Failed to open gorm DB, got error %v", err)
	}

	userStore := &UserStore{db: gormDB}

	var (
		ErrMissingField = errors.New("missing required fields")
		ErrDuplicateKey = errors.New("duplicate key value violates unique constraint")
		ErrConnDone     = errors.New("database connection error")
	)

	tests := []struct {
		name     string
		user     *model.User
		mock     func()
		wantErr  bool
		errorMsg string
	}{
		{
			name: "Successful User Creation",
			user: &model.User{
				Username: "testuser",
				Email:    "testuser@example.com",
				Password: "password",
				Bio:      "Bio",
				Image:    "Image",
			},
			mock: func() {
				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO \"users\"").
					WithArgs(sqlmock.AnyArg(), "testuser", "testuser@example.com", "password", "Bio", "Image").
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
			wantErr: false,
		},
		{
			name: "User Creation with Missing Required Fields",
			user: &model.User{
				Email:    "testuser@example.com",
				Password: "password",
				Bio:      "Bio",
				Image:    "Image",
			},
			mock: func() {
				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO \"users\"").
					WithArgs(sqlmock.AnyArg(), nil, "testuser@example.com", "password", "Bio", "Image").
					WillReturnError(ErrMissingField)
				mock.ExpectRollback()
			},
			wantErr:  true,
			errorMsg: "should fail due to missing fields",
		},
		{
			name: "Duplicate User Creation with Unique Constraints",
			user: &model.User{
				Username: "existinguser",
				Email:    "existinguser@example.com",
				Password: "password",
				Bio:      "Bio",
				Image:    "Image",
			},
			mock: func() {
				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO \"users\"").
					WithArgs(sqlmock.AnyArg(), "existinguser", "existinguser@example.com", "password", "Bio", "Image").
					WillReturnError(ErrDuplicateKey)
				mock.ExpectRollback()
			},
			wantErr:  true,
			errorMsg: "should fail due to duplicate username or email",
		},
		{
			name: "Database Connection Error during User Creation",
			user: &model.User{
				Username: "testuser",
				Email:    "testuser@example.com",
				Password: "password",
				Bio:      "Bio",
				Image:    "Image",
			},
			mock: func() {
				mock.ExpectBegin().WillReturnError(ErrConnDone)
			},
			wantErr:  true,
			errorMsg: "should fail due to database connection error",
		},
		{
			name: "User Creation with Complex Data Structures",
			user: &model.User{
				Username: "complexuser",
				Email:    "complexuser@example.com",
				Password: "password",
				Bio:      "Bio",
				Image:    "Image",
				Follows: []model.User{
					{Username: "follower1", Email: "follower1@example.com"},
					{Username: "follower2", Email: "follower2@example.com"},
				},
			},
			mock: func() {
				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO \"users\"").
					WithArgs(sqlmock.AnyArg(), "complexuser", "complexuser@example.com", "password", "Bio", "Image").
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectExec("INSERT INTO \"follows\"").
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mock()
			err := userStore.Create(tt.user)
			if (err != nil) != tt.wantErr {
				t.Errorf("UserStore.Create() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil {
				t.Logf("Expected error: %s, got: %v", tt.errorMsg, err)
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
		t.Fatalf("failed to open mock sql db, %s", err)
	}
	defer db.Close()

	gormDB, err := gorm.Open("sqlite3", db)
	if err != nil {
		t.Fatalf("failed to open gorm db, %s", err)
	}

	store := &UserStore{db: gormDB}

	tests := []struct {
		name           string
		setup          func()
		id             uint
		expectedUser   *model.User
		expectError    bool
		expectedErrMsg string
	}{
		{
			name: "Successfully Retrieve User by Valid ID",
			setup: func() {
				rows := sqlmock.NewRows([]string{"id", "username", "email", "password", "bio", "image"}).
					AddRow(1, "JohnDoe", "john@example.com", "hashed", "Bio", "image.png")
				mock.ExpectQuery("SELECT \\* FROM users WHERE id = ?").
					WithArgs(1).
					WillReturnRows(rows)
			},
			id: 1,
			expectedUser: &model.User{
				Model:    gorm.Model{ID: 1},
				Username: "JohnDoe",
				Email:    "john@example.com",
				Password: "hashed",
				Bio:      "Bio",
				Image:    "image.png",
			},
			expectError: false,
		},
		{
			name: "User Not Found with Non-Existent ID",
			setup: func() {
				mock.ExpectQuery("SELECT \\* FROM users WHERE id = ?").
					WithArgs(2).
					WillReturnError(gorm.ErrRecordNotFound)
			},
			id:             2,
			expectedUser:   nil,
			expectError:    true,
			expectedErrMsg: "record not found",
		},
		{
			name: "Database Error during Retrieval",
			setup: func() {
				mock.ExpectQuery("SELECT \\* FROM users WHERE id = ?").
					WithArgs(3).
					WillReturnError(errors.New("database error"))
			},
			id:             3,
			expectedUser:   nil,
			expectError:    true,
			expectedErrMsg: "database error",
		},
		{
			name: "Invalid/Zero ID Given",
			setup: func() {
				mock.ExpectQuery("SELECT \\* FROM users WHERE id = ?").
					WithArgs(0).
					WillReturnError(errors.New("invalid or zero ID"))
			},
			id:             0,
			expectedUser:   nil,
			expectError:    true,
			expectedErrMsg: "invalid or zero ID",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			tc.setup()

			user, err := store.GetByID(tc.id)

			if tc.expectError {
				if err == nil || err.Error() != tc.expectedErrMsg {
					t.Errorf("expected error: '%v', got: '%v'", tc.expectedErrMsg, err)
				}
			} else {
				if err != nil {
					t.Errorf("did not expect error, got: %v", err)
				}
				if user.ID != tc.expectedUser.ID || user.Username != tc.expectedUser.Username {
					t.Errorf("expected user ID: '%v', Username: '%v', got user ID: '%v', Username: '%v'",
						tc.expectedUser.ID, tc.expectedUser.Username, user.ID, user.Username)
				}
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

FUNCTION_DEF=func (s *UserStore) GetByUsername(username string) (*model.User, error) 

 */
func TestUserStoreGetByUsername(t *testing.T) {
	type testCase struct {
		description       string
		username          string
		setupMock         func(mock sqlmock.Sqlmock)
		expectedUser      *model.User
		expectedErrString string
	}

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock database: %v", err)
	}
	defer db.Close()

	gormDB, err := gorm.Open("mysql", db)
	if err != nil {
		t.Fatalf("failed to create GORM DB from mock: %v", err)
	}

	userStore := &UserStore{db: gormDB}

	testCases := []testCase{
		{
			description: "Successful Retrieval of User by Username",
			username:    "validUser",
			setupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "username", "email", "password", "bio", "image"}).
					AddRow(1, "validUser", "user@example.com", "hashedpassword", "bio", "image_url")
				mock.ExpectQuery(`^SELECT \* FROM "users" WHERE \(username = \?\) ORDER BY id ASC LIMIT 1$`).
					WithArgs("validUser").WillReturnRows(rows)
			},
			expectedUser: &model.User{
				Model:    gorm.Model{ID: 1},
				Username: "validUser",
				Email:    "user@example.com",
				Password: "hashedpassword",
				Bio:      "bio",
				Image:    "image_url",
			},
			expectedErrString: "",
		},
		{
			description: "User Not Found by Username",
			username:    "nonExistentUser",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`^SELECT \* FROM "users" WHERE \(username = \?\) ORDER BY id ASC LIMIT 1$`).
					WithArgs("nonExistentUser").WillReturnError(gorm.ErrRecordNotFound)
			},
			expectedUser:      nil,
			expectedErrString: gorm.ErrRecordNotFound.Error(),
		},
		{
			description: "Retrieval with Special Characters in Username",
			username:    "special!User",
			setupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "username", "email", "password", "bio", "image"}).
					AddRow(2, "special!User", "special@example.com", "specialhashedpassword", "special bio", "special_image_url")
				mock.ExpectQuery(`^SELECT \* FROM "users" WHERE \(username = \?\) ORDER BY id ASC LIMIT 1$`).
					WithArgs("special!User").WillReturnRows(rows)
			},
			expectedUser: &model.User{
				Model:    gorm.Model{ID: 2},
				Username: "special!User",
				Email:    "special@example.com",
				Password: "specialhashedpassword",
				Bio:      "special bio",
				Image:    "special_image_url",
			},
			expectedErrString: "",
		},
		{
			description: "Database Connection Error",
			username:    "someUser",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`^SELECT \* FROM "users" WHERE \(username = \?\) ORDER BY id ASC LIMIT 1$`).
					WithArgs("someUser").WillReturnError(fmt.Errorf("connection error"))
			},
			expectedUser:      nil,
			expectedErrString: "connection error",
		},
		{
			description: "Username is Empty String",
			username:    "",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`^SELECT \* FROM "users" WHERE \(username = \?\) ORDER BY id ASC LIMIT 1$`).
					WithArgs("").WillReturnError(gorm.ErrRecordNotFound)
			},
			expectedUser:      nil,
			expectedErrString: gorm.ErrRecordNotFound.Error(),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			tc.setupMock(mock)
			user, err := userStore.GetByUsername(tc.username)

			if err != nil && err.Error() != tc.expectedErrString {
				t.Errorf("unexpected error: got %v want %v", err, tc.expectedErrString)
			}

			if fmt.Sprintf("%v", user) != fmt.Sprintf("%v", tc.expectedUser) {
				t.Errorf("unexpected user: got %v want %v", user, tc.expectedUser)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %v", err)
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
	type args struct {
		a *model.User
		b *model.User
	}

	tests := []struct {
		name      string
		args      args
		dbSetup   func(mock sqlmock.Sqlmock)
		wantError bool
	}{
		{
			name: "Scenario 1: Successfully Follow a User",
			args: args{
				a: &model.User{Model: gorm.Model{ID: 1}, Username: "userA"},
				b: &model.User{Model: gorm.Model{ID: 2}, Username: "userB"},
			},
			dbSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec("^INSERT INTO").WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
			wantError: false,
		},
		{
			name: "Scenario 2: Attempt to Follow a User Twice",
			args: args{
				a: &model.User{Model: gorm.Model{ID: 1}, Username: "userA"},
				b: &model.User{Model: gorm.Model{ID: 2}, Username: "userB"},
			},
			dbSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec("^INSERT INTO").WillReturnResult(sqlmock.NewResult(1, 0))
				mock.ExpectCommit()
			},
			wantError: false,
		},
		{
			name: "Scenario 3: Attempt to Follow a Non-Existent User",
			args: args{
				a: &model.User{Model: gorm.Model{ID: 1}, Username: "userA"},
				b: nil,
			},
			dbSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec("^INSERT INTO").WillReturnError(errors.New("user does not exist"))
				mock.ExpectRollback()
			},
			wantError: true,
		},
		{
			name: "Scenario 4: Follow Operation with Database Failure",
			args: args{
				a: &model.User{Model: gorm.Model{ID: 1}, Username: "userA"},
				b: &model.User{Model: gorm.Model{ID: 2}, Username: "userB"},
			},
			dbSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec("^INSERT INTO").WillReturnError(errors.New("some db failure"))
				mock.ExpectRollback()
			},
			wantError: true,
		},
		{
			name: "Scenario 5: Follow User with Null Database Reference",
			args: args{
				a: &model.User{Model: gorm.Model{ID: 1}, Username: "userA"},
				b: &model.User{Model: gorm.Model{ID: 2}, Username: "userB"},
			},
			dbSetup: func(mock sqlmock.Sqlmock) {

			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("Error initializing sqlmock: %s", err)
			}
			defer db.Close()

			gormDB, err := gorm.Open("mysql", db)
			if err != nil {
				t.Fatalf("Error opening gorm DB: %s", err)
			}

			if tt.name == "Scenario 5: Follow User with Null Database Reference" {
				gormDB = nil
			}

			if tt.dbSetup != nil {
				tt.dbSetup(mock)
			}

			userStore := &UserStore{
				db: gormDB,
			}

			err = userStore.Follow(tt.args.a, tt.args.b)

			if (err != nil) != tt.wantError {
				t.Errorf("Follow() error = %v, wantError %v", err, tt.wantError)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("not all expectations were met: %s", err)
			}

			if tt.wantError {
				t.Logf("Expected error occurred: %v", err)
			} else {
				t.Logf("Follow operation successful for %v following %v", tt.args.a.Username, tt.args.b.Username)
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
	t.Run("Scenario 1: Successfully Unfollow a User", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("Failed to open mock sql db, got error: %v", err)
		}
		defer db.Close()

		gormDB, err := gorm.Open("_", db)
		if err != nil {
			t.Fatalf("Failed to open gorm db, got error: %v", err)
		}

		userA := &model.User{Model: gorm.Model{ID: 1}}
		userB := &model.User{Model: gorm.Model{ID: 2}}

		mock.ExpectBegin()
		mock.ExpectExec("DELETE FROM follows WHERE").WithArgs(userA.ID, userB.ID).WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		userStore := &UserStore{db: gormDB}
		err = userStore.Unfollow(userA, userB)

		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}

		t.Log("Success: User A has unfollowed User B")
	})

	t.Run("Scenario 2: Attempt to Unfollow a User Not Being Followed", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("Failed to open mock sql db, got error: %v", err)
		}
		defer db.Close()

		gormDB, err := gorm.Open("_", db)
		if err != nil {
			t.Fatalf("Failed to open gorm db, got error: %v", err)
		}

		userA := &model.User{Model: gorm.Model{ID: 1}}
		userB := &model.User{Model: gorm.Model{ID: 2}}

		mock.ExpectBegin()
		mock.ExpectExec("DELETE FROM follows WHERE").WithArgs(userA.ID, userB.ID).WillReturnResult(sqlmock.NewResult(0, 0))
		mock.ExpectCommit()

		userStore := &UserStore{db: gormDB}
		err = userStore.Unfollow(userA, userB)

		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}

		t.Log("Success: User A was not following User B, and no changes occurred")
	})

	t.Run("Scenario 3: Unfollow with a Nil Database", func(t *testing.T) {
		userA := &model.User{Model: gorm.Model{ID: 1}}
		userB := &model.User{Model: gorm.Model{ID: 2}}

		userStore := &UserStore{db: nil}
		err := userStore.Unfollow(userA, userB)

		if err == nil {
			t.Error("Expected an error due to nil database, but got none")
		}

		t.Log("Success: Properly handled attempt to access nil database")
	})

	t.Run("Scenario 4: Unfollow Non-existent Users", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("Failed to open mock sql db, got error: %v", err)
		}
		defer db.Close()

		gormDB, err := gorm.Open("_", db)
		if err != nil {
			t.Fatalf("Failed to open gorm db, got error: %v", err)
		}

		userA := &model.User{Model: gorm.Model{ID: 9999}}
		userB := &model.User{Model: gorm.Model{ID: 8888}}

		mock.ExpectBegin()
		mock.ExpectExec("DELETE FROM follows WHERE").WithArgs(userA.ID, userB.ID).WillReturnError(gorm.ErrRecordNotFound)
		mock.ExpectRollback()

		userStore := &UserStore{db: gormDB}
		err = userStore.Unfollow(userA, userB)

		if err != gorm.ErrRecordNotFound {
			t.Errorf("Expected record not found error, got: %v", err)
		}

		t.Log("Success: Attempt to unfollow non-existent users handled correctly")
	})

	t.Run("Scenario 5: Database Error on Unfollow", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("Failed to open mock sql db, got error: %v", err)
		}
		defer db.Close()

		gormDB, err := gorm.Open("_", db)
		if err != nil {
			t.Fatalf("Failed to open gorm db, got error: %v", err)
		}

		userA := &model.User{Model: gorm.Model{ID: 1}}
		userB := &model.User{Model: gorm.Model{ID: 2}}

		mock.ExpectBegin()
		mock.ExpectExec("DELETE FROM follows WHERE").WithArgs(userA.ID, userB.ID).WillReturnError(errors.New("simulated db error"))
		mock.ExpectRollback()

		userStore := &UserStore{db: gormDB}
		err = userStore.Unfollow(userA, userB)

		if err == nil || err.Error() != "simulated db error" {
			t.Errorf("Expected simulated db error, got: %v", err)
		}

		t.Log("Success: Simulated database error during unfollow was properly propagated")
	})
}


/*
ROOST_METHOD_HASH=GetFollowingUserIDs_ba3670aa2c
ROOST_METHOD_SIG_HASH=GetFollowingUserIDs_55ccc2afd7

FUNCTION_DEF=func (s *UserStore) GetFollowingUserIDs(m *model.User) ([]uint, error) 

 */
func TestUserStoreGetFollowingUserIDs(t *testing.T) {

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error initializing sqlmock: %v", err)
	}
	defer db.Close()

	gormDB, err := gorm.Open("mysql", db)
	if err != nil {
		t.Fatalf("Error opening gorm DB with mock: %v", err)
	}
	defer gormDB.Close()

	userStore := &UserStore{db: gormDB}

	tests := []struct {
		name        string
		user        *model.User
		setupMock   func()
		expectedIDs []uint
		expectError bool
	}{
		{
			name: "Successfully Retrieve Following User IDs",
			user: &model.User{Model: gorm.Model{ID: 1}},
			setupMock: func() {
				rows := sqlmock.NewRows([]string{"to_user_id"}).AddRow(2).AddRow(3)
				mock.ExpectQuery("^SELECT (.+) FROM follows WHERE").WithArgs(1).WillReturnRows(rows)
			},
			expectedIDs: []uint{2, 3},
		},
		{
			name: "User Follows No Other Users",
			user: &model.User{Model: gorm.Model{ID: 2}},
			setupMock: func() {
				rows := sqlmock.NewRows([]string{"to_user_id"})
				mock.ExpectQuery("^SELECT (.+) FROM follows WHERE").WithArgs(2).WillReturnRows(rows)
			},
			expectedIDs: []uint{},
		},
		{
			name: "Handle Database Query Error",
			user: &model.User{Model: gorm.Model{ID: 3}},
			setupMock: func() {
				mock.ExpectQuery("^SELECT (.+) FROM follows WHERE").WithArgs(3).WillReturnError(gorm.ErrInvalidSQL)
			},
			expectedIDs: []uint{},
			expectError: true,
		},
		{
			name: "Handle SQL Injection Attempt",
			user: &model.User{Model: gorm.Model{ID: 4}},
			setupMock: func() {
				mock.ExpectQuery("^SELECT (.+) FROM follows WHERE").WithArgs(4).WillReturnError(gorm.ErrRecordNotFound)
			},
			expectedIDs: []uint{},
			expectError: true,
		},
		{
			name: "Function Called with Nil User Instance",
			user: nil,
			setupMock: func() {

			},
			expectedIDs: []uint{},
			expectError: true,
		},
		{
			name: "Large Number of Followings",
			user: &model.User{Model: gorm.Model{ID: 5}},
			setupMock: func() {
				rows := sqlmock.NewRows([]string{"to_user_id"})
				for i := 0; i < 1000; i++ {
					rows.AddRow(i)
				}
				mock.ExpectQuery("^SELECT (.+) FROM follows WHERE").WithArgs(5).WillReturnRows(rows)
			},
			expectedIDs: func() []uint {
				var ids []uint
				for i := 0; i < 1000; i++ {
					ids = append(ids, uint(i))
				}
				return ids
			}(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			tt.setupMock()
			ids, err := userStore.GetFollowingUserIDs(tt.user)

			if (err != nil) != tt.expectError {
				t.Errorf("Expected error: %v, got: %v", tt.expectError, err)
			} else if err == nil {
				if len(ids) != len(tt.expectedIDs) {
					t.Errorf("Expected IDs: %v, got: %v", tt.expectedIDs, ids)
				}
			}

			if err = mock.ExpectationsWereMet(); err != nil {
				t.Errorf("Expectations were not met: %v", err)
			}
		})
	}
}

