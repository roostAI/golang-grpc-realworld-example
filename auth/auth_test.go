package auth

import (
	testing "testing"
	time "time"
	jwt "github.com/dgrijalva/jwt-go"
	errors "errors"
	debug "runtime/debug"
)








/*
ROOST_METHOD_HASH=generateToken_b034dbdde5
ROOST_METHOD_SIG_HASH=generateToken_9de4114fe8

FUNCTION_DEF=func generateToken(id uint, now time.Time) (string, error) 

*/
func TestGenerateToken(t *testing.T) {

	testCases := []struct {
		name          string
		id            uint
		now           time.Time
		expectedError error
	}{
		{
			name:          "Successful Token Generation",
			id:            1,
			now:           time.Now(),
			expectedError: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Logf("Panic encountered so failing test. %v\n", r)
					t.Fail()
				}
			}()

			token, err := generateToken(tc.id, tc.now)

			if err != tc.expectedError {
				t.Errorf("Expected error %v, but got %v", tc.expectedError, err)
			}

			if err == nil {

				tokenClaims, err := jwt.ParseWithClaims(token, &claims{}, func(token *jwt.Token) (interface{}, error) {
					return jwtSecret, nil
				})

				if err != nil {
					t.Errorf("Failed to parse token: %v", err)
				}

				claims, ok := tokenClaims.Claims.(*claims)
				if !ok {
					t.Errorf("Failed to parse claims")
				}

				if claims.UserID != tc.id {
					t.Errorf("Expected UserID %v, but got %v", tc.id, claims.UserID)
				}

				if claims.ExpiresAt != tc.now.Add(time.Hour*72).Unix() {
					t.Errorf("Expected ExpiresAt %v, but got %v", tc.now.Add(time.Hour*72).Unix(), claims.ExpiresAt)
				}
			}
		})
	}
}


/*
ROOST_METHOD_HASH=GenerateTokenWithTime_aa6cebe464
ROOST_METHOD_SIG_HASH=GenerateTokenWithTime_96ec3b7507

FUNCTION_DEF=func GenerateTokenWithTime(id uint, t time.Time) (string, error) // GenerateTokenWithTime generates a new token with expired date computed withspecified time


*/
func TestGenerateTokenWithTime(t *testing.T) {

	testCases := []struct {
		name          string
		id            uint
		time          time.Time
		expectError   bool
		errorExpected error
	}{
		{
			name:        "Successful Token Generation",
			id:          1,
			time:        time.Now(),
			expectError: false,
		},
		{
			name:          "Invalid User ID",
			id:            0,
			time:          time.Now(),
			expectError:   true,
			errorExpected: errors.New("Invalid User ID"),
		},
		{
			name:          "Invalid Time",
			id:            1,
			time:          time.Time{},
			expectError:   true,
			errorExpected: errors.New("Invalid Time"),
		},
		{
			name:          "Null Inputs",
			id:            0,
			time:          time.Time{},
			expectError:   true,
			errorExpected: errors.New("Null Inputs"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			defer func() {
				if r := recover(); r != nil {
					t.Logf("Panic encountered so failing test. %v\n%s", r, string(debug.Stack()))
					t.Fail()
				}
			}()

			token, err := GenerateTokenWithTime(tc.id, tc.time)

			if tc.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				} else if err.Error() != tc.errorExpected.Error() {
					t.Errorf("Expected error: %v, but got: %v", tc.errorExpected, err)
				}
			} else {
				if err != nil {
					t.Errorf("Did not expect error but got: %v", err)
				} else if token == "" {
					t.Errorf("Expected token but got none")
				}
			}
		})
	}
}

