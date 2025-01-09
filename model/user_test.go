package model

import (
	"testing"
	"github.com/raahii/golang-grpc-realworld-example/proto"
	"sync"
	"golang.org/x/crypto/bcrypt"
	"strings"
	"errors"
	"github.com/go-ozzo/ozzo-validation"
)

const errUsernameInvalid = "username must be in a valid format"errEmailInvalid = "email must be in a valid format"errPasswordRequired = "password is required"errAllFieldsRequired = "all fields must be valid"
const errUsernameInvalid = "username must be in a valid format"errEmailInvalid = "email must be in a valid format"
const errUsernameInvalid = "username must be in a valid format"errEmailInvalid = "email must be in a valid format"errPasswordRequired = "password is required"
const errUsernameInvalid = "username must be in a valid format"






/*
ROOST_METHOD_HASH=ProtoProfile_c70e154ff1
ROOST_METHOD_SIG_HASH=ProtoProfile_def254b98c

FUNCTION_DEF=func (u *User) ProtoProfile(following bool) *pb.Profile 

 */
func TestUserProtoProfile(t *testing.T) {
	tests := []struct {
		name      string
		user      *User
		following bool
		expected  *pb.Profile
	}{
		{
			name: "Basic Profile Conversion",
			user: &User{
				Username: "john_doe",
				Bio:      "Software Developer",
				Image:    "http://example.com/image.jpg",
			},
			following: true,
			expected: &pb.Profile{
				Username:  "john_doe",
				Bio:       "Software Developer",
				Image:     "http://example.com/image.jpg",
				Following: true,
			},
		},
		{
			name: "Profile Conversion with Following Set to False",
			user: &User{
				Username: "jane_doe",
				Bio:      "Graphic Designer",
				Image:    "http://example.com/jane.jpg",
			},
			following: false,
			expected: &pb.Profile{
				Username:  "jane_doe",
				Bio:       "Graphic Designer",
				Image:     "http://example.com/jane.jpg",
				Following: false,
			},
		},
		{
			name: "Conversion for User with Empty Fields",
			user: &User{
				Username: "",
				Bio:      "",
				Image:    "",
			},
			following: true,
			expected: &pb.Profile{
				Username:  "",
				Bio:       "",
				Image:     "",
				Following: true,
			},
		},
		{
			name: "Conversion Handling Maximum Length Strings",
			user: &User{
				Username: "a" + string(make([]byte, 255)),
				Bio:      "b" + string(make([]byte, 255)),
				Image:    "http://example.com/" + string(make([]byte, 255)) + ".jpg",
			},
			following: false,
			expected: &pb.Profile{
				Username:  "a" + string(make([]byte, 255)),
				Bio:       "b" + string(make([]byte, 255)),
				Image:     "http://example.com/" + string(make([]byte, 255)) + ".jpg",
				Following: false,
			},
		},
		{
			name: "Default Following Behavior Verification",
			user: &User{
				Username: "random_user",
				Bio:      "A random bio",
				Image:    "http://example.com/random.jpg",
			},
			following: false,
			expected: &pb.Profile{
				Username:  "random_user",
				Bio:       "A random bio",
				Image:     "http://example.com/random.jpg",
				Following: false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			profile := tt.user.ProtoProfile(tt.following)

			if profile.Username != tt.expected.Username || profile.Bio != tt.expected.Bio || profile.Image != tt.expected.Image || profile.Following != tt.expected.Following {
				t.Errorf("ProtoProfile() = %v, want %v", profile, tt.expected)
			} else {
				t.Logf("Test %s passed successfully", tt.name)
			}
		})
	}
}


/*
ROOST_METHOD_HASH=ProtoUser_440c1b101c
ROOST_METHOD_SIG_HASH=ProtoUser_fb8c4736ee

FUNCTION_DEF=func (u *User) ProtoUser(token string) *pb.User 

 */
