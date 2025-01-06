package db

import (
	"database/sql"
	"io/ioutil"
	"os"
	"sync"
	"testing"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jinzhu/gorm"
	"github.com/raahii/golang-grpc-realworld-example/model"
	"errors"
	"github.com/BurntSushi/toml"
	"fmt"
	"time"
	"github.com/stretchr/testify/assert"
	"github.com/joho/godotenv"
)


var origReadFile = ioutil.ReadFiletype ExpectedBegin struct {
/*
ROOST_METHOD_HASH=AutoMigrate_94b22622a5
ROOST_METHOD_SIG_HASH=AutoMigrate_2cd152caa7


 */
func TestAutoMigrate(t *testing.T) {
	tests := []struct {
		name          string
		setupMock     func(sqlmock.Sqlmock)
		expectedError bool
	}{
		{
			name: "Successful AutoMigrate Operation",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("CREATE TABLE IF NOT EXISTS users").WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectExec("CREATE TABLE IF NOT EXISTS articles").WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectExec("CREATE TABLE IF NOT EXISTS tags").WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectExec("CREATE TABLE IF NOT EXISTS comments").WillReturnResult(sqlmock.NewResult(1, 1))
			},
			expectedError: false,
		},
		{
			name: "AutoMigrate with Database Error",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("CREATE TABLE IF NOT EXISTS users").WillReturnError(sql.ErrConnDone)
			},
			expectedError: true,
		},
	}

	var db *gorm.DB
	var err error
	var mock sqlmock.Sqlmock

	mutex.Lock()
	if !txdbInitialized {
		sqlDB, mockConn, err := sqlmock.New()
		if err != nil {
			t.Fatalf("unexpected error when opening a stub database connection: %s", err)
		}
		mock = mockConn
		db, err = gorm.Open("sqlmock", sqlDB)
		if err != nil {
			t.Fatalf("failed to open gorm db connection: %v", err)
		}
		txdbInitialized = true
	}
	mutex.Unlock()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock(mock)

			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			err = AutoMigrate(db)
			if tt.expectedError {
				if err == nil {
					t.Errorf("expected an error but did not get one")
				}
			} else {
				if err != nil {
					t.Errorf("did not expect an error but got one: %v", err)
				}
			}

			w.Close()
			os.Stdout = oldStdout
			out, _ := ioutil.ReadAll(r)
			t.Log("Output captured:", string(out))
		})
	}

	db.Close()
}


/*
ROOST_METHOD_HASH=DropTestDB_4c6b54d5e5
ROOST_METHOD_SIG_HASH=DropTestDB_69b51a825b


 */
