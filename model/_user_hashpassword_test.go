package model

import (
	"testing"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
	"regexp"
	"errors"
)





type T struct {
	common
	isParallel bool
	isEnvSet   bool
	context    *testContext // For running tests and subtests.
}

func TestUserHashPassword(t *testing.T) {
	type testCase struct {
		name        string
		user        User
		mockBcrypt  func()
		wantError   bool
		errorString string
		changed     bool
	}

	tests := []testCase{
		{
			name: "Scenario 1: Hashing a Valid Password",
			user: User{Password: "ValidPassword123!"},
			mockBcrypt: func() {

			},
			wantError: false,
			changed:   true,
		},
		{
			name: "Scenario 2: Handling an Empty Password",
			user: User{Password: ""},
			mockBcrypt: func() {

			},
			wantError:   true,
			errorString: "password should not be empty",
			changed:     false,
		},
		{
			name: "Scenario 3: Simulating Hash Generation Failure",
			user: User{Password: "ValidPassword123!"},
			mockBcrypt: func() {

				bcryptGenerateFromPassword = func(password []byte, cost int) ([]byte, error) {
					return nil, errors.New("mocked error")
				}
			},
			wantError:   true,
			errorString: "mocked error",
			changed:     false,
		},
		{
			name: "Scenario 4: Preserving Non-Password Fields",
			user: User{Username: "TestUser", Email: "test@example.com", Password: "AnotherValidPassword", Bio: "A User Bio", Image: "ImageURL"},
			mockBcrypt: func() {

			},
			wantError: false,
			changed:   true,
		},
		{
			name: "Scenario 5: Verification of Hashed Password Format",
			user: User{Password: "SamplePassword"},
			mockBcrypt: func() {

			},
			wantError: false,
			changed:   true,
		},
	}

	originalBcryptGenerateFromPassword := bcrypt.GenerateFromPassword
	defer func() { bcrypt.GenerateFromPassword = originalBcryptGenerateFromPassword }()

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			initialPassword := tc.user.Password
			tc.mockBcrypt()

			err := tc.user.HashPassword()

			if tc.wantError {
				assert.Error(t, err)
				if err != nil {
					assert.EqualError(t, err, tc.errorString)
				}
			} else {
				assert.NoError(t, err)
				if tc.changed {
					assert.NotEqual(t, initialPassword, tc.user.Password, "Password should be hashed and therefore changed")
					matched, _ := regexp.MatchString(`^\$2[a-z]\$[\d]+\$[./A-Za-z0-9]{53}$`, tc.user.Password)
					assert.True(t, matched, "Password should be in bcrypt format")
				}
			}

			if tc.name == "Scenario 4: Preserving Non-Password Fields" {
				assert.Equal(t, "TestUser", tc.user.Username)
				assert.Equal(t, "test@example.com", tc.user.Email)
				assert.Equal(t, "A User Bio", tc.user.Bio)
				assert.Equal(t, "ImageURL", tc.user.Image)
			}
		})
	}
}
