package db

import (
	"fmt"
	"os"
	"testing"
)

func Testdsn(t *testing.T) {
	type testCase struct {
		description     string
		envVariables    map[string]string
		expectedDSN     string
		expectedErr     error
	}

	// Define the test scenarios
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
			// Setting up the environment variables
			for key, value := range tc.envVariables {
				if err := os.Setenv(key, value); err != nil {
					t.Fatalf("Error setting env var %s: %v", key, err)
				}
			}

			// Act
			dsn, err := dsn()

			// Assert
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

			// Clean up environment variables after each test case
			for key := range tc.envVariables {
				// Using `Unsetenv` in defer to ensure cleanup even if a test fails
				if err := os.Unsetenv(key); err != nil {
					t.Fatalf("Error unsetting env var %s: %v", key, err)
				}
			}
		})
	}
}
