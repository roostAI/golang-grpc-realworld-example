package model

import "testing"








/*
ROOST_METHOD_HASH=Validate_1df97b5695
ROOST_METHOD_SIG_HASH=Validate_0591f679fe

FUNCTION_DEF=func (c Comment) Validate() error 

 */
func TestCommentValidate(t *testing.T) {
	tests := []struct {
		name          string
		comment       Comment
		expectedError bool
	}{
		{
			name: "Scenario 1: Valid Comment Body",
			comment: Comment{
				Body: "This is a valid comment body.",
			},
			expectedError: false,
		},
		{
			name: "Scenario 2: Empty Comment Body",
			comment: Comment{
				Body: "",
			},
			expectedError: true,
		},
		{
			name: "Scenario 3: Long Comment Body String",
			comment: Comment{
				Body: createLongString(5000),
			},
			expectedError: false,
		},
		{
			name: "Scenario 4: Comment Body with Special Characters",
			comment: Comment{
				Body: "!@#$%^&*()_+{}:?><",
			},
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.comment.Validate()
			if (err != nil) != tt.expectedError {
				t.Errorf("%s failed: expected error=%v, got %v", tt.name, tt.expectedError, err)
			}

			if tt.expectedError == false {
				t.Logf("%s passed: Comment body '%s' has passed validation as expected.", tt.name, tt.comment.Body)
			} else {
				t.Logf("%s passed: Comment body '%s' has been correctly identified as invalid.", tt.name, tt.comment.Body)
			}
		})
	}
}

func createLongString(n int) string {
	longStr := ""
	for i := 0; i < n; i++ {
		longStr += "a"
	}
	return longStr
}

