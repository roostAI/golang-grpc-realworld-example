package model

import (
	"testing"
	pb "github.com/raahii/golang-grpc-realworld-example/proto"
	"github.com/stretchr/testify/assert"
	"strings"
	"golang.org/x/crypto/bcrypt"
	"errors"
	validation "github.com/go-ozzo/ozzo-validation"
)









/*
ROOST_METHOD_HASH=ProtoProfile_c70e154ff1
ROOST_METHOD_SIG_HASH=ProtoProfile_def254b98c

FUNCTION_DEF=func (u *User) ProtoProfile(following bool) *pb.Profile 

 */
func TestUserProtoProfile(t *testing.T) {
	type testCase struct {
		name      string
		user      User
		following bool
		expected  pb.Profile
	}

	testCases := []testCase{
		{
			name: "Scenario 1: Convert a User Object to Profile without Following",
			user: User{
				Username: "testuser",
				Bio:      "bio here",
				Image:    "imageURL",
			},
			following: false,
			expected: pb.Profile{
				Username:  "testuser",
				Bio:       "bio here",
				Image:     "imageURL",
				Following: false,
			},
		},
		{
			name: "Scenario 2: Convert a User Object to Profile with Following",
			user: User{
				Username: "testuser",
				Bio:      "bio here",
				Image:    "imageURL",
			},
			following: true,
			expected: pb.Profile{
				Username:  "testuser",
				Bio:       "bio here",
				Image:     "imageURL",
				Following: true,
			},
		},
		{
			name: "Scenario 3: Convert User with Complex Bio and Image Data",
			user: User{
				Username: "complexUser",
				Bio:      "This is a very complex bio with various elements like symbols *&^% and longer texts",
				Image:    "http://www.example.com/image?param=1&other=complexImageURL",
			},
			following: true,
			expected: pb.Profile{
				Username:  "complexUser",
				Bio:       "This is a very complex bio with various elements like symbols *&^% and longer texts",
				Image:     "http://www.example.com/image?param=1&other=complexImageURL",
				Following: true,
			},
		},
		{
			name: "Scenario 4: Convert User with Minimal Data",
			user: User{
				Username: "a",
				Bio:      "",
				Image:    "",
			},
			following: false,
			expected: pb.Profile{
				Username:  "a",
				Bio:       "",
				Image:     "",
				Following: false,
			},
		},
		{
			name: "Scenario 5: Verify User Conversion Error Handling",
			user: User{
				Username: "invalidUser",
				Bio:      "\x00\x01\x02\x03\x04",
				Image:    "\x00\x01",
			},
			following: false,
			expected: pb.Profile{
				Username:  "invalidUser",
				Bio:       "\x00\x01\x02\x03\x04",
				Image:     "\x00\x01",
				Following: false,
			},
		},
		{
			name: "Scenario 6: Check for Inconsistent Follower States",
			user: User{
				Username: "userWithFollows",
				Bio:      "Normal Bio",
				Image:    "normalImageURL",
				Follows: []User{
					{Username: "followedUser1"},
					{Username: "followedUser2"},
				},
			},
			following: false,
			expected: pb.Profile{
				Username:  "userWithFollows",
				Bio:       "Normal Bio",
				Image:     "normalImageURL",
				Following: false,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			profile := tc.user.ProtoProfile(tc.following)

			assert.Equal(t, tc.expected.Username, profile.Username, "Unexpected Username")
			assert.Equal(t, tc.expected.Bio, profile.Bio, "Unexpected Bio")
			assert.Equal(t, tc.expected.Image, profile.Image, "Unexpected Image")
			assert.Equal(t, tc.expected.Following, profile.Following, "Unexpected Following state")

			t.Logf("Test case '%s' passed", tc.name)
		})
	}
}


/*
ROOST_METHOD_HASH=ProtoUser_440c1b101c
ROOST_METHOD_SIG_HASH=ProtoUser_fb8c4736ee

FUNCTION_DEF=func (u *User) ProtoUser(token string) *pb.User 

 */