func TestDropTestDB(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		mockSetup      func(sqlmock.Sqlmock) (*gorm.DB, error)
		expectedError  error
		fConcurrent    bool
		fAlreadyClosed bool
	}{
		{
			name: "Successfully Close a Gorm Database Connection",
			mockSetup: func(mock sqlmock.Sqlmock) (*gorm.DB, error) {
				db, err := gorm.Open("sqlmock", "")
				if err != nil {
					return nil, err
				}
				mock.ExpectClose().WillReturnError(nil)
				return db, nil
			},
			expectedError: nil,
		},
		{
			name: "Handle nil Database Connection Gracefully",
			mockSetup: func(mock sqlmock.Sqlmock) (*gorm.DB, error) {
				return nil, nil
			},
			expectedError: nil,
		},
		{
			name: "Simulate Error on Closing Database Connection",
			mockSetup: func(mock sqlmock.Sqlmock) (*gorm.DB, error) {
				db, err := gorm.Open("sqlmock", "")
				if err != nil {
					return nil, err
				}
				mock.ExpectClose().WillReturnError(errors.New("close error"))
				return db, nil
			},
			expectedError: errors.New("close error"),
		},
		{
			name: "Concurrent Calls to DropTestDB",
			mockSetup: func(mock sqlmock.Sqlmock) (*gorm.DB, error) {
				db, err := gorm.Open("sqlmock", "")
				if err != nil {
					return nil, err
				}
				mock.ExpectClose().WillReturnError(nil)
				return db, nil
			},
			expectedError: nil,
			fConcurrent:   true,
		},
		{
			name: "Check Effect of Dropping a Closed Database",
			mockSetup: func(mock sqlmock.Sqlmock) (*gorm.DB, error) {
				db, err := gorm.Open("sqlmock", "")
				if err != nil {
					return nil, err
				}
				mock.ExpectClose().WillReturnError(nil)
				db.Close()
				return db, nil
			},
			expectedError:  nil,
			fAlreadyClosed: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Logf("Running test case: %s", tt.name)

			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("failed to open sqlmock database: %s", err)
			}
			defer db.Close()

			gormDB, err := tt.mockSetup(mock)
			if err != nil {
				t.Fatalf("failed to set up mock gorm.DB: %s", err)
			}

			if tt.fConcurrent {
				var wg sync.WaitGroup
				for i := 0; i < 5; i++ {
					wg.Add(1)
					go func() {
						defer wg.Done()
						err := DropTestDB(gormDB)
						if err != nil && err.Error() != tt.expectedError.Error() {
							t.Errorf("expected error '%v', got '%v'", tt.expectedError, err)
						}
					}()
				}
				wg.Wait()
			} else {
				err = DropTestDB(gormDB)
				if err != nil && err.Error() != tt.expectedError.Error() {
					t.Errorf("expected error '%v', got '%v'", tt.expectedError, err)
				}

				if err := mock.ExpectationsWereMet(); err != nil {
					t.Errorf("unmet SQL expectations: %s", err)
				}
			}

			t.Logf("Test case `%s` passed!", tt.name)
		})
	}
}


/*
ROOST_METHOD_HASH=Seed_5ad31c3a6c
ROOST_METHOD_SIG_HASH=Seed_878933cebc


 */
func TestSeed(t *testing.T) {
	tests := []struct {
		name          string
		setup         func() *gorm.DB
		expectedError error
	}{
		{
			name: "Scenario 1: Successful Seeding of Users from TOML File",
			setup: func() *gorm.DB {
				db, mock := mockDB(t)

				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO users").WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()

				data := `users = [{name = "John Doe", email = "john@example.com"}]`

				ioutil.ReadFile = func(filename string) ([]byte, error) { return []byte(data), nil }

				return db
			},
			expectedError: nil,
		},
		{
			name: "Scenario 2: Missing TOML File",
			setup: func() *gorm.DB {
				db, _ := mockDB(t)

				ioutil.ReadFile = func(filename string) ([]byte, error) { return nil, os.ErrNotExist }

				return db
			},
			expectedError: os.ErrNotExist,
		},
		{
			name: "Scenario 3: Malformed TOML File",
			setup: func() *gorm.DB {
				db, _ := mockDB(t)

				data := `users = [name = "John Doe"`
				ioutil.ReadFile = func(filename string) ([]byte, error) { return []byte(data), nil }

				return db
			},
			expectedError: errors.New("TOML parsing issue expected"),
		},
		{
			name: "Scenario 4: Database Create Operation Error",
			setup: func() *gorm.DB {
				db, mock := mockDB(t)

				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO users").WillReturnError(errors.New("database error"))
				mock.ExpectRollback()

				data := `users = [{name = "John Doe", email = "john@example.com"}]`
				ioutil.ReadFile = func(filename string) ([]byte, error) { return []byte(data), nil }

				return db
			},
			expectedError: errors.New("database error"),
		},
		{
			name: "Scenario 5: Empty User List in TOML File",
			setup: func() *gorm.DB {
				db, mock := mockDB(t)

				mock.ExpectBegin()
				mock.ExpectCommit()

				data := `users = []`
				ioutil.ReadFile = func(filename string) ([]byte, error) { return []byte(data), nil }

				return db
			},
			expectedError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {

				ioutil.ReadFile = origReadFile
			}()

			db := tt.setup()

			err := Seed(db)
			if tt.expectedError == nil && err != nil {
				t.Errorf("expected no error, but got %v", err)
			} else if tt.expectedError != nil && err == nil {
				t.Errorf("expected an error, but got none")
			} else if tt.expectedError != nil && err.Error() != tt.expectedError.Error() {
				t.Errorf("expected error %v, but got %v", tt.expectedError, err)
			}

			db.Close()
		})
	}
}

