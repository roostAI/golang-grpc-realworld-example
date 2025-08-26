package store

import (
	testing "testing"
	gorm "github.com/jinzhu/gorm"
	assert "github.com/stretchr/testify/assert"
	errors "errors"
	gosqlmock "github.com/DATA-DOG/go-sqlmock"
	model "github.com/raahii/golang-grpc-realworld-example/model"
	debug "runtime/debug"
)





type Comment struct {
	ID   uint
	Body string
}
type MockArticleStore struct {
	db *gorm.DB
}


/*
ROOST_METHOD_HASH=NewArticleStore_85784abca5
ROOST_METHOD_SIG_HASH=NewArticleStore_436ae9c986

FUNCTION_DEF=func NewArticleStore(db *gorm.DB) *ArticleStore // NewArticleStore returns a new ArticleStore


*/
func TestNewArticleStore(t *testing.T) {

	testCases := []struct {
		name     string
		db       *gorm.DB
		expected *ArticleStore
	}{
		{
			name:     "Successful creation of a new ArticleStore",
			db:       &gorm.DB{},
			expected: &ArticleStore{db: &gorm.DB{}},
		},
		{
			name:     "Creation of a new ArticleStore with a nil gorm.DB instance",
			db:       nil,
			expected: &ArticleStore{db: nil},
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

			result := NewArticleStore(tc.db)

			assert.NotNil(t, result, "The returned ArticleStore should not be nil")
			assert.Equal(t, tc.expected.db, result.db, "The db field of the returned ArticleStore should be the same as the provided gorm.DB instance")

			if result.db == tc.expected.db {
				t.Logf("Success: The db field of the returned ArticleStore is the same as the provided gorm.DB instance")
			} else {
				t.Logf("Failure: The db field of the returned ArticleStore is not the same as the provided gorm.DB instance")
			}
		})
	}
}


/*
ROOST_METHOD_HASH=ArticleStore.Create_1273475ade
ROOST_METHOD_SIG_HASH=ArticleStore.Create_a27282cad5

FUNCTION_DEF=func (s *ArticleStore) Create(m *model.Article) error // Create creates an article


*/
func TestArticleStoreCreate(t *testing.T) {

	testCases := []struct {
		name          string
		article       *model.Article
		mockDBFunc    func(mock sqlmock.Sqlmock)
		expectedError error
	}{
		{
			name:    "Successful Article Creation",
			article: &model.Article{Title: "Test Article", Description: "Test Description", Body: "Test Body"},
			mockDBFunc: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO").WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
			expectedError: nil,
		},
		{
			name:    "Article Creation with Invalid Data",
			article: &model.Article{Title: "", Description: "Test Description", Body: "Test Body"},
			mockDBFunc: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO").WillReturnError(errors.New("invalid data"))
				mock.ExpectRollback()
			},
			expectedError: errors.New("invalid data"),
		},
		{
			name:    "Article Creation with Database Error",
			article: &model.Article{Title: "Test Article", Description: "Test Description", Body: "Test Body"},
			mockDBFunc: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO").WillReturnError(errors.New("database error"))
				mock.ExpectRollback()
			},
			expectedError: errors.New("database error"),
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

			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
			}
			defer db.Close()

			tc.mockDBFunc(mock)

			gormDB, err := gorm.Open("postgres", db)
			if err != nil {
				t.Fatalf("an error '%s' was not expected when opening gorm database", err)
			}

			store := &ArticleStore{db: gormDB}

			err = store.Create(tc.article)

			if tc.expectedError != nil {
				if err == nil || err.Error() != tc.expectedError.Error() {
					t.Errorf("expected error '%v', got '%v'", tc.expectedError, err)
				}
			} else if err != nil {
				t.Errorf("expected no error, got error '%v'", err)
			}
		})
	}
}


