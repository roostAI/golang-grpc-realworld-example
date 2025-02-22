package store

import (
	debug "runtime/debug"
	testing "testing"
	go-sqlmock "github.com/DATA-DOG/go-sqlmock"
	gorm "github.com/jinzhu/gorm"
	_ "github.com/mattn/go-sqlite3"
	sql "database/sql"
	mysql "github.com/go-sql-driver/mysql"
	model "github.com/raahii/golang-grpc-realworld-example/model"
)








/*
ROOST_METHOD_HASH=NewUserStore_fb599438e5
ROOST_METHOD_SIG_HASH=NewUserStore_c0075221af

FUNCTION_DEF=func NewUserStore(db *gorm.DB) *UserStore // NewUserStore returns a new UserStore


*/
func TestNewUserStore(t *testing.T) {
	tests := []struct {
		name     string
		db       *gorm.DB
		wantNil  bool
		scenario string
	}{
		{
			name:     "Successful UserStore Creation",
			scenario: "Create new UserStore with valid DB connection",
			wantNil:  false,
		},
		{
			name:     "Nil Database Connection",
			db:       nil,
			scenario: "Create new UserStore with nil DB connection",
			wantNil:  true,
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

			t.Logf("Testing Scenario: %s", tt.scenario)

			var db *gorm.DB
			if tt.db == nil && !tt.wantNil {
				sqlDB, mock, err := sqlmock.New()
				if err != nil {
					t.Fatalf("Failed to create mock DB: %v", err)
				}
				defer sqlDB.Close()

				mock.ExpectBegin()

				db, err = gorm.Open("sqlite3", sqlDB)
				if err != nil {
					t.Fatalf("Failed to open gorm connection: %v", err)
				}
				defer db.Close()
			} else {
				db = tt.db
			}

			got := NewUserStore(db)

			if (got == nil) != tt.wantNil {
				t.Errorf("NewUserStore() returned nil: %v, want nil: %v", got == nil, tt.wantNil)
				return
			}

			if !tt.wantNil && got != nil {
				if got.db != db {
					t.Errorf("NewUserStore().db = %v, want %v", got.db, db)
				}
				t.Logf("Successfully created UserStore with DB connection")

				store2 := NewUserStore(db)
				if store2 == nil {
					t.Error("Second NewUserStore() call returned nil")
				}
				if got == store2 {
					t.Error("Multiple UserStore instances should be different objects")
				}
				if got.db != store2.db {
					t.Error("Database connections should be the same for stores created with same DB")
				}
				t.Logf("Successfully verified UserStore instance independence")
			}
		})
	}
}


/*
ROOST_METHOD_HASH=UserStore_Create_9495ddb29d
ROOST_METHOD_SIG_HASH=UserStore_Create_18451817fe

FUNCTION_DEF=func (s *UserStore) Create(m *model.User) error // Create create a user


*/
func TestUserStoreCreate(t *testing.T) {

	type testCase struct {
		name    string
		user    *model.User
		dbSetup func(mock sqlmock.Sqlmock)
		wantErr bool
	}

	tests := []testCase{
		{
			name: "Successful User Creation",
			user: &model.User{
				Username: "testuser",
				Email:    "test@example.com",
				Password: "password123",
				Bio:      "Test bio",
				Image:    "test.jpg",
			},
			dbSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO `users`").
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
			wantErr: false,
		},
		{
			name: "Duplicate Username",
			user: &model.User{
				Username: "existinguser",
				Email:    "new@example.com",
				Password: "password123",
				Bio:      "Test bio",
				Image:    "test.jpg",
			},
			dbSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO `users`").
					WillReturnError(&mysql.MySQLError{Number: 1062})
				mock.ExpectRollback()
			},
			wantErr: true,
		},
		{
			name: "Null Required Fields",
			user: &model.User{
				Username: "",
				Email:    "",
				Password: "",
			},
			dbSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO `users`").
					WillReturnError(&mysql.MySQLError{Number: 1048})
				mock.ExpectRollback()
			},
			wantErr: true,
		},
		{
			name: "Database Connection Failure",
			user: &model.User{
				Username: "testuser",
				Email:    "test@example.com",
				Password: "password123",
			},
			dbSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO `users`").
					WillReturnError(sql.ErrConnDone)
				mock.ExpectRollback()
			},
			wantErr: true,
		},
		{
			name: "Special Characters in Fields",
			user: &model.User{
				Username: "test@user#123",
				Email:    "test+special@example.com",
				Password: "pass!@#$%^&*()",
				Bio:      "Bio with 特殊字符",
				Image:    "image-!@#.jpg",
			},
			dbSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO `users`").
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
			wantErr: false,
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
				t.Fatalf("Failed to create mock DB: %v", err)
			}
			defer db.Close()

			gormDB, err := gorm.Open("mysql", db)
			if err != nil {
				t.Fatalf("Failed to open GORM DB: %v", err)
			}
			defer gormDB.Close()

			tt.dbSetup(mock)

			store := &UserStore{
				db: gormDB,
			}

			err = store.Create(tt.user)

			if (err != nil) != tt.wantErr {
				t.Errorf("UserStore.Create() error = %v, wantErr %v", err, tt.wantErr)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("Unfulfilled expectations: %v", err)
			}

			if err != nil {
				t.Logf("Test '%s' failed with error: %v", tt.name, err)
			} else {
				t.Logf("Test '%s' passed successfully", tt.name)
			}
		})
	}
}


