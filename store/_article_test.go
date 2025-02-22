package store

import (
	fmt "fmt"
	debug "runtime/debug"
	testing "testing"
	go-sqlmock "github.com/DATA-DOG/go-sqlmock"
	gorm "github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	model "github.com/raahii/golang-grpc-realworld-example/model"
	time "time"
	sql "database/sql"
	errors "errors"
)








/*
ROOST_METHOD_HASH=ArticleStore_AddFavorite_9460fca478
ROOST_METHOD_SIG_HASH=ArticleStore_AddFavorite_c13a109f91

FUNCTION_DEF=func (s *ArticleStore) AddFavorite(a *model.Article, u *model.User) error // AddFavorite favorite an article


*/
func TestArticleStoreAddFavorite(t *testing.T) {
	type testCase struct {
		name          string
		article       *model.Article
		user          *model.User
		setupMock     func(mock sqlmock.Sqlmock)
		expectedError bool
	}

	tests := []testCase{
		{
			name: "Successfully Add Favorite",
			article: &model.Article{
				Model: gorm.Model{ID: 1},
				Title: "Test Article",
			},
			user: &model.User{
				Model: gorm.Model{ID: 1},
			},
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO `favorite_articles`").
					WithArgs(1, 1).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectExec("UPDATE `articles`").
					WithArgs(1, 1).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
			expectedError: false,
		},
		{
			name: "Failed Association Update",
			article: &model.Article{
				Model: gorm.Model{ID: 1},
			},
			user: &model.User{
				Model: gorm.Model{ID: 1},
			},
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO `favorite_articles`").
					WillReturnError(fmt.Errorf("association error"))
				mock.ExpectRollback()
			},
			expectedError: true,
		},
		{
			name:          "Nil Article Parameter",
			article:       nil,
			user:          &model.User{Model: gorm.Model{ID: 1}},
			setupMock:     func(mock sqlmock.Sqlmock) {},
			expectedError: true,
		},
		{
			name:          "Nil User Parameter",
			article:       &model.Article{Model: gorm.Model{ID: 1}},
			user:          nil,
			setupMock:     func(mock sqlmock.Sqlmock) {},
			expectedError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Logf("Panic encountered so failing test. %v\n%s", r, string(debug.Stack()))
					t.Fail()
				}
			}()

			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("Failed to create mock DB: %v", err)
			}
			defer db.Close()

			gormDB, err := gorm.Open("_", db)
			if err != nil {
				t.Fatalf("Failed to create GORM DB: %v", err)
			}
			defer gormDB.Close()

			gormDB.LogMode(true)

			tc.setupMock(mock)

			store := &ArticleStore{
				db: gormDB,
			}

			err = store.AddFavorite(tc.article, tc.user)

			if tc.expectedError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tc.expectedError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("Unfulfilled expectations: %v", err)
			}

			t.Logf("Test case '%s' completed successfully", tc.name)
		})
	}
}


/*
ROOST_METHOD_HASH=ArticleStore_Create_1273475ade
ROOST_METHOD_SIG_HASH=ArticleStore_Create_a27282cad5

FUNCTION_DEF=func (s *ArticleStore) Create(m *model.Article) error // Create creates an article


*/
func TestArticleStoreCreate(t *testing.T) {

	type testCase struct {
		name    string
		article *model.Article
		setupDB func(mock sqlmock.Sqlmock)
		wantErr bool
	}

	validUserID := uint(1)
	now := time.Now()

	tests := []testCase{
		{
			name: "Successfully Create Article",
			article: &model.Article{
				Title:       "Test Article",
				Description: "Test Description",
				Body:        "Test Body Content",
				UserID:      validUserID,
			},
			setupDB: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO `articles`").
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
			wantErr: false,
		},
		{
			name: "Missing Required Fields",
			article: &model.Article{

				Description: "Test Description",
				Body:        "Test Body",
				UserID:      validUserID,
			},
			setupDB: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO `articles`").
					WillReturnError(fmt.Errorf("title cannot be null"))
				mock.ExpectRollback()
			},
			wantErr: true,
		},
		{
			name: "Create Article with Tags",
			article: &model.Article{
				Title:       "Test Article with Tags",
				Description: "Test Description",
				Body:        "Test Body Content",
				UserID:      validUserID,
				Tags: []model.Tag{
					{Model: gorm.Model{ID: 1}, Name: "tag1"},
					{Model: gorm.Model{ID: 2}, Name: "tag2"},
				},
			},
			setupDB: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO `articles`").
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectExec("INSERT INTO `article_tags`").
					WillReturnResult(sqlmock.NewResult(1, 2))
				mock.ExpectCommit()
			},
			wantErr: false,
		},
		{
			name: "Database Connection Error",
			article: &model.Article{
				Title:       "Test Article",
				Description: "Test Description",
				Body:        "Test Body",
				UserID:      validUserID,
			},
			setupDB: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO `articles`").
					WillReturnError(fmt.Errorf("database connection lost"))
				mock.ExpectRollback()
			},
			wantErr: true,
		},
		{
			name: "Maximum Field Lengths",
			article: &model.Article{
				Title:       string(make([]byte, 255)),
				Description: string(make([]byte, 1000)),
				Body:        string(make([]byte, 5000)),
				UserID:      validUserID,
			},
			setupDB: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO `articles`").
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			defer func() {
				if r := recover(); r != nil {
					t.Logf("Panic encountered so failing test. %v\n%s", r, string(debug.Stack()))
					t.Fail()
				}
			}()

			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("Failed to create mock DB: %v", err)
			}
			defer db.Close()

			gormDB, err := gorm.Open("mysql", db)
			if err != nil {
				t.Fatalf("Failed to open gorm DB: %v", err)
			}
			defer gormDB.Close()

			tt.setupDB(mock)

			store := &ArticleStore{
				db: gormDB,
			}

			tt.article.CreatedAt = now
			tt.article.UpdatedAt = now

			err = store.Create(tt.article)

			if (err != nil) != tt.wantErr {
				t.Errorf("ArticleStore.Create() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("Unfulfilled expectations: %v", err)
			}

			if err != nil {
				t.Logf("Test case '%s' failed with error: %v", tt.name, err)
			} else {
				t.Logf("Test case '%s' succeeded", tt.name)
			}
		})
	}
}


