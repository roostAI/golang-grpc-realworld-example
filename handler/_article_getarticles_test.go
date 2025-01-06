package handler

import (
	"context"
	"errors"
	"testing"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/raahii/golang-grpc-realworld-example/auth"
	pb "github.com/raahii/golang-grpc-realworld-example/proto"
	"github.com/raahii/golang-grpc-realworld-example/store"
	"github.com/rs/zerolog"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)



type ExpectedQuery struct {
	queryBasedExpectation
	rows             driver.Rows
	delay            time.Duration
	rowsMustBeClosed bool
	rowsWereClosed   bool
}

type Rows struct {
	converter driver.ValueConverter
	cols      []string
	def       []*Column
	rows      [][]driver.Value
	pos       int
	nextErr   map[int]error
	closeErr  error
}







type T struct {
	common
	isParallel bool
	isEnvSet   bool
	context    *testContext // For running tests and subtests.
}
func TestHandlerGetArticles(t *testing.T) {

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to initialize mock database connection: %s", err)
	}
	defer db.Close()

	gormDB, err := gorm.Open(postgres.New(postgres.Config{Conn: db}), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to initialize gorm database connection: %s", err)
	}

	articleStore := &store.ArticleStore{DB: gormDB}
	userStore := &store.UserStore{DB: gormDB}
	logger := zerolog.New(nil)
	handler := &Handler{as: articleStore, us: userStore, logger: &logger}

	type testCase struct {
		name           string
		request        *pb.GetArticlesRequest
		mock           func()
		expectedError  error
		expectedResult *pb.ArticlesResponse
	}

	testCases := []testCase{
		{
			name: "Retrieve Articles Successfully Without Filters",
			request: &pb.GetArticlesRequest{
				Limit:  20,
				Offset: 0,
			},
			mock: func() {
				rows := mock.NewRows([]string{"id", "title", "description", "body"}).
					AddRow(1, "Title1", "Desc1", "Body1").
					AddRow(2, "Title2", "Desc2", "Body2")
				mock.ExpectQuery("^SELECT .+ FROM articles").
					WillReturnRows(rows)
			},
			expectedError: nil,
			expectedResult: &pb.ArticlesResponse{
				Articles: []*pb.Article{
					{Slug: "1", Title: "Title1", Description: "Desc1", Body: "Body1", Favorited: false},
					{Slug: "2", Title: "Title2", Description: "Desc2", Body: "Body2", Favorited: false},
				},
				ArticlesCount: 2,
			},
		},
		{
			name: "Retrieve Articles Filtered by Tag",
			request: &pb.GetArticlesRequest{
				Tag: "golang",
			},
			mock: func() {
				rows := mock.NewRows([]string{"id", "title", "description", "body"}).
					AddRow(1, "Title1", "Desc1", "Body1")
				mock.ExpectQuery("^SELECT .+ FROM articles").
					WillReturnRows(rows)
			},
			expectedError: nil,
			expectedResult: &pb.ArticlesResponse{
				Articles: []*pb.Article{
					{Slug: "1", Title: "Title1", Description: "Desc1", Body: "Body1", Favorited: false},
				},
				ArticlesCount: 1,
			},
		},
		{
			name: "Retrieve Articles Favorite by a User",
			request: &pb.GetArticlesRequest{
				Favorited: "userA",
			},
			mock: func() {

				mock.ExpectQuery("^SELECT .+ FROM users WHERE").
					WillReturnRows(mock.NewRows([]string{"id", "username"}).AddRow(1, "userA"))

				rows := mock.NewRows([]string{"id", "title", "description", "body"}).
					AddRow(1, "Title1", "Desc1", "Body1").
					AddRow(2, "Title2", "Desc2", "Body2")
				mock.ExpectQuery("^SELECT .+ FROM articles").
					WillReturnRows(rows)
			},
			expectedError: nil,
			expectedResult: &pb.ArticlesResponse{
				Articles: []*pb.Article{
					{Slug: "1", Title: "Title1", Description: "Desc1", Body: "Body1", Favorited: true},
					{Slug: "2", Title: "Title2", Description: "Desc2", Body: "Body2", Favorited: true},
				},
				ArticlesCount: 2,
			},
		},
		{
			name: "Handle Database Retrieval Error",
			request: &pb.GetArticlesRequest{
				Limit: 20,
			},
			mock: func() {
				mock.ExpectQuery("^SELECT .+ FROM articles").
					WillReturnError(errors.New("db error"))
			},
			expectedError:  status.Error(codes.Aborted, "internal server error"),
			expectedResult: nil,
		},
		{
			name: "Unauthenticated User Request",
			request: &pb.GetArticlesRequest{
				Limit: 10,
			},
			mock: func() {

				mockWithUnauthenticatedUser := func(ctx context.Context) (uint, error) {
					return 0, status.Error(codes.Unauthenticated, "unauthenticated")
				}
				auth.GetUserID = mockWithUnauthenticatedUser
			},
			expectedError:  nil,
			expectedResult: &pb.ArticlesResponse{Articles: nil, ArticlesCount: 0},
		},
	}

	for _, tc := range testCases {
		tc.mock()
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			resp, err := handler.GetArticles(ctx, tc.request)

			if code := status.Code(err); code != status.Code(tc.expectedError) {
				t.Errorf("Expected error code %v, got %v", status.Code(tc.expectedError), code)
			}

			if tc.expectedResult != nil && len(resp.Articles) != int(tc.expectedResult.ArticlesCount) {
				t.Errorf("Expected response count %d, got %d", tc.expectedResult.ArticlesCount, len(resp.Articles))
			}
		})
	}
}
