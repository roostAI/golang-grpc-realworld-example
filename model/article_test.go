package github

import (
	"testing"
	"time"
	"github.com/jinzhu/gorm"
	pb "github.com/raahii/golang-grpc-realworld-example/proto"
	"github.com/stretchr/testify/assert"
	validation "github.com/go-ozzo/ozzo-validation"
)









/*
ROOST_METHOD_HASH=Overwrite_3d4db6693d
ROOST_METHOD_SIG_HASH=Overwrite_22e8730976

FUNCTION_DEF=func (a *Article) Overwrite(title, description, body string) 

 */
func TestArticleOverwrite(t *testing.T) {

	type test struct {
		name        string
		initial     Article
		title       string
		description string
		body        string
		expected    Article
	}

	tests := []test{
		{
			name: "Scenario 1: Overwrite All Fields of an Article with Non-Empty Strings",
			initial: Article{
				Title:       "Initial Title",
				Description: "Initial Description",
				Body:        "Initial Body",
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
			name: "Scenario 2: Partial Overwrite of an Article's Fields",
			initial: Article{
				Title:       "Initial Title",
				Description: "Initial Description",
				Body:        "Initial Body",
			},
			title:       "",
			description: "Updated Description",
			body:        "",
			expected: Article{
				Title:       "Initial Title",
				Description: "Updated Description",
				Body:        "Initial Body",
			},
		},
		{
			name: "Scenario 3: No Changes When All Inputs are Empty Strings",
			initial: Article{
				Title:       "Initial Title",
				Description: "Initial Description",
				Body:        "Initial Body",
			},
			title:       "",
			description: "",
			body:        "",
			expected: Article{
				Title:       "Initial Title",
				Description: "Initial Description",
				Body:        "Initial Body",
			},
		},
		{
			name: "Scenario 4: Overwrite with Whitespace",
			initial: Article{
				Title:       "Initial Title",
				Description: "Initial Description",
				Body:        "Initial Body",
			},
			title:       "    ",
			description: "    ",
			body:        "    ",
			expected: Article{
				Title:       "Initial Title",
				Description: "Initial Description",
				Body:        "Initial Body",
			},
		},
		{
			name: "Scenario 5: Confirm Correct Behavior With Complex Strings",
			initial: Article{
				Title:       "Simple Title",
				Description: "Simple Description",
				Body:        "Simple Body",
			},
			title:       "@NewTitle!",
			description: "A new, comprehensively-detailed: Description?",
			body:        "Updated Body$",
			expected: Article{
				Title:       "@NewTitle!",
				Description: "A new, comprehensively-detailed: Description?",
				Body:        "Updated Body$",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			article := tt.initial

			article.Overwrite(tt.title, tt.description, tt.body)

			if article.Title != tt.expected.Title {
				t.Errorf("Title: Expected '%s', but got '%s'", tt.expected.Title, article.Title)
			}
			if article.Description != tt.expected.Description {
				t.Errorf("Description: Expected '%s', but got '%s'", tt.expected.Description, article.Description)
			}
			if article.Body != tt.expected.Body {
				t.Errorf("Body: Expected '%s', but got '%s'", tt.expected.Body, article.Body)
			}

			t.Logf("Success: %s", tt.name)
		})
	}
}


/*
ROOST_METHOD_HASH=ProtoArticle_4b12477d53
ROOST_METHOD_SIG_HASH=ProtoArticle_31d9b4d726

FUNCTION_DEF=func (a *Article) ProtoArticle(favorited bool) *pb.Article 

 */
func TestArticleProtoArticle(t *testing.T) {
	tests := []struct {
		name       string
		article    Article
		favorited  bool
		expectedPA *pb.Article
	}{
		{
			name: "Basic Conversion of Article to Proto Article",
			article: Article{
				Model: gorm.Model{
					ID:        1,
					CreatedAt: time.Date(2023, 10, 1, 10, 0, 0, 0, time.UTC),
					UpdatedAt: time.Date(2023, 10, 1, 10, 0, 0, 0, time.UTC),
				},
				Title:          "Test Title",
				Description:    "Test Description",
				Body:           "Test Body",
				Tags:           []Tag{},
				FavoritesCount: 0,
			},
			favorited: false,
			expectedPA: &pb.Article{
				Slug:           "1",
				Title:          "Test Title",
				Description:    "Test Description",
				Body:           "Test Body",
				TagList:        []string{},
				FavoritesCount: 0,
				Favorited:      false,
				CreatedAt:      "2023-10-01T10:00:00+0000Z",
				UpdatedAt:      "2023-10-01T10:00:00+0000Z",
			},
		},
		{
			name: "Conversion with Tags",
			article: Article{
				Model: gorm.Model{
					ID:        2,
					CreatedAt: time.Date(2023, 10, 1, 10, 0, 0, 0, time.UTC),
					UpdatedAt: time.Date(2023, 10, 1, 10, 0, 0, 0, time.UTC),
				},
				Title:       "Tagged Article",
				Description: "Article with Tags",
				Body:        "Body with Tags",
				Tags: []Tag{
					{Name: "Tech"},
					{Name: "Golang"},
				},
				FavoritesCount: 0,
			},
			favorited: false,
			expectedPA: &pb.Article{
				Slug:           "2",
				Title:          "Tagged Article",
				Description:    "Article with Tags",
				Body:           "Body with Tags",
				TagList:        []string{"Tech", "Golang"},
				FavoritesCount: 0,
				Favorited:      false,
				CreatedAt:      "2023-10-01T10:00:00+0000Z",
				UpdatedAt:      "2023-10-01T10:00:00+0000Z",
			},
		},
		{
			name: "Conversion with Favorited Status",
			article: Article{
				Model: gorm.Model{
					ID:        3,
					CreatedAt: time.Date(2023, 10, 1, 10, 0, 0, 0, time.UTC),
					UpdatedAt: time.Date(2023, 10, 1, 10, 0, 0, 0, time.UTC),
				},
				Title:          "Favorited Article",
				Description:    "Article marked as favorited",
				Body:           "Favorited Article Body",
				Tags:           []Tag{},
				FavoritesCount: 0,
			},
			favorited: true,
			expectedPA: &pb.Article{
				Slug:           "3",
				Title:          "Favorited Article",
				Description:    "Article marked as favorited",
				Body:           "Favorited Article Body",
				TagList:        []string{},
				FavoritesCount: 0,
				Favorited:      true,
				CreatedAt:      "2023-10-01T10:00:00+0000Z",
				UpdatedAt:      "2023-10-01T10:00:00+0000Z",
			},
		},
		{
			name: "Edge Case - Empty Article",
			article: Article{
				Model: gorm.Model{
					ID:        4,
					CreatedAt: time.Time{},
					UpdatedAt: time.Time{},
				},
				Title:          "",
				Description:    "",
				Body:           "",
				Tags:           []Tag{},
				FavoritesCount: 0,
			},
			favorited: false,
			expectedPA: &pb.Article{
				Slug:           "4",
				Title:          "",
				Description:    "",
				Body:           "",
				TagList:        []string{},
				FavoritesCount: 0,
				Favorited:      false,
				CreatedAt:      "0001-01-01T00:00:00+0000Z",
				UpdatedAt:      "0001-01-01T00:00:00+0000Z",
			},
		},
		{
			name: "Edge Case - Maximum Tags",
			article: Article{
				Model: gorm.Model{
					ID:        5,
					CreatedAt: time.Date(2023, 10, 1, 10, 0, 0, 0, time.UTC),
					UpdatedAt: time.Date(2023, 10, 1, 10, 0, 0, 0, time.UTC),
				},
				Title:          "Article with many tags",
				Description:    "Description with many tags",
				Body:           "Body with many tags",
				Tags:           make([]Tag, 1000),
				FavoritesCount: 0,
			},
			favorited: false,
			expectedPA: &pb.Article{
				Slug:           "5",
				Title:          "Article with many tags",
				Description:    "Description with many tags",
				Body:           "Body with many tags",
				TagList:        make([]string, 1000),
				FavoritesCount: 0,
				Favorited:      false,
				CreatedAt:      "2023-10-01T10:00:00+0000Z",
				UpdatedAt:      "2023-10-01T10:00:00+0000Z",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			protoArticle := tt.article.ProtoArticle(tt.favorited)

			assert.Equal(t, tt.expectedPA.Slug, protoArticle.Slug)
			assert.Equal(t, tt.expectedPA.Title, protoArticle.Title)
			assert.Equal(t, tt.expectedPA.Description, protoArticle.Description)
			assert.Equal(t, tt.expectedPA.Body, protoArticle.Body)
			assert.Equal(t, tt.expectedPA.TagList, protoArticle.TagList)
			assert.Equal(t, tt.expectedPA.FavoritesCount, protoArticle.FavoritesCount)
			assert.Equal(t, tt.expectedPA.Favorited, protoArticle.Favorited)
			assert.Equal(t, tt.expectedPA.CreatedAt, protoArticle.CreatedAt)
			assert.Equal(t, tt.expectedPA.UpdatedAt, protoArticle.UpdatedAt)

			t.Log("Success: Completed test for scenario", tt.name)
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
		name     string
		article  Article
		wantErr  bool
		errField string
	}{
		{
			name: "Scenario 1: Successful Validation with All Required Fields Present",
			article: Article{
				Title: "Sample Title",
				Body:  "Sample Body",
				Tags:  []Tag{{Name: "Go"}, {Name: "Testing"}},
			},
			wantErr: false,
		},
		{
			name: "Scenario 2: Validation Failure When Title Is Missing",
			article: Article{
				Title: "",
				Body:  "Sample Body",
				Tags:  []Tag{{Name: "Go"}, {Name: "Testing"}},
			},
			wantErr:  true,
			errField: "Title",
		},
		{
			name: "Scenario 3: Validation Failure When Body Is Missing",
			article: Article{
				Title: "Sample Title",
				Body:  "",
				Tags:  []Tag{{Name: "Go"}, {Name: "Testing"}},
			},
			wantErr:  true,
			errField: "Body",
		},
		{
			name: "Scenario 4: Validation Failure When Tags Are Missing",
			article: Article{
				Title: "Sample Title",
				Body:  "Sample Body",
				Tags:  nil,
			},
			wantErr:  true,
			errField: "Tags",
		},
		{
			name: "Scenario 5: Validation with All Fields Optional (Edge Case No Tags Required)",

			article: Article{
				Title: "Sample Title",
				Body:  "Sample Body",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			err := tt.article.Validate()

			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantErr {
				if verrs, ok := err.(validation.Errors); ok {
					if verrs[tt.errField] == nil {
						t.Errorf("Expected validation error for field '%s', but got none", tt.errField)
					}
				} else {
					t.Errorf("Expected validation.Errors type for error, got %T", err)
				}
			}

			t.Logf("Scenario \"%s\": Passed", tt.name)
		})
	}
}

