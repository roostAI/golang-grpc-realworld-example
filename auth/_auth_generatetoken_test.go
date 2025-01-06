package auth

import (
	"errors"
	"os"
	"testing"
	"time"
	"github.com/dgrijalva/jwt-go"
	"github.com/stretchr/testify/assert"
)





type T struct {
	common
	isParallel bool
	isEnvSet   bool
	context    *testContext // For running tests and subtests.
}
func TestGenerateToken(t *testing.T) {
	type testCase struct {
		name          string
		userID        uint
		setupEnv      func()
		expectedError error
		validate      func(t *testing.T, token string, err error)
	}

	testCases := []testCase{
		{
			name:   "Scenario 1: Successful Token Generation for Valid User ID",
			userID: 123,
			setupEnv: func() {
				os.Setenv("JWT_SECRET", "testsecret")
			},
			expectedError: nil,
			validate: func(t *testing.T, token string, err error) {
				assert.NotEmpty(t, token, "Token should not be empty")
				assert.NoError(t, err, "Error should be nil")
			},
		},
		{
			name:   "Scenario 2: Error Handling for Missing JWT Secret",
			userID: 123,
			setupEnv: func() {
				os.Setenv("JWT_SECRET", "")
			},
			expectedError: errors.New("secret key is missing"),
			validate: func(t *testing.T, token string, err error) {
				assert.Error(t, err, "Error should be returned when JWT_SECRET is missing")
			},
		},
		{
			name:   "Scenario 3: Error Handling for Generation with Invalid User ID",
			userID: 0,
			setupEnv: func() {
				os.Setenv("JWT_SECRET", "testsecret")
			},
			expectedError: errors.New("invalid user ID"),
			validate: func(t *testing.T, token string, err error) {
				assert.Error(t, err, "Error should be returned for invalid user ID")
			},
		},
		{
			name:   "Scenario 4: Token Generation Timing Check",
			userID: 456,
			setupEnv: func() {
				os.Setenv("JWT_SECRET", "testsecret")
			},
			expectedError: nil,
			validate: func(t *testing.T, token string, err error) {
				claims := jwt.StandardClaims{}
				_, err := jwt.ParseWithClaims(token, &claims, func(token *jwt.Token) (interface{}, error) {
					return jwtSecret, nil
				})
				assert.NoError(t, err, "Error parsing token should be nil")
				assert.WithinDuration(t, time.Now(), time.Unix(claims.IssuedAt, 0), time.Second, "IssuedAt should be close to the current time")
			},
		},
		{
			name:   "Scenario 5: Token Content Verification",
			userID: 789,
			setupEnv: func() {
				os.Setenv("JWT_SECRET", "testsecret")
			},
			expectedError: nil,
			validate: func(t *testing.T, token string, err error) {
				claims := &claims{}
				_, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
					return jwtSecret, nil
				})
				assert.NoError(t, err, "No error should occur while parsing token")
				assert.Equal(t, claims.UserID, uint(789), "Expected userID claims to be present in token")
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.setupEnv()
			token, err := GenerateToken(tc.userID)
			tc.validate(t, token, err)
			if tc.expectedError != nil {
				assert.Error(t, err, tc.expectedError.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}

	defer os.Unsetenv("JWT_SECRET")
}