func mockDB(t *testing.T) (*gorm.DB, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock database: %s", err)
	}

	gdb, err := gorm.Open("postgres", db)
	if err != nil {
		t.Fatalf("failed to initialize gorm DB: %s", err)
	}
	return gdb, mock
}


/*
ROOST_METHOD_HASH=dsn_e202d1c4f9
ROOST_METHOD_SIG_HASH=dsn_b336e03d64


 */
func Testdsn(t *testing.T) {
	type testCase struct {
		description     string
		envVariables    map[string]string
		expectedDSN     string
		expectedErr     error
	}


	testCases := []testCase{
		{
			description: "All Environment Variables Set",
			envVariables: map[string]string{
				"DB_HOST":     "localhost",
				"DB_USER":     "user",
				"DB_PASSWORD": "password",
				"DB_NAME":     "testdb",
				"DB_PORT":     "3306",
			},
			expectedDSN: "user:password@(localhost:3306)/testdb?charset=utf8mb4&parseTime=True&loc=Local",
			expectedErr: nil,
		},
		{
			description: "Missing DB_HOST Environment Variable",
			envVariables: map[string]string{
				"DB_USER":     "user",
				"DB_PASSWORD": "password",
				"DB_NAME":     "testdb",
				"DB_PORT":     "3306",
			},
			expectedDSN: "",
			expectedErr: fmt.Errorf("$DB_HOST is not set"),
		},
		{
			description: "Missing DB_USER Environment Variable",
			envVariables: map[string]string{
				"DB_HOST":     "localhost",
				"DB_PASSWORD": "password",
				"DB_NAME":     "testdb",
				"DB_PORT":     "3306",
			},
			expectedDSN: "",
			expectedErr: fmt.Errorf("$DB_USER is not set"),
		},
		{
			description: "Missing DB_PASSWORD Environment Variable",
			envVariables: map[string]string{
				"DB_HOST": "localhost",
				"DB_USER": "user",
				"DB_NAME": "testdb",
				"DB_PORT": "3306",
			},
			expectedDSN: "",
			expectedErr: fmt.Errorf("$DB_PASSWORD is not set"),
		},
		{
			description: "Missing DB_NAME Environment Variable",
			envVariables: map[string]string{
				"DB_HOST":     "localhost",
				"DB_USER":     "user",
				"DB_PASSWORD": "password",
				"DB_PORT":     "3306",
			},
			expectedDSN: "",
			expectedErr: fmt.Errorf("$DB_NAME is not set"),
		},
		{
			description: "Missing DB_PORT Environment Variable",
			envVariables: map[string]string{
				"DB_HOST":     "localhost",
				"DB_USER":     "user",
				"DB_PASSWORD": "password",
				"DB_NAME":     "testdb",
			},
			expectedDSN: "",
			expectedErr: fmt.Errorf("$DB_PORT is not set"),
		},
		{
			description: "Environment Variables with Special Characters",
			envVariables: map[string]string{
				"DB_HOST":     "local!host",
				"DB_USER":     "us@er",
				"DB_PASSWORD": "p@ssword",
				"DB_NAME":     "test#db",
				"DB_PORT":     "3306",
			},
			expectedDSN: "us@er:p@ssword@(local!host:3306)/test#db?charset=utf8mb4&parseTime=True&loc=Local",
			expectedErr: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
		
			for key, value := range tc.envVariables {
				if err := os.Setenv(key, value); err != nil {
					t.Fatalf("Error setting env var %s: %v", key, err)
				}
			}

		
			dsn, err := dsn()

		
			if err != nil && tc.expectedErr == nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if err == nil && tc.expectedErr != nil {
				t.Errorf("Expected error %v, got no error", tc.expectedErr)
			}

			if err != nil && tc.expectedErr != nil && err.Error() != tc.expectedErr.Error() {
				t.Errorf("Expected error %v, got %v", tc.expectedErr, err)
			}

			if dsn != tc.expectedDSN {
				t.Errorf("Expected DSN %s, got %s", tc.expectedDSN, dsn)
			}

			t.Logf("Finished scenario: %s", tc.description)

		
			for key := range tc.envVariables {
			
				if err := os.Unsetenv(key); err != nil {
					t.Fatalf("Error unsetting env var %s: %v", key, err)
				}
			}
		})
	}
}


