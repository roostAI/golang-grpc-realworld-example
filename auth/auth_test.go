package github

import (
	"os"
	"testing"
	"github.com/dgrijalva/jwt-go"
	"github.com/stretchr/testify/assert"
	"time"
	"errors"
)









/*
ROOST_METHOD_HASH=GenerateToken_b7f5ef3740
ROOST_METHOD_SIG_HASH=GenerateToken_d10a3e47a3

FUNCTION_DEF=func GenerateToken(id uint) (string, error) 

 */
func TestGenerateToken(t *testing.T) {
	tests := []struct {
		name          string
		userID        uint
		jwtSecret     string
		expectedError bool
		validateFunc  func(t *testing.T, token string, err error)
	}{
		{
			name:          "Successful Token Generation for Valid User ID",
			userID:        1234,
			jwtSecret:     "mySecretKey",
			expectedError: false,
			validateFunc: func(t *testing.T, token string, err error) {
				assert.NoError(t, err)
				assert.NotEmpty(t, token)
				t.Log("Successful Token Generation for Valid User ID passed.")
			},
		},
		{
			name:          "Token Generation Fails due to Missing JWT Secret",
			userID:        1234,
			jwtSecret:     "",
			expectedError: true,
			validateFunc: func(t *testing.T, token string, err error) {
				assert.Error(t, err)
				t.Log("Token Generation Fails due to Missing JWT Secret passed.")
			},
		},
		{
			name:          "Token Expiry Verification",
			userID:        1234,
			jwtSecret:     "mySecretKey",
			expectedError: false,
			validateFunc: func(t *testing.T, token string, err error) {
				assert.NoError(t, err)
				parsedToken, _ := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
					return []byte("mySecretKey"), nil
				})
				claims, _ := parsedToken.Claims.(jwt.MapClaims)
				assert.NotZero(t, claims["exp"])
				t.Log("Token Expiry Verification passed.")
			},
		},
		{
			name:          "Generation with Invalid User ID",
			userID:        0,
			jwtSecret:     "mySecretKey",
			expectedError: false,
			validateFunc: func(t *testing.T, token string, err error) {
				assert.NoError(t, err)
				assert.NotEmpty(t, token)
				t.Log("Generation with Invalid User ID passed.")
			},
		},
		{
			name:          "Verify Claim Data Integrity",
			userID:        5678,
			jwtSecret:     "mySecretKey",
			expectedError: false,
			validateFunc: func(t *testing.T, token string, err error) {
				assert.NoError(t, err)
				parsedToken, _ := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
					return []byte("mySecretKey"), nil
				})
				claims, _ := parsedToken.Claims.(jwt.MapClaims)
				assert.Equal(t, float64(5678), claims["user_id"])
				t.Log("Verify Claim Data Integrity passed.")
			},
		},
		{
			name:          "Retry on JWT Secret Invalid Error",
			userID:        1234,
			jwtSecret:     "invalidSecret",
			expectedError: true,
			validateFunc: func(t *testing.T, token string, err error) {
				assert.Error(t, err)
				t.Log("Retry on JWT Secret Invalid Error passed.")
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			originalSecret := os.Getenv("JWT_SECRET")
			defer os.Setenv("JWT_SECRET", originalSecret)

			os.Setenv("JWT_SECRET", tc.jwtSecret)
			token, err := GenerateToken(tc.userID)

			tc.validateFunc(t, token, err)
		})
	}
}


/*
ROOST_METHOD_HASH=GenerateTokenWithTime_d0df64aa69
ROOST_METHOD_SIG_HASH=GenerateTokenWithTime_72dd09cde6

FUNCTION_DEF=func GenerateTokenWithTime(id uint, t time.Time) (string, error) 

 */
