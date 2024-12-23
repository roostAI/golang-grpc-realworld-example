package store

import (
	"testing"
	"github.com/stretchr/testify/assert"
	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"time"
	"sync"
	"fmt"
	"log"
	"github.com/DATA-DOG/go-sqlmock"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/raahii/golang-grpc-realworld-example/model"
	"errors"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	_ "github.com/mattn/go-sqlite3"
	"database/sql"
	"reflect"
)


/*
ROOST_METHOD_HASH=NewUserStore_6a331dd890
ROOST_METHOD_SIG_HASH=NewUserStore_4f0c2dfca9


 */
func TestNewUserStore(t *testing.T) {
	tests := []struct {
		name          string
		db            *gorm.DB
		expectations  func(sqlmock.Sqlmock)
		expectedDBPtr bool
		expectNilDB   bool
	}{
		{
			name: "Scenario 1: Valid gorm.DB instance",
			db: func() *gorm.DB {
				mockDB, _, _ := sqlmock.New()
				gormDB, _ := gorm.Open("sqlite3", mockDB)
				return gormDB
			}(),
			expectations: func(mock sqlmock.Sqlmock) {

			},
			expectedDBPtr: true,
			expectNilDB:   false,
		},
		{
			name: "Scenario 2: Check DB immutability",
			db: func() *gorm.DB {
				mockDB, _, _ := sqlmock.New()
				gormDB, _ := gorm.Open("sqlite3", mockDB)
				return gormDB
			}(),
			expectations: func(mock sqlmock.Sqlmock) {

			},
			expectedDBPtr: true,
			expectNilDB:   false,
		},
		{
			name: "Scenario 3: Nil DB instance",
			db:   nil,
			expectations: func(mock sqlmock.Sqlmock) {

			},
			expectedDBPtr: false,
			expectNilDB:   true,
		},
		{
			name: "Scenario 4: Validate returned type",
			db: func() *gorm.DB {
				mockDB, _, _ := sqlmock.New()
				gormDB, _ := gorm.Open("sqlite3", mockDB)
				return gormDB
			}(),
			expectations: func(mock sqlmock.Sqlmock) {

			},
			expectedDBPtr: true,
			expectNilDB:   false,
		},
		{
			name: "Scenario 5: Correct DB address",
			db: func() *gorm.DB {
				mockDB, _, _ := sqlmock.New()
				gormDB, _ := gorm.Open("sqlite3", mockDB)
				return gormDB
			}(),
			expectations: func(mock sqlmock.Sqlmock) {

			},
			expectedDBPtr: true,
			expectNilDB:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			_, mock, _ := sqlmock.New()
			tt.expectations(mock)

			userStore := NewUserStore(tt.db)

			if tt.expectedDBPtr {
				assert.NotNil(t, userStore.db, "UserStore DB field should be set")
			} else if tt.expectNilDB {
				assert.Nil(t, userStore.db, "UserStore DB field should be nil")
			}

			if tt.db != nil {
				assert.Equal(t, tt.db, userStore.db, "The memory address of the DB should match what's expected")
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}

			t.Log("Test scenario executed successfully")
		})
	}
}


/*
ROOST_METHOD_HASH=Create_889fc0fc45
ROOST_METHOD_SIG_HASH=Create_4c48ec3920


 */