/*
ROOST_METHOD_HASH=UserStore_Follow_fe0976e4eb
ROOST_METHOD_SIG_HASH=UserStore_Follow_0e703b23f8

FUNCTION_DEF=func (s *UserStore) Follow(a *model.User, b *model.User) error // Follow create follow relashionship to User B from user A


*/
func TestUserStoreFollow(t *testing.T) {

	db, mock, err := go-sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock DB: %v", err)
	}
	defer db.Close()

	gormDB, err := gorm.Open("mysql", db)
	if err != nil {
		t.Fatalf("Failed to open GORM DB: %v", err)
	}
	defer gormDB.Close()

	store := &UserStore{db: gormDB}

	tests := []struct {
		name    string
		userA   *model.User
		userB   *model.User
		mockSQL func()
		wantErr bool
	}{
		{
			name: "Successful Follow",
			userA: &model.User{
				Model:    gorm.Model{ID: 1},
				Username: "userA",
				Email:    "userA@test.com",
			},
			userB: &model.User{
				Model:    gorm.Model{ID: 2},
				Username: "userB",
				Email:    "userB@test.com",
			},
			mockSQL: func() {
				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO `follows`").
					WithArgs(1, 2).
					WillReturnResult(go-sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
			wantErr: false,
		},
		{
			name:  "Follow with nil User A",
			userA: nil,
			userB: &model.User{
				Model:    gorm.Model{ID: 2},
				Username: "userB",
				Email:    "userB@test.com",
			},
			mockSQL: func() {

			},
			wantErr: true,
		},
		{
			name: "Follow with nil User B",
			userA: &model.User{
				Model:    gorm.Model{ID: 1},
				Username: "userA",
				Email:    "userA@test.com",
			},
			userB:   nil,
			mockSQL: func() {},
			wantErr: true,
		},
		{
			name: "Database Error",
			userA: &model.User{
				Model:    gorm.Model{ID: 1},
				Username: "userA",
				Email:    "userA@test.com",
			},
			userB: &model.User{
				Model:    gorm.Model{ID: 2},
				Username: "userB",
				Email:    "userB@test.com",
			},
			mockSQL: func() {
				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO `follows`").
					WithArgs(1, 2).
					WillReturnError(gorm.ErrInvalidSQL)
				mock.ExpectRollback()
			},
			wantErr: true,
		},
		{
			name: "Self Follow",
			userA: &model.User{
				Model:    gorm.Model{ID: 1},
				Username: "userA",
				Email:    "userA@test.com",
			},
			userB: &model.User{
				Model:    gorm.Model{ID: 1},
				Username: "userA",
				Email:    "userA@test.com",
			},
			mockSQL: func() {
				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO `follows`").
					WithArgs(1, 1).
					WillReturnResult(go-sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
			wantErr: false,
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

			t.Log("Starting test case:", tt.name)

			tt.mockSQL()

			err := store.Follow(tt.userA, tt.userB)

			if (err != nil) != tt.wantErr {
				t.Errorf("UserStore.Follow() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("Unfulfilled expectations: %v", err)
			}

			t.Log("Test case completed successfully")
		})
	}
}


/*
ROOST_METHOD_HASH=UserStore_GetByEmail_fda09af5c4
ROOST_METHOD_SIG_HASH=UserStore_GetByEmail_9e84f3286b

FUNCTION_DEF=func (s *UserStore) GetByEmail(email string) (*model.User, error) // GetByEmail finds a user from email


*/
func TestUserStoreGetByEmail(t *testing.T) {

	type testCase struct {
		name          string
		email         string
		mockSetup     func(mock sqlmock.Sqlmock)
		expectedError error
		expectedUser  *model.User
	}

	tests := []testCase{
		{
			name:  "Successfully retrieve user by valid email",
			email: "test@example.com",
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "username", "email", "password", "bio", "image"}).
					AddRow(1, "testuser", "test@example.com", "hashedpass", "test bio", "image.jpg")
				mock.ExpectQuery(`SELECT \* FROM "users"`).
					WithArgs("test@example.com").
					WillReturnRows(rows)
			},
			expectedError: nil,
			expectedUser: &model.User{
				Model: gorm.Model{ID: 1},
				Email: "test@example.com",
			},
		},
		{
			name:  "Handle non-existent email",
			email: "nonexistent@example.com",
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT \* FROM "users"`).
					WithArgs("nonexistent@example.com").
					WillReturnError(gorm.ErrRecordNotFound)
			},
			expectedError: gorm.ErrRecordNotFound,
			expectedUser:  nil,
		},
		{
			name:  "Handle empty email parameter",
			email: "",
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT \* FROM "users"`).
					WithArgs("").
					WillReturnError(gorm.ErrRecordNotFound)
			},
			expectedError: gorm.ErrRecordNotFound,
			expectedUser:  nil,
		},
		{
			name:  "Handle database connection error",
			email: "test@example.com",
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT \* FROM "users"`).
					WithArgs("test@example.com").
					WillReturnError(sql.ErrConnDone)
			},
			expectedError: sql.ErrConnDone,
			expectedUser:  nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {

			defer func() {
				if r := recover(); r != nil {
					t.Logf("Panic encountered so failing test. %v\n%s", r, string(debug.Stack()))
					t.Fail()
				}
			}()

			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("Failed to create mock DB: %v", err)
			}
			defer db.Close()

			gormDB, err := gorm.Open("mysql", db)
			if err != nil {
				t.Fatalf("Failed to open GORM DB: %v", err)
			}
			defer gormDB.Close()

			tc.mockSetup(mock)

			store := &UserStore{
				db: gormDB,
			}

			user, err := store.GetByEmail(tc.email)

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("Unfulfilled expectations: %v", err)
			}

			if tc.expectedError != nil {
				if err == nil {
					t.Errorf("Expected error %v but got nil", tc.expectedError)
				} else if err != tc.expectedError {
					t.Errorf("Expected error %v but got %v", tc.expectedError, err)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got %v", err)
				}
			}

			if tc.expectedUser != nil {
				if user == nil {
					t.Error("Expected user but got nil")
				} else if user.Email != tc.expectedUser.Email {
					t.Errorf("Expected user email %s but got %s", tc.expectedUser.Email, user.Email)
				}
			} else if user != nil {
				t.Error("Expected nil user but got a user")
			}

			t.Logf("Test case '%s' completed successfully", tc.name)
		})
	}
}


/*
ROOST_METHOD_HASH=UserStore_GetByID_1f5f06165b
ROOST_METHOD_SIG_HASH=UserStore_GetByID_2a864916bb

FUNCTION_DEF=func (s *UserStore) GetByID(id uint) (*model.User, error) // GetByID finds a user from id


*/
func TestUserStoreGetById(t *testing.T) {

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer db.Close()

	gormDB, err := gorm.Open("mysql", db)
	if err != nil {
		t.Fatalf("Failed to open GORM connection: %v", err)
	}
	defer gormDB.Close()

	userStore := &UserStore{
		db: gormDB,
	}

	tests := []struct {
		name     string
		userID   uint
		mockFunc func(sqlmock.Sqlmock)
		want     *model.User
		wantErr  error
	}{
		{
			name:   "Successfully retrieve existing user",
			userID: 1,
			mockFunc: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "username", "email", "password", "bio", "image"}).
					AddRow(1, "testuser", "test@example.com", "hashedpass", "test bio", "image.jpg")
				mock.ExpectQuery("^SELECT (.+) FROM `users`").
					WithArgs(1).
					WillReturnRows(rows)
			},
			want: &model.User{
				Model:    gorm.Model{ID: 1},
				Username: "testuser",
				Email:    "test@example.com",
				Password: "hashedpass",
				Bio:      "test bio",
				Image:    "image.jpg",
			},
			wantErr: nil,
		},
		{
			name:   "User not found",
			userID: 999,
			mockFunc: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("^SELECT (.+) FROM `users`").
					WithArgs(999).
					WillReturnError(gorm.ErrRecordNotFound)
			},
			want:    nil,
			wantErr: gorm.ErrRecordNotFound,
		},
		{
			name:   "Database error",
			userID: 2,
			mockFunc: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("^SELECT (.+) FROM `users`").
					WithArgs(2).
					WillReturnError(sql.ErrConnDone)
			},
			want:    nil,
			wantErr: sql.ErrConnDone,
		},
		{
			name:   "Zero ID",
			userID: 0,
			mockFunc: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("^SELECT (.+) FROM `users`").
					WithArgs(0).
					WillReturnError(gorm.ErrRecordNotFound)
			},
			want:    nil,
			wantErr: gorm.ErrRecordNotFound,
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

			t.Logf("Running test case: %s", tt.name)

			tt.mockFunc(mock)

			got, err := userStore.GetByID(tt.userID)

			if (err != nil) != (tt.wantErr != nil) {
				t.Errorf("UserStore.GetByID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err != nil && tt.wantErr != nil && err.Error() != tt.wantErr.Error() {
				t.Errorf("UserStore.GetByID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if got != nil && tt.want != nil {
				if got.ID != tt.want.ID ||
					got.Username != tt.want.Username ||
					got.Email != tt.want.Email ||
					got.Password != tt.want.Password ||
					got.Bio != tt.want.Bio ||
					got.Image != tt.want.Image {
					t.Errorf("UserStore.GetByID() = %v, want %v", got, tt.want)
				}
			} else if (got == nil) != (tt.want == nil) {
				t.Errorf("UserStore.GetByID() = %v, want %v", got, tt.want)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("Unfulfilled expectations: %s", err)
			}

			t.Logf("Test case completed successfully: %s", tt.name)
		})
	}
}


/*
ROOST_METHOD_HASH=UserStore_GetByUsername_622b1b9e41
ROOST_METHOD_SIG_HASH=UserStore_GetByUsername_992f00baec

FUNCTION_DEF=func (s *UserStore) GetByUsername(username string) (*model.User, error) // GetByUsername finds a user from username


*/
func TestUserStoreGetByUsername(t *testing.T) {

	type testCase struct {
		name      string
		username  string
		mockSetup func(mock sqlmock.Sqlmock)
		validate  func(*testing.T, *model.User, error)
	}

	tests := []testCase{
		{
			name:     "Successfully retrieve user by valid username",
			username: "testuser",
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "username", "email", "password", "bio", "image"}).
					AddRow(1, "testuser", "test@example.com", "hashedpass", "bio", "image.jpg")
				mock.ExpectQuery(`SELECT \* FROM "users"`).
					WithArgs("testuser").
					WillReturnRows(rows)
			},
			validate: func(t *testing.T, user *model.User, err error) {
				if err != nil {
					t.Errorf("expected no error, got %v", err)
				}
				if user == nil {
					t.Error("expected user to not be nil")
					return
				}
				if user.Username != "testuser" {
					t.Errorf("expected username 'testuser', got %s", user.Username)
				}
			},
		},
		{
			name:     "Username not found in database",
			username: "nonexistent",
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT \* FROM "users"`).
					WithArgs("nonexistent").
					WillReturnError(gorm.ErrRecordNotFound)
			},
			validate: func(t *testing.T, user *model.User, err error) {
				if err != gorm.ErrRecordNotFound {
					t.Errorf("expected ErrRecordNotFound, got %v", err)
				}
				if user != nil {
					t.Error("expected user to be nil")
				}
			},
		},
		{
			name:     "Database connection error",
			username: "testuser",
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT \* FROM "users"`).
					WithArgs("testuser").
					WillReturnError(sql.ErrConnDone)
			},
			validate: func(t *testing.T, user *model.User, err error) {
				if err != sql.ErrConnDone {
					t.Errorf("expected sql.ErrConnDone, got %v", err)
				}
				if user != nil {
					t.Error("expected user to be nil")
				}
			},
		},
		{
			name:     "Empty username parameter",
			username: "",
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT \* FROM "users"`).
					WithArgs("").
					WillReturnError(gorm.ErrRecordNotFound)
			},
			validate: func(t *testing.T, user *model.User, err error) {
				if err != gorm.ErrRecordNotFound {
					t.Errorf("expected ErrRecordNotFound, got %v", err)
				}
				if user != nil {
					t.Error("expected user to be nil")
				}
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {

			defer func() {
				if r := recover(); r != nil {
					t.Logf("Panic encountered so failing test. %v\n%s", r, string(debug.Stack()))
					t.Fail()
				}
			}()

			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("failed to create mock db: %v", err)
			}
			defer db.Close()

			gormDB, err := gorm.Open("mysql", db)
			if err != nil {
				t.Fatalf("failed to create gorm db: %v", err)
			}
			defer gormDB.Close()

			tc.mockSetup(mock)

			store := &UserStore{db: gormDB}

			t.Logf("Testing scenario: %s", tc.name)
			user, err := store.GetByUsername(tc.username)

			tc.validate(t, user, err)

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
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

	tests := []struct {
		name     string
		userID   uint
		mockFunc func(mock sqlmock.Sqlmock)
		want     []uint
		wantErr  bool
	}{
		{
			name:   "Success - Multiple Following Users",
			userID: 1,
			mockFunc: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"to_user_id"}).
					AddRow(2).
					AddRow(3).
					AddRow(4)
				mock.ExpectQuery("SELECT (.+) FROM `follows` WHERE").
					WithArgs(1).
					WillReturnRows(rows)
			},
			want:    []uint{2, 3, 4},
			wantErr: false,
		},
		{
			name:   "Success - Empty Following List",
			userID: 1,
			mockFunc: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"to_user_id"})
				mock.ExpectQuery("SELECT (.+) FROM `follows` WHERE").
					WithArgs(1).
					WillReturnRows(rows)
			},
			want:    []uint{},
			wantErr: false,
		},
		{
			name:   "Error - Database Error",
			userID: 1,
			mockFunc: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT (.+) FROM `follows` WHERE").
					WithArgs(1).
					WillReturnError(sql.ErrConnDone)
			},
			want:    []uint{},
			wantErr: true,
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
				t.Fatalf("Failed to create mock DB: %v", err)
			}
			defer db.Close()

			gormDB, err := gorm.Open("mysql", db)
			if err != nil {
				t.Fatalf("Failed to create GORM DB: %v", err)
			}
			defer gormDB.Close()

			tt.mockFunc(mock)

			store := &UserStore{
				db: gormDB,
			}

			user := &model.User{
				Model: gorm.Model{ID: tt.userID},
			}

			got, err := store.GetFollowingUserIDs(user)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetFollowingUserIDs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if len(got) != len(tt.want) {
				t.Errorf("GetFollowingUserIDs() got = %v, want %v", got, tt.want)
				return
			}

			for i, id := range got {
				if id != tt.want[i] {
					t.Errorf("GetFollowingUserIDs() got[%d] = %v, want[%d] = %v", i, id, i, tt.want[i])
				}
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("Unfulfilled expectations: %v", err)
			}

			t.Logf("Test '%s' completed successfully", tt.name)
		})
	}
}


