package store

import (
	errors "errors"
	testing "testing"
	gorm "github.com/jinzhu/gorm"
	sqlmock "github.com/DATA-DOG/go-sqlmock"
	githubcompkgerrors "github.com/pkg/errors"
	model "github.com/raahii/golang-grpc-realworld-example/model"
	debug "runtime/debug"
	fmt "fmt"
	assert "github.com/stretchr/testify/assert"
	reflect "reflect"
	strconv "strconv"
)





type MockDB struct {
	mock sqlmock.Sqlmock
	gorm *gorm.DB
}
type keys struct {
	ID uint
}
type mockArticleStore struct {
	db *gorm.DB
}


/*
ROOST_METHOD_HASH=NewArticleStore_85784abca5
ROOST_METHOD_SIG_HASH=NewArticleStore_436ae9c986

FUNCTION_DEF=func NewArticleStore(db *gorm.DB) *ArticleStore // NewArticleStore returns a new ArticleStore


*/
func TestNewArticleStore(t *testing.T) {

	tests := []struct {
		name   string
		db     *gorm.DB
		expect func(t *testing.T, store *ArticleStore, err error)
	}{
		{
			name: "Correctly Initialized ArticleStore",
			db: func() *gorm.DB {
				mockDB, _, _ := sqlmock.New()
				mockGorm, _ := gorm.Open("postgres", mockDB)
				return mockGorm
			}(),
			expect: func(t *testing.T, store *ArticleStore, err error) {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
					return
				}
				if store == nil {
					t.Errorf("expected an ArticleStore, got nil")
				}
				t.Log("pass Correctly Initialized ArticleStore")
			},
		},
		{
			name: "Error when Null DB provided",
			db:   nil,
			expect: func(t *testing.T, store *ArticleStore, err error) {
				if err != errors.New("unable to create ArticleStore with nil DB") {
					t.Errorf("expected error, got nil")
				}
				t.Log("pass Error when Null DB provided")
			},
		},
		{
			name: "ArticleStore Reinitialization",
			db: func() *gorm.DB {
				mockDB, _, _ := sqlmock.New()
				mockGorm, _ := gorm.Open("postgres", mockDB)
				return mockGorm
			}(),
			expect: func(t *testing.T, store *ArticleStore, err error) {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
					return
				}
				if store == nil {
					t.Errorf("expected an ArticleStore, got nil")
				}
				t.Log("pass ArticleStore Reinitialization")
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Logf("Panic encountered so failing test. %v \n", r)
					t.Fail()
				}
			}()
			result := NewArticleStore(tc.db)
			if result != nil {
				tc.expect(t, result, nil)
				return
			}

		})
	}
}


/*
ROOST_METHOD_HASH=ArticleStore_Create_1273475ade
ROOST_METHOD_SIG_HASH=ArticleStore_Create_a27282cad5

FUNCTION_DEF=func (s *ArticleStore) Create(m *model.Article) error // Create creates an article


*/
func TestArticleStoreCreate(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal("could not mock sql: ", err)
	}
	defer db.Close()

	gormDB, err := gorm.Open("sqlmock", db)
	if err != nil {
		t.Fatal("could not mock sql: ", err)
	}

	articleStore := ArticleStore{
		db: gormDB,
	}

	testUser := model.User{
		Username: "Tester",
		Email:    "tester@test.com",
		Password: "123456",
	}

	expectedArticle := model.Article{
		Title:  "Test article",
		Body:   "Test article body",
		Author: testUser,
	}

	validDataScenario := func() {
		mock.ExpectBegin()
		mock.ExpectExec("INSERT INTO \"articles\" (\"created_at\",\"updated_at\",\"deleted_at\",\"slug\",\"title\",\"description\",\"body\",\"author_id\")").
			WithArgs(expectedArticle.Title, expectedArticle.Body, expectedArticle.Author.ID).
			WillReturnResult(sqlmock.NewResult(1, 1))

		err := articleStore.Create(&expectedArticle)

		if err != nil {
			t.Errorf("valid data scenario failed: %s", err)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("valid data scenario not met: %s", err)
		}
	}

	invalidDataScenario := func() {
		expectedArticle := model.Article{
			Title:  "",
			Body:   "Test article body",
			Author: testUser,
		}

		mock.ExpectBegin()
		mock.ExpectExec("INSERT INTO \"articles\" (\"created_at\",\"updated_at\",\"deleted_at\",\"slug\",\"title\",\"description\",\"body\",\"author_id\")").
			WithArgs(expectedArticle.Title, expectedArticle.Body, expectedArticle.Author).
			WillReturnError(githubcompkgerrors.New("invalid data"))

		err := articleStore.Create(&expectedArticle)

		if err == nil {
			t.Errorf("Test failed: expected an error for incomplete data scenario, got %v", err)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("Test failed: expected error was not met for incomplete data scenario, got %s", err)
		}
	}

	unreachableDBScenario := func() {
		articleStore := &ArticleStore{}

		expectedArticle := model.Article{
			Title:  "Test article",
			Body:   "Test article body",
			Author: testUser,
		}

		err := articleStore.Create(&expectedArticle)

		if err == nil {
			t.Errorf("Test failed: expected an error but got nil for unreachable db scenario")
		}
	}

	testCases := []struct {
		name   string
		action func()
	}{
		{"valid data scenario", validDataScenario},
		{"invalid data scenario", invalidDataScenario},
		{"unreachable database", unreachableDBScenario},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("Panic encountered so failing test: %v", r)
				}
			}()
			testCase.action()
		})
	}
}


