package github

import (
	"os"
	"testing"
	"time"
	"github.com/dgrijalva/jwt-go"
	"github.com/stretchr/testify/assert"
)









/*
ROOST_METHOD_HASH=GenerateToken_b7f5ef3740
ROOST_METHOD_SIG_HASH=GenerateToken_d10a3e47a3

FUNCTION_DEF=func GenerateToken(id uint) (string, error) 

 */
func TestGenerateToken(t *testing.T) {

	originalJwtSecret := os.Getenv("JWT_SECRET")
	defer os.Setenv("JWT_SECRET", originalJwtSecret)

	testCases := []struct {
		name       string
		userID     uint
		expectErr  bool
		setup      func()
		assertFunc func(t *testing.T, token string, err error)
	}{
		{
			name:   "Generate Token with Valid UserID",
			userID: 1,
			setup: func() {
				os.Setenv("JWT_SECRET", "test_secret")
			},
			expectErr: false,
			assertFunc: func(t *testing.T, token string, err error) {
				if err != nil {
					t.Errorf("expected no error, got %v", err)
				}
				if token == "" {
					t.Errorf("expected valid token, got empty string")
				}

				parsedClaims := &claims{}
				_, parseErr := jwt.ParseWithClaims(token, parsedClaims, func(token *jwt.Token) (interface{}, error) {
					return []byte(os.Getenv("JWT_SECRET")), nil
				})
				if parseErr != nil || parsedClaims.UserID != 1 {
					t.Errorf("expected token to encode user ID 1, got: %v", parsedClaims)
				}
			},
		},
		{
			name:   "Generate Token with Zero UserID",
			userID: 0,
			setup: func() {
				os.Setenv("JWT_SECRET", "test_secret")
			},
			expectErr: false,
			assertFunc: func(t *testing.T, token string, err error) {
				if err != nil {
					t.Errorf("expected no error, got %v", err)
				}
				if token == "" {
					t.Errorf("expected valid token, got empty string")
				}
			},
		},
		{
			name:   "Error Handling for Missing JWT Secret",
			userID: 1,
			setup: func() {
				os.Setenv("JWT_SECRET", "")
			},
			expectErr: true,
			assertFunc: func(t *testing.T, token string, err error) {
				if err == nil {
					t.Errorf("expected an error due to missing secret, got nil")
				}
			},
		},
		{
			name:   "Validate Token Expiry in Generated Token",
			userID: 1,
			setup: func() {
				os.Setenv("JWT_SECRET", "test_secret")
			},
			expectErr: false,
			assertFunc: func(t *testing.T, token string, err error) {
				if err != nil {
					t.Errorf("expected no error, got %v", err)
				}
				parsedClaims := &claims{}
				_, parseErr := jwt.ParseWithClaims(token, parsedClaims, func(token *jwt.Token) (interface{}, error) {
					return []byte(os.Getenv("JWT_SECRET")), nil
				})
				if parseErr != nil {
					t.Errorf("failed to parse token: %v", err)
				}
				expectedExpiry := time.Now().Add(time.Hour * 72).Unix()
				if parsedClaims.ExpiresAt < expectedExpiry-30 || parsedClaims.ExpiresAt > expectedExpiry+30 {
					t.Errorf("unexpected expiry time, got: %v", parsedClaims.ExpiresAt)
				}
			},
		},
		{
			name:   "Generate Token with Large UserID",
			userID: 4294967295,
			setup: func() {
				os.Setenv("JWT_SECRET", "test_secret")
			},
			expectErr: false,
			assertFunc: func(t *testing.T, token string, err error) {
				if err != nil {
					t.Errorf("expected no error, got %v", err)
				}
				if token == "" {
					t.Errorf("expected valid token, got empty string")
				}
			},
		},
		{
			name:   "Concurrent Token Generation",
			userID: 1,
			setup: func() {
				os.Setenv("JWT_SECRET", "test_secret")
			},
			expectErr: false,
			assertFunc: func(t *testing.T, token string, err error) {
				if err != nil {
					t.Errorf("expected no error, got %v", err)
				}
				if token == "" {
					t.Errorf("expected valid token, got empty string")
				}
				tokens := make(chan string, 10)
				errors := make(chan error, 10)
				for i := 0; i < 10; i++ {
					go func() {
						tok, tokErr := GenerateToken(1)
						tokens <- tok
						errors <- tokErr
					}()
				}

				for i := 0; i < 10; i++ {
					if tokErr := <-errors; tokErr != nil {
						t.Errorf("expected no error in concurrent generation, got %v", tokErr)
					}
					if tok := <-tokens; tok == "" {
						t.Errorf("expected non-empty token in concurrent generation")
					}
				}
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.setup()
			token, err := GenerateToken(tc.userID)
			tc.assertFunc(t, token, err)
			t.Logf("Test %s completed successfully", tc.name)
		})
	}
}


/*
ROOST_METHOD_HASH=GenerateTokenWithTime_d0df64aa69
ROOST_METHOD_SIG_HASH=GenerateTokenWithTime_72dd09cde6

FUNCTION_DEF=func GenerateTokenWithTime(id uint, t time.Time) (string, error) 

 */