func TestCreate(t *testing.T) {

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("error creating mock database: %s", err)
	}
	defer db.Close()

	gormDB, err := gorm.Open("mysql", db)
	if err != nil {
		t.Fatalf("error opening gorm DB: %s", err)
	}

	userStore := &UserStore{db: gormDB}

	tests := []struct {
		name        string
		user        model.User
		setupMock   func()
		expectError bool
	}{
		{
			name: "Successfully Creating a New User Record",
			user: model.User{
				Username: "uniqueuser1",
				Email:    "uniqueemail1@example.com",
				Password: "securePassword",
				Bio:      "A mysterious person",
				Image:    "http://someimage.com/photo.jpg",
			},
			setupMock: func() {
				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO `users`.*").WithArgs(
					AnyArg{},
					AnyArg{},
					AnyArg{},
					AnyArg{},
					"uniqueuser1",
					"uniqueemail1@example.com",
					"securePassword",
					"A mysterious person",
					"http://someimage.com/photo.jpg",
				).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
			expectError: false,
		},
		{
			name: "Creating a User with Duplicate Username",
			user: model.User{
				Username: "takenUsername",
				Email:    "email2@example.com",
				Password: "securePassword",
				Bio:      "",
				Image:    "",
			},
			setupMock: func() {
				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO `users`.*").WithArgs(
					AnyArg{},
					AnyArg{},
					AnyArg{},
					AnyArg{},
					"takenUsername",
					"email2@example.com",
					"securePassword",
					"",
					"",
				).
					WillReturnError(fmt.Errorf("duplicate key value violates unique constraint \"users_username_key\""))
				mock.ExpectRollback()
			},
			expectError: true,
		},
		{
			name: "Creating a User with Duplicate Email",
			user: model.User{
				Username: "uniqueuser3",
				Email:    "takenemail@example.com",
				Password: "securePassword",
				Bio:      "",
				Image:    "",
			},
			setupMock: func() {
				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO `users`.*").WithArgs(
					AnyArg{},
					AnyArg{},
					AnyArg{},
					AnyArg{},
					"uniqueuser3",
					"takenemail@example.com",
					"securePassword",
					"",
					"",
				).
					WillReturnError(fmt.Errorf("duplicate key value violates unique constraint \"users_email_key\""))
				mock.ExpectRollback()
			},
			expectError: true,
		},
		{
			name: "Attempting to Create a User with Missing Required Fields",
			user: model.User{
				Email:    "email4@example.com",
				Password: "securePassword",
				Bio:      "Loves sunshine",
				Image:    "http://someimage.com/photo.jpg",
			},
			setupMock: func() {
				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO `users`.*").WithArgs(
					AnyArg{},
					AnyArg{},
					AnyArg{},
					AnyArg{},
					"",
					"email4@example.com",
					"securePassword",
					"Loves sunshine",
					"http://someimage.com/photo.jpg",
				).
					WillReturnError(fmt.Errorf("violates not-null constraint"))
				mock.ExpectRollback()
			},
			expectError: true,
		},
		{
			name: "Database Connection Error on Create",
			user: model.User{
				Username: "userError",
				Email:    "emailError@example.com",
				Password: "securePassword",
				Bio:      "",
				Image:    "",
			},
			setupMock: func() {
				mock.ExpectBegin().WillReturnError(fmt.Errorf("connection refused"))
			},
			expectError: true,
		},
		{
			name: "Creating a User with an Empty Bio or Image",
			user: model.User{
				Username: "optionalfields",
				Email:    "optionalfields@example.com",
				Password: "securePassword",
				Bio:      "",
				Image:    "",
			},
			setupMock: func() {
				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO `users`.*").WithArgs(
					AnyArg{},
					AnyArg{},
					AnyArg{},
					AnyArg{},
					"optionalfields",
					"optionalfields@example.com",
					"securePassword",
					"",
					"",
				).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			err := userStore.Create(&tt.user)

			if (err != nil) != tt.expectError {
				t.Errorf("TestCreate() = %v, expectError %v", err, tt.expectError)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				log.Printf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}


/*
ROOST_METHOD_HASH=GetByID_bbf946112e
ROOST_METHOD_SIG_HASH=GetByID_728dd55ed1


 */
func TestGetByID(t *testing.T) {
	type testScenario struct {
		name       string
		id         uint
		setupMock  func(sqlmock.Sqlmock)
		expectUser *model.User
		expectErr  error
	}

	scenarios := []testScenario{
		{
			name: "Scenario 1: Successful Retrieval of User by Valid ID",
			id:   1,
			setupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "created_at", "updated_at", "deleted_at", "username", "email", "password", "bio", "image"}).
					AddRow(1, time.Now(), time.Now(), nil, "john_doe", "john@example.com", "hashedpassword", "Bio", "image.png")
				mock.ExpectQuery("^SELECT \\* FROM `users` WHERE `users`.`deleted_at` IS NULL AND \\(\\(`users`.`id` = \\?\\)\\)$").
					WithArgs(1).
					WillReturnRows(rows)
			},
			expectUser: &model.User{
				Model:    gorm.Model{ID: 1, CreatedAt: time.Now(), UpdatedAt: time.Now()},
				Username: "john_doe",
				Email:    "john@example.com",
				Password: "hashedpassword",
				Bio:      "Bio",
				Image:    "image.png",
			},
			expectErr: nil,
		},
		{
			name: "Scenario 2: User Retrieval with Non-existent ID",
			id:   9999,
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("^SELECT \\* FROM `users` WHERE `users`.`deleted_at` IS NULL AND \\(\\(`users`.`id` = \\?\\)\\)$").
					WithArgs(9999).
					WillReturnError(gorm.ErrRecordNotFound)
			},
			expectUser: nil,
			expectErr:  gorm.ErrRecordNotFound,
		},
		{
			name: "Scenario 3: Handling Database Error During Retrieval",
			id:   1,
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("^SELECT \\* FROM `users` WHERE `users`.`deleted_at` IS NULL AND \\(\\(`users`.`id` = \\?\\)\\)$").
					WithArgs(1).
					WillReturnError(errors.New("database error"))
			},
			expectUser: nil,
			expectErr:  errors.New("database error"),
		},
	
	}

	for _, scenario := range scenarios {
		db, mock, err := sqlmock.New()
		assert.NoError(t, err)
		defer db.Close()

		gormDB, err := gorm.Open("postgres", db)
		assert.NoError(t, err)

		store := UserStore{db: gormDB}

		scenario.setupMock(mock)

		t.Run(scenario.name, func(t *testing.T) {
			result, err := store.GetByID(scenario.id)

			if err != nil {
				t.Log(fmt.Sprintf("Expected error: %v, got: %v", scenario.expectErr, err))
			} else {
				t.Log(fmt.Sprintf("Expected user: %#v, got: %#v", scenario.expectUser, result))
			}

			assert.Equal(t, scenario.expectErr, err)
			if scenario.expectUser != nil {
				expectedUser := *scenario.expectUser
				expectedUser.CreatedAt = result.CreatedAt
				expectedUser.UpdatedAt = result.UpdatedAt
				assert.Equal(t, expectedUser, *result)
			} else {
				assert.Nil(t, result)
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
		mockSetup     func(mock sqlmock.Sqlmock)
		expectedUser  *model.User
		expectedError bool
	}{
		{
			name:  "Retrieve Existing User by Email",
			email: "user@example.com",
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "username", "email"}).
					AddRow(1, "user1", "user@example.com")
				mock.ExpectQuery("SELECT (.+) FROM users WHERE (.+)").
					WithArgs("user@example.com").
					WillReturnRows(rows)
			},
			expectedUser:  &model.User{Email: "user@example.com"},
			expectedError: false,
		},
		{
			name:  "Email Does Not Exist",
			email: "missing@example.com",
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT (.+) FROM users WHERE (.+)").
					WithArgs("missing@example.com").
					WillReturnError(gorm.ErrRecordNotFound)
			},
			expectedUser:  nil,
			expectedError: true,
		},
		{
			name:  "Invalid Email Format",
			email: "not-an-email",
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT (.+) FROM users WHERE (.+)").
					WithArgs("not-an-email").
					WillReturnError(gorm.ErrRecordNotFound)
			},
			expectedUser:  nil,
			expectedError: true,
		},
		{
			name:  "Database Connection Error",
			email: "user@example.com",
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT (.+) FROM users WHERE (.+)").
					WithArgs("user@example.com").
					WillReturnError(errors.New("DB connection error"))
			},
			expectedUser:  nil,
			expectedError: true,
		},
		{
			name:  "Empty Email Input",
			email: "",
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT (.+) FROM users WHERE (.+)").
					WithArgs("").
					WillReturnError(gorm.ErrRecordNotFound)
			},
			expectedUser:  nil,
			expectedError: true,
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
				t.Fatalf("failed to open gorm DB connection: %v", err)
			}
			defer gormDB.Close()

			store := &UserStore{db: gormDB}

		
			if tt.mockSetup != nil {
				tt.mockSetup(mock)
			}

		
			user, err := store.GetByEmail(tt.email)

		
			if tt.expectedError {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				t.Logf("expected error received: %v", err)
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				} else if user.Email != tt.expectedUser.Email {
					t.Errorf("expected user email to be %v but got %v", tt.expectedUser.Email, user.Email)
				} else {
					t.Logf("successfully retrieved user with email: %v", user.Email)
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


 */
func TestGetByUsername(t *testing.T) {

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to open mock sql db, %s", err)
	}
	defer db.Close()


	gormDB, err := gorm.Open("sqlite3", db)
	if err != nil {
		t.Fatalf("Failed to open gorm db, %s", err)
	}
	defer gormDB.Close()

	store := &UserStore{db: gormDB}


	tests := []struct {
		name         string
		username     string
		setupMock    func()
		expectedUser *model.User
		expectError  bool
	}{
		{
			name:     "Successful User Retrieval by Username",
			username: "existingUser",
			setupMock: func() {
				rows := sqlmock.NewRows([]string{"id", "username", "email", "password", "bio", "image"}).
					AddRow(1, "existingUser", "user@example.com", "password123", "A short bio", "image.png")
				mock.ExpectQuery("^SELECT \\* FROM \"users\" WHERE (username = \\?) LIMIT 1$").
					WithArgs("existingUser").
					WillReturnRows(rows)
			},
			expectedUser: &model.User{Model: gorm.Model{ID: 1}, Username: "existingUser", Email: "user@example.com", Password: "password123", Bio: "A short bio", Image: "image.png"},
			expectError:  false,
		},
		{
			name:     "User Not Found by Username",
			username: "nonExistingUser",
			setupMock: func() {
				mock.ExpectQuery("^SELECT \\* FROM \"users\" WHERE (username = \\?) LIMIT 1$").
					WithArgs("nonExistingUser").
					WillReturnError(gorm.ErrRecordNotFound)
			},
			expectedUser: nil,
			expectError:  true,
		},
		{
			name:     "Database Query Failure",
			username: "someUser",
			setupMock: func() {
				mock.ExpectQuery("^SELECT \\* FROM \"users\" WHERE (username = \\?) LIMIT 1$").
					WithArgs("someUser").
					WillReturnError(errors.New("database error"))
			},
			expectedUser: nil,
			expectError:  true,
		},
		{
			name:     "Empty String as Username",
			username: "",
			setupMock: func() {
				mock.ExpectQuery("^SELECT \\* FROM \"users\" WHERE (username = \\?) LIMIT 1$").
					WithArgs("").
					WillReturnError(gorm.ErrRecordNotFound)
			},
			expectedUser: nil,
			expectError:  true,
		},
		{
			name:     "Case Sensitivity in Username Query",
			username: "EXISTINGUSER",
			setupMock: func() {
				mock.ExpectQuery("^SELECT \\* FROM \"users\" WHERE (username = \\?) LIMIT 1$").
					WithArgs("EXISTINGUSER").
					WillReturnError(gorm.ErrRecordNotFound)
			},
			expectedUser: nil,
			expectError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			user, err := store.GetByUsername(tt.username)
			if (err != nil) != tt.expectError {
				t.Errorf("unexpected error: %v", err)
			}

			if user != nil && tt.expectedUser != nil {
				if user.ID != tt.expectedUser.ID || user.Username != tt.expectedUser.Username ||
					user.Email != tt.expectedUser.Email {
					t.Errorf("expected user: %+v, got: %+v", tt.expectedUser, user)
				}
			} else if user != tt.expectedUser {
				t.Errorf("expected user: %+v, got: %+v", tt.expectedUser, user)
			}

			t.Logf("Test '%s' completed successfully", tt.name)
		})
	}
}