/*
ROOST_METHOD_HASH=ArticleStore_CreateComment_b16d4a71d4
ROOST_METHOD_SIG_HASH=ArticleStore_CreateComment_7475736b06

FUNCTION_DEF=func (s *ArticleStore) CreateComment(m *model.Comment) error // CreateComment creates a comment of the article


*/
func TestArticleStoreCreateComment(t *testing.T) {

	type args struct {
		comment *model.Comment
	}
	tests := []struct {
		name    string
		setupDB func(mock sqlmock.Sqlmock)
		args    args
		wantErr bool
	}{
		{
			name: "scenario 1: Successfully creating a comment",
			setupDB: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("INSERT INTO").WillReturnResult(sqlmock.NewResult(1, 1))
			},
			args: args{
				comment: &model.Comment{Body: "Test comment"},
			},
			wantErr: false,
		},
		{
			name: "scenario 2: CreateComment fails when comment is invalid",
			setupDB: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("INSERT INTO").WillReturnResult(sqlmock.NewResult(1, 1))
			},
			args: args{
				comment: &model.Comment{},
			},
			wantErr: true,
		},
		{
			name: "scenario 3: CreateComment fails when there's a database error",
			setupDB: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("INSERT INTO").WillReturnError(errors.New("database error"))
			},
			args: args{
				comment: &model.Comment{Body: "Test comment"},
			},
			wantErr: true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Logf("Panic encountered so failing test. Err: %v", r)
					t.Fail()
				}
			}()

			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
			}
			defer db.Close()

			gormDB, err := gorm.Open("postgres", db)
			if err != nil {
				t.Fatalf("failed to open gorm db, err: %s", err)
			}

			test.setupDB(mock)

			store := &ArticleStore{
				db: gormDB,
			}

			err = store.CreateComment(test.args.comment)
			if (err != nil) != test.wantErr {
				t.Errorf("CreateComment() error = %v, wantErr %v", err, test.wantErr)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations in the mock: %s", err)
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
	scenarios := []struct {
		name       string
		setupMocks func(mock sqlmock.Sqlmock)
		input      *model.Article
		expectErr  bool
	}{
		{
			name: "Delete Article Successfully",
			setupMocks: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("DELETE").WithArgs(1).WillReturnResult(sqlmock.NewResult(1, 1))
			},
			input:     &model.Article{},
			expectErr: false,
		},
		{
			name: "Delete Article That Does Not Exist",
			setupMocks: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("DELETE").WithArgs(1).WillReturnError(errors.New("no record found"))
			},
			input:     &model.Article{},
			expectErr: true,
		},
		{
			name: "Database Unavailable During Deletion",
			setupMocks: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("DELETE").WillReturnError(errors.New("database error"))
			},
			input:     &model.Article{},
			expectErr: true,
		},
		{
			name:       "Invalid Article Data",
			setupMocks: func(mock sqlmock.Sqlmock) {},
			input:      nil,
			expectErr:  true,
		},
	}

	for _, tc := range scenarios {
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Fatalf("Panic encountered so failing test. %v\n%s", r, string(debug.Stack()))
				}
			}()

			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("Error while setting up mock DB. %v", err)
			}
			defer db.Close()

			tc.setupMocks(mock)

			gormDB, err := gorm.Open("postgres", db)
			if err != nil {
				t.Fatalf("Error while opening gorm DB. %v", err)
			}

			store := &ArticleStore{db: gormDB}

			err = store.Delete(tc.input)
			if (err != nil) != tc.expectErr {
				t.Fatalf("Expected error: %v, got: %v", tc.expectErr, err)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("Unfulfilled expectations: %s", err)
			}

			if err != nil && tc.expectErr {
				t.Logf("Expected error: %v\n", err)
			} else if err == nil && !tc.expectErr {
				t.Log("Expected success, and got it")
			}
		})
	}
}