func TestGenerateTokenWithTime(t *testing.T) {
	originalJwtSecret := os.Getenv("JWT_SECRET")
	defer os.Setenv("JWT_SECRET", originalJwtSecret)

	validUserID := uint(1)
	excessivelyLargeUserID := uint(^uint(0))
	futureTime := time.Now().Add(24 * time.Hour)
	pastTime := time.Now().Add(-24 * time.Hour)
	currentTime := time.Now()

	tests := []struct {
		name        string
		userID      uint
		time        time.Time
		expectError bool
		jwtSecret   string
		validate    func(t *testing.T, token string, err error)
	}{
		{
			name:      "Valid User ID and Current Time",
			userID:    validUserID,
			time:      currentTime,
			jwtSecret: "test-secret",
			validate: func(t *testing.T, token string, err error) {
				assert.NoError(t, err)
				assert.NotEmpty(t, token, "Token should be non-empty")
			},
		},
		{
			name:      "Valid User ID and Future Time",
			userID:    validUserID,
			time:      futureTime,
			jwtSecret: "test-secret",
			validate: func(t *testing.T, token string, err error) {
				assert.NoError(t, err)
				assert.NotEmpty(t, token, "Token should be non-empty")
			},
		},
		{
			name:      "Valid User ID and Past Time",
			userID:    validUserID,
			time:      pastTime,
			jwtSecret: "test-secret",
			validate: func(t *testing.T, token string, err error) {
				assert.NoError(t, err)
				assert.NotEmpty(t, token, "Token should be non-empty")

			},
		},
		{
			name:      "Invalid User ID (Zero Value)",
			userID:    0,
			time:      currentTime,
			jwtSecret: "test-secret",
			validate: func(t *testing.T, token string, err error) {
				assert.Error(t, err, "Should return an error for invalid user ID")
			},
		},
		{
			name:      "Missing JWT Secret Environment Variable",
			userID:    validUserID,
			time:      currentTime,
			jwtSecret: "",
			validate: func(t *testing.T, token string, err error) {
				assert.Error(t, err, "Should return an error when JWT_SECRET is missing")
			},
		},
		{
			name:      "Excessively Large User ID",
			userID:    excessivelyLargeUserID,
			time:      currentTime,
			jwtSecret: "test-secret",
			validate: func(t *testing.T, token string, err error) {

				assert.Error(t, err, "Should handle excessively large user ID gracefully")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv("JWT_SECRET", tt.jwtSecret)
			token, err := GenerateTokenWithTime(tt.userID, tt.time)
			tt.validate(t, token, err)
			if err != nil {
				t.Logf("Error: %v", err)
			} else {
				t.Logf("Generated Token: %s", token)
			}
		})
	}
}


/*
ROOST_METHOD_HASH=generateToken_2cc40e0108
ROOST_METHOD_SIG_HASH=generateToken_9de4114fe8

FUNCTION_DEF=func generateToken(id uint, now time.Time) (string, error) 

 */
func TestGenerateToken(t *testing.T) {
	type testCase struct {
		name          string
		userID        uint
		now           time.Time
		setupEnv      func()
		expectedError error
	}

	testCases := []testCase{
		{
			name:   "Scenario 1: Successfully Generate Token with Valid User ID and Current Time",
			userID: 1,
			now:    time.Now(),
			setupEnv: func() {
				os.Setenv("JWT_SECRET", "mysecretkey")
			},
			expectedError: nil,
		},
		{
			name:   "Scenario 2: Handle Missing JWT Secret Environment Variable",
			userID: 1,
			now:    time.Now(),
			setupEnv: func() {
				os.Unsetenv("JWT_SECRET")
			},
			expectedError: jwt.ErrSignatureInvalid,
		},
		{
			name:   "Scenario 3: Verify Token Expiration Time Calculation",
			userID: 1,
			now:    time.Unix(1600000000, 0),
			setupEnv: func() {
				os.Setenv("JWT_SECRET", "mysecretkey")
			},
			expectedError: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.setupEnv()
			jwtSecret = []byte(os.Getenv("JWT_SECRET"))

			tokenString, err := generateToken(tc.userID, tc.now)

			if tc.expectedError != nil {
				if err == nil || err.Error() != tc.expectedError.Error() {
					t.Errorf("expected error: %v, but got: %v", tc.expectedError, err)
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if tokenString == "" {
				t.Error("expected non-empty token string, but got empty string")
			}

			token, _ := jwt.ParseWithClaims(tokenString, &claims{}, func(token *jwt.Token) (interface{}, error) {
				return jwtSecret, nil
			})

			if token == nil {
				t.Error("expected valid token, but got nil")
				return
			}

			if claims, ok := token.Claims.(*claims); ok && token.Valid {
				if claims.ExpiresAt != tc.now.Add(72*time.Hour).Unix() {
					t.Errorf("expected expiration time: %v, but got: %v", tc.now.Add(72*time.Hour).Unix(), claims.ExpiresAt)
				}
			} else {
				t.Error("failed to parse token claims")
			}
		})
	}
}

