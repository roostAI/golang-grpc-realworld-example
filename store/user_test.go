package store

import (
	"testing"
	"github.com/jinzhu/gorm"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/raahii/golang-grpc-realworld-example/store"
	"errors"
	"time"
	"github.com/jinzhu/gorm/dialects/postgres"
	"github.com/raahii/golang-grpc-realworld-example/model"
	"github.com/stretchr/testify/assert"
)

type ExpectedQuery struct {
	queryBasedExpectation
	rows             driver.Rows
	delay            time.Duration
	rowsMustBeClosed bool
	rowsWereClosed   bool
}
type T struct {
	common
	isParallel bool
	isEnvSet   bool
	context    *testContext
}
type ExpectedBegin struct {
	commonExpectation
	delay time.Duration
}
type ExpectedCommit struct {
	commonExpectation
}
type ExpectedExec struct {
	queryBasedExpectation
	result driver.Result
	delay  time.Duration
}
type Rows struct {
	converter driver.ValueConverter
	cols      []string
	def       []*Column
	rows      [][]driver.Value
	pos       int
	nextErr   map[int]error
	closeErr  error
}
type ExpectedRollback struct {
	commonExpectation
}
/*
ROOST_METHOD_HASH=NewUserStore_6a331dd890
ROOST_METHOD_SIG_HASH=NewUserStore_4f0c2dfca9


 */
func TestNewUserStore(t *testing.T) {

	t.Run("Scenario 1: Basic Initialization with a Valid gorm.DB Object", func(t *testing.T) {

		db, _, err := sqlmock.New()
		if err != nil {
			t.Fatalf("Failed to create sqlmock: %v", err)
		}
		defer db.Close()
		gormDB, err := gorm.Open("sqlmock", db)
		if err != nil {
			t.Fatalf("Failed to open gorm DB: %v", err)
		}

		userStore := store.NewUserStore(gormDB)

		if userStore.db != gormDB {
			t.Errorf("Expected gorm.DB instance %v, but got %v", gormDB, userStore.db)
		} else {
			t.Log("Successfully initialized UserStore with valid gorm.DB")
		}
	})

	t.Run("Scenario 2: Handling of Nil gorm.DB Object", func(t *testing.T) {

		var nilDB *gorm.DB

		userStore := store.NewUserStore(nilDB)

		if userStore.db != nil {
			t.Errorf("Expected nil gorm.DB, but got %v", userStore.db)
		} else {
			t.Log("Correctly handled nil gorm.DB without panic")
		}
	})

	t.Run("Scenario 3: Integration with Other Components Using a Mocked DB", func(t *testing.T) {

		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("Failed to create sqlmock: %v", err)
		}
		defer db.Close()
		gormDB, err := gorm.Open("sqlmock", db)
		if err != nil {
			t.Fatalf("Failed to open gorm DB: %v", err)
		}

		mock.ExpectQuery("SELECT \\* FROM users").WillReturnRows(sqlmock.NewRows(nil))

		userStore := store.NewUserStore(gormDB)

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("There were unfulfilled expectations: %s", err)
		} else {
			t.Log("UserStore can successfully integrate with mocked DB")
		}
	})

	t.Run("Scenario 4: Multiple Initializations of UserStore", func(t *testing.T) {

		db1, _, err := sqlmock.New()
		if err != nil {
			t.Fatalf("Failed to create sqlmock: %v", err)
		}
		defer db1.Close()
		gormDB1, err := gorm.Open("sqlmock", db1)
		if err != nil {
			t.Fatalf("Failed to open gorm DB: %v", err)
		}

		db2, _, err := sqlmock.New()
		if err != nil {
			t.Fatalf("Failed to create sqlmock: %v", err)
		}
		defer db2.Close()
		gormDB2, err := gorm.Open("sqlmock", db2)
		if err != nil {
			t.Fatalf("Failed to open gorm DB: %v", err)
		}

		userStore1 := store.NewUserStore(gormDB1)
		userStore2 := store.NewUserStore(gormDB2)

		if userStore1 == userStore2 || userStore1.db == userStore2.db {
			t.Errorf("Expected distinct UserStore instances, but got identical instances")
		} else {
			t.Log("Validation succeeded: Distinct UserStore instances with separate DBs")
		}
	})
}


/*
ROOST_METHOD_HASH=Create_889fc0fc45
ROOST_METHOD_SIG_HASH=Create_4c48ec3920


 */