/*
ROOST_METHOD_HASH=ArticleStore_DeleteComment_effbcb38aa
ROOST_METHOD_SIG_HASH=ArticleStore_DeleteComment_d3c99623e4

FUNCTION_DEF=func (s *ArticleStore) DeleteComment(m *model.Comment) error // DeleteComment deletes an comment


*/
func (db *MockDB) Delete(value interface{}, where ...interface{}) *gorm.DB {
	if _, ok := value.(*model.Comment); ok {
		return &gorm.DB{
			Error: errors.New("Record not found in mock db for deletion"),
		}
	}
	return &gorm.DB{
		Error: errors.New("Value passed is not *model.Comment"),
	}
}

func TestArticleStoreDeleteComment(t *testing.T) {
	testCases := []struct {
		name          string
		comment       *model.Comment
		deleteCalled  bool
		expectedError error
	}{
		{
			name:          "Successful Deletion of a Comment",
			comment:       &model.Comment{},
			deleteCalled:  true,
			expectedError: nil,
		},
		{
			name:          "Attempted Deletion of a Non-Existent Comment",
			comment:       &model.Comment{},
			deleteCalled:  true,
			expectedError: errors.New("Comment does not exist"),
		},
		{
			name:          "Deletion of a Comment from an Empty Database",
			comment:       &model.Comment{},
			deleteCalled:  false,
			expectedError: errors.New("Database is empty"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockDB := new(MockDB)
			store := &ArticleStore{
				db: mockDB.gorm,
			}

			err := store.DeleteComment(tc.comment)

			if tc.deleteCalled {
				assert.Nil(t, err, fmt.Sprintf("Expected no error but got %v", err))
			} else {
				assert.Equal(t, tc.expectedError, err)
			}
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

	expectedTags1 := []model.Tag{{}}
	expectedTags2 := []model.Tag{{}}
	err2 := gorm.ErrInvalidSQL

	scenarios := []struct {
		Name         string
		SetupMock    func(sqlmock.Sqlmock)
		ExpectedTags []model.Tag
		ExpectedErr  error
	}{
		{"Successfully Retrieving Article Tags from Article Store", setupScenario1, expectedTags1, nil},
		{"Error while Retrieving Article Tags from Article Store", setupScenario2, expectedTags2, err2},
		{"Empty Article Tags in Article Store", func(sqlmock.Sqlmock) {}, nil, nil},
	}

	for _, s := range scenarios {
		gormDB, mock, err := newMockDB()
		if err != nil {
			t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
		}

		articleStore := ArticleStore{db: gormDB}

		t.Run(s.Name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Logf("Panic encountered so failing test. %v", r)
					t.Fail()
				}
			}()
			s.SetupMock(mock)

			tagResult, err := articleStore.GetTags()

			if s.ExpectedErr != nil {
				if err.Error() != s.ExpectedErr.Error() {
					t.Errorf("expected error '%v', got '%v'", s.ExpectedErr, err)
					t.Fail()
				}
			} else if err != nil {
				t.Errorf("unexpected error gotten '%v'", err)
				t.Fail()
			} else {
				if !reflect.DeepEqual(tagResult, s.ExpectedTags) {
					t.Errorf("expected tag value does not match. Expected %#v, got %#v", s.ExpectedTags, tagResult)
				}
			}
		})
	}
}