func TestUserProtoUser(t *testing.T) {

	type testCase struct {
		description string
		user        User
		token       string
		expected    *pb.User
	}

	testCases := []testCase{
		{
			description: "Scenario 1: Conversion of User Model to Proto User with Valid Data",
			user: User{
				Username: "testuser",
				Email:    "test@example.com",
				Bio:      "This is a test bio",
				Image:    "http://example.com/image.png",
			},
			token: "validToken",
			expected: &pb.User{
				Email:    "test@example.com",
				Token:    "validToken",
				Username: "testuser",
				Bio:      "This is a test bio",
				Image:    "http://example.com/image.png",
			},
		},
		{
			description: "Scenario 2: Conversion of User Model with Empty Fields",
			user: User{
				Username: "testuser",
				Email:    "test@example.com",
				Bio:      "",
				Image:    "",
			},
			token: "validToken",
			expected: &pb.User{
				Email:    "test@example.com",
				Token:    "validToken",
				Username: "testuser",
				Bio:      "",
				Image:    "",
			},
		},
		{
			description: "Scenario 3: Conversion of User Model with Maximum Length Strings",
			user: User{
				Username: strings.Repeat("a", 255),
				Email:    strings.Repeat("b", 255) + "@example.com",
				Bio:      strings.Repeat("c", 1000),
				Image:    "http://example.com/" + strings.Repeat("d", 200),
			},
			token: "maximumLengthToken",
			expected: &pb.User{
				Email:    strings.Repeat("b", 255) + "@example.com",
				Token:    "maximumLengthToken",
				Username: strings.Repeat("a", 255),
				Bio:      strings.Repeat("c", 1000),
				Image:    "http://example.com/" + strings.Repeat("d", 200),
			},
		},
		{
			description: "Scenario 4: Handling Special Characters in User Fields",
			user: User{
				Username: "user!@#$%^&*()",
				Email:    "email!@example.com",
				Bio:      "Bio with special characters !@#$%^&*()",
				Image:    "http://example.com/image!@#$%^&*().png",
			},
			token: "specialCharToken",
			expected: &pb.User{
				Email:    "email!@example.com",
				Token:    "specialCharToken",
				Username: "user!@#$%^&*()",
				Bio:      "Bio with special characters !@#$%^&*()",
				Image:    "http://example.com/image!@#$%^&*().png",
			},
		},
		{
			description: "Scenario 5: Conversion of User Model to Proto User Without Token",
			user: User{
				Username: "testuser",
				Email:    "test@example.com",
				Bio:      "This is a test bio",
				Image:    "http://example.com/image.png",
			},
			token: "",
			expected: &pb.User{
				Email:    "test@example.com",
				Token:    "",
				Username: "testuser",
				Bio:      "This is a test bio",
				Image:    "http://example.com/image.png",
			},
		},
		{
			description: "Scenario 6: User Model with Invalid Email Format",
			user: User{
				Username: "testuser",
				Email:    "invalid-email",
				Bio:      "This is a test bio",
				Image:    "http://example.com/image.png",
			},
			token: "validToken",
			expected: &pb.User{
				Email:    "invalid-email",
				Token:    "validToken",
				Username: "testuser",
				Bio:      "This is a test bio",
				Image:    "http://example.com/image.png",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			result := tc.user.ProtoUser(tc.token)
			assert.Equal(t, tc.expected, result, "Expected and actual ProtoUser outputs should match")
		})
	}

	t.Run("Scenario 7: Stress Testing with Multiple Consecutive Calls", func(t *testing.T) {
		user := User{
			Username: "testuser",
			Email:    "test@example.com",
			Bio:      "This is a test bio",
			Image:    "http://example.com/image.png",
		}
		token := "validToken"
		expected := &pb.User{
			Email:    "test@example.com",
			Token:    "validToken",
			Username: "testuser",
			Bio:      "This is a test bio",
			Image:    "http://example.com/image.png",
		}

		for i := 0; i < 1000; i++ {
			result := user.ProtoUser(token)
			assert.Equal(t, expected, result, "Expected and actual ProtoUser outputs should match on iteration %d", i)
		}
	})
}


/*
ROOST_METHOD_HASH=CheckPassword_377b31181b
ROOST_METHOD_SIG_HASH=CheckPassword_e6e0413d83

FUNCTION_DEF=func (u *User) CheckPassword(plain string) bool 

 */
func TestUserCheckPassword(t *testing.T) {

	createHashedPassword := func(password string, cost int) string {
		hashed, err := bcrypt.GenerateFromPassword([]byte(password), cost)
		if err != nil {
			t.Fatalf("Error hashing password: %v", err)
		}
		return string(hashed)
	}

	tests := []struct {
		name           string
		userPassword   string
		inputPassword  string
		expectedResult bool
	}{
		{
			name:           "Valid Password Validation",
			userPassword:   createHashedPassword("correct_password", bcrypt.DefaultCost),
			inputPassword:  "correct_password",
			expectedResult: true,
		},
		{
			name:           "Invalid Password Validation",
			userPassword:   createHashedPassword("correct_password", bcrypt.DefaultCost),
			inputPassword:  "wrong_password",
			expectedResult: false,
		},
		{
			name:           "Empty Password Validation",
			userPassword:   createHashedPassword("correct_password", bcrypt.DefaultCost),
			inputPassword:  "",
			expectedResult: false,
		},

		{
			name:           "Non-UTF8 Encoded Password",
			userPassword:   createHashedPassword("correct_password", bcrypt.DefaultCost),
			inputPassword:  string([]byte{0xff, 0xfe, 0xfd}),
			expectedResult: false,
		},
		{
			name:           "Hashed Password Field is Empty",
			userPassword:   "",
			inputPassword:  "any_password",
			expectedResult: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := User{
				Password: tt.userPassword,
			}
			result := user.CheckPassword(tt.inputPassword)
			if result != tt.expectedResult {
				t.Errorf("failed %s: expected %v, got %v", tt.name, tt.expectedResult, result)
			} else {
				t.Logf("success %s: expected and got %v", tt.name, result)
			}
		})
	}
}