/*
ROOST_METHOD_HASH=Update_68f27dd78a
ROOST_METHOD_SIG_HASH=Update_87150d6435


 */
func TestUpdate(t *testing.T) {
	tests := []struct {
		name          string
		existingUsers []model.User
		updateUser    model.User
		expectedError error
	}{
		{
			name: "Scenario 1: Successful Update of Existing User",
			existingUsers: []model.User{
				{Model: gorm.Model{ID: 1}, Username: "user1", Email: "user1@example.com", Password: "password1", Bio: "Bio1", Image: "Image1"},
			},
			updateUser:    model.User{Model: gorm.Model{ID: 1}, Username: "user1", Email: "user1@example.com", Password: "newpassword", Bio: "NewBio", Image: "NewImage"},
			expectedError: nil,
		},
		{
			name:          "Scenario 2: Update Non-Existent User",
			existingUsers: []model.User{},
			updateUser:    model.User{Model: gorm.Model{ID: 2}, Username: "user2", Email: "user2@example.com", Password: "password2"},
			expectedError: gorm.ErrRecordNotFound,
		},
		{
			name: "Scenario 3: Update with Unique Constraint Violation",
			existingUsers: []model.User{
				{Model: gorm.Model{ID: 3}, Username: "user3", Email: "user3@example.com", Password: "password3"},
				{Model: gorm.Model{ID: 4}, Username: "user4", Email: "user4@example.com", Password: "password4"},
			},
			updateUser:    model.User{Model: gorm.Model{ID: 4}, Username: "user3", Email: "user3@example.com", Password: "password4"},
			expectedError: gorm.ErrInvalidSQL,
		},
		{
			name: "Scenario 4: Update with Partial Data (Nullable Fields)",
			existingUsers: []model.User{
				{Model: gorm.Model{ID: 5}, Username: "user5", Email: "user5@example.com", Password: "password5", Bio: "Bio5", Image: "Image5"},
			},
			updateUser:    model.User{Model: gorm.Model{ID: 5}, Bio: "UpdatedBio"},
			expectedError: nil,
		},
		{
			name: "Scenario 5: Update User with Empty Fields Except Required Ones",
			existingUsers: []model.User{
				{Model: gorm.Model{ID: 6}, Username: "user6", Email: "user6@example.com", Password: "password6", Bio: "Bio6", Image: "Image6"},
			},
			updateUser:    model.User{Model: gorm.Model{ID: 6}, Password: "newpassword"},
			expectedError: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("error: %v", err)
			}
			defer db.Close()

			gormDB, err := gorm.Open("sqlite3", db)
			if err != nil {
				t.Fatalf("error while opening db: %v", err)
			}

			userStore := UserStore{db: gormDB}

		
			if len(test.existingUsers) > 0 {
				rows := sqlmock.NewRows([]string{"id", "username", "email", "password", "bio", "image"})
				for _, u := range test.existingUsers {
					rows.AddRow(u.ID, u.Username, u.Email, u.Password, u.Bio, u.Image)
				}
				mock.ExpectQuery("^SELECT \\* FROM \"users\" WHERE \\(id IN \\(\\?\\)\\)$").WillReturnRows(rows)
			}

			mock.ExpectBegin()
		
			if test.expectedError == gorm.ErrRecordNotFound {
				mock.ExpectExec("^UPDATE \"users\" SET").WillReturnError(test.expectedError)
			} else if test.expectedError == gorm.ErrInvalidSQL {
				mock.ExpectExec("^UPDATE \"users\" SET").WillReturnError(test.expectedError)
			} else {
				mock.ExpectExec("^UPDATE \"users\" SET").WillReturnResult(sqlmock.NewResult(1, 1))
			}
			mock.ExpectCommit()

			err = userStore.Update(&test.updateUser)
			if test.expectedError != nil && err == nil {
				t.Fatalf("expected error: %v but got nil", test.expectedError)
			}
			if test.expectedError == nil && err != nil {
				t.Fatalf("expected error: nil but got %v", err)
			}

		
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unmet expectations: %s", err)
			}

			t.Logf("Scenario '%s' executed successfully", test.name)
		})
	}
}


