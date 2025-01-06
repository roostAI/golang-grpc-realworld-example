package model

import (
	"testing"
	validation "github.com/go-ozzo/ozzo-validation"
)





type T struct {
	common
	isParallel bool
	isEnvSet   bool
	context    *testContext // For running tests and subtests.
}
func TestArticleValidate(t *testing.T) {
	type testCase struct {
		name           string
		article        Article
		expectedErrMsg string
	}

	tests := []testCase{
		{
			name: "Scenario 1: Title Field is Missing",
			article: Article{
				Body: "Sample body",
				Tags: []Tag{{Name: "Go"}},
			},
			expectedErrMsg: "Title: cannot be blank.",
		},
		{
			name: "Scenario 2: Body Field is Missing",
			article: Article{
				Title: "Sample Title",
				Tags:  []Tag{{Name: "Go"}},
			},
			expectedErrMsg: "Body: cannot be blank.",
		},
		{
			name: "Scenario 3: Tags Field is Missing",
			article: Article{
				Title: "Sample Title",
				Body:  "Sample body",
			},
			expectedErrMsg: "Tags: cannot be blank.",
		},
		{
			name: "Scenario 4: All Required Fields Present",
			article: Article{
				Title: "Sample Title",
				Body:  "Sample body",
				Tags:  []Tag{{Name: "Go"}},
			},
			expectedErrMsg: "",
		},
		{
			name: "Scenario 5: Long Title, but Not Exceeding Any Limit",
			article: Article{
				Title: "This is a very long title that nonetheless adheres to arbitrary constraints since no limit is defined",
				Body:  "Sample body",
				Tags:  []Tag{{Name: "Go"}},
			},
			expectedErrMsg: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.article.Validate()

			if err == nil {
				if tt.expectedErrMsg != "" {
					t.Errorf("Expected error message: '%s', got nil", tt.expectedErrMsg)
				} else {
					t.Logf("Success: No error as expected")
				}
			} else {
				validationErrs, ok := err.(validation.Errors)
				if !ok {
					t.Fatalf("Expected validation.Errors type, but got %T", err)
				}

				errMsg := validationErrs.Error()
				if errMsg != tt.expectedErrMsg {
					t.Errorf("Expected error message: '%s', got: '%s'", tt.expectedErrMsg, errMsg)
				} else {
					t.Logf("Success: Error message matched as expected")
				}
			}
		})
	}
}
