package model

import (
	"regexp"
	"testing"
	"time"
	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/go-ozzo/ozzo-validation/is"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)





type T struct {
	common
	isParallel bool
	isEnvSet   bool
	context    *testContext // For running tests and subtests.
}
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

