package model

import (
	testing "testing"
	time "time"
	pb "github.com/raahii/golang-grpc-realworld-example/proto"
	assert "github.com/stretchr/testify/assert"
	gorm "github.com/jinzhu/gorm"
	debug "runtime/debug"
)








/*
ROOST_METHOD_HASH=Article.Overwrite_f9458afd30
ROOST_METHOD_SIG_HASH=Article.Overwrite_bdf4e0ce1a

FUNCTION_DEF=func (a *Article) Overwrite(title, description, body string) // Overwrite overwrite each field if it's not zero-value


*/
func TestArticleOverwrite(t *testing.T) {

	testCases := []struct {
		name        string
		initial     Article
		newTitle    string
		newDesc     string
		newBody     string
		expected    Article
		description string
	}{
		{
			name: "Overwriting all fields of an article",
			initial: Article{
				Title:       "Initial Title",
				Description: "Initial Description",
				Body:        "Initial Body",
			},
			newTitle: "New Title",
			newDesc:  "New Description",
			newBody:  "New Body",
			expected: Article{
				Title:       "New Title",
				Description: "New Description",
				Body:        "New Body",
			},
			description: "This test is meant to check if the Overwrite function correctly overwrites all fields of an article when provided with non-empty strings for title, description, and body.",
		},
		{
			name: "Overwriting some fields of an article",
			initial: Article{
				Title:       "Initial Title",
				Description: "Initial Description",
				Body:        "Initial Body",
			},
			newTitle: "New Title",
			newDesc:  "",
			newBody:  "",
			expected: Article{
				Title:       "New Title",
				Description: "Initial Description",
				Body:        "Initial Body",
			},
			description: "This test is meant to check if the Overwrite function correctly overwrites some fields of an article when provided with non-empty strings for some parameters and empty strings for others.",
		},
		{
			name: "Not overwriting any fields of an article",
			initial: Article{
				Title:       "Initial Title",
				Description: "Initial Description",
				Body:        "Initial Body",
			},
			newTitle: "",
			newDesc:  "",
			newBody:  "",
			expected: Article{
				Title:       "Initial Title",
				Description: "Initial Description",
				Body:        "Initial Body",
			},
			description: "This test is meant to check if the Overwrite function does not overwrite any fields of an article when provided with empty strings for all parameters.",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Logf("Panic encountered so failing test. %v", r)
					t.Fail()
				}
			}()

			tc.initial.Overwrite(tc.newTitle, tc.newDesc, tc.newBody)

			if tc.initial.Title != tc.expected.Title || tc.initial.Description != tc.expected.Description || tc.initial.Body != tc.expected.Body {
				t.Errorf("Failed %s: Expected %v but got %v", tc.name, tc.expected, tc.initial)
			} else {
				t.Logf("Success %s: Expected %v and got %v", tc.name, tc.expected, tc.initial)
			}
		})
	}
}


/*
ROOST_METHOD_HASH=Article.ProtoArticle_9d0be39c93
ROOST_METHOD_SIG_HASH=Article.ProtoArticle_31d9b4d726

FUNCTION_DEF=func (a *Article) ProtoArticle(favorited bool) *pb.Article 

*/
func TestArticleProtoArticle(t *testing.T) {

	testCases := []struct {
		name      string
		article   Article
		favorited bool
		expected  *pb.Article
	}{
		{
			name: "Test ProtoArticle with a valid Article and favorited set to true",
			article: Article{
				Model: gorm.Model{
					ID:        1,
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				},
				Title:          "Test Title",
				Description:    "Test Description",
				Body:           "Test Body",
				Tags:           []Tag{{Model: gorm.Model{ID: 1}, Name: "Test"}},
				FavoritesCount: 1,
			},
			favorited: true,
			expected: &pb.Article{
				Slug:           "1",
				Title:          "Test Title",
				Description:    "Test Description",
				Body:           "Test Body",
				FavoritesCount: 1,
				Favorited:      true,
				CreatedAt:      time.Now().Format(time.RFC3339),
				UpdatedAt:      time.Now().Format(time.RFC3339),
				TagList:        []string{"Test"},
			},
		},
		{
			name: "Test ProtoArticle with a valid Article and favorited set to false",
			article: Article{
				Model: gorm.Model{
					ID:        2,
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				},
				Title:          "Test Title 2",
				Description:    "Test Description 2",
				Body:           "Test Body 2",
				Tags:           []Tag{{Model: gorm.Model{ID: 2}, Name: "Test 2"}},
				FavoritesCount: 2,
			},
			favorited: false,
			expected: &pb.Article{
				Slug:           "2",
				Title:          "Test Title 2",
				Description:    "Test Description 2",
				Body:           "Test Body 2",
				FavoritesCount: 2,
				Favorited:      false,
				CreatedAt:      time.Now().Format(time.RFC3339),
				UpdatedAt:      time.Now().Format(time.RFC3339),
				TagList:        []string{"Test 2"},
			},
		},
		{
			name: "Test ProtoArticle with an Article having no Tags",
			article: Article{
				Model: gorm.Model{
					ID:        3,
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				},
				Title:          "Test Title 3",
				Description:    "Test Description 3",
				Body:           "Test Body 3",
				FavoritesCount: 3,
			},
			favorited: true,
			expected: &pb.Article{
				Slug:           "3",
				Title:          "Test Title 3",
				Description:    "Test Description 3",
				Body:           "Test Body 3",
				FavoritesCount: 3,
				Favorited:      true,
				CreatedAt:      time.Now().Format(time.RFC3339),
				UpdatedAt:      time.Now().Format(time.RFC3339),
				TagList:        []string{},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Logf("Panic encountered so failing test. %v\n%s", r, string(debug.Stack()))
					t.Fail()
				}
			}()

			result := tc.article.ProtoArticle(tc.favorited)

			assert.Equal(t, tc.expected, result, "Expected and actual results do not match")
		})
	}
}

