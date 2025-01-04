package model

import (
	"testing"
	"fmt"
	"time"
	"github.com/jinzhu/gorm"
	"github.com/raahii/golang-grpc-realworld-example/proto"
	"github.com/go-ozzo/ozzo-validation"
)

type T struct {
	common
	isParallel bool
	isEnvSet   bool
	context    *testContext
}
/*
ROOST_METHOD_HASH=Overwrite_3d4db6693d
ROOST_METHOD_SIG_HASH=Overwrite_22e8730976


 */
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


/*
ROOST_METHOD_HASH=ProtoArticle_4b12477d53
ROOST_METHOD_SIG_HASH=ProtoArticle_31d9b4d726


 */
func TestArticleProtoArticle(t *testing.T) {
	type test struct {
		name      string
		input     Article
		favorited bool
		expected  pb.Article
	}

	tests := []test{
		{
			name: "Convert Article with All Fields Set",
			input: Article{
				Model: gorm.Model{
					ID:        1,
					CreatedAt: time.Now().Add(-24 * time.Hour),
					UpdatedAt: time.Now(),
				},
				Title:          "Test Article",
				Description:    "This is a test article",
				Body:           "Article content...",
				Tags:           []Tag{{Name: "go"}, {Name: "grpc"}},
				Author:         User{Model: gorm.Model{ID: 1}, Username: "author1"},
				UserID:         1,
				FavoritesCount: 10,
			},
			favorited: true,
			expected: pb.Article{
				Slug:           "1",
				Title:          "Test Article",
				Description:    "This is a test article",
				Body:           "Article content...",
				TagList:        []string{"go", "grpc"},
				FavoritesCount: 10,
				Favorited:      true,
			},
		},
		{
			name: "Convert Article with No Tags",
			input: Article{
				Model: gorm.Model{
					ID:        2,
					CreatedAt: time.Now().Add(-48 * time.Hour),
					UpdatedAt: time.Now(),
				},
				Title:          "No Tags Article",
				Description:    "No tags in this article",
				Body:           "Content without tags...",
				Tags:           []Tag{},
				Author:         User{Model: gorm.Model{ID: 2}, Username: "author2"},
				UserID:         2,
				FavoritesCount: 5,
			},
			favorited: false,
			expected: pb.Article{
				Slug:           "2",
				Title:          "No Tags Article",
				Description:    "No tags in this article",
				Body:           "Content without tags...",
				TagList:        []string{},
				FavoritesCount: 5,
				Favorited:      false,
			},
		},
		{
			name: "Validate Date Formatting",
			input: Article{
				Model: gorm.Model{
					ID:        3,
					CreatedAt: time.Date(2023, time.September, 12, 10, 15, 30, 0, time.UTC),
					UpdatedAt: time.Date(2023, time.September, 15, 12, 22, 10, 0, time.UTC),
				},
				Title:          "Date Test Article",
				Description:    "Testing date formats",
				Body:           "Timestamp conversions...",
				Tags:           []Tag{{Name: "testing"}},
				Author:         User{Model: gorm.Model{ID: 3}, Username: "author3"},
				UserID:         3,
				FavoritesCount: 0,
			},
			favorited: false,
			expected: pb.Article{
				Slug:           "3",
				Title:          "Date Test Article",
				Description:    "Testing date formats",
				Body:           "Timestamp conversions...",
				TagList:        []string{"testing"},
				FavoritesCount: 0,
				Favorited:      false,
				CreatedAt:      "2023-09-12T10:15:30+0000Z",
				UpdatedAt:      "2023-09-15T12:22:10+0000Z",
			},
		},
		{
			name: "Convert Article with Zero Favorite Count",
			input: Article{
				Model: gorm.Model{
					ID:        4,
					CreatedAt: time.Now().Add(-72 * time.Hour),
					UpdatedAt: time.Now(),
				},
				Title:          "Zero Fav Count",
				Description:    "No favorites yet",
				Body:           "Fresh article...",
				Tags:           []Tag{},
				Author:         User{Model: gorm.Model{ID: 4}, Username: "author4"},
				UserID:         4,
				FavoritesCount: 0,
			},
			favorited: true,
			expected: pb.Article{
				Slug:           "4",
				Title:          "Zero Fav Count",
				Description:    "No favorites yet",
				Body:           "Fresh article...",
				TagList:        []string{},
				FavoritesCount: 0,
				Favorited:      true,
			},
		},
		{
			name: "Edge Case with Missing Description",
			input: Article{
				Model: gorm.Model{
					ID:        5,
					CreatedAt: time.Now().Add(-96 * time.Hour),
					UpdatedAt: time.Now(),
				},
				Title:          "Article with no Description",
				Description:    "",
				Body:           "Still has content...",
				Tags:           []Tag{},
				Author:         User{Model: gorm.Model{ID: 5}, Username: "author5"},
				UserID:         5,
				FavoritesCount: 1,
			},
			favorited: false,
			expected: pb.Article{
				Slug:           "5",
				Title:          "Article with no Description",
				Description:    "",
				Body:           "Still has content...",
				TagList:        []string{},
				FavoritesCount: 1,
				Favorited:      false,
			},
		},
		{
			name: "Article with Long Title",
			input: Article{
				Model: gorm.Model{
					ID:        6,
					CreatedAt: time.Now().Add(-120 * time.Hour),
					UpdatedAt: time.Now(),
				},
				Title:          "This is an extremely long title that should be tested for cases where it might reach known limits on title length processing",
				Description:    "Testing max title length",
				Body:           "Ensuring long titles work...",
				Tags:           []Tag{},
				Author:         User{Model: gorm.Model{ID: 6}, Username: "author6"},
				UserID:         6,
				FavoritesCount: 3,
			},
			favorited: false,
			expected: pb.Article{
				Slug:           "6",
				Title:          "This is an extremely long title that should be tested for cases where it might reach known limits on title length processing",
				Description:    "Testing max title length",
				Body:           "Ensuring long titles work...",
				TagList:        []string{},
				FavoritesCount: 3,
				Favorited:      false,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			actual := tc.input.ProtoArticle(tc.favorited)

			if actual.Slug != tc.expected.Slug {
				t.Errorf("expected Slug %s, got %s", tc.expected.Slug, actual.Slug)
			}
			if actual.Title != tc.expected.Title {
				t.Errorf("expected Title %s, got %s", tc.expected.Title, actual.Title)
			}
			if actual.Description != tc.expected.Description {
				t.Errorf("expected Description %s, got %s", tc.expected.Description, actual.Description)
			}
			if actual.Body != tc.expected.Body {
				t.Errorf("expected Body %s, got %s", tc.expected.Body, actual.Body)
			}
			if len(actual.TagList) != len(tc.expected.TagList) {
				t.Errorf("expected TagList length %d, got %d", len(tc.expected.TagList), len(actual.TagList))
			}
			for i, tag := range actual.TagList {
				if tag != tc.expected.TagList[i] {
					t.Errorf("expected TagList[%d] %s, got %s", i, tc.expected.TagList[i], tag)
				}
			}
			if actual.CreatedAt != tc.expected.CreatedAt {
				t.Errorf("expected CreatedAt %s, got %s", tc.expected.CreatedAt, actual.CreatedAt)
			}
			if actual.UpdatedAt != tc.expected.UpdatedAt {
				t.Errorf("expected UpdatedAt %s, got %s", tc.expected.UpdatedAt, actual.UpdatedAt)
			}
			if actual.Favorited != tc.expected.Favorited {
				t.Errorf("expected Favorited %v, got %v", tc.expected.Favorited, actual.Favorited)
			}
			if actual.FavoritesCount != tc.expected.FavoritesCount {
				t.Errorf("expected FavoritesCount %d, got %d", tc.expected.FavoritesCount, actual.FavoritesCount)
			}

			t.Logf("Success: %s", tc.name)
		})
	}
}


/*
ROOST_METHOD_HASH=Validate_f6d09c3ac5
ROOST_METHOD_SIG_HASH=Validate_99e41aac91


 */
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