func TestUserStoreCreate(t *testing.T) {
	t.Run("Successful User Creation", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
		}
		defer db.Close()

		mock.ExpectBegin()
		mock.ExpectExec("INSERT INTO \"users\"").
			WithArgs(sqlmock.AnyArg(), "unique_user", "unique_email@example.com", sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		gormDB, err := gorm.Open("postgres", db)
		if err != nil {
			t.Fatalf("failed to initialize gorm db: %v", err)
		}
		defer gormDB.Close()

		userStore := &UserStore{db: gormDB}
		user := &model.User{
			Username: "unique_user",
			Email:    "unique_email@example.com",
			Password: "securepassword",
			Bio:      "User Bio",
			Image:    "User Image",
		}

		err = userStore.Create(user)
		if err != nil {
			t.Errorf("expected no error, got '%v'", err)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %s", err)
		}
	})

	t.Run("Duplicate Email Error", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
		}
		defer db.Close()

		mock.ExpectBegin()
		mock.ExpectExec("INSERT INTO \"users\"").
			WithArgs(sqlmock.AnyArg(), "another_user", "duplicate_email@example.com", sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnError(errors.New("duplicate key value violates unique constraint"))

		gormDB, err := gorm.Open("postgres", db)
		if err != nil {
			t.Fatalf("failed to initialize gorm db: %v", err)
		}
		defer gormDB.Close()

		userStore := &UserStore{db: gormDB}
		user := &model.User{
			Username: "another_user",
			Email:    "duplicate_email@example.com",
			Password: "securepassword",
			Bio:      "User Bio",
			Image:    "User Image",
		}

		err = userStore.Create(user)
		if err == nil {
			t.Error("expected error, got none")
		}
		if err.Error() != "duplicate key value violates unique constraint" {
			t.Errorf("unexpected error message: %s", err)
		}
	})

	t.Run("Duplicate Username Error", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
		}
		defer db.Close()

		mock.ExpectBegin()
		mock.ExpectExec("INSERT INTO \"users\"").
			WithArgs(sqlmock.AnyArg(), "duplicate_user", "new_email@example.com", sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnError(errors.New("duplicate key value violates unique constraint"))

		gormDB, err := gorm.Open("postgres", db)
		if err != nil {
			t.Fatalf("failed to initialize gorm db: %v", err)
		}
		defer gormDB.Close()

		userStore := &UserStore{db: gormDB}
		user := &model.User{
			Username: "duplicate_user",
			Email:    "new_email@example.com",
			Password: "securepassword",
			Bio:      "User Bio",
			Image:    "User Image",
		}

		err = userStore.Create(user)
		if err == nil {
			t.Error("expected error, got none")
		}
		if err.Error() != "duplicate key value violates unique constraint" {
			t.Errorf("unexpected error message: %s", err)
		}
	})

	t.Run("Database Connection Error", func(t *testing.T) {
		userStore := &UserStore{db: nil}
		user := &model.User{
			Username: "new_user",
			Email:    "new_email@example.com",
			Password: "securepassword",
			Bio:      "User Bio",
			Image:    "User Image",
		}

		err := userStore.Create(user)
		if err == nil {
			t.Error("expected error, got none")
		}
	})

	t.Run("Invalid Model Fields", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
		}
		defer db.Close()

		mock.ExpectBegin()
		mock.ExpectExec("INSERT INTO \"users\"").
			WithArgs(sqlmock.AnyArg(), "", "", sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnError(errors.New("not null constraint failed"))

		gormDB, err := gorm.Open("postgres", db)
		if err != nil {
			t.Fatalf("failed to initialize gorm db: %v", err)
		}
		defer gormDB.Close()

		userStore := &UserStore{db: gormDB}
		user := &model.User{}

		err = userStore.Create(user)
		if err == nil {
			t.Error("expected error, got none")
		}
	})

	t.Run("Successful User Creation with Empty Follows and Favorites", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
		}
		defer db.Close()

		mock.ExpectBegin()
		mock.ExpectExec("INSERT INTO \"users\"").
			WithArgs(sqlmock.AnyArg(), "user_with_empty_lists", "email@example.com", sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		gormDB, err := gorm.Open("postgres", db)
		if err != nil {
			t.Fatalf("failed to initialize gorm db: %v", err)
		}
		defer gormDB.Close()

		userStore := &UserStore{db: gormDB}
		user := &model.User{
			Username:         "user_with_empty_lists",
			Email:            "email@example.com",
			Password:         "securepassword",
			Bio:              "User Bio",
			Image:            "User Image",
			Follows:          []model.User{},
			FavoriteArticles: []model.Article{},
		}

		err = userStore.Create(user)
		if err != nil {
			t.Errorf("expected no error, got '%v'", err)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %s", err)
		}
	})

	t.Run("Error Handling in Overridden Time Function", func(t *testing.T) {
		db, _, err := sqlmock.New()
		if err != nil {
			t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
		}
		defer db.Close()

		gormDB, err := gorm.Open("postgres", db)
		if err != nil {
			t.Fatalf("failed to initialize gorm db: %v", err)
		}
		defer gormDB.Close()

		originalNowFunc := gormDB.Set("gorm:now_func", func() time.Time {
			return time.Time{}
		})
		defer gormDB.Set("gorm:now_func", originalNowFunc.(func() time.Time))

		userStore := &UserStore{db: gormDB}
		user := &model.User{
			Username: "test_user",
			Email:    "test_email@example.com",
			Password: "securepassword",
			Bio:      "User Bio",
			Image:    "User Image",
		}

		err = userStore.Create(user)
		if err == nil {
			t.Error("expected error due to invalid timestamp, got none")
		}
	})

	t.Run("Concurrency with Global Locks", func(t *testing.T) {
		t.Log("Concurrency tests might be needed based on the actual implementation specifics")
	})
}


