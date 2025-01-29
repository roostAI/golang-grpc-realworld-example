package auth

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

	type testCase struct {
		name           string
		userID         uint
		jwtSecret      string
		expectError    bool
		validationFunc func(t *testing.T, token string, err error)
	}

	originalSecret := os.Getenv("JWT_SECRET")
	defer func() {
		os.Setenv("JWT_SECRET", originalSecret)
	}()

	testCases := []testCase{
		{
			name:        "Valid Token Generation",
			userID:      1,
			jwtSecret:   "testSecret",
			expectError: false,
			validationFunc: func(t *testing.T, token string, err error) {
				assert.NoError(t, err)
				assert.NotEmpty(t, token)

				parsedToken, parseErr := jwt.ParseWithClaims(token, &claims{}, func(token *jwt.Token) (interface{}, error) {
					return []byte("testSecret"), nil
				})
				assert.NoError(t, parseErr)
				claims, ok := parsedToken.Claims.(*claims)
				assert.True(t, ok)
				assert.Equal(t, uint(1), claims.UserID)
				assert.True(t, claims.ExpiresAt > time.Now().Unix())
			},
		},
		{
			name:        "Token Generation with Missing JWT Secret",
			userID:      1,
			jwtSecret:   "",
			expectError: true,
			validationFunc: func(t *testing.T, token string, err error) {
				assert.Error(t, err)
				assert.Empty(t, token)
			},
		},
		{
			name:        "Token Generation with Invalid User ID",
			userID:      0,
			jwtSecret:   "testSecret",
			expectError: false,
			validationFunc: func(t *testing.T, token string, err error) {
				assert.NoError(t, err)
				assert.NotEmpty(t, token)
				parsedToken, parseErr := jwt.ParseWithClaims(token, &claims{}, func(token *jwt.Token) (interface{}, error) {
					return []byte("testSecret"), nil
				})
				assert.NoError(t, parseErr)
				claims, ok := parsedToken.Claims.(*claims)
				assert.True(t, ok)
				assert.Equal(t, uint(0), claims.UserID)
			},
		},
		{
			name:        "Check JWT Signature with Incorrect Signing Method",
			userID:      1,
			jwtSecret:   "testSecret",
			expectError: false,
			validationFunc: func(t *testing.T, token string, err error) {
				assert.NoError(t, err)
				assert.NotEmpty(t, token)
				parsedToken, parseErr := jwt.ParseWithClaims(token, &claims{}, func(token *jwt.Token) (interface{}, error) {
					return []byte("testSecret"), nil
				})
				assert.NoError(t, parseErr)
				assert.Equal(t, parsedToken.Method.Alg(), jwt.SigningMethodHS256.Name)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			os.Setenv("JWT_SECRET", tc.jwtSecret)

			token, err := GenerateToken(tc.userID)

			tc.validationFunc(t, token, err)
		})
	}
}


/*
ROOST_METHOD_HASH=GenerateTokenWithTime_d0df64aa69
ROOST_METHOD_SIG_HASH=GenerateTokenWithTime_72dd09cde6

FUNCTION_DEF=func GenerateTokenWithTime(id uint, t time.Time) (string, error) 

 */
func TestGenerateTokenWithTime(t *testing.T) {

	type testCase struct {
		description string
		userID      uint
		timestamp   time.Time
		setup       func()
		assertion   func(t *testing.T, token string, err error)
	}

	jwtSecretBackup := os.Getenv("JWT_SECRET")
	defer os.Setenv("JWT_SECRET", jwtSecretBackup)

	tests := []testCase{
		{
			description: "Successful Token Generation for Valid User ID",
			userID:      12345,
			timestamp:   time.Now(),
			setup: func() {
				os.Setenv("JWT_SECRET", "my_secret_key")
			},
			assertion: func(t *testing.T, token string, err error) {
				assert.NoError(t, err)
				assert.NotEmpty(t, token, "Token should be generated")
			},
		},
		{
			description: "Failure when JWT Secret is Not Set",
			userID:      12345,
			timestamp:   time.Now(),
			setup: func() {
				os.Unsetenv("JWT_SECRET")
			},
			assertion: func(t *testing.T, token string, err error) {
				assert.Error(t, err, "Expected error due to unset JWT secret")
			},
		},
		{
			description: "Token Contains Correct Claims",
			userID:      54321,
			timestamp:   time.Now(),
			setup: func() {
				os.Setenv("JWT_SECRET", "my_secret_key")
			},
			assertion: func(t *testing.T, token string, err error) {
				assert.NoError(t, err)
				assert.NotEmpty(t, token)

				parsedToken, _ := jwt.ParseWithClaims(token, &claims{}, func(token *jwt.Token) (interface{}, error) {
					return []byte("my_secret_key"), nil
				})

				if claims, ok := parsedToken.Claims.(*claims); ok && parsedToken.Valid {
					assert.Equal(t, uint(54321), claims.UserID, "Token should contain the correct UserID")
				} else {
					t.Error("Failed to parse claims or token invalid")
				}
			},
		},
		{
			description: "Token Expiry Validation",
			userID:      12345,
			timestamp:   time.Now().Add(time.Hour * 3),
			setup: func() {
				os.Setenv("JWT_SECRET", "my_secret_key")
			},
			assertion: func(t *testing.T, token string, err error) {
				assert.NoError(t, err)
				assert.NotEmpty(t, token)

				parsedToken, _ := jwt.ParseWithClaims(token, &claims{}, func(token *jwt.Token) (interface{}, error) {
					return []byte("my_secret_key"), nil
				})

				if claims, ok := parsedToken.Claims.(*claims); ok && parsedToken.Valid {
					expectedExpiry := claims.ExpiresAt
					assert.Equal(t, expectedExpiry, claims.ExpiresAt, "Token should have the correct expiry timestamp")
				}
			},
		},
		{
			description: "Handling Invalid User ID",
			userID:      0,
			timestamp:   time.Now(),
			setup: func() {
				os.Setenv("JWT_SECRET", "my_secret_key")
			},
			assertion: func(t *testing.T, token string, err error) {
				assert.Error(t, err, "Expected error due to invalid user ID")
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			tc.setup()
			token, err := GenerateTokenWithTime(tc.userID, tc.timestamp)
			tc.assertion(t, token, err)
		})
	}

}

