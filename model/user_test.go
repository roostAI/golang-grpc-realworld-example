package model

import (
	"testing"
	"github.com/raahii/golang-grpc-realworld-example/proto"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
	"regexp"
	"errors"
	"time"
	"github.com/go-ozzo/ozzo-validation"
	"github.com/go-ozzo/ozzo-validation/is"
	"gorm.io/gorm"
)

type T struct {
	common
	isParallel bool
	isEnvSet   bool
	context    *testContext
}
/*
ROOST_METHOD_HASH=ProtoProfile_c70e154ff1
ROOST_METHOD_SIG_HASH=ProtoProfile_def254b98c


 */
func TestUserProtoProfile(t *testing.T) {
	tests := []struct {
		name            string
		user            User
		following       bool
		expectedProfile *pb.Profile
	}{
		{
			name: "Convert User to Profile with Following Status True",
			user: User{
				Username: "testuser",
				Bio:      "test bio",
				Image:    "testimage.png",
			},
			following: true,
			expectedProfile: &pb.Profile{
				Username:  "testuser",
				Bio:       "test bio",
				Image:     "testimage.png",
				Following: true,
			},
		},
		{
			name: "Convert User to Profile with Following Status False",
			user: User{
				Username: "anotheruser",
				Bio:      "another bio",
				Image:    "anotherimage.png",
			},
			following: false,
			expectedProfile: &pb.Profile{
				Username:  "anotheruser",
				Bio:       "another bio",
				Image:     "anotherimage.png",
				Following: false,
			},
		},
		{
			name: "Conversion of User with Empty Values",
			user: User{
				Username: "",
				Bio:      "",
				Image:    "",
			},
			following: false,
			expectedProfile: &pb.Profile{
				Username:  "",
				Bio:       "",
				Image:     "",
				Following: false,
			},
		},
		{
			name: "Conversion of a User with Special Characters",
			user: User{
				Username: "special!@#$%",
				Bio:      "bio*&^%",
				Image:    "special.png",
			},
			following: true,
			expectedProfile: &pb.Profile{
				Username:  "special!@#$%",
				Bio:       "bio*&^%",
				Image:     "special.png",
				Following: true,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			actualProfile := tc.user.ProtoProfile(tc.following)

			if actualProfile.Username != tc.expectedProfile.Username ||
				actualProfile.Bio != tc.expectedProfile.Bio ||
				actualProfile.Image != tc.expectedProfile.Image ||
				actualProfile.Following != tc.expectedProfile.Following {
				t.Logf("Expected: %+v, Got: %+v", tc.expectedProfile, actualProfile)
				t.Error("Profile conversion failed.")
			} else {
				t.Logf("Test '%s' passed.\n", tc.name)
			}
		})
	}
}


/*
ROOST_METHOD_HASH=ProtoUser_440c1b101c
ROOST_METHOD_SIG_HASH=ProtoUser_fb8c4736ee


 */
func TestUserProtoUser(t *testing.T) {
	tests := []struct {
		name      string
		modelUser User
		token     string
		expected  *pb.User
	}{
		{
			name: "Successful Conversion of User to Proto User with Valid Token",
			modelUser: User{
				Username: "testuser",
				Email:    "user@example.com",
				Bio:      "A bio",
				Image:    "image.png",
			},
			token: "validToken123",
			expected: &pb.User{
				Email:    "user@example.com",
				Token:    "validToken123",
				Username: "testuser",
				Bio:      "A bio",
				Image:    "image.png",
			},
		},
		{
			name: "Conversion of User with Empty Fields to Proto User",
			modelUser: User{
				Username: "testuser",
				Email:    "user@example.com",
				Bio:      "",
				Image:    "",
			},
			token: "emptyFieldsToken",
			expected: &pb.User{
				Email:    "user@example.com",
				Token:    "emptyFieldsToken",
				Username: "testuser",
				Bio:      "",
				Image:    "",
			},
		},
		{
			name: "Conversion of User with Special Characters in Fields",
			modelUser: User{
				Username: "tëstuser@!",
				Email:    "usér+123@example.com",
				Bio:      "Bïø: 🎉🚀",
				Image:    "images/✨.jpg",
			},
			token: "specialCharsToken",
			expected: &pb.User{
				Email:    "usér+123@example.com",
				Token:    "specialCharsToken",
				Username: "tëstuser@!",
				Bio:      "Bïø: 🎉🚀",
				Image:    "images/✨.jpg",
			},
		},
		{
			name: "Consistency Across Multiple ProtoUser Conversions",
			modelUser: User{
				Username: "testuser",
				Email:    "user@example.com",
				Bio:      "A consistent bio",
				Image:    "consistent.png",
			},
			token: "consistentToken",
			expected: &pb.User{
				Email:    "user@example.com",
				Token:    "consistentToken",
				Username: "testuser",
				Bio:      "A consistent bio",
				Image:    "consistent.png",
			},
		},
		{
			name: "Conversion of Proto User with Null Token Parameter",
			modelUser: User{
				Username: "testuser",
				Email:    "user@example.com",
				Bio:      "A bio",
				Image:    "image.png",
			},
			token: "",
			expected: &pb.User{
				Email:    "user@example.com",
				Token:    "",
				Username: "testuser",
				Bio:      "A bio",
				Image:    "image.png",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := tt.modelUser.ProtoUser(tt.token)
			assert.Equal(t, tt.expected, actual)
			for fieldName, expectedValue := range map[string]string{
				"Email":    tt.expected.Email,
				"Username": tt.expected.Username,
				"Bio":      tt.expected.Bio,
				"Image":    tt.expected.Image,
				"Token":    tt.expected.Token,
			} {
				t.Logf("Verifying %q field, expected: %q, actual: %q", fieldName, expectedValue, expectedValue)
			}
		})
	}
}


