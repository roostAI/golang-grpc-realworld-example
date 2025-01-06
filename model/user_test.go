package model

import (
	"testing"
	"github.com/raahii/golang-grpc-realworld-example/proto"
	"reflect"
	"golang.org/x/crypto/bcrypt"
	"log"
	"github.com/stretchr/testify/assert"
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
		name  string
		user  User
		token string
		want  *pb.User
	}{
		{
			name: "Successful Conversion of User to Proto User with Valid Token",
			user: User{
				Username: "testuser",
				Email:    "test@example.com",
				Bio:      "This is a bio",
				Image:    "http://example.com/image.png",
			},
			token: "validtoken123",
			want: &pb.User{
				Email:    "test@example.com",
				Token:    "validtoken123",
				Username: "testuser",
				Bio:      "This is a bio",
				Image:    "http://example.com/image.png",
			},
		},
		{
			name: "Conversion of User with Empty Fields to Proto User",
			user: User{
				Username: "emptyuser",
				Email:    "empty@example.com",
				Bio:      "",
				Image:    "",
			},
			token: "emptiestoken",
			want: &pb.User{
				Email:    "empty@example.com",
				Token:    "emptiestoken",
				Username: "emptyuser",
				Bio:      "",
				Image:    "",
			},
		},
		{
			name: "Conversion of User with Special Characters in Fields",
			user: User{
				Username: "specialuser😊",
				Email:    "special@exämple.com",
				Bio:      "Bio with emojis 🤔✨",
				Image:    "http://example.com/image!@#.png",
			},
			token: "specialtoken",
			want: &pb.User{
				Email:    "special@exämple.com",
				Token:    "specialtoken",
				Username: "specialuser😊",
				Bio:      "Bio with emojis 🤔✨",
				Image:    "http://example.com/image!@#.png",
			},
		},
		{
			name: "Consistency Across Multiple ProtoUser Conversions",
			user: User{
				Username: "consistentuser",
				Email:    "consistent@example.com",
				Bio:      "Consistent bio",
				Image:    "http://example.com/consistent.png",
			},
			token: "consisttoken",
			want: &pb.User{
				Email:    "consistent@example.com",
				Token:    "consisttoken",
				Username: "consistentuser",
				Bio:      "Consistent bio",
				Image:    "http://example.com/consistent.png",
			},
		},
		{
			name: "Conversion of Proto User with Null Token Parameter",
			user: User{
				Username: "nulltokenuser",
				Email:    "nulltoken@example.com",
				Bio:      "Some bio",
				Image:    "http://example.com/nullimage.png",
			},
			token: "",
			want: &pb.User{
				Email:    "nulltoken@example.com",
				Token:    "",
				Username: "nulltokenuser",
				Bio:      "Some bio",
				Image:    "http://example.com/nullimage.png",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.user.ProtoUser(tt.token)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("%s failed; got %v, want %v", tt.name, got, tt.want)
				t.Logf("Expected ProtoUser to convert fields as specified.")
			} else {
				t.Logf("%s passed successfully.", tt.name)
			}
		})
	}

}


/*
ROOST_METHOD_HASH=CheckPassword_377b31181b
ROOST_METHOD_SIG_HASH=CheckPassword_e6e0413d83


 */
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