func TestGenerateTokenWithTime(t *testing.T) {
	type test struct {
		name      string
		userID    uint
		inputTime time.Time
		expectErr bool
		envSecret string
	}

	tests := []test{
		{
			name:      "Valid Token Generation with Current Time",
			userID:    123,
			inputTime: time.Now(),
			expectErr: false,
			envSecret: "current_secret",
		},
		{
			name:      "Token Generation with Past Time",
			userID:    123,
			inputTime: time.Now().AddDate(-1, 0, 0),
			expectErr: false,
			envSecret: "current_secret",
		},
		{
			name:      "Token Generation with Future Time",
			userID:    123,
			inputTime: time.Now().AddDate(1, 0, 0),
			expectErr: false,
			envSecret: "current_secret",
		},
		{
			name:      "Token Generation with Invalid User ID",
			userID:    0,
			inputTime: time.Now(),
			expectErr: true,
			envSecret: "current_secret",
		},
		{
			name:      "Environment Variable Not Set",
			userID:    123,
			inputTime: time.Now(),
			expectErr: true,
			envSecret: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			os.Setenv("JWT_SECRET", tt.envSecret)
			defer os.Unsetenv("JWT_SECRET")

			token, err := GenerateTokenWithTime(tt.userID, tt.inputTime)

			if tt.expectErr {
				if err == nil {
					t.Errorf("Expected error but got none")
				} else {
					t.Logf("Received expected error: %v", err)
				}
				return
			} else if err != nil {
				t.Errorf("Did not expect an error but got: %v", err)
				return
			}

			if token == "" {
				t.Error("Generated token should not be empty")
			} else {
				t.Log("Successfully generated token")
			}

			if tt.envSecret != "" {
				parsedToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
					return []byte(tt.envSecret), nil
				})
				if err != nil || !parsedToken.Valid {
					t.Errorf("Failed to validate token: %v", err)
				} else {
					t.Log("Token validation successful")
				}
			} else {
				t.Log("JWT secret not set, skipping token validation")
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
	assert := assert.New(t)

	type test struct {
		name        string
		id          uint
		now         time.Time
		expectedErr bool
		checkToken  func(token string) error
	}

	tests := []test{
		{
			name:        "Generate Valid Token",
			id:          1,
			now:         time.Now(),
			expectedErr: false,
			checkToken: func(token string) error {
				parsedToken, err := jwt.ParseWithClaims(token, &claims{}, func(token *jwt.Token) (interface{}, error) {
					return jwtSecret, nil
				})
				if err != nil {
					return err
				}
				assert.NotEmpty(parsedToken)
				assert.True(parsedToken.Valid)
				return nil
			},
		},
		{
			name:        "Expired Token Generation",
			id:          2,
			now:         time.Date(2023, time.January, 1, 0, 0, 0, 0, time.UTC),
			expectedErr: false,
			checkToken: func(token string) error {
				parsedToken, _ := jwt.ParseWithClaims(token, &claims{}, func(token *jwt.Token) (interface{}, error) {
					return jwtSecret, nil
				})
				claims, ok := parsedToken.Claims.(*claims)
				if !ok {
					return errors.New("unable to parse claims")
				}
				exp := claims.ExpiresAt
				expectedExp := time.Date(2023, time.January, 4, 0, 0, 0, 0, time.UTC).Unix()
				assert.Equal(expectedExp, exp)
				return nil
			},
		},
		{
			name:        "Token Generation with Invalid Secret",
			id:          3,
			now:         time.Now(),
			expectedErr: true,
			checkToken:  nil,
		},
		{
			name:        "Token Generation with Future Date",
			id:          4,
			now:         time.Now().Add(time.Hour * 24 * 7),
			expectedErr: false,
			checkToken: func(token string) error {
				parsedToken, _ := jwt.ParseWithClaims(token, &claims{}, func(token *jwt.Token) (interface{}, error) {
					return jwtSecret, nil
				})
				claims, ok := parsedToken.Claims.(*claims)
				if !ok {
					return errors.New("unable to parse claims")
				}
				iat := claims.IssuedAt
				assert.Equal(iat, time.Now().Add(time.Hour*24*7).Unix())
				return nil
			},
		},
		{
			name:        "Token Generation for Multiple Users",
			id:          5,
			now:         time.Now(),
			expectedErr: false,
			checkToken: func(firstToken string) error {
				secondID := uint(6)
				secondToken, err := generateToken(secondID, time.Now())
				if err != nil {
					return err
				}
				assert.NotEqual(firstToken, secondToken)
				return nil
			},
		},
	}

	originalSecret := os.Getenv("JWT_SECRET")
	defer os.Setenv("JWT_SECRET", originalSecret)

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {

			if tc.name == "Token Generation with Invalid Secret" {
				os.Setenv("JWT_SECRET", "")
			} else {
				os.Setenv("JWT_SECRET", "validsecret")
			}

			token, err := generateToken(tc.id, tc.now)
			if tc.expectedErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
				if tc.checkToken != nil {
					err := tc.checkToken(token)
					assert.NoError(err)
				}
			}

			t.Log("Test:", tc.name)
			if err != nil {
				t.Log("Failure Reason:", err)
			} else {
				t.Log("Token Generated:", token)
			}
		})
	}
}