/*
ROOST_METHOD_HASH=HashPassword_ea0347143c
ROOST_METHOD_SIG_HASH=HashPassword_fc69fabec5

FUNCTION_DEF=func (u *User) HashPassword() error 

 */
func TestUserHashPassword(t *testing.T) {
	tests := []struct {
		name        string
		password    string
		expectedErr error
		verifyFunc  func(user *User) bool
	}{
		{
			name:        "Successful Password Hashing",
			password:    "ValidPassword123",
			expectedErr: nil,
			verifyFunc: func(user *User) bool {

				return user.Password != "ValidPassword123" && len(user.Password) > 0
			},
		},
		{
			name:        "Error on Empty Password",
			password:    "",
			expectedErr: errors.New("password should not be empty"),
			verifyFunc:  nil,
		},
		{
			name: "Error from bcrypt Function",

			password:    "AnyPassword",
			expectedErr: bcrypt.ErrPasswordTooLong,
			verifyFunc:  nil,
		},
		{
			name:        "Verification of Password Integrity after Hashing",
			password:    "SecurePassword",
			expectedErr: nil,
			verifyFunc: func(user *User) bool {

				err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte("SecurePassword"))
				return err == nil
			},
		},
		{
			name:        "Post-Hashing Password Format Check",
			password:    "AnotherPassword",
			expectedErr: nil,
			verifyFunc: func(user *User) bool {

				return !strings.Contains(user.Password, "AnotherPassword")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := &User{Password: tt.password}

			err := user.HashPassword()

			if (err != nil) != (tt.expectedErr != nil) || (err != nil && err.Error() != tt.expectedErr.Error()) {
				t.Errorf("expected error %v, got %v", tt.expectedErr, err)
			}

			if tt.verifyFunc != nil {
				if verified := tt.verifyFunc(user); !verified {
					t.Errorf("verification function failed for test: %s", tt.name)
				}
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
		name    string
		user    User
		wantErr bool
		errMsg  string
	}{
		{
			name: "Valid User Input",
			user: User{
				Username: "ValidUser123",
				Email:    "valid@example.com",
				Password: "securepassword",
			},
			wantErr: false,
		},
		{
			name: "Invalid Username - Contains Special Characters",
			user: User{
				Username: "user@name",
				Email:    "valid@example.com",
				Password: "securepassword",
			},
			wantErr: true,
			errMsg:  "Username: must be in a valid format",
		},
		{
			name: "Empty Username",
			user: User{
				Username: "",
				Email:    "valid@example.com",
				Password: "securepassword",
			},
			wantErr: true,
			errMsg:  "Username: cannot be blank",
		},
		{
			name: "Invalid Email Format",
			user: User{
				Username: "ValidUser123",
				Email:    "user.com",
				Password: "securepassword",
			},
			wantErr: true,
			errMsg:  "Email: must be a valid email address",
		},
		{
			name: "Empty Email",
			user: User{
				Username: "ValidUser123",
				Email:    "",
				Password: "securepassword",
			},
			wantErr: true,
			errMsg:  "Email: cannot be blank",
		},
		{
			name: "Empty Password",
			user: User{
				Username: "ValidUser123",
				Email:    "valid@example.com",
				Password: "",
			},
			wantErr: true,
			errMsg:  "Password: cannot be blank",
		},
		{
			name: "Long Username",
			user: User{
				Username: "ThisIsAVeryLongUsernameThatShouldStillBeValid123",
				Email:    "valid@example.com",
				Password: "securepassword",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.user.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Test case %q - User.Validate() error = %v, wantErr %v", tt.name, err, tt.wantErr)
			}
			if err != nil && tt.wantErr && tt.errMsg != "" {
				if verrs, ok := err.(validation.Errors); ok {
					for _, msg := range verrs {
						if msg.Error() != tt.errMsg {
							t.Errorf("Test case %q - User.Validate() expected message: %v, got: %v", tt.name, tt.errMsg, msg)
						}
					}
				} else {
					t.Errorf("Test case %q - User.Validate() expected a validation.Errors", tt.name)
				}
			}
			t.Logf("Test case %q - finished with error: %v", tt.name, err)
		})
	}
}