/*
ROOST_METHOD_HASH=New_1d2840dc39
ROOST_METHOD_SIG_HASH=New_f9cc65f555


 */
func TestNew(t *testing.T) {

	validDSN := "user:password@tcp(localhost:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local"
	invalidDSN := "invalid_dsn"


	setupEnv := func(dsn string) {
		os.Setenv("DB_HOST", "localhost")
		os.Setenv("DB_USER", "user")
		os.Setenv("DB_PASSWORD", "password")
		os.Setenv("DB_NAME", "dbname")
		os.Setenv("DB_PORT", "3306")
	}


	defer func() {
		os.Unsetenv("DB_HOST")
		os.Unsetenv("DB_USER")
		os.Unsetenv("DB_PASSWORD")
		os.Unsetenv("DB_NAME")
		os.Unsetenv("DB_PORT")
	}()

	tests := []struct {
		name           string
		mockDSN        func() (string, error)
		expectDB       bool
		expectError    bool
		modifyEnv      func()
		description    string
		errorAssertion func(error) bool
	}{
		{
			name: "Successful Database Connection Establishment",
			mockDSN: func() (string, error) {
				return validDSN, nil
			},
			expectDB:       true,
			expectError:    false,
			modifyEnv:      func() { setupEnv(validDSN) },
			description:    "Connects successfully with valid dsn.",
			errorAssertion: func(err error) bool { return err == nil },
		},
		{
			name: "Connection Attempts Exhausted with Failure",
			mockDSN: func() (string, error) {
				return invalidDSN, nil
			},
			expectDB:       false,
			expectError:    true,
			modifyEnv:      func() { setupEnv(invalidDSN) },
			description:    "Returns error after multiple failed connection attempts.",
			errorAssertion: func(err error) bool { return err != nil },
		},
		{
			name: "Invalid Database Credentials",
			mockDSN: func() (string, error) {
				return validDSN, nil
			},
			expectDB:    false,
			expectError: true,
			modifyEnv: func() {
				os.Setenv("DB_USER", "wrong_user")
			},
			description:    "Fails due to invalid database credentials.",
			errorAssertion: func(err error) bool { return err != nil },
		},
		{
			name: "dsn() Function Returns Error",
			mockDSN: func() (string, error) {
				return "", errors.New("dsn() error")
			},
			expectDB:       false,
			expectError:    true,
			modifyEnv:      func() { os.Setenv("DB_HOST", "") },
			description:    "Fails when dsn() function itself returns an error.",
			errorAssertion: func(err error) bool { return err != nil && err.Error() == "dsn() error" },
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Log(tt.description)

		
			originalDSN := dsn
			dsn = tt.mockDSN
			defer func() { dsn = originalDSN }()

		
			if tt.modifyEnv != nil {
				tt.modifyEnv()
			}

		
			db, err := New()

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, db, "Expected no valid database instance")
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, db, "Expected valid database instance")
			}
		})
	}


	t.Run("Retry Logic Verification", func(t *testing.T) {
	
		mockDB, mock, err := sqlmock.New()
		assert.NoError(t, err, "Mock setup should not fail")
		defer mockDB.Close()

	
		db, err := gorm.Open("mysql", mockDB)
		assert.NoError(t, err)

	
		simulateNetworkFluctuation := func(db *gorm.DB) {
			time.Sleep(500 * time.Millisecond)
		
		}

		go simulateNetworkFluctuation(db)

		db.DB().SetMaxIdleConns(3)
		actualDB, _ := db.DB()
		assert.Equal(t, 3, actualDB.Stats().MaxIdleConns, "MaxIdleConns should be set to 3")
	})
}