/*
ROOST_METHOD_HASH=ArticleStore_CreateComment_b16d4a71d4
ROOST_METHOD_SIG_HASH=ArticleStore_CreateComment_7475736b06

FUNCTION_DEF=func (s *ArticleStore) CreateComment(m *model.Comment) error // CreateComment creates a comment of the article


*/
func TestArticleStoreCreateComment(t *testing.T) {

	db, mock, err := go-sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer db.Close()

	gormDB, err := gorm.Open("mysql", db)
	if err != nil {
		t.Fatalf("Failed to open gorm connection: %v", err)
	}
	defer gormDB.Close()

	store := &ArticleStore{db: gormDB}

	tests := []struct {
		name    string
		comment *model.Comment
		mock    func()
		wantErr bool
	}{
		{
			name: "Success - Valid Comment Creation",
			comment: &model.Comment{
				Body:      "Test comment",
				UserID:    1,
				ArticleID: 1,
			},
			mock: func() {
				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO `comments`").
					WithArgs(go-sqlmock.AnyArg(), go-sqlmock.AnyArg(), go-sqlmock.AnyArg(), "Test comment", uint(1), uint(1)).
					WillReturnResult(go-sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
			wantErr: false,
		},
		{
			name: "Failure - Missing Required Fields",
			comment: &model.Comment{
				Body:      "",
				UserID:    1,
				ArticleID: 1,
			},
			mock: func() {
				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO `comments`").
					WillReturnError(sql.ErrNoRows)
				mock.ExpectRollback()
			},
			wantErr: true,
		},
		{
			name: "Failure - Non-existent Article",
			comment: &model.Comment{
				Body:      "Test comment",
				UserID:    1,
				ArticleID: 999,
			},
			mock: func() {
				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO `comments`").
					WillReturnError(gorm.ErrRecordNotFound)
				mock.ExpectRollback()
			},
			wantErr: true,
		},
		{
			name: "Failure - Database Connection Error",
			comment: &model.Comment{
				Body:      "Test comment",
				UserID:    1,
				ArticleID: 1,
			},
			mock: func() {
				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO `comments`").
					WillReturnError(sql.ErrConnDone)
				mock.ExpectRollback()
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			defer func() {
				if r := recover(); r != nil {
					t.Logf("Panic encountered so failing test. %v\n%s", r, string(debug.Stack()))
					t.Fail()
				}
			}()

			t.Log("Starting test case:", tt.name)

			tt.mock()

			err := store.CreateComment(tt.comment)

			if (err != nil) != tt.wantErr {
				t.Errorf("ArticleStore.CreateComment() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("Unfulfilled expectations: %v", err)
			}

			if err != nil {
				t.Logf("Test case completed with expected error: %v", err)
			} else {
				t.Log("Test case completed successfully")
			}
		})
	}
}


/*
ROOST_METHOD_HASH=ArticleStore_Delete_8daad9ff19
ROOST_METHOD_SIG_HASH=ArticleStore_Delete_0e09651031

FUNCTION_DEF=func (s *ArticleStore) Delete(m *model.Article) error // Delete deletes an article


*/
func TestArticleStoreDelete(t *testing.T) {

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock DB: %v", err)
	}
	defer db.Close()

	gormDB, err := gorm.Open("mysql", db)
	if err != nil {
		t.Fatalf("Failed to open GORM DB: %v", err)
	}
	defer gormDB.Close()

	store := &ArticleStore{db: gormDB}

	tests := []struct {
		name    string
		article *model.Article
		mock    func()
		wantErr bool
	}{
		{
			name: "Successfully delete existing article",
			article: &model.Article{
				Model: gorm.Model{ID: 1},
				Title: "Test Article",
			},
			mock: func() {
				mock.ExpectBegin()
				mock.ExpectExec("DELETE FROM `articles`").
					WithArgs(sqlmock.AnyArg()).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
			wantErr: false,
		},
		{
			name: "Delete non-existent article",
			article: &model.Article{
				Model: gorm.Model{ID: 999},
			},
			mock: func() {
				mock.ExpectBegin()
				mock.ExpectExec("DELETE FROM `articles`").
					WithArgs(sqlmock.AnyArg()).
					WillReturnError(gorm.ErrRecordNotFound)
				mock.ExpectRollback()
			},
			wantErr: true,
		},
		{
			name:    "Delete with nil article",
			article: nil,
			mock: func() {

			},
			wantErr: true,
		},
		{
			name: "Delete with DB connection error",
			article: &model.Article{
				Model: gorm.Model{ID: 1},
			},
			mock: func() {
				mock.ExpectBegin()
				mock.ExpectExec("DELETE FROM `articles`").
					WithArgs(sqlmock.AnyArg()).
					WillReturnError(sql.ErrConnDone)
				mock.ExpectRollback()
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			defer func() {
				if r := recover(); r != nil {
					t.Logf("Panic encountered so failing test. %v\n%s", r, string(debug.Stack()))
					t.Fail()
				}
			}()

			t.Logf("Running test case: %s", tt.name)

			tt.mock()

			err := store.Delete(tt.article)

			if (err != nil) != tt.wantErr {
				t.Errorf("ArticleStore.Delete() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("Unfulfilled expectations: %v", err)
			}

			t.Logf("Test case completed: %s", tt.name)
		})
	}
}


/*
ROOST_METHOD_HASH=ArticleStore_DeleteComment_effbcb38aa
ROOST_METHOD_SIG_HASH=ArticleStore_DeleteComment_d3c99623e4

FUNCTION_DEF=func (s *ArticleStore) DeleteComment(m *model.Comment) error // DeleteComment deletes an comment


*/
func TestArticleStoreDeleteComment(t *testing.T) {

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer db.Close()

	gormDB, err := gorm.Open("mysql", db)
	if err != nil {
		t.Fatalf("Failed to open gorm connection: %v", err)
	}
	defer gormDB.Close()

	store := &ArticleStore{
		db: gormDB,
	}

	tests := []struct {
		name    string
		comment *model.Comment
		mockSQL func()
		wantErr bool
	}{
		{
			name: "Successfully delete existing comment",
			comment: &model.Comment{
				Model: gorm.Model{
					ID:        1,
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				},
				Body:      "Test comment",
				UserID:    1,
				ArticleID: 1,
			},
			mockSQL: func() {
				mock.ExpectBegin()
				mock.ExpectExec("DELETE FROM `comments`").
					WithArgs(sqlmock.AnyArg()).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
			wantErr: false,
		},
		{
			name: "Delete non-existent comment",
			comment: &model.Comment{
				Model: gorm.Model{
					ID: 999,
				},
			},
			mockSQL: func() {
				mock.ExpectBegin()
				mock.ExpectExec("DELETE FROM `comments`").
					WithArgs(sqlmock.AnyArg()).
					WillReturnResult(sqlmock.NewResult(0, 0))
				mock.ExpectCommit()
			},
			wantErr: false,
		},
		{
			name:    "Delete with nil comment",
			comment: nil,
			mockSQL: func() {

			},
			wantErr: true,
		},
		{
			name: "Database error during deletion",
			comment: &model.Comment{
				Model: gorm.Model{
					ID: 1,
				},
			},
			mockSQL: func() {
				mock.ExpectBegin()
				mock.ExpectExec("DELETE FROM `comments`").
					WithArgs(sqlmock.AnyArg()).
					WillReturnError(sql.ErrConnDone)
				mock.ExpectRollback()
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			defer func() {
				if r := recover(); r != nil {
					t.Logf("Panic encountered so failing test. %v\n%s", r, string(debug.Stack()))
					t.Fail()
				}
			}()

			t.Log("Starting test case:", tt.name)

			tt.mockSQL()

			err := store.DeleteComment(tt.comment)

			if (err != nil) != tt.wantErr {
				t.Errorf("DeleteComment() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("Unfulfilled expectations: %v", err)
			}

			t.Log("Test case completed successfully")
		})
	}
}


/*
ROOST_METHOD_HASH=ArticleStore_DeleteFavorite_29c18a04a8
ROOST_METHOD_SIG_HASH=ArticleStore_DeleteFavorite_53deb5e792

FUNCTION_DEF=func (s *ArticleStore) DeleteFavorite(a *model.Article, u *model.User) error // DeleteFavorite unfavorite an article


*/
func TestArticleStoreDeleteFavorite(t *testing.T) {

	type testCase struct {
		name          string
		article       *model.Article
		user          *model.User
		setupMock     func(mock sqlmock.Sqlmock)
		expectedError error
	}

	tests := []testCase{
		{
			name: "Successful unfavorite",
			article: &model.Article{
				Model:          gorm.Model{ID: 1},
				FavoritesCount: 2,
			},
			user: &model.User{
				Model: gorm.Model{ID: 1},
			},
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec("DELETE FROM `favorite_articles`").
					WithArgs(1, 1).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectExec("UPDATE `articles`").
					WithArgs(1, 1).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
			expectedError: nil,
		},
		{
			name: "Failed association removal",
			article: &model.Article{
				Model:          gorm.Model{ID: 1},
				FavoritesCount: 2,
			},
			user: &model.User{
				Model: gorm.Model{ID: 1},
			},
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec("DELETE FROM `favorite_articles`").
					WithArgs(1, 1).
					WillReturnError(errors.New("association removal failed"))
				mock.ExpectRollback()
			},
			expectedError: errors.New("association removal failed"),
		},
		{
			name:    "Nil article parameter",
			article: nil,
			user: &model.User{
				Model: gorm.Model{ID: 1},
			},
			setupMock:     func(mock sqlmock.Sqlmock) {},
			expectedError: errors.New("article cannot be nil"),
		},
		{
			name: "Nil user parameter",
			article: &model.Article{
				Model:          gorm.Model{ID: 1},
				FavoritesCount: 2,
			},
			user:          nil,
			setupMock:     func(mock sqlmock.Sqlmock) {},
			expectedError: errors.New("user cannot be nil"),
		},
		{
			name: "Failed favorites count update",
			article: &model.Article{
				Model:          gorm.Model{ID: 1},
				FavoritesCount: 2,
			},
			user: &model.User{
				Model: gorm.Model{ID: 1},
			},
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec("DELETE FROM `favorite_articles`").
					WithArgs(1, 1).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectExec("UPDATE `articles`").
					WithArgs(1, 1).
					WillReturnError(errors.New("update failed"))
				mock.ExpectRollback()
			},
			expectedError: errors.New("update failed"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {

			defer func() {
				if r := recover(); r != nil {
					t.Logf("Panic encountered so failing test. %v\n%s", r, string(debug.Stack()))
					t.Fail()
				}
			}()

			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("Failed to create mock DB: %v", err)
			}
			defer db.Close()

			gormDB, err := gorm.Open("mysql", db)
			if err != nil {
				t.Fatalf("Failed to create GORM DB: %v", err)
			}
			defer gormDB.Close()

			tc.setupMock(mock)

			store := &ArticleStore{
				db: gormDB,
			}

			err = store.DeleteFavorite(tc.article, tc.user)

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("Unfulfilled expectations: %v", err)
			}

			if tc.expectedError == nil && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			} else if tc.expectedError != nil && err == nil {
				t.Errorf("Expected error %v but got none", tc.expectedError)
			} else if tc.expectedError != nil && err != nil && tc.expectedError.Error() != err.Error() {
				t.Errorf("Expected error %v but got %v", tc.expectedError, err)
			}

			if tc.expectedError == nil && tc.article != nil {
				expectedCount := tc.article.FavoritesCount - 1
				if tc.article.FavoritesCount != expectedCount {
					t.Errorf("Expected favorites count to be %d but got %d", expectedCount, tc.article.FavoritesCount)
				}
			}
		})
	}
}


/*
ROOST_METHOD_HASH=ArticleStore_GetArticles_101b7250e8
ROOST_METHOD_SIG_HASH=ArticleStore_GetArticles_91bc0a6760

FUNCTION_DEF=func (s *ArticleStore) GetArticles(tagName, username string, favoritedBy *model.User, limit, offset int64) ([ // GetArticles get global articles
]model.Article, error) 

*/
func TestArticleStoreGetArticles(t *testing.T) {

	db, mock, err := go-sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock DB: %v", err)
	}
	defer db.Close()

	gormDB, err := gorm.Open("mysql", db)
	if err != nil {
		t.Fatalf("Failed to open GORM DB: %v", err)
	}
	defer gormDB.Close()

	store := &ArticleStore{db: gormDB}

	testUser := &model.User{
		Model:    gorm.Model{ID: 1},
		Username: "testuser",
		Email:    "test@example.com",
	}

	tests := []struct {
		name        string
		tagName     string
		username    string
		favoritedBy *model.User
		limit       int64
		offset      int64
		mockSetup   func(go-sqlmock.Sqlmock)
		wantErr     bool
		wantLen     int
	}{
		{
			name: "Success - No Filters",
			mockSetup: func(mock go-sqlmock.Sqlmock) {
				rows := go-sqlmock.NewRows([]string{"id", "title", "description", "body", "user_id", "created_at", "updated_at"}).
					AddRow(1, "Test Article", "Test Description", "Test Body", 1, time.Now(), time.Now())

				mock.ExpectQuery("^SELECT (.+) FROM `articles`").
					WillReturnRows(rows)
			},
			wantErr: false,
			wantLen: 1,
		},
		{
			name:     "Filter by Username",
			username: "testuser",
			mockSetup: func(mock go-sqlmock.Sqlmock) {
				rows := go-sqlmock.NewRows([]string{"id", "title", "description", "body", "user_id", "created_at", "updated_at"}).
					AddRow(1, "Test Article", "Test Description", "Test Body", 1, time.Now(), time.Now())

				mock.ExpectQuery("^SELECT (.+) FROM `articles` join users").
					WithArgs("testuser").
					WillReturnRows(rows)
			},
			wantErr: false,
			wantLen: 1,
		},
		{
			name:    "Filter by Tag",
			tagName: "testtag",
			mockSetup: func(mock go-sqlmock.Sqlmock) {
				rows := go-sqlmock.NewRows([]string{"id", "title", "description", "body", "user_id", "created_at", "updated_at"}).
					AddRow(1, "Test Article", "Test Description", "Test Body", 1, time.Now(), time.Now())

				mock.ExpectQuery("^SELECT (.+) FROM `articles` join article_tags").
					WithArgs("testtag").
					WillReturnRows(rows)
			},
			wantErr: false,
			wantLen: 1,
		},
		{
			name:        "Filter by Favorited",
			favoritedBy: testUser,
			mockSetup: func(mock go-sqlmock.Sqlmock) {
				favRows := go-sqlmock.NewRows([]string{"article_id"}).AddRow(1)
				mock.ExpectQuery("^SELECT article_id FROM `favorite_articles`").
					WithArgs(testUser.ID).
					WillReturnRows(favRows)

				articleRows := go-sqlmock.NewRows([]string{"id", "title", "description", "body", "user_id", "created_at", "updated_at"}).
					AddRow(1, "Test Article", "Test Description", "Test Body", 1, time.Now(), time.Now())
				mock.ExpectQuery("^SELECT (.+) FROM `articles`").
					WillReturnRows(articleRows)
			},
			wantErr: false,
			wantLen: 1,
		},
		{
			name: "Database Error",
			mockSetup: func(mock go-sqlmock.Sqlmock) {
				mock.ExpectQuery("^SELECT (.+) FROM `articles`").
					WillReturnError(sql.ErrConnDone)
			},
			wantErr: true,
			wantLen: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			defer func() {
				if r := recover(); r != nil {
					t.Logf("Panic encountered so failing test. %v\n%s", r, string(debug.Stack()))
					t.Fail()
				}
			}()

			tt.mockSetup(mock)

			articles, err := store.GetArticles(tt.tagName, tt.username, tt.favoritedBy, tt.limit, tt.offset)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetArticles() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if len(articles) != tt.wantLen {
				t.Errorf("GetArticles() got %v articles, want %v", len(articles), tt.wantLen)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("Unfulfilled expectations: %v", err)
			}

			t.Logf("Test '%s' completed successfully", tt.name)
		})
	}
}