/*
ROOST_METHOD_HASH=ArticleStore.Delete_8daad9ff19
ROOST_METHOD_SIG_HASH=ArticleStore.Delete_0e09651031

FUNCTION_DEF=func (s *ArticleStore) Delete(m *model.Article) error // Delete deletes an article


*/
func TestArticleStoreDelete(t *testing.T) {

	testCases := []struct {
		name    string
		article *model.Article
		wantErr bool
	}{
		{
			name: "Successful Deletion of an Article",
			article: &model.Article{
				Model: gorm.Model{
					ID: 1,
				},
				Title:       "Test Article",
				Description: "This is a test article",
				Body:        "This is the body of the test article",
			},
			wantErr: false,
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

			db, mock, err := gosqlmock.New()
			if err != nil {
				t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
			}
			defer db.Close()

			gormDB, err := gorm.Open("postgres", db)
			if err != nil {
				t.Fatalf("an error '%s' was not expected when opening gorm database", err)
			}

			mock.ExpectBegin()
			mock.ExpectExec("DELETE FROM \"articles\" WHERE \"articles\".\"id\" = $1").WithArgs(tc.article.ID).WillReturnResult(gosqlmock.NewResult(1, 1))
			mock.ExpectCommit()

			store := &ArticleStore{db: gormDB}

			err = store.Delete(tc.article)

			if (err != nil) != tc.wantErr {
				t.Errorf("ArticleStore.Delete() error = %v, wantErr %v", err, tc.wantErr)
				return
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}


/*
ROOST_METHOD_HASH=ArticleStore.DeleteComment_effbcb38aa
ROOST_METHOD_SIG_HASH=ArticleStore.DeleteComment_d3c99623e4

FUNCTION_DEF=func (s *ArticleStore) DeleteComment(m *model.Comment) error // DeleteComment deletes an comment


*/
func TestArticleStoreDeleteComment(t *testing.T) {

	testCases := []struct {
		name    string
		comment *model.Comment
		wantErr bool
	}{
		{
			name: "Successful Deletion of a Comment",
			comment: &model.Comment{
				Model: gorm.Model{
					ID: 1,
				},
				Body: "This is a test comment",
			},
			wantErr: false,
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

			db, mock, err := gosqlmock.New()
			if err != nil {
				t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
			}
			defer db.Close()

			gormDB, err := gorm.Open("postgres", db)
			if err != nil {
				t.Fatalf("an error '%s' was not expected when opening gorm database", err)
			}

			mock.ExpectExec("DELETE FROM \"comments\" WHERE \"comments\".\"id\" = $1").
				WithArgs(tc.comment.ID).
				WillReturnResult(gosqlmock.NewResult(1, 1))

			store := &ArticleStore{db: gormDB}

			err = store.DeleteComment(tc.comment)

			if tc.wantErr {
				if err == nil {
					t.Errorf("DeleteComment() error = %v, wantErr %v", err, tc.wantErr)
					return
				}
			} else {
				if err != nil {
					t.Errorf("DeleteComment() error = %v, wantErr %v", err, tc.wantErr)
					return
				}
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}


/*
ROOST_METHOD_HASH=ArticleStore.GetCommentByID_7ecaa81f20
ROOST_METHOD_SIG_HASH=ArticleStore.GetCommentByID_f6f8a51973

FUNCTION_DEF=func (s *ArticleStore) GetCommentByID(id uint) (*model.Comment, error) // GetCommentByID finds an comment from id


*/
func TestArticleStoreGetCommentById(t *testing.T) {

	testCases := []struct {
		name          string
		mockDBFunc    func() (*gorm.DB, sqlmock.Sqlmock, error)
		id            uint
		expectedError error
	}{
		{
			name: "Successful retrieval of a comment by ID",
			mockDBFunc: func() (*gorm.DB, sqlmock.Sqlmock, error) {
				db, mock, err := sqlmock.New()
				if err != nil {
					return nil, nil, err
				}
				gormDB, err := gorm.Open("postgres", db)
				if err != nil {
					return nil, nil, err
				}
				mock.ExpectQuery("^SELECT (.+) FROM \"comments\" WHERE \"comments\".\"id\" = \\$1$").
					WithArgs(1).
					WillReturnRows(sqlmock.NewRows([]string{"id", "body"}).
						AddRow(1, "Test comment"))
				return gormDB, mock, nil
			},
			id:            1,
			expectedError: nil,
		},
		{
			name: "Attempt to retrieve a comment with an ID that does not exist",
			mockDBFunc: func() (*gorm.DB, sqlmock.Sqlmock, error) {
				db, mock, err := sqlmock.New()
				if err != nil {
					return nil, nil, err
				}
				gormDB, err := gorm.Open("postgres", db)
				if err != nil {
					return nil, nil, err
				}
				mock.ExpectQuery("^SELECT (.+) FROM \"comments\" WHERE \"comments\".\"id\" = \\$1$").
					WithArgs(2).
					WillReturnError(gorm.ErrRecordNotFound)
				return gormDB, mock, nil
			},
			id:            2,
			expectedError: gorm.ErrRecordNotFound,
		},
		{
			name: "Database error during comment retrieval",
			mockDBFunc: func() (*gorm.DB, sqlmock.Sqlmock, error) {
				db, mock, err := sqlmock.New()
				if err != nil {
					return nil, nil, err
				}
				gormDB, err := gorm.Open("postgres", db)
				if err != nil {
					return nil, nil, err
				}
				mock.ExpectQuery("^SELECT (.+) FROM \"comments\" WHERE \"comments\".\"id\" = \\$1$").
					WithArgs(3).
					WillReturnError(errors.New("database error"))
				return gormDB, mock, nil
			},
			id:            3,
			expectedError: errors.New("database error"),
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

			db, mock, err := tc.mockDBFunc()
			if err != nil {
				t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
			}
			defer db.Close()

			store := &ArticleStore{db: db}
			comment, err := store.GetCommentByID(tc.id)

			if err != tc.expectedError {
				t.Errorf("expected error: %v, got: %v", tc.expectedError, err)
			}

			if tc.expectedError == nil && comment.ID != tc.id {
				t.Errorf("expected comment ID: %v, got: %v", tc.id, comment.ID)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}


/*
ROOST_METHOD_HASH=ArticleStore.GetTags_45f5cdc4bb
ROOST_METHOD_SIG_HASH=ArticleStore.GetTags_fb0aefcdd2

FUNCTION_DEF=func (s *ArticleStore) GetTags() ([ // GetTags creates a article tag
]model.Tag, error) 

*/
func TestArticleStoreGetTags(t *testing.T) {

	testCases := []struct {
		name          string
		mockDBHandler func(mock gosqlmock.Sqlmock)
		expectedTags  []model.Tag
		expectedError error
	}{
		{
			name: "Successful retrieval of tags from the ArticleStore",
			mockDBHandler: func(mock gosqlmock.Sqlmock) {
				rows := gosqlmock.NewRows([]string{"ID", "Name"}).
					AddRow(1, "tag1").
					AddRow(2, "tag2")
				mock.ExpectQuery("^SELECT (.+) FROM \"tags\"$").WillReturnRows(rows)
			},
			expectedTags: []model.Tag{{Model: gorm.Model{ID: 1}, Name: "tag1"}, {Model: gorm.Model{ID: 2}, Name: "tag2"}},
		},
		{
			name: "Database error when retrieving tags",
			mockDBHandler: func(mock gosqlmock.Sqlmock) {
				mock.ExpectQuery("^SELECT (.+) FROM \"tags\"$").WillReturnError(gorm.ErrRecordNotFound)
			},
			expectedError: gorm.ErrRecordNotFound,
		},
		{
			name: "No tags in the ArticleStore",
			mockDBHandler: func(mock gosqlmock.Sqlmock) {
				rows := gosqlmock.NewRows([]string{"ID", "Name"})
				mock.ExpectQuery("^SELECT (.+) FROM \"tags\"$").WillReturnRows(rows)
			},
			expectedTags: []model.Tag{},
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

			db, mock, err := gosqlmock.New()
			if err != nil {
				t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
			}
			defer db.Close()

			tc.mockDBHandler(mock)

			gdb, err := gorm.Open("postgres", db)
			if err != nil {
				t.Fatalf("Failed to open gorm db: %v", err)
			}
			store := &ArticleStore{db: gdb}

			tags, err := store.GetTags()

			assert.Equal(t, tc.expectedTags, tags)
			assert.Equal(t, tc.expectedError, err)
		})
	}
}


/*
ROOST_METHOD_HASH=ArticleStore.GetByID_6fe18728fc
ROOST_METHOD_SIG_HASH=ArticleStore.GetByID_bb488e542f

FUNCTION_DEF=func (s *ArticleStore) GetByID(id uint) (*model.Article, error) // GetByID finds an article from id


*/
func TestArticleStoreGetById(t *testing.T) {

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	gormDB, err := gorm.Open("postgres", db)
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening gorm database", err)
	}

	store := &ArticleStore{
		db: gormDB,
	}

	mockArticle := &model.Article{}

	mock.ExpectQuery("^SELECT (.+) FROM \"articles\" WHERE \"articles\".\"id\" = \\$1").
		WithArgs(mockArticle.ID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "title", "description", "body", "created_at", "updated_at"}).
			AddRow(mockArticle.ID, mockArticle.Title, mockArticle.Description, mockArticle.Body, mockArticle.CreatedAt, mockArticle.UpdatedAt))

	article, err := store.GetByID(mockArticle.ID)

	assert.NoError(t, err)
	assert.Equal(t, mockArticle, article)
}


/*
ROOST_METHOD_HASH=ArticleStore.Update_3cddacb803
ROOST_METHOD_SIG_HASH=ArticleStore.Update_e245edd177

FUNCTION_DEF=func (s *ArticleStore) Update(m *model.Article) error // Update updates an article


*/
func (s *MockArticleStore) Update(m *model.Article) error {
	if m.ID == 0 {
		return errors.New("article does not exist")
	}
	if m.Title == "" || m.Description == "" || m.Body == "" {
		return errors.New("invalid article model")
	}
	return nil
}

func TestArticleStoreUpdate(t *testing.T) {

	sqlDB, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer sqlDB.Close()

	db, err := gorm.Open("postgres", sqlDB)
	if err != nil {
		t.Fatalf("Failed to open DB: %v", err)
	}

	store := &MockArticleStore{
		db: db,
	}

	scenarios := []struct {
		desc     string
		article  *model.Article
		expected error
	}{
		{
			desc: "Successful Article Update",
			article: &model.Article{
				Model: gorm.Model{
					ID: 1,
				},
				Title:       "Test Title",
				Description: "Test Description",
				Body:        "Test Body",
			},
			expected: nil,
		},
		{
			desc: "Article Update with Non-Existent Article",
			article: &model.Article{
				Model: gorm.Model{
					ID: 0,
				},
			},
			expected: errors.New("article does not exist"),
		},
		{
			desc: "Article Update with Invalid Article Model",
			article: &model.Article{
				Model: gorm.Model{
					ID: 1,
				},
				Title: "",
			},
			expected: errors.New("invalid article model"),
		},
	}

	for _, s := range scenarios {
		t.Run(s.desc, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Logf("Panic encountered so failing test. %v", r)
					t.Fail()
				}
			}()

			err := store.Update(s.article)

			if err != nil && err.Error() != s.expected.Error() {
				t.Errorf("Expected error: %v, got: %v", s.expected, err)
			}
			if err == nil && s.expected != nil {
				t.Errorf("Expected error: %v, got: %v", s.expected, err)
			}
		})
	}
}


/*
ROOST_METHOD_HASH=ArticleStore.GetFeedArticles_a37e1934b6
ROOST_METHOD_SIG_HASH=ArticleStore.GetFeedArticles_f5f09c020e

FUNCTION_DEF=func (s *ArticleStore) GetFeedArticles(userIDs [ // GetFeedArticles returns following users' articles
]uint, limit, offset int64) ([]model.Article, error) 

*/
func TestArticleStoreGetFeedArticles(t *testing.T) {

	testCases := []struct {
		name     string
		userIDs  []uint
		limit    int64
		offset   int64
		expected []model.Article
		err      error
	}{
		{
			name:     "Retrieve Feed Articles Successfully",
			userIDs:  []uint{1, 2, 3},
			limit:    10,
			offset:   0,
			expected: []model.Article{{Model: gorm.Model{ID: 1}}, {Model: gorm.Model{ID: 2}}, {Model: gorm.Model{ID: 3}}},
			err:      nil,
		},
		{
			name:     "Retrieve Feed Articles with Offset",
			userIDs:  []uint{1, 2, 3},
			limit:    10,
			offset:   1,
			expected: []model.Article{{Model: gorm.Model{ID: 2}}, {Model: gorm.Model{ID: 3}}},
			err:      nil,
		},
		{
			name:     "Retrieve Feed Articles with Limit",
			userIDs:  []uint{1, 2, 3},
			limit:    2,
			offset:   0,
			expected: []model.Article{{Model: gorm.Model{ID: 1}}, {Model: gorm.Model{ID: 2}}},
			err:      nil,
		},
		{
			name:     "Retrieve Feed Articles with Empty User IDs",
			userIDs:  []uint{},
			limit:    10,
			offset:   0,
			expected: []model.Article{},
			err:      nil,
		},
		{
			name:     "Retrieve Feed Articles with Invalid User IDs",
			userIDs:  []uint{100, 101, 102},
			limit:    10,
			offset:   0,
			expected: []model.Article{},
			err:      nil,
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

			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("Failed to open sqlmock database: %v", err)
			}
			defer db.Close()

			gormDB, err := gorm.Open("postgres", db)
			if err != nil {
				t.Fatalf("Failed to open gorm database: %v", err)
			}

			rows := sqlmock.NewRows([]string{"id"})
			for _, article := range tc.expected {
				rows.AddRow(article.Model.ID)
			}
			mock.ExpectQuery("^SELECT (.+) FROM \"articles\" WHERE (.+)").
				WithArgs(tc.userIDs, tc.limit, tc.offset).
				WillReturnRows(rows)

			store := &ArticleStore{db: gormDB}

			articles, err := store.GetFeedArticles(tc.userIDs, tc.limit, tc.offset)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			assert.Equal(t, tc.expected, articles, "Expected and actual articles do not match")
		})
	}
}


