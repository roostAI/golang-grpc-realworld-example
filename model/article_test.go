package model

import (
	"testing"
	"time"
	pb "github.com/raahii/golang-grpc-realworld-example/proto"
	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/assert"
)









/*
ROOST_METHOD_HASH=Overwrite_3d4db6693d
ROOST_METHOD_SIG_HASH=Overwrite_22e8730976

FUNCTION_DEF=func (a *Article) Overwrite(title, description, body string) 

 */
func TestArticleOverwrite(t *testing.T) {
	tests := []struct {
		name                                      string
		initialTitle, initialDesc, initialBody    string
		newTitle, newDesc, newBody                string
		expectedTitle, expectedDesc, expectedBody string
	}{
		{
			name:          "Overwrite Title, Description, and Body",
			initialTitle:  "Old Title",
			initialDesc:   "Old Description",
			initialBody:   "Old Body",
			newTitle:      "New Title",
			newDesc:       "New Description",
			newBody:       "New Body",
			expectedTitle: "New Title",
			expectedDesc:  "New Description",
			expectedBody:  "New Body",
		},
		{
			name:          "Overwrite Title Only",
			initialTitle:  "Old Title",
			initialDesc:   "Old Description",
			initialBody:   "Old Body",
			newTitle:      "New Title",
			newDesc:       "",
			newBody:       "",
			expectedTitle: "New Title",
			expectedDesc:  "Old Description",
			expectedBody:  "Old Body",
		},
		{
			name:          "Overwrite Description Only",
			initialTitle:  "Old Title",
			initialDesc:   "Old Description",
			initialBody:   "Old Body",
			newTitle:      "",
			newDesc:       "New Description",
			newBody:       "",
			expectedTitle: "Old Title",
			expectedDesc:  "New Description",
			expectedBody:  "Old Body",
		},
		{
			name:          "Overwrite Body Only",
			initialTitle:  "Old Title",
			initialDesc:   "Old Description",
			initialBody:   "Old Body",
			newTitle:      "",
			newDesc:       "",
			newBody:       "New Body",
			expectedTitle: "Old Title",
			expectedDesc:  "Old Description",
			expectedBody:  "New Body",
		},
		{
			name:          "No Overwrite on Empty Inputs",
			initialTitle:  "Old Title",
			initialDesc:   "Old Description",
			initialBody:   "Old Body",
			newTitle:      "",
			newDesc:       "",
			newBody:       "",
			expectedTitle: "Old Title",
			expectedDesc:  "Old Description",
			expectedBody:  "Old Body",
		},
		{
			name:          "Overwrite with Whitespace",
			initialTitle:  "Old Title",
			initialDesc:   "Old Description",
			initialBody:   "Old Body",
			newTitle:      "   ",
			newDesc:       "   ",
			newBody:       "   ",
			expectedTitle: "Old Title",
			expectedDesc:  "Old Description",
			expectedBody:  "Old Body",
		},
	}

	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {

			article := Article{
				Title:       tt.initialTitle,
				Description: tt.initialDesc,
				Body:        tt.initialBody,
			}

			article.Overwrite(tt.newTitle, tt.newDesc, tt.newBody)

			if article.Title != tt.expectedTitle {
				t.Errorf("Title = %v; want %v", article.Title, tt.expectedTitle)
			}
			if article.Description != tt.expectedDesc {
				t.Errorf("Description = %v; want %v", article.Description, tt.expectedDesc)
			}
			if article.Body != tt.expectedBody {
				t.Errorf("Body = %v; want %v", article.Body, tt.expectedBody)
			}

			t.Logf("Scenario %q passed successfully.", tt.name)
		})
	}
}


/*
ROOST_METHOD_HASH=ProtoArticle_4b12477d53
ROOST_METHOD_SIG_HASH=ProtoArticle_31d9b4d726

FUNCTION_DEF=func (a *Article) ProtoArticle(favorited bool) *pb.Article 

 */