/*
ROOST_METHOD_HASH=ArticleStore_GetByID_6fe18728fc
ROOST_METHOD_SIG_HASH=ArticleStore_GetByID_bb488e542f

FUNCTION_DEF=func (s *ArticleStore) GetByID(id uint) (*model.Article, error) // GetByID finds an article from id


*/
func TestArticleStoreGetById(t *testing.T) {

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock DB: %v", err)
	}
	defer db.Close()

	gormDB, err := gorm.Open("mysql", db)
	if err != nil {
		t.Fatalf("Failed to open GORM connection: %v", err)
	}
	defer gormDB.Close()

	store := &ArticleStore{db: gormDB}

	tests := []struct {
		name          string
		id            uint
		mockSetup     func(sqlmock.Sqlmock)
		expectedError error
		expectArticle bool
	}{
		{
			name: "Successfully retrieve article",
			id:   1,
			mockSetup: func(mock sqlmock.Sqlmock) {

				rows := sqlmock.NewRows([]string{"id", "created_at", "updated_at", "deleted_at", "title", "description", "body", "user_id", "favorites_count"}).
					AddRow(1, time.Now(), time.Now(), nil, "Test Title", "Test Description", "Test Body", 1, 0)
				mock.ExpectQuery("SELECT").WillReturnRows(rows)

				tagRows := sqlmock.NewRows([]string{"id", "name"}).
					AddRow(1, "tag1")
				mock.ExpectQuery("SELECT").WillReturnRows(tagRows)

				authorRows := sqlmock.NewRows([]string{"id", "username", "email", "bio", "image"}).
					AddRow(1, "testuser", "test@example.com", "bio", "image.jpg")
				mock.ExpectQuery("SELECT").WillReturnRows(authorRows)
			},
			expectedError: nil,
			expectArticle: true,
		},
		{
			name: "Article not found",
			id:   999,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT").WillReturnError(gorm.ErrRecordNotFound)
			},
			expectedError: gorm.ErrRecordNotFound,
			expectArticle: false,
		},
		{
			name: "Database error",
			id:   1,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT").WillReturnError(sql.ErrConnDone)
			},
			expectedError: sql.ErrConnDone,
			expectArticle: false,
		},
		{
			name: "Zero ID handling",
			id:   0,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT").WillReturnError(gorm.ErrRecordNotFound)
			},
			expectedError: gorm.ErrRecordNotFound,
			expectArticle: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			defer func() {
				if r := recover(); r != nil {
					t.Logf("Panic encountered so failing test. %v\n%s", r, string(debug.Stack()))
					t.Fail()
				}
			}()

			tt.mockSetup(mock)

			article, err := store.GetByID(tt.id)

			if err != tt.expectedError {
				t.Errorf("Expected error %v, got %v", tt.expectedError, err)
			}

			if tt.expectArticle && article == nil {
				t.Error("Expected article to be returned, got nil")
			}

			if !tt.expectArticle && article != nil {
				t.Error("Expected nil article, got non-nil")
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("Unfulfilled expectations: %s", err)
			}

			t.Logf("Test '%s' completed successfully", tt.name)
		})
	}
}