/*
ROOST_METHOD_HASH=GetByID_bbf946112e
ROOST_METHOD_SIG_HASH=GetByID_728dd55ed1


 */
func TestUserStoreGetByID(t *testing.T) {
	type testCase struct {
		name          string
		setupMock     func(sqlmock.Sqlmock)
		id            uint
		expectedUser  *model.User
		expectedError error
	}

	t.Log("Initializing sqlmock for database mocking.")
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("An error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	gormDB, err := gorm.Open("postgres", db)
	if err != nil {
		t.Fatalf("Failed to open gorm DB with sqlmock, error: %s", err)
	}
	defer gormDB.Close()

	store := &UserStore{db: gormDB}

	tests := []testCase{
		{
			name: "Retrieve User Successfully by Valid ID",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("^SELECT (.+) FROM \"users\" WHERE (.+)$").
					WithArgs(1).
					WillReturnRows(sqlmock.NewRows([]string{"id", "username", "email", "password", "bio", "image"}).
						AddRow(1, "user1", "user1@example.com", "password", "bio", "image"))
			},
			id: 1,
			expectedUser: &model.User{
				Model:    gorm.Model{ID: 1},
				Username: "user1",
				Email:    "user1@example.com",
				Password: "password",
				Bio:      "bio",
				Image:    "image",
			},
			expectedError: nil,
		},
		{
			name: "Fail to Retrieve User for Non-Existent ID",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("^SELECT (.+) FROM \"users\" WHERE (.+)$").
					WithArgs(9999).
					WillReturnError(gorm.ErrRecordNotFound)
			},
			id:            9999,
			expectedUser:  nil,
			expectedError: gorm.ErrRecordNotFound,
		},
		{
			name: "Return Error Due to Database Connection Issue",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("^SELECT (.+) FROM \"users\" WHERE (.+)$").
					WithArgs(1).
					WillReturnError(errors.New("connection kill"))
			},
			id:            1,
			expectedUser:  nil,
			expectedError: errors.New("connection kill"),
		},
		{
			name: "Handle Large Numeric User ID Inputs",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("^SELECT (.+) FROM \"users\" WHERE (.+)$").
					WithArgs(9223372036854775807).
					WillReturnError(gorm.ErrRecordNotFound)
			},
			id:            9223372036854775807,
			expectedUser:  nil,
			expectedError: gorm.ErrRecordNotFound,
		},
		{
			name: "Validate Function's Response to Zero as User ID",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("^SELECT (.+) FROM \"users\" WHERE (.+)$").
					WithArgs(0).
					WillReturnError(gorm.ErrRecordNotFound)
			},
			id:            0,
			expectedUser:  nil,
			expectedError: gorm.ErrRecordNotFound,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Logf("Running test case: %s", tc.name)
			tc.setupMock(mock)

			actualUser, actualError := store.GetByID(tc.id)

			assert.Equal(t, tc.expectedUser, actualUser, "The user returned does not match the expected value.")
			assert.Equal(t, tc.expectedError, actualError, "The error returned does not match the expected value.")

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}


/*
ROOST_METHOD_HASH=GetByEmail_3574af40e5
ROOST_METHOD_SIG_HASH=GetByEmail_5731b833c1


 */
