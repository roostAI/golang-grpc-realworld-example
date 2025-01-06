package model

import (
	"testing"
	"golang.org/x/crypto/bcrypt"
	"log"
)




type T struct {
	common
	isParallel bool
	isEnvSet   bool
	context    *testContext // For running tests and subtests.
}
func TestUserCheckPassword(t *testing.T) {
	type testCase struct {
		name         string
		userPassword string
		plainTextPwd string
		expected     bool
	}

	hashPassword := func(password string) string {
		hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			log.Fatalf("Error hashing password: %v", err)
		}
		return string(hashed)
	}

	testCases := []testCase{
		{
			name:         "Scenario 1: Correct Password",
			userPassword: hashPassword("correctpass"),
			plainTextPwd: "correctpass",
			expected:     true,
		},
		{
			name:         "Scenario 2: Incorrect Password",
			userPassword: hashPassword("correctpass"),
			plainTextPwd: "wrongpass",
			expected:     false,
		},
		{
			name:         "Scenario 3: Empty Password",
			userPassword: hashPassword("correctpass"),
			plainTextPwd: "",
			expected:     false,
		},
		{
			name:         "Scenario 4: Empty Hashed Password in User",
			userPassword: "",
			plainTextPwd: "somepass",
			expected:     false,
		},
		{
			name:         "Scenario 5: Maximum Length Password",
			userPassword: hashPassword("a" + string(make([]byte, 71))),
			plainTextPwd: "a" + string(make([]byte, 71)),
			expected:     true,
		},
		{
			name:         "Scenario 6: Unicode Characters in Password",
			userPassword: hashPassword("pāsswörd😊"),
			plainTextPwd: "pāsswörd😊",
			expected:     true,
		},
		{
			name:         "Scenario 7: Special Characters in Password",
			userPassword: hashPassword("@Sp3c!@lCh@r$"),
			plainTextPwd: "@Sp3c!@lCh@r$",
			expected:     true,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			user := &User{Password: tc.userPassword}
			t.Logf("Running scenario: %s", tc.name)

			actual := user.CheckPassword(tc.plainTextPwd)

			if actual != tc.expected {
				t.Errorf("Test %s failed: Expected %v, got %v", tc.name, tc.expected, actual)
			} else {
				t.Logf("Test %s passed", tc.name)
			}
		})
	}
}