/*
ROOST_METHOD_HASH=ArticleStore_GetCommentByID_7ecaa81f20
ROOST_METHOD_SIG_HASH=ArticleStore_GetCommentByID_f6f8a51973

FUNCTION_DEF=func (s *ArticleStore) GetCommentByID(id uint) (*model.Comment, error) // GetCommentByID finds an comment from id


*/
func TestArticleStoreGetCommentById(t *testing.T) {

	type testCase struct {
		name          string
		commentID     uint
		mockSetup     func(mock sqlmock.Sqlmock)
		expectedError error
		expectComment bool
	}

	tests := []testCase{
		{
			name:      "Successfully retrieve existing comment",
			commentID: 1,
			mockSetup: func(mock sqlmock.Sqlmock) {
				columns := []string{"id", "created_at", "updated_at", "deleted_at", "body", "user_id", "article_id"}
				mock.ExpectQuery("^SELECT (.+) FROM `comments` WHERE `comments`.`deleted_at` IS NULL AND ((.+))").
					WithArgs(1).
					WillReturnRows(sqlmock.NewRows(columns).
						AddRow(1, "2023-01-01", "2023-01-01", nil, "Test comment", 1, 1))
			},
			expectedError: nil,
			expectComment: true,
		},
		{
			name:      "Comment not found",
			commentID: 999,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("^SELECT (.+) FROM `comments`").
					WithArgs(999).
					WillReturnError(gorm.ErrRecordNotFound)
			},
			expectedError: gorm.ErrRecordNotFound,
			expectComment: false,
		},
		{
			name:      "Database error",
			commentID: 1,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("^SELECT (.+) FROM `comments`").
					WithArgs(1).
					WillReturnError(sql.ErrConnDone)
			},
			expectedError: sql.ErrConnDone,
			expectComment: false,
		},
		{
			name:      "Zero ID handling",
			commentID: 0,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("^SELECT (.+) FROM `comments`").
					WithArgs(0).
					WillReturnError(gorm.ErrRecordNotFound)
			},
			expectedError: gorm.ErrRecordNotFound,
			expectComment: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {

			defer func() {
				if r := recover(); r != nil {
					t.Logf("Panic encountered so failing test. %v\n%s", r, string(debug.Stack()))
					t.Fail()
				}
			}()

			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("Failed to create mock DB: %v", err)
			}
			defer db.Close()

			gormDB, err := gorm.Open("mysql", db)
			if err != nil {
				t.Fatalf("Failed to open GORM DB: %v", err)
			}
			defer gormDB.Close()

			tc.mockSetup(mock)

			store := &ArticleStore{
				db: gormDB,
			}

			comment, err := store.GetCommentByID(tc.commentID)

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("Unfulfilled expectations: %s", err)
			}

			if tc.expectedError != nil {
				if err == nil {
					t.Errorf("Expected error %v but got nil", tc.expectedError)
				} else if !errors.Is(err, tc.expectedError) {
					t.Errorf("Expected error %v but got %v", tc.expectedError, err)
				}
			} else if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if tc.expectComment {
				if comment == nil {
					t.Error("Expected comment but got nil")
				}
			} else {
				if comment != nil {
					t.Error("Expected nil comment but got a comment")
				}
			}

			t.Logf("Test case '%s' completed successfully", tc.name)
		})
	}
}