/*
ROOST_METHOD_HASH=UserStore_IsFollowing_309135b43a
ROOST_METHOD_SIG_HASH=UserStore_IsFollowing_4644b1529c

FUNCTION_DEF=func (s *UserStore) IsFollowing(a *model.User, b *model.User) (bool, error) // IsFollowing returns whether user A follows user B or not


*/
func TestUserStoreIsFollowing(t *testing.T) {

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer db.Close()

	gormDB, err := gorm.Open("mysql", db)
	if err != nil {
		t.Fatalf("Failed to open GORM connection: %v", err)
	}
	defer gormDB.Close()

	store := &UserStore{db: gormDB}

	tests := []struct {
		name     string
		userA    *model.User
		userB    *model.User
		mockFunc func()
		want     bool
		wantErr  bool
	}{
		{
			name:  "Valid Following Relationship",
			userA: &model.User{Model: gorm.Model{ID: 1}},
			userB: &model.User{Model: gorm.Model{ID: 2}},
			mockFunc: func() {
				mock.ExpectQuery("SELECT count\\(\\*\\) FROM `follows`").
					WithArgs(1, 2).
					WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
			},
			want:    true,
			wantErr: false,
		},
		{
			name:  "Non-Existent Following Relationship",
			userA: &model.User{Model: gorm.Model{ID: 1}},
			userB: &model.User{Model: gorm.Model{ID: 2}},
			mockFunc: func() {
				mock.ExpectQuery("SELECT count\\(\\*\\) FROM `follows`").
					WithArgs(1, 2).
					WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
			},
			want:    false,
			wantErr: false,
		},
		{
			name:     "Nil User A",
			userA:    nil,
			userB:    &model.User{Model: gorm.Model{ID: 2}},
			mockFunc: func() {},
			want:     false,
			wantErr:  false,
		},
		{
			name:     "Nil User B",
			userA:    &model.User{Model: gorm.Model{ID: 1}},
			userB:    nil,
			mockFunc: func() {},
			want:     false,
			wantErr:  false,
		},
		{
			name:  "Database Error",
			userA: &model.User{Model: gorm.Model{ID: 1}},
			userB: &model.User{Model: gorm.Model{ID: 2}},
			mockFunc: func() {
				mock.ExpectQuery("SELECT count\\(\\*\\) FROM `follows`").
					WithArgs(1, 2).
					WillReturnError(sql.ErrConnDone)
			},
			want:    false,
			wantErr: true,
		},
		{
			name:  "Same User Check",
			userA: &model.User{Model: gorm.Model{ID: 1}},
			userB: &model.User{Model: gorm.Model{ID: 1}},
			mockFunc: func() {
				mock.ExpectQuery("SELECT count\\(\\*\\) FROM `follows`").
					WithArgs(1, 1).
					WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
			},
			want:    false,
			wantErr: false,
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

			t.Logf("Testing scenario: %s", tt.name)

			tt.mockFunc()

			got, err := store.IsFollowing(tt.userA, tt.userB)

			if (err != nil) != tt.wantErr {
				t.Errorf("IsFollowing() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("IsFollowing() = %v, want %v", got, tt.want)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("Unfulfilled expectations: %s", err)
			}

			t.Logf("Test completed successfully")
		})
	}
}


