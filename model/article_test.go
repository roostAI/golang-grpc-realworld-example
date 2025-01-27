package github

import (
	"strings"
	"testing"
	validation "github.com/go-ozzo/ozzo-validation"
)









/*
ROOST_METHOD_HASH=Overwrite_3d4db6693d
ROOST_METHOD_SIG_HASH=Overwrite_22e8730976

FUNCTION_DEF=func (a *Article) Overwrite(title, description, body string) 

 */
func TestArticleOverwrite(t *testing.T) {
	type fields struct {
		Title       string
		Description string
		Body        string
	}
	type args struct {
		title       string
		description string
		body        string
	}

	tests := []struct {
		name       string
		initial    fields
		args       args
		wantFields fields
	}{
		{
			name: "Update all fields of an Article",
			initial: fields{
				Title:       "Old Title",
				Description: "Old Description",
				Body:        "Old Body",
			},
			args: args{
				title:       "New Title",
				description: "New Description",
				body:        "New Body",
			},
			wantFields: fields{
				Title:       "New Title",
				Description: "New Description",
				Body:        "New Body",
			},
		},
		{
			name: "Update only the Title of an Article",
			initial: fields{
				Title:       "Old Title",
				Description: "Old Description",
				Body:        "Old Body",
			},
			args: args{
				title:       "New Title",
				description: "",
				body:        "",
			},
			wantFields: fields{
				Title:       "New Title",
				Description: "Old Description",
				Body:        "Old Body",
			},
		},
		{
			name: "Update only the Description of an Article",
			initial: fields{
				Title:       "Old Title",
				Description: "Old Description",
				Body:        "Old Body",
			},
			args: args{
				title:       "",
				description: "New Description",
				body:        "",
			},
			wantFields: fields{
				Title:       "Old Title",
				Description: "New Description",
				Body:        "Old Body",
			},
		},
		{
			name: "Update only the Body of an Article",
			initial: fields{
				Title:       "Old Title",
				Description: "Old Description",
				Body:        "Old Body",
			},
			args: args{
				title:       "",
				description: "",
				body:        "New Body",
			},
			wantFields: fields{
				Title:       "Old Title",
				Description: "Old Description",
				Body:        "New Body",
			},
		},
		{
			name: "No Fields Updated When All Strings are Empty",
			initial: fields{
				Title:       "Old Title",
				Description: "Old Description",
				Body:        "Old Body",
			},
			args: args{
				title:       "",
				description: "",
				body:        "",
			},
			wantFields: fields{
				Title:       "Old Title",
				Description: "Old Description",
				Body:        "Old Body",
			},
		},
		{
			name: "Update an Article with Long Strings",
			initial: fields{
				Title:       "Old Title",
				Description: "Old Description",
				Body:        "Old Body",
			},
			args: args{
				title:       strings.Repeat("Long Title ", 100),
				description: strings.Repeat("Long Description ", 100),
				body:        strings.Repeat("Long Body ", 100),
			},
			wantFields: fields{
				Title:       strings.Repeat("Long Title ", 100),
				Description: strings.Repeat("Long Description ", 100),
				Body:        strings.Repeat("Long Body ", 100),
			},
		},
		{
			name: "Update with Special Characters and Unicode Input",
			initial: fields{
				Title:       "Old Title",
				Description: "Old Description",
				Body:        "Old Body",
			},
			args: args{
				title:       "New Title ⚡",
				description: "New Description 🚀",
				body:        "New Body 🌟",
			},
			wantFields: fields{
				Title:       "New Title ⚡",
				Description: "New Description 🚀",
				Body:        "New Body 🌟",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			article := &Article{
				Title:       tt.initial.Title,
				Description: tt.initial.Description,
				Body:        tt.initial.Body,
			}
			article.Overwrite(tt.args.title, tt.args.description, tt.args.body)

			if article.Title != tt.wantFields.Title {
				t.Errorf("Overwrite() Title = %v, want %v", article.Title, tt.wantFields.Title)
			}
			if article.Description != tt.wantFields.Description {
				t.Errorf("Overwrite() Description = %v, want %v", article.Description, tt.wantFields.Description)
			}
			if article.Body != tt.wantFields.Body {
				t.Errorf("Overwrite() Body = %v, want %v", article.Body, tt.wantFields.Body)
			}

			t.Logf("Test '%s' passed", tt.name)
		})
	}
}


/*
ROOST_METHOD_HASH=Validate_f6d09c3ac5
ROOST_METHOD_SIG_HASH=Validate_99e41aac91

FUNCTION_DEF=func (a Article) Validate() error 

 */
func TestArticleValidate(t *testing.T) {
	tests := []struct {
		name         string
		article      Article
		expectError  bool
		errorMessage []string
	}{
		{
			name: "Scenario 1: Validation Passes When All Required Fields Are Present",
			article: Article{
				Title: "Valid Title",
				Body:  "Valid body content.",
				Tags:  []Tag{{Name: "tag1"}},
			},
			expectError: false,
		},
		{
			name: "Scenario 2: Missing Title Field",
			article: Article{
				Title: "",
				Body:  "Valid body content.",
				Tags:  []Tag{{Name: "tag1"}},
			},
			expectError:  true,
			errorMessage: []string{"Title"},
		},
		{
			name: "Scenario 3: Missing Body Field",
			article: Article{
				Title: "Valid Title",
				Body:  "",
				Tags:  []Tag{{Name: "tag1"}},
			},
			expectError:  true,
			errorMessage: []string{"Body"},
		},
		{
			name: "Scenario 4: Missing Tags",
			article: Article{
				Title: "Valid Title",
				Body:  "Valid body content.",
				Tags:  nil,
			},
			expectError:  true,
			errorMessage: []string{"Tags"},
		},
		{
			name: "Scenario 5: All Required Fields Missing",
			article: Article{
				Title: "",
				Body:  "",
				Tags:  nil,
			},
			expectError:  true,
			errorMessage: []string{"Title", "Body", "Tags"},
		},
		{
			name: "Scenario 6: Boundary Case with Empty Tag Name",
			article: Article{
				Title: "Valid Title",
				Body:  "Valid body content.",
				Tags:  []Tag{{Name: ""}},
			},
			expectError:  true,
			errorMessage: []string{"Tags"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.article.Validate()

			if tt.expectError {
				if err == nil {
					t.Errorf("expected an error, got nil")
				} else {
					t.Logf("expected error: %v", err)
					if !errorMessageContains(err, tt.errorMessage) {
						t.Errorf("error message does not contain expected field(s): %v", strings.Join(tt.errorMessage, ", "))
					}
				}
			} else {
				if err != nil {
					t.Errorf("expected no error, got %v", err)
				} else {
					t.Log("validation passed successfully")
				}
			}
		})
	}
}

func errorMessageContains(err error, fields []string) bool {
	errors, ok := err.(validation.Errors)
	if !ok {
		return false
	}

	for _, field := range fields {
		if _, exists := errors[field]; !exists {
			return false
		}
	}
	return true
}

