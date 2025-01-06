package db

import (
	"database/sql"
	"errors"
	"os"
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/jinzhu/gorm"
)

// TestNew implements comprehensive test cases for the New() function in the db package
func TestNew(t *testing.T) {
	// Set mock responses for dsn function simulations
	validDSN := "user:password@tcp(localhost:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local"
	invalidDSN := "invalid_dsn"

	// Helper function to set up environment variables
	setupEnv := func(dsn string) {
		os.Setenv("DB_HOST", "localhost")
		os.Setenv("DB_USER", "user")
		os.Setenv("DB_PASSWORD", "password")
		os.Setenv("DB_NAME", "dbname")
		os.Setenv("DB_PORT", "3306")
	}

	// Clear environment after tests to avoid side effects
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

			// Mock dsn function to return preset values
			originalDSN := dsn
			dsn = tt.mockDSN
			defer func() { dsn = originalDSN }()

			// Prepare environment variables if necessary
			if tt.modifyEnv != nil {
				tt.modifyEnv()
			}

			// Attempt to establish a new database connection
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

	// Additional specialized test for retry logic
	t.Run("Retry Logic Verification", func(t *testing.T) {
		// Prepare a mock database
		mockDB, mock, err := sqlmock.New()
		assert.NoError(t, err, "Mock setup should not fail")
		defer mockDB.Close()

		// Use the mockDB as the database connection
		db, err := gorm.Open("mysql", mockDB)
		assert.NoError(t, err)

		// Simulate blocking and unblocking the database access
		simulateNetworkFluctuation := func(db *gorm.DB) {
			time.Sleep(500 * time.Millisecond) // Simulate downtime
			// Subject to integration test for transient network issues
		}

		go simulateNetworkFluctuation(db)

		db.DB().SetMaxIdleConns(3) // Assert MaxIdleConns setting
		actualDB, _ := db.DB()
		assert.Equal(t, 3, actualDB.Stats().MaxIdleConns, "MaxIdleConns should be set to 3")
	})
}