func TestUserStoreGetByEmail(t *testing.T) {

	tests := []struct {
		name          string
		email         string
		setupMock     func(sqlmock.Sqlmock)
		expectedUser  *model.User
		expectedError error
	}{
		{
			name:  "Retrieve Existing User by Email",
			email: "existinguser@example.com",
			setupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "username", "email", "password", "bio", "image"}).
					AddRow(1, "ExistingUser", "existinguser@example.com", "hashedpassword", "Bio", "ImageURL")
				mock.ExpectQuery("^SELECT \\* FROM \"users\" WHERE \\(email = \\$1\\) ORDER BY \"users\".\"id\" ASC LIMIT 1$").
					WithArgs("existinguser@example.com").WillReturnRows(rows)
			},
			expectedUser: &model.User{
				Model:    gorm.Model{ID: 1},
				Username: "ExistingUser",
				Email:    "existinguser@example.com",
				Password: "hashedpassword",
				Bio:      "Bio",
				Image:    "ImageURL",
			},
			expectedError: nil,
		},
		{
			name:  "Email Does Not Exist in Database",
			email: "nonexistent@example.com",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("^SELECT \\* FROM \"users\" WHERE \\(email = \\$1\\) ORDER BY \"users\".\"id\" ASC LIMIT 1$").
					WithArgs("nonexistent@example.com").WillReturnError(gorm.ErrRecordNotFound)
			},
			expectedUser:  nil,
			expectedError: gorm.ErrRecordNotFound,
		},
		{
			name:  "Invalid Email Format",
			email: "invalid-email",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("^SELECT \\* FROM \"users\" WHERE \\(email = \\$1\\) ORDER BY \"users\".\"id\" ASC LIMIT 1$").
					WithArgs("invalid-email").WillReturnError(gorm.ErrRecordNotFound)
			},
			expectedUser:  nil,
			expectedError: gorm.ErrRecordNotFound,
		},
		{
			name:  "Database Connection Error",
			email: "dberror@example.com",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("^SELECT \\* FROM \"users\" WHERE \\(email = \\$1\\) ORDER BY \"users\".\"id\" ASC LIMIT 1$").
					WithArgs("dberror@example.com").WillReturnError(errors.New("connection error"))
			},
			expectedUser:  nil,
			expectedError: errors.New("connection error"),
		},
		{
			name:  "Multiple Users with the Same Email",
			email: "duplicateemail@example.com",
			setupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "username", "email", "password", "bio", "image"}).
					AddRow(1, "User1", "duplicateemail@example.com", "password1", "Bio1", "Image1").
					AddRow(2, "User2", "duplicateemail@example.com", "password2", "Bio2", "Image2")
				mock.ExpectQuery("^SELECT \\* FROM \"users\" WHERE \\(email = \\$1\\) ORDER BY \"users\".\"id\" ASC LIMIT 1$").
					WithArgs("duplicateemail@example.com").WillReturnRows(rows)
			},
			expectedUser: &model.User{
				Model:    gorm.Model{ID: 1},
				Username: "User1",
				Email:    "duplicateemail@example.com",
				Password: "password1",
				Bio:      "Bio1",
				Image:    "Image1",
			},
			expectedError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("error while opening a stub database connection: %s", err)
			}
			defer db.Close()

			mockSQL, gormErr := gorm.Open("postgres", db)
			if gormErr != nil {
				t.Fatalf("error initializing gorm db mock: %s", gormErr)
			}

			tt.setupMock(mock)

			store := &UserStore{db: mockSQL}
			user, err := store.GetByEmail(tt.email)

			if tt.expectedError != nil && err != nil {
				if tt.expectedError.Error() != err.Error() {
					t.Errorf("unexpected error. expected: %v, got: %v", tt.expectedError, err)
				}
			} else if (tt.expectedUser == nil) != (user == nil) {
				t.Fatalf("unexpected user result. expected: %v, got: %v", tt.expectedUser, user)
			} else if tt.expectedUser != nil && user != nil {

				if tt.expectedUser.Model.ID != user.Model.ID ||
					tt.expectedUser.Username != user.Username ||
					tt.expectedUser.Email != user.Email ||
					tt.expectedUser.Password != user.Password ||
					tt.expectedUser.Bio != user.Bio ||
					tt.expectedUser.Image != user.Image {
					t.Errorf("mismatched user. expected: %v, got: %v", tt.expectedUser, user)
				}
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatalf("there were unfulfilled expectations: %s", err)
			}

			t.Log("Test case executed:", tt.name)
		})
	}
}


/*
ROOST_METHOD_HASH=GetByUsername_f11f114df2
ROOST_METHOD_SIG_HASH=GetByUsername_954d096e24


 */