func TestUserProtoUser(t *testing.T) {

	tests := []struct {
		name   string
		user   *User
		token  string
		expect *pb.User
	}{
		{
			name: "Valid User Data with Valid Token",
			user: &User{
				Username: "testuser",
				Email:    "testuser@example.com",
				Bio:      "This is a bio",
				Image:    "link-to-image",
			},
			token: "validToken123",
			expect: &pb.User{
				Email:    "testuser@example.com",
				Token:    "validToken123",
				Username: "testuser",
				Bio:      "This is a bio",
				Image:    "link-to-image",
			},
		},
		{
			name: "User with Empty Fields and Valid Token",
			user: &User{
				Username: "testuser2",
				Email:    "testuser2@example.com",
				Bio:      "",
				Image:    "",
			},
			token: "validToken456",
			expect: &pb.User{
				Email:    "testuser2@example.com",
				Token:    "validToken456",
				Username: "testuser2",
				Bio:      "",
				Image:    "",
			},
		},
		{
			name: "Special Characters in User Fields",
			user: &User{
				Username: "user*#&$",
				Email:    "special@chars.com",
				Bio:      "Bio with special chars !@#",
				Image:    "image-link",
			},
			token: "specialCharToken",
			expect: &pb.User{
				Email:    "special@chars.com",
				Token:    "specialCharToken",
				Username: "user*#&$",
				Bio:      "Bio with special chars !@#",
				Image:    "image-link",
			},
		},
		{
			name:   "Handling Nil User Reference",
			user:   nil,
			token:  "nilToken",
			expect: nil,
		},
		{
			name: "Token Edge Cases with Empty and Long String",
			user: &User{
				Username: "edgeUser",
				Email:    "edge@example.com",
				Bio:      "Edge case bio",
				Image:    "edge-image",
			},
			token: "",
			expect: &pb.User{
				Email:    "edge@example.com",
				Token:    "",
				Username: "edgeUser",
				Bio:      "Edge case bio",
				Image:    "edge-image",
			},
		},
		{
			name: "Token Edge Cases with Long String",
			user: &User{
				Username: "edgeUser",
				Email:    "edge@example.com",
				Bio:      "Edge case bio",
				Image:    "edge-image",
			},
			token: "longTokenStringLongTokenStringLongTokenStringLongToken",
			expect: &pb.User{
				Email:    "edge@example.com",
				Token:    "longTokenStringLongTokenStringLongTokenStringLongToken",
				Username: "edgeUser",
				Bio:      "Edge case bio",
				Image:    "edge-image",
			},
		},
		{
			name: "ProtoUser Call with Invalid User Data",
			user: &User{
				Username: "invalidEmailUser",
				Email:    "invalid-email-format",
				Bio:      "Invalid email",
				Image:    "invalid-image",
			},
			token: "invalidDataToken",
			expect: &pb.User{
				Email:    "invalid-email-format",
				Token:    "invalidDataToken",
				Username: "invalidEmailUser",
				Bio:      "Invalid email",
				Image:    "invalid-image",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.user == nil {
				if result := tt.user.ProtoUser(tt.token); result != nil {
					t.Errorf("expected nil, got %v", result)
				} else {
					t.Log("handling of nil user passed")
				}
			} else {
				result := tt.user.ProtoUser(tt.token)
				if result.Email != tt.expect.Email || result.Token != tt.expect.Token ||
					result.Username != tt.expect.Username || result.Bio != tt.expect.Bio ||
					result.Image != tt.expect.Image {
					t.Errorf("expected %+v, got %+v", tt.expect, result)
				} else {
					t.Logf("test %s passed", tt.name)
				}
			}
		})
	}

	t.Run("Concurrent ProtoUser Calls", func(t *testing.T) {
		u := &User{
			Username: "concurrentUser",
			Email:    "concurrent@example.com",
			Bio:      "Concurrent bio",
			Image:    "concurrent-image",
		}
		token := "concurrentToken"
		var wg sync.WaitGroup
		wg.Add(10)

		for i := 0; i < 10; i++ {
			go func() {
				defer wg.Done()
				result := u.ProtoUser(token)
				if result.Email != u.Email || result.Username != u.Username || result.Token != token {
					t.Errorf("Concurrency test failed: expected email %s, username %s, token %s",
						u.Email, u.Username, token)
				} else {
					t.Log("concurrency test success")
				}
			}()
		}
		wg.Wait()
	})
}


/*
ROOST_METHOD_HASH=CheckPassword_377b31181b
ROOST_METHOD_SIG_HASH=CheckPassword_e6e0413d83

FUNCTION_DEF=func (u *User) CheckPassword(plain string) bool 

 */