/*
ROOST_METHOD_HASH=ArticleStore.AddFavorite_9460fca478
ROOST_METHOD_SIG_HASH=ArticleStore.AddFavorite_c13a109f91

FUNCTION_DEF=func (s *ArticleStore) AddFavorite(a *model.Article, u *model.User) error // AddFavorite favorite an article


*/
func TestArticleStoreAddFavorite(t *testing.T) {

	testCases := []struct {
		name           string
		article        *model.Article
		user           *model.User
		expectedError  error
		expectedResult int
	}{
		{
			name:           "Successfully adding a favorite article",
			article:        &model.Article{FavoritesCount: 0},
			user:           &model.User{},
			expectedError:  nil,
			expectedResult: 1,
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

			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
			}
			defer db.Close()

			gdb, err := gorm.Open("postgres", db)
			if err != nil {
				t.Fatalf("an error '%s' was not expected when opening gorm database", err)
			}

			mock.ExpectBegin()
			mock.ExpectCommit()

			store := &ArticleStore{db: gdb}

			err = store.AddFavorite(tc.article, tc.user)

			assert.Equal(t, tc.expectedError, err)

			assert.Equal(t, tc.expectedResult, tc.article.FavoritesCount)
		})
	}
}


/*
ROOST_METHOD_HASH=ArticleStore.DeleteFavorite_29c18a04a8
ROOST_METHOD_SIG_HASH=ArticleStore.DeleteFavorite_53deb5e792

FUNCTION_DEF=func (s *ArticleStore) DeleteFavorite(a *model.Article, u *model.User) error // DeleteFavorite unfavorite an article


*/
func TestArticleStoreDeleteFavorite(t *testing.T) {

	testCases := []struct {
		name          string
		mockDBFunc    func() (*gorm.DB, gosqlmock.Sqlmock)
		article       *model.Article
		user          *model.User
		expectedError error
	}{
		{
			name: "Successful Deletion of Favorite Article",
			mockDBFunc: func() (*gorm.DB, gosqlmock.Sqlmock) {
				db, mock, _ := gosqlmock.New()
				gormDB, _ := gorm.Open("postgres", db)
				mock.ExpectBegin()
				mock.ExpectExec("DELETE FROM").WillReturnResult(gosqlmock.NewResult(1, 1))
				mock.ExpectExec("UPDATE").WillReturnResult(gosqlmock.NewResult(1, 1))
				mock.ExpectCommit()
				return gormDB, mock
			},
			article: &model.Article{FavoritesCount: 1},
			user:    &model.User{},
		},
		{
			name: "Deletion of Non-Favorite Article",
			mockDBFunc: func() (*gorm.DB, gosqlmock.Sqlmock) {
				db, mock, _ := gosqlmock.New()
				gormDB, _ := gorm.Open("postgres", db)
				mock.ExpectBegin()
				mock.ExpectExec("DELETE FROM").WillReturnResult(gosqlmock.NewResult(1, 0))
				mock.ExpectRollback()
				return gormDB, mock
			},
			article:       &model.Article{FavoritesCount: 0},
			user:          &model.User{},
			expectedError: errors.New("record not found"),
		},
		{
			name: "Deletion of Favorite Article with Database Error",
			mockDBFunc: func() (*gorm.DB, gosqlmock.Sqlmock) {
				db, mock, _ := gosqlmock.New()
				gormDB, _ := gorm.Open("postgres", db)
				mock.ExpectBegin()
				mock.ExpectExec("DELETE FROM").WillReturnError(errors.New("database error"))
				mock.ExpectRollback()
				return gormDB, mock
			},
			article:       &model.Article{FavoritesCount: 1},
			user:          &model.User{},
			expectedError: errors.New("database error"),
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

			db, _ := tc.mockDBFunc()
			store := &ArticleStore{db: db}

			err := store.DeleteFavorite(tc.article, tc.user)

			if tc.expectedError != nil {
				if err == nil || err.Error() != tc.expectedError.Error() {
					t.Errorf("Expected error %v, but got %v", tc.expectedError, err)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error %v", err)
				}
				if tc.article.FavoritesCount != 0 {
					t.Errorf("Expected FavoritesCount to be decremented, but got %v", tc.article.FavoritesCount)
				}
			}
		})
	}
}

