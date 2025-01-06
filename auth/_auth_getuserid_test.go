package auth

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"
	"github.com/dgrijalva/jwt-go"
	grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
	"github.com/stretchr/testify/assert"
)







type T struct {
	common
	isParallel bool
	isEnvSet   bool
	context    *testContext // For running tests and subtests.
}

func TestGetUserID(t *testing.T) {
	type testCase struct {
		name           string
		tokenString    string
		expectedUserID uint
		expectError    bool
		errorMessage   string
	}

	var testCases = []testCase{
		{
			name:           "Valid Token with Correct Claims",
			tokenString:    generateValidToken(t, time.Now().Add(time.Minute), 1234),
			expectedUserID: 1234,
			expectError:    false,
		},
		{
			name:         "Invalid Token Format",
			tokenString:  "invalid_token",
			expectError:  true,
			errorMessage: "invalid token: it's not even a token",
		},
		{
			name:         "Token with Expired Claims",
			tokenString:  generateValidToken(t, time.Now().Add(-time.Minute), 1234),
			expectError:  true,
			errorMessage: "token expired",
		},
		{
			name:         "Missing Token in Context",
			tokenString:  "",
			expectError:  true,
			errorMessage: "Request unauthenticated with Token",
		},
		{
			name:         "Token Claims Unable to Map",
			tokenString:  generateTokenWithIncorrectClaims(t),
			expectError:  true,
			errorMessage: "invalid token: cannot map token to claims",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := createContextWithToken(tc.tokenString)
			userID, err := GetUserID(ctx)

			if tc.expectError {
				assert.Error(t, err)
				assert.Equal(t, tc.errorMessage, err.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedUserID, userID)
			}

			if err != nil {
				t.Log(fmt.Sprintf("Expected error: %s", err.Error()))
			} else {
				t.Log(fmt.Sprintf("Successfully retrieved User ID: %d", userID))
			}
		})
	}
}
func createContextWithToken(token string) context.Context {
	if token != "" {
		return grpc_auth.AddAuthToIncomingContext(context.Background(), "authorization", "Token "+token)
	}
	return context.Background()
}
func generateTokenWithIncorrectClaims(t *testing.T) string {
	type invalidClaims struct {
		InvalidID string `json:"invalid_id"`
		jwt.StandardClaims
	}
	claims := &invalidClaims{
		InvalidID: "invalid",
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Minute).Unix(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		t.Fatalf("Unable to sign invalid token: %v", err)
	}
	return tokenString
}
func generateValidToken(t *testing.T, expiry time.Time, userID uint) string {
	claims := &claims{
		UserID: userID,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expiry.Unix(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		t.Fatalf("Unable to sign valid token: %v", err)
	}
	return tokenString
}
