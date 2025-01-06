package auth

import (
	"errors"
	"os"
	"reflect"
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

func TestgenerateToken(t *testing.T) {
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