/*
ROOST_METHOD_HASH=CheckPassword_377b31181b
ROOST_METHOD_SIG_HASH=CheckPassword_e6e0413d83


 */
func TestUserCheckPassword(t *testing.T) {

	hashPassword := func(password string) string {
		hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			t.Fatalf("Failed to hash password: %v", err)
		}
		return string(hashed)
	}

	type testCase struct {
		name     string
		user     User
		password string
		expected bool
	}

	testCases := []testCase{
		{
			name:     "Correct Password",
			user:     User{Password: hashPassword("securepassword123")},
			password: "securepassword123",
			expected: true,
		},
		{
			name:     "Incorrect Password",
			user:     User{Password: hashPassword("securepassword123")},
			password: "wrongpassword",
			expected: false,
		},
		{
			name:     "Empty Password",
			user:     User{Password: hashPassword("anotherpassword")},
			password: "",
			expected: false,
		},
		{
			name:     "Empty Hashed Password in User",
			user:     User{Password: ""},
			password: "somepassword",
			expected: false,
		},
		{
			name:     "Maximum Length Password",
			user:     User{Password: hashPassword(string(make([]byte, 72)))},
			password: string(make([]byte, 72)),
			expected: true,
		},
		{
			name:     "Unicode Characters in Password",
			user:     User{Password: hashPassword("pāsswørd✨🌟")},
			password: "pāsswørd✨🌟",
			expected: true,
		},
		{
			name:     "Special Characters in Password",
			user:     User{Password: hashPassword("!@#$%^&*()_+{}:\"<>?")},
			password: "!@#$%^&*()_+{}:\"<>?",
			expected: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {

			result := testCase.user.CheckPassword(testCase.password)

			if result != testCase.expected {
				t.Errorf("TestCase %s - Expected: %v, Got: %v", testCase.name, testCase.expected, result)
			} else {
				t.Logf("TestCase %s - Success: %v", testCase.name, result)
			}
		})
	}
}


/*
ROOST_METHOD_HASH=HashPassword_ea0347143c
ROOST_METHOD_SIG_HASH=HashPassword_fc69fabec5


 */
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


/*
ROOST_METHOD_HASH=Validate_532ff0c623
ROOST_METHOD_SIG_HASH=Validate_663e136f97


 */
func TestUserValidate(t *testing.T) {
	tests := []struct {
		name    string
		user    User
		wantErr bool
		errMsg  string
	}{
		{
			name: "Validate Correct User",
			user: User{
				Username: "validUser123",
				Email:    "test@example.com",
				Password: "validPassword",
			},
			wantErr: false,
			errMsg:  "",
		},
		{
			name: "Validate Missing Username",
			user: User{
				Username: "",
				Email:    "test@example.com",
				Password: "validPassword",
			},
			wantErr: true,
			errMsg:  "Username: cannot be blank.",
		},
		{
			name: "Validate Missing Email",
			user: User{
				Username: "validUser123",
				Email:    "",
				Password: "validPassword",
			},
			wantErr: true,
			errMsg:  "Email: cannot be blank.",
		},
		{
			name: "Validate Incorrect Email Format",
			user: User{
				Username: "validUser123",
				Email:    "invalid-email",
				Password: "validPassword",
			},
			wantErr: true,
			errMsg:  "Email: must be a valid email address.",
		},
		{
			name: "Validate Missing Password",
			user: User{
				Username: "validUser123",
				Email:    "test@example.com",
				Password: "",
			},
			wantErr: true,
			errMsg:  "Password: cannot be blank.",
		},
		{
			name: "Validate Invalid Username Characters",
			user: User{
				Username: "invalid!@#User",
				Email:    "test@example.com",
				Password: "validPassword",
			},
			wantErr: true,
			errMsg:  "Username: must be in a valid format.",
		},
		{
			name: "Validate Minimal Acceptable Input",
			user: User{
				Username: "u",
				Email:    "u@e.co",
				Password: "1",
			},
			wantErr: false,
			errMsg:  "",
		},
		{
			name: "Combined Field Validation Failure",
			user: User{
				Username: "",
				Email:    "invalid-email",
				Password: "validPassword",
			},
			wantErr: true,
			errMsg:  "Username: cannot be blank; Email: must be a valid email address.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.user.Validate()
			if tt.wantErr {
				assert.Error(t, err)
				if ve, ok := err.(validation.Errors); ok {
					for field, msg := range ve {
						assert.Contains(t, msg.Error(), field)
						assert.Contains(t, msg.Error(), tt.errMsg)
					}
				}
				t.Logf("Expected error: \"%v\", got error: \"%v\"", tt.errMsg, err)
			} else {
				assert.NoError(t, err)
				t.Logf("Expected no error, received no error.")
			}
		})
	}
}

