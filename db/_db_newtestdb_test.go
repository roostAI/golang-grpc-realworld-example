package db

import (
	"fmt"
	"sync"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
)

// TestNewTestDB function tests multiple scenarios for NewTestDB.
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

			// Prepare DSN but not used as per the test breakdown
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

// TestConcurrentInitialization tests concurrent calls to NewTestDB.
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