/*
ROOST_METHOD_HASH=UserStore_Unfollow_29d3ef7f50
ROOST_METHOD_SIG_HASH=UserStore_Unfollow_31d9214353

FUNCTION_DEF=func (s *UserStore) Unfollow(a *model.User, b *model.User) error // Unfollow delete follow relashionship to User B from user A


*/
func TestUserStoreUnfollow(t *testing.T) {

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer db.Close()

	gormDB, err := gorm.Open("mysql", db)
	if err != nil {
		t.Fatalf("Failed to open gorm connection: %v", err)
	}
	defer gormDB.Close()

	store := &UserStore{db: gormDB}

	tests := []struct {
		name    string
		userA   *model.User
		userB   *model.User
		mockSQL func()
		wantErr bool
	}{
		{
			name: "Successful unfollow",
			userA: &model.User{
				Model:    gorm.Model{ID: 1},
				Username: "userA",
				Email:    "userA@test.com",
			},
			userB: &model.User{
				Model:    gorm.Model{ID: 2},
				Username: "userB",
				Email:    "userB@test.com",
			},
			mockSQL: func() {
				mock.ExpectBegin()
				mock.ExpectExec("DELETE FROM `follows`").
					WithArgs(1, 2).
					WillReturnResult(sqlmock.NewResult(0, 1))
				mock.ExpectCommit()
			},
			wantErr: false,
		},
		{
			name:  "Invalid userA (nil)",
			userA: nil,
			userB: &model.User{
				Model:    gorm.Model{ID: 2},
				Username: "userB",
				Email:    "userB@test.com",
			},
			mockSQL: func() {

			},
			wantErr: true,
		},
		{
			name: "Invalid userB (nil)",
			userA: &model.User{
				Model:    gorm.Model{ID: 1},
				Username: "userA",
				Email:    "userA@test.com",
			},
			userB: nil,
			mockSQL: func() {

			},
			wantErr: true,
		},
		{
			name: "Database error",
			userA: &model.User{
				Model:    gorm.Model{ID: 1},
				Username: "userA",
				Email:    "userA@test.com",
			},
			userB: &model.User{
				Model:    gorm.Model{ID: 2},
				Username: "userB",
				Email:    "userB@test.com",
			},
			mockSQL: func() {
				mock.ExpectBegin()
				mock.ExpectExec("DELETE FROM `follows`").
					WithArgs(1, 2).
					WillReturnError(gorm.ErrInvalidSQL)
				mock.ExpectRollback()
			},
			wantErr: true,
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

			t.Logf("Running test case: %s", tt.name)

			tt.mockSQL()

			err := store.Unfollow(tt.userA, tt.userB)

			if (err != nil) != tt.wantErr {
				t.Errorf("UserStore.Unfollow() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("Unfulfilled expectations: %v", err)
			}

			if err != nil {
				t.Logf("Got expected error: %v", err)
			} else {
				t.Log("Successfully executed unfollow operation")
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
		name    string
		user    *model.User
		setupDB func(mock sqlmock.Sqlmock)
		wantErr bool
	}

	tests := []testCase{
		{
			name: "Successful Update",
			user: &model.User{
				Model:    gorm.Model{ID: 1},
				Username: "updateduser",
				Email:    "updated@example.com",
				Password: "newpassword",
				Bio:      "Updated bio",
				Image:    "updated.jpg",
			},
			setupDB: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec("UPDATE `users`").
					WithArgs(
						sqlmock.AnyArg(),
						"updateduser",
						"updated@example.com",
						"newpassword",
						"Updated bio",
						"updated.jpg",
						1,
					).WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
			wantErr: false,
		},
		{
			name: "Update Non-Existent User",
			user: &model.User{
				Model:    gorm.Model{ID: 999},
				Username: "nonexistent",
				Email:    "nonexistent@example.com",
			},
			setupDB: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec("UPDATE `users`").
					WillReturnError(gorm.ErrRecordNotFound)
				mock.ExpectRollback()
			},
			wantErr: true,
		},
		{
			name: "Update with Duplicate Email",
			user: &model.User{
				Model:    gorm.Model{ID: 1},
				Username: "user1",
				Email:    "duplicate@example.com",
			},
			setupDB: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec("UPDATE `users`").
					WillReturnError(gorm.ErrInvalidTransaction)
				mock.ExpectRollback()
			},
			wantErr: true,
		},
		{
			name: "Update with Empty User",
			user: &model.User{},
			setupDB: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec("UPDATE `users`").
					WillReturnError(gorm.ErrInvalidSQL)
				mock.ExpectRollback()
			},
			wantErr: true,
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
				t.Fatalf("Failed to create mock DB: %v", err)
			}
			defer db.Close()

			gormDB, err := gorm.Open("mysql", db)
			if err != nil {
				t.Fatalf("Failed to open GORM DB: %v", err)
			}
			defer gormDB.Close()

			tt.setupDB(mock)

			store := &UserStore{
				db: gormDB,
			}

			err = store.Update(tt.user)

			if (err != nil) != tt.wantErr {
				t.Errorf("UserStore.Update() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("Unfulfilled expectations: %v", err)
			}

			if err != nil {
				t.Logf("Test '%s' completed with expected error: %v", tt.name, err)
			} else {
				t.Logf("Test '%s' completed successfully", tt.name)
			}
		})
	}
}

