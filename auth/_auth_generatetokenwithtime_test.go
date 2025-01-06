package auth

import (
	"os"
	"testing"
	"time"
	"github.com/dgrijalva/jwt-go"
)


type T struct {
	common
	isParallel bool
	isEnvSet   bool
	context    *testContext // For running tests and subtests.
}


func TestGenerateTokenWithTime(t *testing.T) {

	type testCase struct {
		desc     string
		userID   uint
		time     time.Time
		expected bool
		envVar   string
	}

	originalSecret := os.Getenv("JWT_SECRET")
	defer os.Setenv("JWT_SECRET", originalSecret)

	testCases := []testCase{
		{
			desc:     "Successful Token Generation",
			userID:   1,
			time:     time.Now(),
			expected: true,
			envVar:   "secret-key",
		},
		{
			desc:     "Token Generation with Future Date",
			userID:   1,
			time:     time.Now().Add(24 * time.Hour),
			expected: true,
			envVar:   "secret-key",
		},
		{
			desc:     "Token Generation with Past Date",
			userID:   1,
			time:     time.Now().Add(-24 * time.Hour),
			expected: true,
			envVar:   "secret-key",
		},
		{
			desc:     "Handling Invalid User ID",
			userID:   0,
			time:     time.Now(),
			expected: false,
			envVar:   "secret-key",
		},
		{
			desc:     "Undefined JWT Secret Environment Variable",
			userID:   1,
			time:     time.Now(),
			expected: false,
			envVar:   "",
		},
		{
			desc:     "Maximal User ID and Current Time",
			userID:   ^uint(0),
			time:     time.Now(),
			expected: true,
			envVar:   "secret-key",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			os.Setenv("JWT_SECRET", tc.envVar)
			token, err := GenerateTokenWithTime(tc.userID, tc.time)

			if tc.expected {
				if err != nil || token == "" {
					t.Errorf("expected valid token, got error: %v", err)
				} else {
					t.Logf("Test Passed: %s", tc.desc)
				}
			} else {
				if err == nil || token != "" {
					t.Errorf("expected error or empty token, got: %s", token)
				} else {
					t.Logf("Test Passed: %s", tc.desc)
				}
			}
		})
	}
}
