package model

import "testing"





type T struct {
	common
	isParallel bool
	isEnvSet   bool
	context    *testContext // For running tests and subtests.
}
func TestArticleOverwrite(t *testing.T) {
	tests := []struct {
		name        string
		initial     Article
		title       string
		description string
		body        string
		expected    Article
		explanation string
	}{
		{
			name: "Overwrite All Fields",
			initial: Article{
				Title:       "Old Title",
				Description: "Old Description",
				Body:        "Old Body",
			},
			title:       "New Title",
			description: "New Description",
			body:        "New Body",
			expected: Article{
				Title:       "New Title",
				Description: "New Description",
				Body:        "New Body",
			},
			explanation: "All fields are provided, expecting all to be updated",
		},
		{
			name: "Overwrite Only Title",
			initial: Article{
				Title:       "Old Title",
				Description: "Old Description",
				Body:        "Old Body",
			},
			title:       "New Title",
			description: "",
			body:        "",
			expected: Article{
				Title:       "New Title",
				Description: "Old Description",
				Body:        "Old Body",
			},
			explanation: "Only title is provided, expecting only title to be updated",
		},
		{
			name: "Overwrite Only Description",
			initial: Article{
				Title:       "Old Title",
				Description: "Old Description",
				Body:        "Old Body",
			},
			title:       "",
			description: "New Description",
			body:        "",
			expected: Article{
				Title:       "Old Title",
				Description: "New Description",
				Body:        "Old Body",
			},
			explanation: "Only description is provided, expecting only description to be updated",
		},
		{
			name: "Overwrite Only Body",
			initial: Article{
				Title:       "Old Title",
				Description: "Old Description",
				Body:        "Old Body",
			},
			title:       "",
			description: "",
			body:        "New Body",
			expected: Article{
				Title:       "Old Title",
				Description: "Old Description",
				Body:        "New Body",
			},
			explanation: "Only body is provided, expecting only body to be updated",
		},
		{
			name: "No Change When All Empty",
			initial: Article{
				Title:       "Old Title",
				Description: "Old Description",
				Body:        "Old Body",
			},
			title:       "",
			description: "",
			body:        "",
			expected: Article{
				Title:       "Old Title",
				Description: "Old Description",
				Body:        "Old Body",
			},
			explanation: "All inputs are empty, expecting no changes to initial state",
		},
		{
			name: "Mixed Values Update",
			initial: Article{
				Title:       "Old Title",
				Description: "Old Description",
				Body:        "Old Body",
			},
			title:       "New Title",
			description: "",
			body:        "New Body",
			expected: Article{
				Title:       "New Title",
				Description: "Old Description",
				Body:        "New Body",
			},
			explanation: "Mixed values provided, expecting specified fields to be updated",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			article := tt.initial
			article.Overwrite(tt.title, tt.description, tt.body)

			if article.Title != tt.expected.Title || article.Description != tt.expected.Description || article.Body != tt.expected.Body {
				t.Errorf("Overwrite() = %v, expected %v", article, tt.expected)
			} else {
				t.Logf("Success: %s. Explanation: %s", tt.name, tt.explanation)
			}
		})
	}

}