func TestUserStoreGetByUsername(t *testing.T) {
	type testCase struct {
		description  string
		username     string
		mockSetup    func(sqlmock.Sqlmock)
		expectedUser *model.User
		expectedErr  error
	}

	tests := []testCase{
		{
			description: "Retrieve Existing User by Username",
			username:    "existing_user",
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "username", "email", "password", "bio", "image"}).AddRow(1, "existing_user", "user@example.com", "password123", "bio data", "image.png")
				mock.ExpectQuery(`SELECT \* FROM "users" WHERE \(username = \?\) ORDER BY "users"\."id" ASC LIMIT 1`).WithArgs("existing_user").WillReturnRows(rows)
			},
			expectedUser: &model.User{
				Model:    gorm.Model{ID: 1},
				Username: "existing_user",
				Email:    "user@example.com",
				Password: "password123",
				Bio:      "bio data",
				Image:    "image.png",
			},
			expectedErr: nil,
		},
		{
			description: "Username Not Found",
			username:    "non_existent_user",
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT \* FROM "users" WHERE \(username = \?\) ORDER BY "users"\."id" ASC LIMIT 1`).WithArgs("non_existent_user").WillReturnError(gorm.ErrRecordNotFound)
			},
			expectedUser: nil,
			expectedErr:  gorm.ErrRecordNotFound,
		},
		{
			description: "Database Connectivity Issues",
			username:    "any_user",
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT \* FROM "users" WHERE \(username = \?\) ORDER BY "users"\."id" ASC LIMIT 1`).WithArgs("any_user").WillReturnError(errors.New("db connection error"))
			},
			expectedUser: nil,
			expectedErr:  errors.New("db connection error"),
		},
		{
			description: "Invalid Username Input",
			username:    "",
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT \* FROM "users" WHERE \(username = \?\) ORDER BY "users"\."id" ASC LIMIT 1`).WithArgs("").WillReturnError(gorm.ErrRecordNotFound)
			},
			expectedUser: nil,
			expectedErr:  gorm.ErrRecordNotFound,
		},
		{
			description: "Multiple Users with the Same Username",
			username:    "duplicate_user",
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "username", "email", "password", "bio", "image"}).
					AddRow(1, "duplicate_user", "user1@example.com", "password1", "bio1", "image1.png").
					AddRow(2, "duplicate_user", "user2@example.com", "password2", "bio2", "image2.png")
				mock.ExpectQuery(`SELECT \* FROM "users" WHERE \(username = \?\) ORDER BY "users"\."id" ASC LIMIT 1`).WithArgs("duplicate_user").WillReturnRows(rows)
			},
			expectedUser: &model.User{
				Model:    gorm.Model{ID: 1},
				Username: "duplicate_user",
				Email:    "user1@example.com",
				Password: "password1",
				Bio:      "bio1",
				Image:    "image1.png",
			},
			expectedErr: nil,
		},
	}

	for _, tc := range tests {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("error opening stub database connection: %v", err)
		}

		defer db.Close()

		if tc.mockSetup != nil {
			tc.mockSetup(mock)
		}

		gormDB, err := gorm.Open("postgres", db)
		if err != nil {
			t.Fatalf("failed to create GORM DB from sqlmock: %v", err)
		}

		store := &UserStore{db: gormDB}

		user, err := store.GetByUsername(tc.username)

		if tc.expectedErr != nil && err != nil && errors.Is(err, tc.expectedErr) {
			t.Logf("%s: Expected error matches actual error: %v", tc.description, err)
		} else if tc.expectedErr == nil && err == nil && equalUsers(tc.expectedUser, user) {
			t.Logf("%s: User data matches expectation", tc.description)
		} else {
			t.Errorf("%s: unexpected result, got (user: %v, err: %v)", tc.description, user, err)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("There were unmet expectations: %s", err)
		}
	}
}

func equalUsers(expected, actual *model.User) bool {
	if expected == nil || actual == nil {
		return expected == actual
	}
	return expected.ID == actual.ID &&
		expected.Username == actual.Username &&
		expected.Email == actual.Email &&
		expected.Password == actual.Password &&
		expected.Bio == actual.Bio &&
		expected.Image == actual.Image
}


/*
ROOST_METHOD_HASH=Update_68f27dd78a
ROOST_METHOD_SIG_HASH=Update_87150d6435


 */
func TestUserStoreUpdate(t *testing.T) {

	type testCase struct {
		name         string
		setup        func(sqlmock.Sqlmock)
		inputUser    model.User
		expectedErr  bool
		expectedRows int64
	}

	validUser := model.User{
		Username: "validUser",
		Email:    "valid@example.com",
		Password: "Password123",
		Bio:      "This is a bio",
		Image:    "http://example.com/image.jpg",
	}

	invalidUser := model.User{
		Username: "",
		Email:    "",
		Password: "Password123",
		Bio:      "This is a bio",
		Image:    "http://example.com/image.jpg",
	}

	testCases := []testCase{
		{
			name: "Update with Valid User Data",
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec(`^UPDATE "users" SET (.+)$`).
					WithArgs("newEmail@example.com", "Password123", "This is a bio", "http://example.com/image.jpg", "validUser").
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
			inputUser: model.User{
				Username: "validUser",
				Email:    "newEmail@example.com",
				Password: "Password123",
				Bio:      "This is a bio",
				Image:    "http://example.com/image.jpg",
			},
			expectedErr:  false,
			expectedRows: 1,
		},
		{
			name: "Update a Non-Existing User",
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec(`^UPDATE "users" SET (.+)$`).
					WithArgs("nonExistingUser@example.com", "Password123", "This is a bio", "http://example.com/image.jpg", "nonExistingUser").
					WillReturnResult(sqlmock.NewResult(0, 0))
				mock.ExpectCommit()
			},
			inputUser: model.User{
				Username: "nonExistingUser",
				Email:    "nonExistingUser@example.com",
				Password: "Password123",
				Bio:      "This is a bio",
				Image:    "http://example.com/image.jpg",
			},
			expectedErr:  true,
			expectedRows: 0,
		},
		{
			name: "Update with Invalid User Data",
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec(`^UPDATE "users" SET (.+)$`).
					WithArgs(invalidUser.Email, invalidUser.Password, invalidUser.Bio, invalidUser.Image, invalidUser.Username).
					WillReturnResult(sqlmock.NewResult(0, 0))
				mock.ExpectCommit()
			},
			inputUser:    invalidUser,
			expectedErr:  true,
			expectedRows: 0,
		},
		{
			name: "Update When DB Connection Fails",
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec(`^UPDATE "users" SET (.+)$`).
					WithArgs(validUser.Email, validUser.Password, validUser.Bio, validUser.Image, validUser.Username).
					WillReturnError(gorm.ErrInvalidSQL)
				mock.ExpectRollback()
			},
			inputUser:   validUser,
			expectedErr: true,
		},
		{
			name: "No Changes in Update Operation",
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec(`^UPDATE "users" SET (.+)$`).
					WithArgs(validUser.Email, validUser.Password, validUser.Bio, validUser.Image, validUser.Username).
					WillReturnResult(sqlmock.NewResult(1, 0))
				mock.ExpectCommit()
			},
			inputUser:    validUser,
			expectedErr:  false,
			expectedRows: 0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			assert.NoError(t, err)
			defer db.Close()

			gormDB, err := gorm.Open("postgres", db)
			assert.NoError(t, err)

			tc.setup(mock)

			store := &UserStore{db: gormDB}

			err = store.Update(&tc.inputUser)

			assert.Equal(t, tc.expectedErr, err != nil)

			if !tc.expectedErr {
				assert.Equal(t, tc.expectedRows, store.db.RowsAffected)
			}

			assert.Nil(t, mock.ExpectationsWereMet(), "there were unfulfilled expectations")
		})
	}
}


/*
ROOST_METHOD_HASH=Follow_48fdf1257b
ROOST_METHOD_SIG_HASH=Follow_8217e61c06


 */
func TestUserStoreFollow(t *testing.T) {
	type testScenario struct {
		name          string
		setupMocks    func(sqlmock.Sqlmock, *model.User, *model.User)
		expectedError error
	}

	tests := []testScenario{
		{
			name: "Successfully Follow Another User",
			setupMocks: func(mock sqlmock.Sqlmock, userA, userB *model.User) {
				mock.ExpectBegin()
				mock.ExpectExec(`INSERT INTO follows`).WithArgs(userA.ID, userB.ID).WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
		},
		{
			name: "Fail to Follow a User Due to Database Error",
			setupMocks: func(mock sqlmock.Sqlmock, userA, userB *model.User) {
				mock.ExpectBegin()
				mock.ExpectExec(`INSERT INTO follows`).WithArgs(userA.ID, userB.ID).
					WillReturnError(errors.New("database error"))
				mock.ExpectRollback()
			},
			expectedError: errors.New("database error"),
		},
		{
			name: "Follow Yourself Operation",
			setupMocks: func(mock sqlmock.Sqlmock, userA, _ *model.User) {
				mock.ExpectBegin()
				mock.ExpectExec(`INSERT INTO follows`).WithArgs(userA.ID, userA.ID).
					WillReturnError(errors.New("cannot follow yourself"))
				mock.ExpectRollback()
			},
			expectedError: errors.New("cannot follow yourself"),
		},
		{
			name: "User Does Not Exist in Database",
			setupMocks: func(mock sqlmock.Sqlmock, userA, userB *model.User) {
				mock.ExpectBegin()
				mock.ExpectExec(`INSERT INTO follows`).WithArgs(userA.ID, userB.ID).
					WillReturnError(errors.New("user does not exist"))
				mock.ExpectRollback()
			},
			expectedError: errors.New("user does not exist"),
		},
		{
			name: "Follow Operation Produces a Cycle in 'Following' Graph",
			setupMocks: func(mock sqlmock.Sqlmock, userA, userB *model.User) {
				mock.ExpectBegin()
				mock.ExpectExec(`INSERT INTO follows`).WithArgs(userA.ID, userB.ID).
					WillReturnError(errors.New("cycle detected"))
				mock.ExpectRollback()
			},
			expectedError: errors.New("cycle detected"),
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
				t.Fatalf("failed to initialize gorm DB: %v", err)
			}
			defer gormDB.Close()

			gormDB.LogMode(true)

			userA := &model.User{Model: gorm.Model{ID: 1}, Username: "UserA"}
			userB := &model.User{Model: gorm.Model{ID: 2}, Username: "UserB"}

			tc.setupMocks(mock, userA, userB)

			store := &UserStore{db: gormDB}
			err = store.Follow(userA, userB)

			if tc.expectedError != nil {
				if err == nil || err.Error() != tc.expectedError.Error() {
					t.Errorf("expected error: %v, got: %v", tc.expectedError, err)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}

			t.Logf("Test case '%s' executed", tc.name)
		})
	}
}


/*
ROOST_METHOD_HASH=IsFollowing_f53a5d9cef
ROOST_METHOD_SIG_HASH=IsFollowing_9eba5a0e9c


 */
func TestUserStoreIsFollowing(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	gormDB, err := gorm.Open("mysql", db)
	if err != nil {
		t.Fatalf("unexpected error when opening gorm: %v", err)
	}
	defer gormDB.Close()

	userStore := &UserStore{db: gormDB}

	type TestCase struct {
		name          string
		userA         *model.User
		userB         *model.User
		mockBehaviour func()
		expected      bool
		expectError   bool
	}

	testCases := []TestCase{
		{
			name: "Scenario 1: Check If User A Follows User B",
			userA: &model.User{
				Model: gorm.Model{ID: 1},
			},
			userB: &model.User{
				Model: gorm.Model{ID: 2},
			},
			mockBehaviour: func() {
				mock.ExpectQuery("^SELECT count(.+) FROM `follows` WHERE (.+)$").
					WithArgs(1, 2).
					WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
			},
			expected:    true,
			expectError: false,
		},
		{
			name: "Scenario 2: User A Does Not Follow User B",
			userA: &model.User{
				Model: gorm.Model{ID: 1},
			},
			userB: &model.User{
				Model: gorm.Model{ID: 2},
			},
			mockBehaviour: func() {
				mock.ExpectQuery("^SELECT count(.+) FROM `follows` WHERE (.+)$").
					WithArgs(1, 2).
					WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
			},
			expected:    false,
			expectError: false,
		},
		{
			name:          "Scenario 3: User A or User B is Nil",
			userA:         nil,
			userB:         nil,
			mockBehaviour: func() {},
			expected:      false,
			expectError:   false,
		},
		{
			name: "Scenario 4: Database Error Handling",
			userA: &model.User{
				Model: gorm.Model{ID: 1},
			},
			userB: &model.User{
				Model: gorm.Model{ID: 2},
			},
			mockBehaviour: func() {
				mock.ExpectQuery("^SELECT count(.+) FROM `follows` WHERE (.+)$").
					WithArgs(1, 2).
					WillReturnError(gorm.ErrInvalidSQL)
			},
			expected:    false,
			expectError: true,
		},
		{
			name: "Scenario 5: Check for Self-following",
			userA: &model.User{
				Model: gorm.Model{ID: 1},
			},
			userB: &model.User{
				Model: gorm.Model{ID: 1},
			},
			mockBehaviour: func() {
				mock.ExpectQuery("^SELECT count(.+) FROM `follows` WHERE (.+)$").
					WithArgs(1, 1).
					WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
			},
			expected:    false,
			expectError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.mockBehaviour()

			result, err := userStore.IsFollowing(tc.userA, tc.userB)

			if tc.expectError {
				if err == nil {
					t.Logf("expected error but got nil")
					t.Fail()
				}
			} else {
				if err != nil {
					t.Logf("did not expect error but got: %v", err)
					t.Fail()
				}
			}

			if result != tc.expected {
				t.Errorf("expected %v, got %v", tc.expected, result)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %v", err)
			}
		})
	}
}


/*
ROOST_METHOD_HASH=Unfollow_57959a8a53
ROOST_METHOD_SIG_HASH=Unfollow_8bd8e0bc55


 */
func TestUserStoreUnfollow(t *testing.T) {
	tests := []struct {
		name          string
		follower      model.User
		followee      model.User
		setupMockFunc func(mock sqlmock.Sqlmock)
		expectedError error
	}{
		{
			name: "Successfully Unfollow a User",
			follower: model.User{
				Model:   gorm.Model{ID: 1},
				Follows: []model.User{{Model: gorm.Model{ID: 2}}},
			},
			followee: model.User{
				Model: gorm.Model{ID: 2},
			},
			setupMockFunc: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec(`DELETE FROM "follows"`).
					WithArgs(1, 2).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
			expectedError: nil,
		},
		{
			name: "Attempt to Unfollow a Non-Followed User",
			follower: model.User{
				Model:   gorm.Model{ID: 1},
				Follows: []model.User{},
			},
			followee: model.User{
				Model: gorm.Model{ID: 2},
			},
			setupMockFunc: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec(`DELETE FROM "follows"`).
					WithArgs(1, 2).
					WillReturnResult(sqlmock.NewResult(1, 0))
				mock.ExpectCommit()
			},
			expectedError: nil,
		},
		{
			name: "Unfollow with a Non-Existent User in the Database",
			follower: model.User{
				Model: gorm.Model{ID: 1},
			},
			followee: model.User{
				Model: gorm.Model{ID: 999},
			},
			setupMockFunc: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec(`DELETE FROM "follows"`).
					WithArgs(1, 999).
					WillReturnError(gorm.ErrRecordNotFound)
				mock.ExpectRollback()
			},
			expectedError: gorm.ErrRecordNotFound,
		},
		{
			name: "Database Error During Unfollow Operation",
			follower: model.User{
				Model:   gorm.Model{ID: 1},
				Follows: []model.User{{Model: gorm.Model{ID: 2}}},
			},
			followee: model.User{
				Model: gorm.Model{ID: 2},
			},
			setupMockFunc: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec(`DELETE FROM "follows"`).
					WithArgs(1, 2).
					WillReturnError(gorm.ErrCantStartTransaction)
				mock.ExpectRollback()
			},
			expectedError: gorm.ErrCantStartTransaction,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			assert.NoError(t, err)
			defer db.Close()
			gormDB, err := gorm.Open("postgres", db)
			assert.NoError(t, err)

			tt.setupMockFunc(mock)

			userStore := &UserStore{db: gormDB}
			err = userStore.Unfollow(&tt.follower, &tt.followee)

			if tt.expectedError != nil {
				assert.EqualError(t, err, tt.expectedError.Error())
			} else {
				assert.NoError(t, err)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}


/*
ROOST_METHOD_HASH=GetFollowingUserIDs_ba3670aa2c
ROOST_METHOD_SIG_HASH=GetFollowingUserIDs_55ccc2afd7


 */
func TestUserStoreGetFollowingUserIDs(t *testing.T) {

	setupMockDB := func() (*UserStore, sqlmock.Sqlmock) {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
		}

		return &UserStore{db: &gorm.DB{CommonDB: db}}, mock
	}

	testCases := []struct {
		name          string
		prepareMock   func(mock sqlmock.Sqlmock)
		user          model.User
		expectedIDs   []uint
		expectedError bool
	}{
		{
			name: "Successfully Retrieve Following User IDs",
			prepareMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"to_user_id"}).
					AddRow(uint(1)).
					AddRow(uint(2)).
					AddRow(uint(3))

				mock.ExpectQuery("^SELECT (.+) FROM follows WHERE (.+)$").
					WithArgs(uint(42)).
					WillReturnRows(rows)
			},
			user:        model.User{Model: gorm.Model{ID: uint(42)}},
			expectedIDs: []uint{1, 2, 3},
		},
		{
			name: "User Is Not Following Anyone",
			prepareMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("^SELECT (.+) FROM follows WHERE (.+)$").
					WithArgs(uint(42)).
					WillReturnRows(sqlmock.NewRows([]string{"to_user_id"}))
			},
			user:        model.User{Model: gorm.Model{ID: uint(42)}},
			expectedIDs: []uint{},
		},
		{
			name: "Database Error While Retrieving Followers",
			prepareMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("^SELECT (.+) FROM follows WHERE (.+)$").
					WithArgs(uint(42)).
					WillReturnError(errors.New("db error"))
			},
			user:          model.User{Model: gorm.Model{ID: uint(42)}},
			expectedIDs:   []uint{},
			expectedError: true,
		},
		{
			name: "Large Number of Followings",
			prepareMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"to_user_id"})
				for i := 0; i < 1000; i++ {
					rows.AddRow(uint(i + 1))
				}
				mock.ExpectQuery("^SELECT (.+) FROM follows WHERE (.+)$").
					WithArgs(uint(42)).
					WillReturnRows(rows)
			},
			user:        model.User{Model: gorm.Model{ID: uint(42)}},
			expectedIDs: createSequentialSlice(1000),
		},
		{
			name: "User ID Does Not Exist",
			prepareMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("^SELECT (.+) FROM follows WHERE (.+)$").
					WithArgs(uint(100)).
					WillReturnRows(sqlmock.NewRows([]string{"to_user_id"}))
			},
			user:        model.User{Model: gorm.Model{ID: uint(100)}},
			expectedIDs: []uint{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			store, mock := setupMockDB()
			defer store.db.Close()

			tc.prepareMock(mock)
			ids, err := store.GetFollowingUserIDs(&tc.user)

			if tc.expectedError {
				if err == nil {
					t.Errorf("expected error but got nil")
				}
			} else {
				if err != nil {
					t.Errorf("did not expect error but got %v", err)
				}
				if !equalSlices(ids, tc.expectedIDs) {
					t.Errorf("expected %v but got %v", tc.expectedIDs, ids)
				}
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func createSequentialSlice(n int) []uint {
	slice := make([]uint, n)
	for i := 0; i < n; i++ {
		slice[i] = uint(i + 1)
	}
	return slice
}

func equalSlices(a, b []uint) bool {
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

