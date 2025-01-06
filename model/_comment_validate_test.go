package model

import (
	"strings"
	"testing"
	validation "github.com/go-ozzo/ozzo-validation"
)







type T struct {
	common
	isParallel bool
	isEnvSet   bool
	context    *testContext // For running tests and subtests.
}
func TestCommentValidate(t *testing.T) {
	tests := []struct {
		name        string
		comment     Comment
		expectError bool
		errorMsg    string
	}{
		{
			name: "Valid Comment Body",
			comment: Comment{
				Body: "This is a valid comment.",
			},
			expectError: false,
			errorMsg:    "expected no error for a valid comment body",
		},
		{
			name: "Missing Comment Body",
			comment: Comment{
				Body: "",
			},
			expectError: true,
			errorMsg:    "expected an error for a missing body",
		},
		{
			name: "Whitespace Comment Body",
			comment: Comment{
				Body: "   ",
			},
			expectError: true,
			errorMsg:    "expected an error for a body containing only whitespace",
		},
		{
			name: "Extremely Large Comment Body",
			comment: Comment{
				Body: strings.Repeat("a", 10000),
			},
			expectError: false,
			errorMsg:    "expected no error for a large comment body",
		},
		{
			name: "Special Characters in Comment Body",
			comment: Comment{
				Body: "This is a comment with special characters! 😃✨💡",
			},
			expectError: false,
			errorMsg:    "expected no error for a body with special characters",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.comment.Validate()

			if tc.expectError {
				if err == nil {
					t.Errorf("%s - %s", tc.name, tc.errorMsg)
				} else {
					t.Logf("%s - passed as expected, error: %v", tc.name, err)
				}
			} else {
				if err != nil {
					t.Errorf("%s - unexpected error: %v", tc.name, err)
				} else {
					t.Logf("%s - validation passed with no errors", tc.name)
				}
			}
		})
	}
}