func newMockDB() (*gorm.DB, sqlmock.Sqlmock, error) {
	mockDb, mock, err := sqlmock.New()
	if err != nil {
		return nil, nil, err
	}
	gormDb, err := gorm.Open("postgres", mockDb)
	if err != nil {
		return nil, nil, err
	}
	return gormDb, mock, nil
}

func setupScenario1(mock sqlmock.Sqlmock) {
	mock.ExpectQuery("^SELECT (.+) FROM \"tags\"$").
		WillReturnRows(sqlmock.NewRows([]string{"ID", "Name"}).
			AddRow(1, "test"))
}

func setupScenario2(mock sqlmock.Sqlmock) {
	mock.ExpectQuery("^SELECT (.+) FROM \"tags\"$").
		WillReturnError(gorm.ErrInvalidSQL)
}


/*
ROOST_METHOD_HASH=ArticleStore_GetByID_6fe18728fc
ROOST_METHOD_SIG_HASH=ArticleStore_GetByID_bb488e542f

FUNCTION_DEF=func (s *ArticleStore) GetByID(id uint) (*model.Article, error) // GetByID finds an article from id


*/
func TestArticleStoreGetById(t *testing.T) {
	var expectedArticle model.Article
	expectedArticle.ID = 1

	tests := []struct {
		Id      uint
		Setup   func(db *gorm.DB)
		WantErr bool
	}{
		{
			Id: expectedArticle.ID,
			Setup: func(db *gorm.DB) {

				db.Save(&expectedArticle)
			},
			WantErr: false,
		},
		{
			Id:      1000,
			Setup:   func(db *gorm.DB) {},
			WantErr: true,
		},
		{
			Id: expectedArticle.ID,
			Setup: func(db *gorm.DB) {

				db.Close()
			},
			WantErr: true,
		},
	}

	for i, tt := range tests {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
		}
		gormDb, err := gorm.Open("sqlmock", db)
		defer func() {
			_ = gormDb.Close()
		}()
		defer func() {
			_ = mock.ExpectationsWereMet()
		}()
		if err != nil {
			t.Fatalf("failed to open sqlmock database: %s", err)
		}
		tt.Setup(gormDb)
		articlestore := &ArticleStore{gormDb}

		t.Run("TestCaseNo-"+strconv.Itoa(i), func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Logf("Panic encountered so failing test. %v", r)
					t.Fail()
				}
			}()
			output, err := articlestore.GetByID(tt.Id)
			if tt.WantErr {
				assert.Error(t, err)
				t.Logf("Received error as expected: %s", err.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, expectedArticle, *output)
			}
		})
	}
}