/*
ROOST_METHOD_HASH=ArticleStore_GetComments_d7c78dda64
ROOST_METHOD_SIG_HASH=ArticleStore_GetComments_af08ddd59e

FUNCTION_DEF=func (s *ArticleStore) GetComments(m *model.Article) ([ // GetComments gets coments of the article
]model.Comment, error) 

*/
func TestArticleStoreGetComments(t *testing.T) {

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock DB: %v", err)
	}
	defer db.Close()

	gormDB, err := gorm.Open("mysql", db)
	if err != nil {
		t.Fatalf("Failed to open GORM DB: %v", err)
	}
	defer gormDB.Close()

	store := &ArticleStore{
		db: gormDB,
	}

	tests := []struct {
		name         string
		article      *model.Article
		mockSetup    func(sqlmock.Sqlmock)
		wantComments []model.Comment
		wantErr      bool
	}{
		{
			name: "Successful retrieval of comments",
			article: &model.Article{
				Model: gorm.Model{ID: 1},
			},
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{
					"id", "created_at", "updated_at", "deleted_at",
					"body", "user_id", "article_id",
					"author.id", "author.username", "author.email",
				}).AddRow(
					1, time.Now(), time.Now(), nil,
					"Test comment", 1, 1,
					1, "testuser", "test@example.com",
				)

				mock.ExpectQuery("SELECT").WillReturnRows(rows)
			},
			wantComments: []model.Comment{
				{
					Model: gorm.Model{ID: 1},
					Body:  "Test comment",
					Author: model.User{
						Model:    gorm.Model{ID: 1},
						Username: "testuser",
						Email:    "test@example.com",
					},
					UserID:    1,
					ArticleID: 1,
				},
			},
			wantErr: false,
		},
		{
			name: "No comments found",
			article: &model.Article{
				Model: gorm.Model{ID: 2},
			},
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{
					"id", "created_at", "updated_at", "deleted_at",
					"body", "user_id", "article_id",
					"author.id", "author.username", "author.email",
				})
				mock.ExpectQuery("SELECT").WillReturnRows(rows)
			},
			wantComments: []model.Comment{},
			wantErr:      false,
		},
		{
			name: "Database error",
			article: &model.Article{
				Model: gorm.Model{ID: 3},
			},
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT").WillReturnError(sql.ErrConnDone)
			},
			wantComments: nil,
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			defer func() {
				if r := recover(); r != nil {
					t.Logf("Panic encountered so failing test. %v\n%s", r, string(debug.Stack()))
					t.Fail()
				}
			}()

			t.Logf("Running test case: %s", tt.name)

			tt.mockSetup(mock)

			gotComments, err := store.GetComments(tt.article)

			if (err != nil) != tt.wantErr {
				t.Errorf("ArticleStore.GetComments() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if len(gotComments) != len(tt.wantComments) {
					t.Errorf("ArticleStore.GetComments() returned %d comments, want %d",
						len(gotComments), len(tt.wantComments))
					return
				}

				if len(tt.wantComments) > 0 {
					for i, want := range tt.wantComments {
						got := gotComments[i]
						if got.ID != want.ID || got.Body != want.Body {
							t.Errorf("Comment mismatch at index %d", i)
						}
						if got.Author.Username != want.Author.Username {
							t.Errorf("Author mismatch at index %d", i)
						}
					}
				}
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("Unfulfilled expectations: %v", err)
			}

			t.Logf("Test case completed successfully: %s", tt.name)
		})
	}
}


