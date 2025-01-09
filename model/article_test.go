package undefined

import "testing"








/*
ROOST_METHOD_HASH=Overwrite_3d4db6693d
ROOST_METHOD_SIG_HASH=Overwrite_22e8730976

FUNCTION_DEF=func (a *Article) Overwrite(title, description, body string) 

 */
func TestArticleOverwrite(t *testing.T) {

	tests := []struct {
		name        string
		initial     Article
		title       string
		description string
		body        string
		expected    Article
	}{
		{
			name: "Overwrite all fields",
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
		},
		{
			name: "Partial overwrite with only title",
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
		},
		{
			name: "No change with all parameters empty",
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
		},
		{
			name: "Overwrite body and description, leave title unchanged",
			initial: Article{
				Title:       "Old Title",
				Description: "Old Description",
				Body:        "Old Body",
			},
			title:       "",
			description: "New Description",
			body:        "New Body",
			expected: Article{
				Title:       "Old Title",
				Description: "New Description",
				Body:        "New Body",
			},
		},
		{
			name: "Overwrite with special characters",
			initial: Article{
				Title:       "Old Title",
				Description: "Old Description",
				Body:        "Old Body",
			},
			title:       "!@#%^&*()",
			description: "{}[]:;'<>,.?/",
			body:        "~`+=-|\\",
			expected: Article{
				Title:       "!@#%^&*()",
				Description: "{}[]:;'<>,.?/",
				Body:        "~`+=-|\\",
			},
		},
		{
			name: "Large input values",
			initial: Article{
				Title:       "Old Title",
				Description: "Old Description",
				Body:        "Old Body",
			},
			title:       "T" + string(make([]byte, 1000)) + "tle",
			description: "D" + string(make([]byte, 1000)) + "scription",
			body:        "B" + string(make([]byte, 1000)) + "ody",
			expected: Article{
				Title:       "T" + string(make([]byte, 1000)) + "tle",
				Description: "D" + string(make([]byte, 1000)) + "scription",
				Body:        "B" + string(make([]byte, 1000)) + "ody",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			tt.initial.Overwrite(tt.title, tt.description, tt.body)

			if tt.initial.Title != tt.expected.Title {
				t.Errorf("Title mismatch, got: %s, want: %s", tt.initial.Title, tt.expected.Title)
			}
			if tt.initial.Description != tt.expected.Description {
				t.Errorf("Description mismatch, got: %s, want: %s", tt.initial.Description, tt.expected.Description)
			}
			if tt.initial.Body != tt.expected.Body {
				t.Errorf("Body mismatch, got: %s, want: %s", tt.initial.Body, tt.expected.Body)
			}
			t.Logf("Article fields correctly overwritten in scenario: %s", tt.name)
		})
	}
}