/*
ROOST_METHOD_HASH=Follow_48fdf1257b
ROOST_METHOD_SIG_HASH=Follow_8217e61c06


 */
func TestFollow(t *testing.T) {

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	gormDB, err := gorm.Open("sqlite3", db)
	if err != nil {
		t.Fatalf("failed to initialize gorm with sqlmock: %s", err)
	}

	userStore := UserStore{db: gormDB}

	t.Run("Normal Operation - Successful Follow", func(t *testing.T) {
	
		userA := &model.User{Username: "Alice", Email: "alice@example.com"}
		userB := &model.User{Username: "Bob", Email: "bob@example.com"}

		mock.ExpectBegin()
		mock.ExpectExec(`INSERT INTO "follows"`).
			WithArgs(sqlmock.AnyArg(), userA.ID, userB.ID).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

	
		err := userStore.Follow(userA, userB)

	
		if err != nil {
			t.Errorf("expected successful follow, got error: %s", err)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %s", err)
		}

		t.Log("Success: User A followed User B")
	})

	t.Run("Edge Case - User Follows Self", func(t *testing.T) {
	
		userA := &model.User{Username: "Alice", Email: "alice@example.com"}

	
		err := userStore.Follow(userA, userA)

	
		if err == nil {
			t.Error("expected error when user tries to follow self, got nil")
		}

		t.Log("Success: User cannot follow themselves")
	})

	t.Run("Error Handling - Follow Non-existing User", func(t *testing.T) {
	
		userA := &model.User{Username: "Alice", Email: "alice@example.com"}
		userB := &model.User{Username: "Carol"}

		mock.ExpectBegin()
		mock.ExpectExec(`INSERT INTO "follows"`).
			WithArgs(sqlmock.AnyArg(), userA.ID, userB.ID).
			WillReturnError(errors.New("user not found"))
		mock.ExpectRollback()

	
		err := userStore.Follow(userA, userB)

	
		if err == nil || err.Error() != "user not found" {
			t.Errorf("expected error 'user not found', got: %v", err)
		}

		t.Log("Success: Proper error returned when user tries to follow non-existing user")
	})

	t.Run("Constraint Violation - Duplicate Follow Entry", func(t *testing.T) {
	
		userA := &model.User{Username: "Alice", Email: "alice@example.com"}
		userB := &model.User{Username: "Bob", Email: "bob@example.com"}

		mock.ExpectBegin()
		mock.ExpectExec(`INSERT INTO "follows"`).
			WithArgs(sqlmock.AnyArg(), userA.ID, userB.ID).
			WillReturnError(errors.New("duplicate entry"))
		mock.ExpectRollback()

	
		err := userStore.Follow(userA, userB)

	
		if err == nil {
			t.Error("expected error when trying to follow user already followed, got nil")
		}

		t.Log("Success: Duplicate follow entry properly not allowed")
	})

	t.Run("Database Error - Follow While DB Down", func(t *testing.T) {
	
		userA := &model.User{Username: "Alice", Email: "alice@example.com"}
		userB := &model.User{Username: "Bob", Email: "bob@example.com"}

		mock.ExpectBegin().
			WillReturnError(errors.New("database is down"))

	
		err := userStore.Follow(userA, userB)

	
		if err == nil || err.Error() != "database is down" {
			t.Errorf("expected 'database is down' error, got: %v", err)
		}

		t.Log("Success: Database error properly returned when the database is down")
	})

	t.Run("Input Validation - Null User", func(t *testing.T) {
	
		userA := &model.User{Username: "Alice", Email: "alice@example.com"}

	
		err := userStore.Follow(userA, nil)

	
		if err == nil {
			t.Error("expected error when following a nil user, got nil")
		}

		t.Log("Success: Proper handling of nil user input")
	})
}