/*
ROOST_METHOD_HASH=ArticleStore_GetFeedArticles_a37e1934b6
ROOST_METHOD_SIG_HASH=ArticleStore_GetFeedArticles_f5f09c020e

FUNCTION_DEF=func (s *ArticleStore) GetFeedArticles(userIDs [ // GetFeedArticles returns following users' articles
]uint, limit, offset int64) ([]model.Article, error) 

*/
func TestArticleStoreGetFeedArticles(t *testing.T) {

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock DB: %v", err)
	}
	defer db.Close()

	gormDB, err := gorm.Open("mysql", db)
	if err != nil {
		t.Fatalf("Failed to open GORM DB: %v", err)
	}
	defer gormDB.Close()

	store := &ArticleStore{db: gormDB}

	tests := []struct {
		name      string
		userIDs   []uint
		limit     int64
		offset    int64
		mockSetup func(sqlmock.Sqlmock)
		wantLen   int
		wantErr   bool
	}{
		{
			name:    "Success - Multiple Users Articles",
			userIDs: []uint{1, 2, 3},
			limit:   10,
			offset:  0,
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{
					"id", "created_at", "updated_at", "deleted_at",
					"title", "description", "body", "user_id", "favorites_count",
				}).
					AddRow(1, time.Now(), time.Now(), nil, "Title1", "Desc1", "Body1", 1, 0).
					AddRow(2, time.Now(), time.Now(), nil, "Title2", "Desc2", "Body2", 2, 0)

				mock.ExpectQuery("SELECT").WillReturnRows(rows)

				authorRows := sqlmock.NewRows([]string{
					"id", "created_at", "updated_at", "deleted_at",
					"username", "email", "password", "bio", "image",
				}).
					AddRow(1, time.Now(), time.Now(), nil, "user1", "user1@test.com", "pass", "bio", "img").
					AddRow(2, time.Now(), time.Now(), nil, "user2", "user2@test.com", "pass", "bio", "img")

				mock.ExpectQuery("SELECT").WillReturnRows(authorRows)
			},
			wantLen: 2,
			wantErr: false,
		},
		{
			name:    "Empty Result Set",
			userIDs: []uint{4, 5},
			limit:   10,
			offset:  0,
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{
					"id", "created_at", "updated_at", "deleted_at",
					"title", "description", "body", "user_id", "favorites_count",
				})
				mock.ExpectQuery("SELECT").WillReturnRows(rows)
			},
			wantLen: 0,
			wantErr: false,
		},
		{
			name:    "Database Error",
			userIDs: []uint{1},
			limit:   10,
			offset:  0,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT").WillReturnError(sql.ErrConnDone)
			},
			wantLen: 0,
			wantErr: true,
		},
		{
			name:    "Empty UserIDs Array",
			userIDs: []uint{},
			limit:   10,
			offset:  0,
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{
					"id", "created_at", "updated_at", "deleted_at",
					"title", "description", "body", "user_id", "favorites_count",
				})
				mock.ExpectQuery("SELECT").WillReturnRows(rows)
			},
			wantLen: 0,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			defer func() {
				if r := recover(); r != nil {
					t.Logf("Panic encountered so failing test. %v\n%s", r, string(debug.Stack()))
					t.Fail()
				}
			}()

			tt.mockSetup(mock)

			got, err := store.GetFeedArticles(tt.userIDs, tt.limit, tt.offset)

			if (err != nil) != tt.wantErr {
				t.Errorf("ArticleStore.GetFeedArticles() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if len(got) != tt.wantLen {
				t.Errorf("ArticleStore.GetFeedArticles() returned %d articles, want %d", len(got), tt.wantLen)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("Unfulfilled mock expectations: %s", err)
			}

			t.Logf("Test '%s' completed successfully", tt.name)
		})
	}
}


