package github

import (
	"fmt"
	"strings"
	"testing"
	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/stretchr/testify/assert"
)









/*
ROOST_METHOD_HASH=Validate_1df97b5695
ROOST_METHOD_SIG_HASH=Validate_0591f679fe

FUNCTION_DEF=func (c Comment) Validate() error 

 */
func TestCommentValidate(t *testing.T) {
	tests := []struct {
		name       string
		comment    Comment
		expectErr  bool
		errMessage string
	}{
		{
			name: "Valid comment with non-empty body",
			comment: Comment{
				Body: "This is a valid comment.",
			},
			expectErr:  false,
			errMessage: "",
		},
		{
			name: "Invalid comment with empty body",
			comment: Comment{
				Body: "",
			},
			expectErr:  true,
			errMessage: "Body: cannot be blank.",
		},
		{
			name: "Invalid comment with whitespace body",
			comment: Comment{
				Body: "   ",
			},
			expectErr:  true,
			errMessage: "Body: cannot be blank.",
		},
		{
			name: "Valid comment with special characters",
			comment: Comment{
				Body: "!@#$%^&*()",
			},
			expectErr:  false,
			errMessage: "",
		},
		{
			name: "Valid comment with long body",
			comment: Comment{
				Body: strings.Repeat("a", 1000),
			},
			expectErr:  false,
			errMessage: "",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := tt.comment.Validate()

			if tt.expectErr {
				assert.Error(t, err, fmt.Sprintf("Expected error for test case: %s", tt.name))
				if err != nil {
					validationErr, ok := err.(validation.Errors)
					assert.True(t, ok, "Error is of type validation.Errors")
					assert.EqualError(t, validationErr, tt.errMessage, fmt.Sprintf("Error message should match for test case: %s", tt.name))
				}
			} else {
				assert.NoError(t, err, fmt.Sprintf("Expected no error for test case: %s", tt.name))
			}
		})
	}
}