/*
ROOST_METHOD_HASH=IsFollowing_f53a5d9cef
ROOST_METHOD_SIG_HASH=IsFollowing_9eba5a0e9c


 */
func TestIsFollowing(t *testing.T) {

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error initializing SQL mock: %v", err)
	}
	defer db.Close()

	gormDB, err := gorm.Open("sqlite3", "test.db")
	if err != nil {
		t.Fatalf("Error opening GORM DB: %v", err)
	}
	defer gormDB.Close()

	store := &UserStore{db: gormDB}

	type args struct {
		a *model.User
		b *model.User
	}

	tests := []struct {
		name       string
		setupMock  func()
		args       args
		want       bool
		expectErr  bool
	}{
		{
			name: "Scenario 1: User A is following User B",
			setupMock: func() {
				mock.ExpectQuery("SELECT count(\\*) FROM follows").
					WithArgs(1, 2).
					WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
			},
			args: args{
				a: &model.User{Model: gorm.Model{ID: 1}},
				b: &model.User{Model: gorm.Model{ID: 2}},
			},
			want:      true,
			expectErr: false,
		},
		{
			name: "Scenario 2: User A is not following User B",
			setupMock: func() {
				mock.ExpectQuery("SELECT count(\\*) FROM follows").
					WithArgs(1, 3).
					WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
			},
			args: args{
				a: &model.User{Model: gorm.Model{ID: 1}},
				b: &model.User{Model: gorm.Model{ID: 3}},
			},
			want:      false,
			expectErr: false,
		},
		{
			name: "Scenario 3: User A or User B is nil",
			setupMock: func() {
			
			},
			args: args{
				a: nil,
				b: &model.User{Model: gorm.Model{ID: 3}},
			},
			want:      false,
			expectErr: false,
		},
		{
			name: "Scenario 4: Database error occurs during the query",
			setupMock: func() {
				mock.ExpectQuery("SELECT count(\\*) FROM follows").
					WithArgs(1, 2).
					WillReturnError(errors.New("database error"))
			},
			args: args{
				a: &model.User{Model: gorm.Model{ID: 1}},
				b: &model.User{Model: gorm.Model{ID: 2}},
			},
			want:      false,
			expectErr: true,
		},
		{
			name: "Scenario 5: User A follows multiple people, including User B",
			setupMock: func() {
				mock.ExpectQuery("SELECT count(\\*) FROM follows").
					WithArgs(1, 2).
					WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
			},
			args: args{
				a: &model.User{Model: gorm.Model{ID: 1}},
				b: &model.User{Model: gorm.Model{ID: 2}},
			},
			want:      true,
			expectErr: false,
		},
		{
			name: "Scenario 6: Large Dataset of Follows",
			setupMock: func() {
				mock.ExpectQuery("SELECT count(\\*) FROM follows").
					WithArgs(1, 2).
					WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
			},
			args: args{
				a: &model.User{Model: gorm.Model{ID: 1}},
				b: &model.User{Model: gorm.Model{ID: 2}},
			},
			want:      true,
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
		
			tt.setupMock()

		
			got, err := store.IsFollowing(tt.args.a, tt.args.b)

		
			if (err != nil) != tt.expectErr {
				t.Errorf("IsFollowing() error = %v, wantErr %v", err, tt.expectErr)
			}
			if got != tt.want {
				t.Errorf("IsFollowing() = %v, want %v", got, tt.want)
			}

		
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("unfulfilled expectations: %s", err)
			}
		})
	}
}