/*
ROOST_METHOD_HASH=ArticleStore_GetTags_45f5cdc4bb
ROOST_METHOD_SIG_HASH=ArticleStore_GetTags_fb0aefcdd2

FUNCTION_DEF=func (s *ArticleStore) GetTags() ([ // GetTags creates a article tag
]model.Tag, error) 

*/
func TestArticleStoreGetTags(t *testing.T) {

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer db.Close()

	gormDB, err := gorm.Open("mysql", db)
	if err != nil {
		t.Fatalf("Failed to open GORM connection: %v", err)
	}
	defer gormDB.Close()

	store := &ArticleStore{
		db: gormDB,
	}

	tests := []struct {
		name    string
		mock    func()
		want    []model.Tag
		wantErr bool
	}{
		{
			name: "Success - Empty Database",
			mock: func() {
				mock.ExpectQuery("SELECT (.+) FROM `tags`").
					WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at", "deleted_at", "name"}))
			},
			want:    []model.Tag{},
			wantErr: false,
		},
		{
			name: "Success - Multiple Tags",
			mock: func() {
				rows := sqlmock.NewRows([]string{"id", "created_at", "updated_at", "deleted_at", "name"}).
					AddRow(1, "2024-01-01", "2024-01-01", nil, "golang").
					AddRow(2, "2024-01-01", "2024-01-01", nil, "testing")
				mock.ExpectQuery("SELECT (.+) FROM `tags`").WillReturnRows(rows)
			},
			want: []model.Tag{
				{Model: gorm.Model{ID: 1}, Name: "golang"},
				{Model: gorm.Model{ID: 2}, Name: "testing"},
			},
			wantErr: false,
		},
		{
			name: "Error - Database Connection Error",
			mock: func() {
				mock.ExpectQuery("SELECT (.+) FROM `tags`").
					WillReturnError(sql.ErrConnDone)
			},
			want:    []model.Tag{},
			wantErr: true,
		},
		{
			name: "Success - Special Characters in Tags",
			mock: func() {
				rows := sqlmock.NewRows([]string{"id", "created_at", "updated_at", "deleted_at", "name"}).
					AddRow(1, "2024-01-01", "2024-01-01", nil, "C++").
					AddRow(2, "2024-01-01", "2024-01-01", nil, "日本語")
				mock.ExpectQuery("SELECT (.+) FROM `tags`").WillReturnRows(rows)
			},
			want: []model.Tag{
				{Model: gorm.Model{ID: 1}, Name: "C++"},
				{Model: gorm.Model{ID: 2}, Name: "日本語"},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			defer func() {
				if r := recover(); r != nil {
					t.Logf("Panic encountered so failing test. %v\n%s", r, string(debug.Stack()))
					t.Fail()
				}
			}()

			tt.mock()

			got, err := store.GetTags()

			t.Logf("Test: %s\nInput: GetTags()\nOutput: %+v\nError: %v", tt.name, got, err)

			if (err != nil) != tt.wantErr {
				t.Errorf("ArticleStore.GetTags() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if len(got) != len(tt.want) {
					t.Errorf("ArticleStore.GetTags() returned %d tags, want %d", len(got), len(tt.want))
					return
				}

				for i, wantTag := range tt.want {
					if got[i].ID != wantTag.ID || got[i].Name != wantTag.Name {
						t.Errorf("ArticleStore.GetTags()[%d] = %v, want %v", i, got[i], wantTag)
					}
				}
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("Unfulfilled mock expectations: %s", err)
			}
		})
	}
}


/*
ROOST_METHOD_HASH=ArticleStore_IsFavorited_799826fee5
ROOST_METHOD_SIG_HASH=ArticleStore_IsFavorited_f6d5e67492

FUNCTION_DEF=func (s *ArticleStore) IsFavorited(a *model.Article, u *model.User) (bool, error) // IsFavorited returns whether the article is favorited by the user


*/
func TestArticleStoreIsFavorited(t *testing.T) {

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer db.Close()

	gormDB, err := gorm.Open("mysql", db)
	if err != nil {
		t.Fatalf("Failed to open GORM connection: %v", err)
	}
	defer gormDB.Close()

	store := &ArticleStore{
		db: gormDB,
	}

	tests := []struct {
		name        string
		article     *model.Article
		user        *model.User
		mockSetup   func(sqlmock.Sqlmock)
		wantFav     bool
		wantErr     bool
		description string
	}{
		{
			name:    "Nil Article and User",
			article: nil,
			user:    nil,
			mockSetup: func(mock sqlmock.Sqlmock) {

			},
			wantFav:     false,
			wantErr:     false,
			description: "Should handle nil inputs gracefully",
		},
		{
			name: "Valid Article and User with Favorite",
			article: &model.Article{
				Model: gorm.Model{ID: 1},
			},
			user: &model.User{
				Model: gorm.Model{ID: 1},
			},
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT count\\(\\*\\) FROM `favorite_articles`").
					WithArgs(1, 1).
					WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
			},
			wantFav:     true,
			wantErr:     false,
			description: "Should return true for favorited article",
		},
		{
			name: "Valid Article and User without Favorite",
			article: &model.Article{
				Model: gorm.Model{ID: 2},
			},
			user: &model.User{
				Model: gorm.Model{ID: 2},
			},
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT count\\(\\*\\) FROM `favorite_articles`").
					WithArgs(2, 2).
					WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
			},
			wantFav:     false,
			wantErr:     false,
			description: "Should return false for non-favorited article",
		},
		{
			name: "Database Error",
			article: &model.Article{
				Model: gorm.Model{ID: 3},
			},
			user: &model.User{
				Model: gorm.Model{ID: 3},
			},
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT count\\(\\*\\) FROM `favorite_articles`").
					WithArgs(3, 3).
					WillReturnError(sql.ErrConnDone)
			},
			wantFav:     false,
			wantErr:     true,
			description: "Should handle database errors properly",
		},
		{
			name: "Zero ID Values",
			article: &model.Article{
				Model: gorm.Model{ID: 0},
			},
			user: &model.User{
				Model: gorm.Model{ID: 0},
			},
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT count\\(\\*\\) FROM `favorite_articles`").
					WithArgs(0, 0).
					WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
			},
			wantFav:     false,
			wantErr:     false,
			description: "Should handle zero ID values appropriately",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			defer func() {
				if r := recover(); r != nil {
					t.Logf("Panic encountered so failing test. %v\n%s", r, string(debug.Stack()))
					t.Fail()
				}
			}()

			t.Log(tt.description)

			tt.mockSetup(mock)

			gotFav, err := store.IsFavorited(tt.article, tt.user)

			if (err != nil) != tt.wantErr {
				t.Errorf("IsFavorited() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotFav != tt.wantFav {
				t.Errorf("IsFavorited() = %v, want %v", gotFav, tt.wantFav)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("Unfulfilled expectations: %v", err)
			}

			t.Logf("Test case '%s' completed successfully", tt.name)
		})
	}
}