func TestArticleProtoArticle(t *testing.T) {
	type test struct {
		name      string
		article   Article
		favorited bool
		expected  pb.Article
	}

	tests := []test{
		{
			name: "Fully Populated Article",
			article: Article{
				Model: gorm.Model{
					ID:        1,
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				},
				Title:          "Test Title",
				Description:    "Test Description",
				Body:           "Test Body",
				Tags:           []Tag{{Name: "Tag1"}, {Name: "Tag2"}},
				FavoritesCount: 10,
			},
			favorited: true,
			expected: pb.Article{
				Slug:           "1",
				Title:          "Test Title",
				Description:    "Test Description",
				Body:           "Test Body",
				TagList:        []string{"Tag1", "Tag2"},
				FavoritesCount: 10,
				Favorited:      true,
			},
		},
		{
			name: "Article Without Tags",
			article: Article{
				Model: gorm.Model{
					ID:        2,
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				},
				Title:       "No Tags Title",
				Description: "No Tags Description",
				Body:        "No Tags Body",
			},
			favorited: false,
			expected: pb.Article{
				Slug:        "2",
				Title:       "No Tags Title",
				Description: "No Tags Description",
				Body:        "No Tags Body",
				TagList:     []string{},
				Favorited:   false,
			},
		},
		{
			name: "Article with Zero Values",
			article: Article{
				Model: gorm.Model{
					ID:        3,
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				},
				Title:       "",
				Description: "",
				Body:        "",
			},
			favorited: false,
			expected: pb.Article{
				Slug:        "3",
				Title:       "",
				Description: "",
				Body:        "",
				TagList:     []string{},
				Favorited:   false,
			},
		},
		{
			name: "Minimum Required Fields",
			article: Article{
				Model: gorm.Model{
					ID:        4,
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				},
				Title:       "Minimal Title",
				Description: "Minimal Description",
				Body:        "Minimal Body",
			},
			favorited: false,
			expected: pb.Article{
				Slug:        "4",
				Title:       "Minimal Title",
				Description: "Minimal Description",
				Body:        "Minimal Body",
				TagList:     []string{},
				Favorited:   false,
			},
		},
		{
			name: "Article with Future Dates",
			article: Article{
				Model: gorm.Model{
					ID:        5,
					CreatedAt: time.Now().Add(time.Hour * 24),
					UpdatedAt: time.Now().Add(time.Hour * 24),
				},
				Title:       "Future Title",
				Description: "Future Description",
				Body:        "Future Body",
			},
			favorited: true,
			expected: pb.Article{
				Slug:        "5",
				Title:       "Future Title",
				Description: "Future Description",
				Body:        "Future Body",
				TagList:     []string{},
				Favorited:   true,
			},
		},
	}

	for _, tst := range tests {
		t.Run(tst.name, func(t *testing.T) {
			actual := tst.article.ProtoArticle(tst.favorited)

			assert.Equal(t, tst.expected.Slug, actual.Slug, "The Slug should be equal")
			assert.Equal(t, tst.expected.Title, actual.Title, "The Title should be equal")
			assert.Equal(t, tst.expected.Description, actual.Description, "The Description should be equal")
			assert.Equal(t, tst.expected.Body, actual.Body, "The Body should be equal")
			assert.ElementsMatch(t, tst.expected.TagList, actual.TagList, "The TagList should be equal")
			assert.Equal(t, tst.expected.FavoritesCount, actual.FavoritesCount, "The FavoritesCount should be equal")
			assert.Equal(t, tst.expected.Favorited, actual.Favorited, "The Favorited should be equal")

			_, err := time.Parse(ISO8601, actual.CreatedAt)
			assert.NoError(t, err, "CreatedAt should parse without error")
			_, err = time.Parse(ISO8601, actual.UpdatedAt)
			assert.NoError(t, err, "UpdatedAt should parse without error")
		})
	}
}