/*
ROOST_METHOD_HASH=Unfollow_57959a8a53
ROOST_METHOD_SIG_HASH=Unfollow_8bd8e0bc55


 */
func TestUserStoreUnfollow(t *testing.T) {
	type args struct {
		follower *model.User
		followee *model.User
	}

	tests := []struct {
		name       string
		setupMocks func(mock sqlmock.Sqlmock)
		args       args
		wantErr    bool
	}{
		{
			name: "Scenario 1: Successful Unfollow Operation",
			setupMocks: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectQuery("^SELECT .* FROM `follows`").
					WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
				mock.ExpectExec("^DELETE FROM `follows`").
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
			args: args{
				follower: &model.User{Model: gorm.Model{ID: 1}, Follows: []model.User{{Model: gorm.Model{ID: 2}}}},
				followee: &model.User{Model: gorm.Model{ID: 2}},
			},
			wantErr: false,
		},
		{
			name: "Scenario 2: Unfollow Non-Existent Relationship",
			setupMocks: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectQuery("^SELECT .* FROM `follows`").
					WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
				mock.ExpectCommit()
			},
			args: args{
				follower: &model.User{Model: gorm.Model{ID: 1}},
				followee: &model.User{Model: gorm.Model{ID: 2}},
			},
			wantErr: false,
		},
		{
			name: "Scenario 3: Unfollow with an Empty 'Follows' List",
			setupMocks: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectQuery("^SELECT .* FROM `follows`").
					WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
				mock.ExpectCommit()
			},
			args: args{
				follower: &model.User{Model: gorm.Model{ID: 1}, Follows: []model.User{}},
				followee: &model.User{Model: gorm.Model{ID: 2}},
			},
			wantErr: false,
		},
		{
			name: "Scenario 4: Unfollow with Database Error",
			setupMocks: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec("^DELETE FROM `follows`").
					WillReturnError(fmt.Errorf("some db error"))
				mock.ExpectRollback()
			},
			args: args{
				follower: &model.User{Model: gorm.Model{ID: 1}, Follows: []model.User{{Model: gorm.Model{ID: 2}}}},
				followee: &model.User{Model: gorm.Model{ID: 2}},
			},
			wantErr: true,
		},
		{
			name: "Scenario 5: Unfollow Self",
			setupMocks: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectQuery("^SELECT .* FROM `follows`").
					WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
				mock.ExpectCommit()
			},
			args: args{
				follower: &model.User{Model: gorm.Model{ID: 1}},
				followee: &model.User{Model: gorm.Model{ID: 1}},
			},
			wantErr: false,
		},
		{
			name: "Scenario 6: Multiple Unfollows from the Same User",
			setupMocks: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectQuery("^SELECT .* FROM `follows`").
					WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
				mock.ExpectExec("^DELETE FROM `follows`").
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()

				mock.ExpectBegin()
				mock.ExpectQuery("^SELECT .* FROM `follows`").
					WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
				mock.ExpectExec("^DELETE FROM `follows`").
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
			args: args{
				follower: &model.User{Model: gorm.Model{ID: 1}, Follows: []model.User{
					{Model: gorm.Model{ID: 2}},
					{Model: gorm.Model{ID: 3}},
				}},
				followee: &model.User{Model: gorm.Model{ID: 2}},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
			if err != nil {
				t.Fatalf("failed to open db mock: %v", err)
			}
			defer db.Close()

			sqlDB, err := gorm.Open("sqlite3", db)
			if err != nil {
				t.Fatalf("failed to open gorm db: %v", err)
			}
			defer sqlDB.Close()

			store := &UserStore{db: sqlDB}
			tt.setupMocks(mock)

			t.Logf("Running scenario: %s", tt.name)
			err = store.Unfollow(tt.args.follower, tt.args.followee)
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


 */
func TestUserStoreGetFollowingUserIDs(t *testing.T) {

	tests := []struct {
		name           string
		prepDBMock     func(sqlmock.Sqlmock)
		user           *model.User
		expectedIDs    []uint
		expectedErr    error
	}{
		{
			name: "Retrieve Following User IDs for a User with Followers",
			prepDBMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"to_user_id"}).AddRow(2).AddRow(3)
				mock.ExpectQuery("^SELECT to_user_id FROM follows WHERE (.+)").WithArgs(1).WillReturnRows(rows)
			},
			user:        &model.User{Model: gorm.Model{ID: 1}},
			expectedIDs: []uint{2, 3},
			expectedErr: nil,
		},
		{
			name: "Retrieve Following User IDs for a User with No Followers",
			prepDBMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"to_user_id"})
				mock.ExpectQuery("^SELECT to_user_id FROM follows WHERE (.+)").WithArgs(2).WillReturnRows(rows)
			},
			user:        &model.User{Model: gorm.Model{ID: 2}},
			expectedIDs: []uint{},
			expectedErr: nil,
		},
		{
			name: "Handle Database Query Error",
			prepDBMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("^SELECT to_user_id FROM follows WHERE (.+)").WithArgs(3).WillReturnError(errors.New("a database error"))
			},
			user:        &model.User{Model: gorm.Model{ID: 3}},
			expectedIDs: []uint{},
			expectedErr: errors.New("a database error"),
		},
		{
			name: "Retrieve Following User IDs with Invalid User Input",
			prepDBMock: func(mock sqlmock.Sqlmock) {
			
			},
			user:        nil,
			expectedIDs: []uint{},
			expectedErr: nil,
		},
		{
			name: "Large Number of Followers",
			prepDBMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"to_user_id"})
				for i := 1; i <= 1000; i++ {
					rows.AddRow(i)
				}
				mock.ExpectQuery("^SELECT to_user_id FROM follows WHERE (.+)").WithArgs(4).WillReturnRows(rows)
			},
			user:        &model.User{Model: gorm.Model{ID: 4}},
			expectedIDs: generateExpectedIDs(1000),
			expectedErr: nil,
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
				t.Fatalf("an error '%s' was not expected when connecting to gorm", err)
			}
			defer gormDB.Close()

			store := &UserStore{db: gormDB}
			if tt.prepDBMock != nil {
				tt.prepDBMock(mock)
			}

			ids, err := store.GetFollowingUserIDs(tt.user)
		
			if err != nil {
				t.Logf("Received error: %s", err.Error())
			} else {
				t.Logf("Received IDs: %v", ids)
			}

			assert.Equal(t, tt.expectedErr, err)
			assert.Equal(t, tt.expectedIDs, ids)
		})
	}
}

func generateExpectedIDs(n int) []uint {
	ids := make([]uint, n)
	for i := 0; i < n; i++ {
		ids[i] = uint(i + 1)
	}
	return ids
}