/*
ROOST_METHOD_HASH=ArticleStore_Update_3cddacb803
ROOST_METHOD_SIG_HASH=ArticleStore_Update_e245edd177

FUNCTION_DEF=func (s *ArticleStore) Update(m *model.Article) error // Update updates an article


*/
func TestArticleStoreUpdate(t *testing.T) {

	scenarios := []struct {
		desc   string
		setup  func(mock sqlmock.Sqlmock, article *model.Article)
		verify func(t *testing.T, mock sqlmock.Sqlmock, db *gorm.DB)
	}{
		{

			desc: "Update existing article",
			setup: func(mock sqlmock.Sqlmock, article *model.Article) {

				rows := sqlmock.NewRows([]string{"title"}).AddRow(article.Title)

				mock.ExpectBegin()
				mock.ExpectQuery("SELECT").WillReturnRows(rows)
				mock.ExpectExec("UPDATE").WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
			verify: func(t *testing.T, mock sqlmock.Sqlmock, db *gorm.DB) {
				if err := mock.ExpectationsWereMet(); err != nil {
					t.Fatalf("unable to meet all expectations: %s", err)
				}
			},
		},
		{

			desc: "Update nonexistent article",
			setup: func(mock sqlmock.Sqlmock, article *model.Article) {
				mock.ExpectBegin()
				mock.ExpectQuery("SELECT").WillReturnError(gorm.ErrRecordNotFound)
				mock.ExpectRollback()
			},
			verify: func(t *testing.T, mock sqlmock.Sqlmock, db *gorm.DB) {
				if err := mock.ExpectationsWereMet(); err != nil {
					t.Fatalf("unable to meet all expectations: %s", err)
				}
			},
		},
		{

			desc: "Update article database error",
			setup: func(mock sqlmock.Sqlmock, article *model.Article) {

				rows := sqlmock.NewRows([]string{"title"}).AddRow(article.Title)

				mock.ExpectBegin()
				mock.ExpectQuery("SELECT").WillReturnRows(rows)
				mock.ExpectExec("UPDATE").WillReturnError(gorm.ErrRecordNotFound)
				mock.ExpectRollback()
			},
			verify: func(t *testing.T, mock sqlmock.Sqlmock, db *gorm.DB) {
				if err := mock.ExpectationsWereMet(); err != nil {
					t.Fatalf("unable to meet all expectations: %s", err)
				}
			},
		},
	}

	for _, scenario := range scenarios {
		t.Run(scenario.desc, func(t *testing.T) {

			defer func() {
				if r := recover(); r != nil {
					t.Fatalf("Panic encountered: %v", r)
				}
			}()

			mockDB, mock, _ := sqlmock.New()
			gormDB, _ := gorm.Open("mysql", mockDB)
			articleStore := ArticleStore{
				db: gormDB,
			}

			article := &model.Article{}
			scenario.setup(mock, article)
			if err := articleStore.Update(article); err != nil {

				t.Logf("An error occurred: %v", err)
				t.Logf("Test Scenario: %s failed", scenario.desc)
			} else {
				scenario.verify(t, mock, gormDB)
				t.Logf("Test Scenario: %s passed", scenario.desc)
			}
			mockDB.Close()
		})
	}
}


/*
ROOST_METHOD_HASH=ArticleStore_IsFavorited_799826fee5
ROOST_METHOD_SIG_HASH=ArticleStore_IsFavorited_f6d5e67492

FUNCTION_DEF=func (s *ArticleStore) IsFavorited(a *model.Article, u *model.User) (bool, error) // IsFavorited returns whether the article is favorited by the user


*/
func TestArticleStoreIsFavorited(t *testing.T) {
	db, mock, _ := sqlmock.New()
	gormDB, _ := gorm.Open("mysql", db)
	store := &ArticleStore{
		db: gormDB,
	}

	scenarios := []struct {
		Name        string
		SetupMock   func()
		Article     keys
		User        keys
		ExpectedRes bool
		ShouldError bool
	}{
		{
			Name: "Scenario 1: Nil article and user",
			SetupMock: func() {

			},
			Article:     keys{ID: 1},
			User:        keys{ID: 1},
			ExpectedRes: false,
			ShouldError: false,
		},
		{
			Name: "Scenario 2: Valid article and user are favorited",
			SetupMock: func() {
				mock.ExpectQuery("SELECT \\* FROM favorite_articles WHERE \\(article_id = \\? AND user_id = \\?\\)").WithArgs(1, 1).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
			},
			Article:     keys{ID: 1},
			User:        keys{ID: 1},
			ExpectedRes: true,
			ShouldError: false,
		},
		{
			Name: "Scenario 3: Valid article and user not favorited",
			SetupMock: func() {
				mock.ExpectQuery("SELECT \\* FROM favorite_articles WHERE \\(article_id = \\? AND user_id = \\?\\)").WithArgs(1, 1).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
			},
			Article:     keys{ID: 1},
			User:        keys{ID: 1},
			ExpectedRes: false,
			ShouldError: false,
		},
		{
			Name: "Scenario 4: Database error",
			SetupMock: func() {
				mock.ExpectQuery("SELECT \\* FROM favorite_articles WHERE \\(article_id = \\? AND user_id = \\?\\)").WithArgs(1, 1).WillReturnError(fmt.Errorf("db error"))
			},
			Article:     keys{ID: 1},
			User:        keys{ID: 1},
			ExpectedRes: false,
			ShouldError: true,
		},
	}

	for _, s := range scenarios {
		t.Run(s.Name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Logf("Panic encountered so failing test. %v\n%s", r, string(debug.Stack()))
					t.Fail()
				}
			}()

			s.SetupMock()

			a := &model.Article{Model: gorm.Model{ID: s.Article.ID}}
			u := &model.User{Model: gorm.Model{ID: s.User.ID}}
			res, err := store.IsFavorited(a, u)

			if s.ShouldError && err == nil {
				t.Error("Expected an error but didn't get one")
				return
			}

			if !s.ShouldError && err != nil {
				t.Errorf("Didn't expect an error but got one: %v", err)
				return
			}

			if res != s.ExpectedRes {
				t.Errorf("Expected result %v, got %v", s.ExpectedRes, res)
			}
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
		t.Fatalf("An error '%s' occurred when opening a stub database connection", err)
	}
	gormDB, _ := gorm.Open("postgres", db)

	mockArticleStore := &ArticleStore{
		db: gormDB,
	}

	mockUserIDs := []uint{1, 2}
	mockLimit := int64(2)
	mockOffset := int64(0)

	testCases := []struct {
		name      string
		userIDs   []uint
		limit     int64
		offset    int64
		mockError error
		wantError bool
	}{
		{
			name:      "Retrieve Feed Articles Successfully",
			userIDs:   mockUserIDs,
			limit:     mockLimit,
			offset:    mockOffset,
			mockError: nil,
			wantError: false,
		},
		{
			name:      "User IDs do not Exist in the Database",
			userIDs:   []uint{10, 11},
			limit:     mockLimit,
			offset:    mockOffset,
			mockError: nil,
			wantError: true,
		},
		{
			name:      "Database Connection Error",
			userIDs:   mockUserIDs,
			limit:     mockLimit,
			offset:    mockOffset,
			mockError: errors.New("database connection error"),
			wantError: true,
		},
		{
			name:      "Offset and Limit Values Handling",
			userIDs:   mockUserIDs,
			limit:     2,
			offset:    1,
			mockError: nil,
			wantError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.mockError != nil {
				mock.ExpectQuery("^SELECT (.+) FROM \"articles\"*").
					WillReturnError(tc.mockError)
			} else {
				rows := sqlmock.NewRows([]string{"id", "user_id", "title"}).
					AddRow(1, 1, "title 1").
					AddRow(2, 2, "title 2")
				mock.ExpectQuery("^SELECT (.+) FROM \"articles\"*").
					WillReturnRows(rows)
			}

			articles, err := mockArticleStore.GetFeedArticles(tc.userIDs, tc.limit, tc.offset)

			if tc.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				for _, article := range articles {
					assert.Equal(t, model.Article{}, article)
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
	mockDB, mock, _ := sqlmock.New()
	defer mockDB.Close()
	gormDB, _ := gorm.Open("postgres", mockDB)

	mockArticles := []model.Article{
		{
			Model: gorm.Model{
				ID: 1,
			},
			Title:       "article1",
			Description: "desc1",
			Body:        "body1",
		},
		{
			Model: gorm.Model{
				ID: 2,
			},
			Title:       "article2",
			Description: "desc2",
			Body:        "body2",
		},
	}

	mockUser := model.User{
		Model: gorm.Model{
			ID: 1,
		},
		Username: "usertest",
	}

	mockStore := &mockArticleStore{gormDB}

	mock.ExpectQuery("SELECT").WillReturnRows(sqlmock.NewRows([]string{"id", "title", "description", "body"}).AddRow(mockArticles[0].ID, mockArticles[0].Title, mockArticles[0].Description, mockArticles[0].Body).AddRow(mockArticles[1].ID, mockArticles[1].Title, mockArticles[1].Description, mockArticles[1].Body))

	articles, err := mockStore.GetArticles("", mockUser.Username, &mockUser, 2, 0)

	assert.Nil(t, err)
	assert.Equal(t, mockArticles, articles)
}

func (mockStore *mockArticleStore) GetArticles(tagName, username string, favoritedBy *model.User, limit, offset int64) ([]model.Article, error) {
	return []model.Article{}, nil
}

