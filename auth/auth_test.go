package auth

import (
	"os"
	"testing"
	"time"
	"github.com/dgrijalva/jwt-go"
	"errors"
	"github.com/stretchr/testify/assert"
	"reflect"
	"context"
	"github.com/grpc-ecosystem/go-grpc-middleware/auth"
	"google.golang.org/grpc/metadata"
)

type T struct {
	common
	isParallel bool
	isEnvSet   bool
	context    *testContext
}
/*
ROOST_METHOD_HASH=GenerateTokenWithTime_d0df64aa69
ROOST_METHOD_SIG_HASH=GenerateTokenWithTime_72dd09cde6


 */
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


/*
ROOST_METHOD_HASH=GenerateToken_b7f5ef3740
ROOST_METHOD_SIG_HASH=GenerateToken_d10a3e47a3


 */
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


/*
ROOST_METHOD_HASH=generateToken_2cc40e0108
ROOST_METHOD_SIG_HASH=generateToken_9de4114fe8


 */
func TestgenerateToken(t *testing.T) {

	jwtSecret = []byte(os.Getenv("JWT_SECRET"))

	scenarios := []struct {
		name           string
		userID         uint
		now            time.Time
		expectErr      bool
		secret         string
		modifySecret   bool
		expectedErr    error
		validationFunc func(string) error
	}{
		{
			name:           "Successful Token Generation",
			userID:         123,
			now:            time.Now(),
			expectErr:      false,
			secret:         os.Getenv("JWT_SECRET"),
			modifySecret:   false,
			validationFunc: nil,
		},
		{
			name:           "Handle Empty JWT Secret",
			userID:         123,
			now:            time.Now(),
			expectErr:      true,
			secret:         "",
			modifySecret:   true,
			expectedErr:    jwt.ErrSignatureInvalid,
			validationFunc: nil,
		},
		{
			name:         "Token Expiration Time Check",
			userID:       123,
			now:          time.Now(),
			expectErr:    false,
			secret:       os.Getenv("JWT_SECRET"),
			modifySecret: false,
			validationFunc: func(tokenStr string) error {
				token, err := jwt.ParseWithClaims(tokenStr, &claims{}, func(token *jwt.Token) (interface{}, error) {
					return jwtSecret, nil
				})
				if err != nil {
					return err
				}

				if c, ok := token.Claims.(*claims); ok && token.Valid {
					expExpected := c.IssuedAt + 72*60*60
					if !reflect.DeepEqual(expExpected, c.ExpiresAt) {
						return errors.New("Expiration time mismatch")
					}
				}
				return nil
			},
		},
		{
			name:           "Error Handling During Signing",
			userID:         123,
			now:            time.Now(),
			expectErr:      true,
			secret:         "corruptedSecret",
			modifySecret:   true,
			expectedErr:    jwt.ErrSignatureInvalid,
			validationFunc: nil,
		},
		{
			name:           "Handling Maximum User ID",
			userID:         uint(^uint(0)),
			now:            time.Now(),
			expectErr:      false,
			secret:         os.Getenv("JWT_SECRET"),
			modifySecret:   false,
			validationFunc: nil,
		},
		{
			name:         "Custom IssuedAt Validation",
			userID:       123,
			now:          time.Now(),
			expectErr:    false,
			secret:       os.Getenv("JWT_SECRET"),
			modifySecret: false,
			validationFunc: func(tokenStr string) error {
				token, err := jwt.ParseWithClaims(tokenStr, &claims{}, func(token *jwt.Token) (interface{}, error) {
					return jwtSecret, nil
				})
				if err != nil {
					return err
				}
				if c, ok := token.Claims.(*claims); ok && token.Valid {
					if c.IssuedAt != c.NotBefore {
						return errors.New("IssuedAt time mismatch")
					}
				}
				return nil
			},
		},
	}

	for _, sc := range scenarios {
		t.Run(sc.name, func(t *testing.T) {
			if sc.modifySecret {
				jwtSecret = []byte(sc.secret)
			}

			token, err := generateToken(sc.userID, sc.now)
			if sc.expectErr {
				if err == nil {
					t.Errorf("Expected error but got none")
					return
				}
				if !errors.Is(err, sc.expectedErr) {
					t.Errorf("Expected error %v, but got %v", sc.expectedErr, err)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
					return
				}

				if sc.validationFunc != nil {
					err := sc.validationFunc(token)
					if err != nil {
						t.Error(err.Error())
					}
				}

				if token == "" {
					t.Errorf("Expected valid token, got empty string")
				}
			}
		})
	}
}


/*
ROOST_METHOD_HASH=GetUserID_f2dd680cb2
ROOST_METHOD_SIG_HASH=GetUserID_e739312e3d


 */
func TestGetUserID(t *testing.T) {
	jwtSecret = []byte("test_secret")

	type test struct {
		name           string
		ctx            context.Context
		expectedUserID uint
		expectedErr    string
	}

	tests := []test{
		{
			name:           "Valid Token with Correct Claims",
			ctx:            createContextWithToken(createValidToken(1, time.Now().Add(time.Hour).Unix(), t)),
			expectedUserID: 1,
			expectedErr:    "",
		},
		{
			name:           "Invalid Token Format",
			ctx:            createContextWithToken("invalid.token.format"),
			expectedUserID: 0,
			expectedErr:    "invalid token: it's not even a token",
		},
		{
			name:           "Token with Expired Claims",
			ctx:            createContextWithToken(createValidToken(1, time.Now().Add(-time.Hour).Unix(), t)),
			expectedUserID: 0,
			expectedErr:    "token expired",
		},
		{
			name:           "Valid Token with Expired Time Check in Code",
			ctx:            createContextWithToken(createValidToken(1, time.Now().Add(-time.Hour).Unix(), t)),
			expectedUserID: 0,
			expectedErr:    "token expired",
		},
		{
			name:           "Missing Token in Context",
			ctx:            context.Background(),
			expectedUserID: 0,
			expectedErr:    "Request unauthenticated with Token",
		},
		{
			name:           "Token Claims Unable to Map",
			ctx:            createContextWithToken(createInvalidClaimsToken(t)),
			expectedUserID: 0,
			expectedErr:    "invalid token: cannot map token to claims",
		},
		{
			name:           "Invalid Signature",
			ctx:            createContextWithToken(createInvalidSignatureToken(t)),
			expectedUserID: 0,
			expectedErr:    "invalid signature",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			userID, err := GetUserID(tc.ctx)
			if tc.expectedErr != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.expectedErr)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tc.expectedUserID, userID)
		})
	}
}

func createContextWithToken(token string) context.Context {
	md := metadata.Pairs("authorization", "Token "+token)
	return metadata.NewIncomingContext(context.Background(), md)
}

func createInvalidClaimsToken(t *testing.T) string {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"wrong_field": "wrong_value",
	})
	ss, err := token.SignedString(jwtSecret)
	if err != nil {
		t.Fatalf("failed to sign token: %v", err)
	}
	return ss
}

func createInvalidSignatureToken(t *testing.T) string {
	claims := claims{
		UserID: 1,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Hour).Unix(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	ss, err := token.SignedString([]byte("invalid_secret"))
	if err != nil {
		t.Fatalf("failed to sign token: %v", err)
	}

	return ss
}

func createValidToken(userID uint, exp int64, t *testing.T) string {
	claims := claims{
		UserID: userID,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: exp,
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	ss, err := token.SignedString(jwtSecret)
	if err != nil {
		t.Fatalf("failed to sign token: %v", err)
	}

	return ss
}

