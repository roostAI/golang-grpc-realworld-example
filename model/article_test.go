package undefined










/*
ROOST_METHOD_HASH=Overwrite_3d4db6693d
ROOST_METHOD_SIG_HASH=Overwrite_22e8730976

FUNCTION_DEF=func (a *Article) Overwrite(title, description, body string) 

 */
func TestArticleOverwrite(t *testing.T) {
	tests := []struct {
		name           string
		initialArticle Article
		newTitle       string
		newDescription string
		newBody        string
		expectedTitle  string
		expectedDesc   string
		expectedBody   string
	}{
		{
			name: "Overwrite All Fields",
			initialArticle: Article{
				Title:       "Old Title",
				Description: "Old Description",
				Body:        "Old Body",
			},
			newTitle:       "New Title",
			newDescription: "New Description",
			newBody:        "New Body",
			expectedTitle:  "New Title",
			expectedDesc:   "New Description",
			expectedBody:   "New Body",
		},
		{
			name: "Overwrite Title Only",
			initialArticle: Article{
				Title:       "Old Title",
				Description: "Old Description",
				Body:        "Old Body",
			},
			newTitle:       "New Title",
			newDescription: "",
			newBody:        "",
			expectedTitle:  "New Title",
			expectedDesc:   "Old Description",
			expectedBody:   "Old Body",
		},
		{
			name: "Overwrite with Empty Strings",
			initialArticle: Article{
				Title:       "Old Title",
				Description: "Old Description",
				Body:        "Old Body",
			},
			newTitle:       "",
			newDescription: "",
			newBody:        "",
			expectedTitle:  "Old Title",
			expectedDesc:   "Old Description",
			expectedBody:   "Old Body",
		},
		{
			name: "Overwrite Description Only",
			initialArticle: Article{
				Title:       "Old Title",
				Description: "Old Description",
				Body:        "Old Body",
			},
			newTitle:       "",
			newDescription: "New Description",
			newBody:        "",
			expectedTitle:  "Old Title",
			expectedDesc:   "New Description",
			expectedBody:   "Old Body",
		},
		{
			name: "Overwrite Body Only",
			initialArticle: Article{
				Title:       "Old Title",
				Description: "Old Description",
				Body:        "Old Body",
			},
			newTitle:       "",
			newDescription: "",
			newBody:        "New Body",
			expectedTitle:  "Old Title",
			expectedDesc:   "Old Description",
			expectedBody:   "New Body",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			article := tt.initialArticle
			article.Overwrite(tt.newTitle, tt.newDescription, tt.newBody)

			if article.Title != tt.expectedTitle {
				t.Errorf("Test '%s' failed: expected Title '%s', got '%s'", tt.name, tt.expectedTitle, article.Title)
			} else {
				t.Logf("Test '%s': Title overwrite successful", tt.name)
			}

			if article.Description != tt.expectedDesc {
				t.Errorf("Test '%s' failed: expected Description '%s', got '%s'", tt.name, tt.expectedDesc, article.Description)
			} else {
				t.Logf("Test '%s': Description overwrite successful", tt.name)
			}

			if article.Body != tt.expectedBody {
				t.Errorf("Test '%s' failed: expected Body '%s', got '%s'", tt.name, tt.expectedBody, article.Body)
			} else {
				t.Logf("Test '%s': Body overwrite successful", tt.name)
			}
		})
	}
}