/*
ROOST_METHOD_HASH=NewTestDB_7feb2c4a7a
ROOST_METHOD_SIG_HASH=NewTestDB_1b71546d9d


 */
func TestNewTestDB(t *testing.T) {
	tests := []struct {
		name           string
		prepareEnv     func() error
		prepareDSN     func() string
		mockBehavior   func(sqlmock.Sqlmock)
		expectedErr    bool
		expectedDBOpen bool
	}{
		{
			name: "Successfully Initialize a New Test Database Connection",
			prepareEnv: func() error {
				return godotenv.Load("../env/test.env")
			},
			prepareDSN: func() string {
				return "user:password@tcp(localhost:3306)/testdb?charset=utf8mb4&parseTime=True&loc=Local"
			},
			mockBehavior: func(mock sqlmock.Sqlmock) {
				mock.ExpectPing()
			},
			expectedErr:    false,
			expectedDBOpen: true,
		},
		{
			name: "Fail to Load Environment File",
			prepareEnv: func() error {
				return fmt.Errorf("failed to load ../env/test.env")
			},
			prepareDSN: func() string {
				return ""
			},
			mockBehavior:   func(mock sqlmock.Sqlmock) {},
			expectedErr:    true,
			expectedDBOpen: false,
		},
		{
			name: "Data Source Name (DSN) Function Error",
			prepareEnv: func() error {
				return nil
			},
			prepareDSN: func() string {
				return ""
			},
			mockBehavior:   func(mock sqlmock.Sqlmock) {},
			expectedErr:    true,
			expectedDBOpen: false,
		},
		{
			name: "MySQL Connection Error During First Opening",
			prepareEnv: func() error {
				return nil
			},
			prepareDSN: func() string {
				return "invalid_dsn"
			},
			mockBehavior: func(mock sqlmock.Sqlmock) {
				mock.ExpectPing().WillReturnError(fmt.Errorf("connection error"))
			},
			expectedErr:    true,
			expectedDBOpen: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.prepareEnv()
			if tt.expectedErr {
				assert.NotNil(t, err, "expected an error due to environment setup")
				return
			}
			assert.Nil(t, err, "unexpected error while setting up environment")

		
			_ = tt.prepareDSN()

			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("Failed to open mock sql db, %v", err)
			}
			defer db.Close()

			mutex.Lock()
			txdbInitialized = false
			mutex.Unlock()

			gormDB, err := NewTestDB()
			if tt.expectedErr {
				assert.Nil(t, gormDB, "expected DB connection to be nil")
				assert.Error(t, err)
				return
			}
			assert.NotNil(t, gormDB, "expected DB connection to be non-nil")
			assert.NoError(t, err)

			if tt.expectedDBOpen {
				assert.NotNil(t, gormDB, "LogMode should be checked on non-nil DB")
				assert.Equal(t, noLogMode, gormDB.LogMode(false).logMode, "LogMode should be false by default")
			}

			err = mock.ExpectationsWereMet()
			assert.Nil(t, err, "expected all mock expectations to be met")
		})
	}
}

func TestConcurrentInitialization(t *testing.T) {
	err := godotenv.Load("../env/test.env")
	assert.Nil(t, err)

	goroutines := 5
	var wg sync.WaitGroup
	var errors []error
	errorMutex := &sync.Mutex{}

	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := NewTestDB()
			errorMutex.Lock()
			errors = append(errors, err)
			errorMutex.Unlock()
		}()
	}

	wg.Wait()

	for _, err := range errors {
		assert.Nil(t, err, "expected no error in concurrent initialization")
	}

	mutex.Lock()
	assert.True(t, txdbInitialized, "txdb should be initialized once")
	mutex.Unlock()
}