func TestUserCheckPassword(t *testing.T) {
	tests := []struct {
		name            string
		initialPassword string
		testPassword    string
		expectedResult  bool
		scenario        string
	}{
		{
			name:            "Valid Password Check",
			initialPassword: "valid_password",
			testPassword:    "valid_password",
			expectedResult:  true,
			scenario:        "When the plain text password matches the hashed password",
		},
		{
			name:            "Invalid Password Check",
			initialPassword: "valid_password",
			testPassword:    "invalid_password",
			expectedResult:  false,
			scenario:        "When the provided password does not match the hashed password",
		},
		{
			name:            "Empty Password Input Check",
			initialPassword: "valid_password",
			testPassword:    "",
			expectedResult:  false,
			scenario:        "When an empty string password is given for comparison",
		},
		{
			name:            "Malformed Password Check",
			initialPassword: "valid_password",
			testPassword:    "\x99\x8F\x01",
			expectedResult:  false,
			scenario:        "When malformed password data is present in input",
		},
		{
			name:            "Long Password Handling",
			initialPassword: "short_password",
			testPassword:    strings.Repeat("a", 1000),
			expectedResult:  false,
			scenario:        "When handling extraordinarily long password inputs",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash, err := bcrypt.GenerateFromPassword([]byte(tt.initialPassword), bcrypt.DefaultCost)
			if err != nil {
				t.Fatalf("Failed to hash password: %v", err)
			}
			user := User{Password: string(hash)}

			result := user.CheckPassword(tt.testPassword)
			if result != tt.expectedResult {
				t.Errorf("Failed at scenario: %s, expected: %v, got: %v", tt.scenario, tt.expectedResult, result)
			} else {
				t.Logf("Passed scenario: %s", tt.scenario)
			}
		})
	}
}


/*
ROOST_METHOD_HASH=Validate_532ff0c623
ROOST_METHOD_SIG_HASH=Validate_663e136f97

FUNCTION_DEF=func (u User) Validate() error 

 */
func TestUserValidate(t *testing.T) {
	tests := []struct {
		name        string
		user        User
		expectError bool
		errorMsg    string
	}{
		{
			name: "Valid User Data",
			user: User{
				Username: "ValidUser123",
				Email:    "user@example.com",
				Password: "password",
			},
			expectError: false,
		},
		{
			name: "Invalid Username",
			user: User{
				Username: "Invalid!User",
				Email:    "user@example.com",
				Password: "password",
			},
			expectError: true,
			errorMsg:    errUsernameInvalid,
		},
		{
			name: "Missing Email",
			user: User{
				Username: "ValidUser123",
				Email:    "",
				Password: "password",
			},
			expectError: true,
			errorMsg:    errEmailInvalid,
		},
		{
			name: "Invalid Email Format",
			user: User{
				Username: "ValidUser123",
				Email:    "incorrectemailformat",
				Password: "password",
			},
			expectError: true,
			errorMsg:    errEmailInvalid,
		},
		{
			name: "Missing Password",
			user: User{
				Username: "ValidUser123",
				Email:    "user@example.com",
				Password: "",
			},
			expectError: true,
			errorMsg:    errPasswordRequired,
		},
		{
			name: "Multiple Validation Errors",
			user: User{
				Username: "Invalid!User",
				Email:    "incorrectemailformat",
				Password: "",
			},
			expectError: true,
			errorMsg:    errAllFieldsRequired,
		},
		{
			name: "Boundary Username Length",
			user: User{
				Username: "A",
				Email:    "user@example.com",
				Password: "password",
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.user.Validate()

			if tt.expectError {
				if err == nil {
					t.Errorf("%s: expected error but got none", tt.name)
				} else {
					var validationErrors validation.Errors
					if errors.As(err, &validationErrors) {
						errorCount := len(validationErrors)
						if tt.name == "Multiple Validation Errors" && errorCount < 2 {
							t.Errorf("%s: expected multiple errors, but got %d", tt.name, errorCount)
						} else if validationErrors[tt.name] != nil && validationErrors[tt.name].Error() != tt.errorMsg {
							t.Errorf("%s: expected error message %q, but got %q", tt.name, tt.errorMsg, validationErrors[tt.name].Error())
						}
					}
				}
			} else if err != nil {
				t.Errorf("%s: unexpected error: %v", tt.name, err)
			}
		})
	}
}