/*
ROOST_METHOD_HASH=ArticleStore_Update_3cddacb803
ROOST_METHOD_SIG_HASH=ArticleStore_Update_e245edd177

FUNCTION_DEF=func (s *ArticleStore) Update(m *model.Article) error // Update updates an article


*/
func TestArticleStoreUpdate(t *testing.T) {
	tests := []struct {
		name    string
		article *model.Article
		mockFn  func(mock sqlmock.Sqlmock)
		wantErr bool
	}{
		{
			name: "Successful Update",
			article: &model.Article{
				Model: gorm.Model{
					ID:        1,
					UpdatedAt: time.Now(),
				},
				Title:       "Updated Title",
				Description: "Updated Description",
				Body:        "Updated Body",
				UserID:      1,
			},
			mockFn: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec("UPDATE `articles`").
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
			wantErr: false,
		},
		{
			name: "Update Non-Existent Article",
			article: &model.Article{
				Model: gorm.Model{
					ID: 999,
				},
			},
			mockFn: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec("UPDATE `articles`").
					WillReturnError(gorm.ErrRecordNotFound)
				mock.ExpectRollback()
			},
			wantErr: true,
		},
		{
			name: "Database Error",
			article: &model.Article{
				Model: gorm.Model{
					ID: 1,
				},
			},
			mockFn: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec("UPDATE `articles`").
					WillReturnError(errors.New("database error"))
				mock.ExpectRollback()
			},
			wantErr: true,
		},
		{
			name: "Invalid Article Data",
			article: &model.Article{
				Model: gorm.Model{
					ID: 1,
				},
				Title: "",
			},
			mockFn: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec("UPDATE `articles`").
					WillReturnError(errors.New("validation error"))
				mock.ExpectRollback()
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			defer func() {
				if r := recover(); r != nil {
					t.Logf("Panic encountered so failing test. %v\n%s", r, string(debug.Stack()))
					t.Fail()
				}
			}()

			store, mock := mockArticleStore(t)
			defer store.db.Close()

			tt.mockFn(mock)

			t.Logf("Testing scenario: %s", tt.name)
			err := store.Update(tt.article)

			if (err != nil) != tt.wantErr {
				t.Errorf("ArticleStore.Update() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("Unfulfilled expectations: %v", err)
			}

			if err != nil {
				t.Logf("Expected error occurred: %v", err)
			} else {
				t.Log("Update successful")
			}
		})
	}
}

func mockArticleStore(t *testing.T) (*ArticleStore, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock DB: %v", err)
	}

	gormDB, err := gorm.Open("mysql", db)
	if err != nil {
		t.Fatalf("Failed to open GORM DB: %v", err)
	}

	return &ArticleStore{db: gormDB}, mock
}


/*
ROOST_METHOD_HASH=NewArticleStore_85784abca5
ROOST_METHOD_SIG_HASH=NewArticleStore_436ae9c986

FUNCTION_DEF=func NewArticleStore(db *gorm.DB) *ArticleStore // NewArticleStore returns a new ArticleStore


*/
func TestNewArticleStore(t *testing.T) {

	type testCase struct {
		name     string
		db       *gorm.DB
		wantNil  bool
		validate func(*testing.T, *ArticleStore)
	}

	mockDB, _, err := go-sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock DB: %v", err)
	}
	defer mockDB.Close()

	gormDB, err := gorm.Open("mysql", mockDB)
	if err != nil {
		t.Fatalf("Failed to create GORM DB: %v", err)
	}
	defer gormDB.Close()

	tests := []testCase{
		{
			name: "Successful ArticleStore Creation",
			db:   gormDB,
			validate: func(t *testing.T, store *ArticleStore) {
				if store == nil {
					t.Error("Expected non-nil ArticleStore")
				}
				if store.db != gormDB {
					t.Error("DB reference mismatch")
				}
			},
		},
		{
			name: "Nil Database Connection",
			db:   nil,
			validate: func(t *testing.T, store *ArticleStore) {
				if store == nil {
					t.Error("Expected non-nil ArticleStore even with nil DB")
				}
				if store.db != nil {
					t.Error("Expected nil DB reference")
				}
			},
		},
		{
			name: "DB Reference Integrity",
			db:   gormDB,
			validate: func(t *testing.T, store *ArticleStore) {
				if store.db != gormDB {
					t.Error("DB reference not maintained correctly")
				}
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {

			defer func() {
				if r := recover(); r != nil {
					t.Logf("Panic encountered so failing test. %v\n%s", r, string(debug.Stack()))
					t.Fail()
				}
			}()

			t.Logf("Running test case: %s", tc.name)

			store := NewArticleStore(tc.db)

			tc.validate(t, store)

			t.Logf("Test case completed successfully: %s", tc.name)
		})
	}
}

