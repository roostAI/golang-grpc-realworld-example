package handler

import (
	"context"
	"os"
	"strconv"
	"testing"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/raahii/golang-grpc-realworld-example/auth"
	"github.com/raahii/golang-grpc-realworld-example/model"
	pb "github.com/raahii/golang-grpc-realworld-example/proto"
	"github.com/raahii/golang-grpc-realworld-example/store"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)


type ExpectedQuery struct {
	queryBasedExpectation
	rows             driver.Rows
	delay            time.Duration
	rowsMustBeClosed bool
	rowsWereClosed   bool
}







type T struct {
	common
	isParallel bool
	isEnvSet   bool
	context    *testContext // For running tests and subtests.
}
func TestHandlerGetArticle(t *testing.T) {
	logger := zerolog.New(zerolog.ConsoleWriter{Out: os.Stdout})
	us, as, mock := createMockStores(t)
	h := Handler{logger: &logger, us: us, as: as}

	tests := []struct {
		name          string
		setupMocks    func(req *pb.GetArticleRequest)
		req           *pb.GetArticleRequest
		expectedResp  *pb.ArticleResponse
		expectedError codes.Code
		contextFunc   func() context.Context
	}{
		{
			name: "Valid Article Slug and Authenticated User",
			setupMocks: func(req *pb.GetArticleRequest) {
				mock.ExpectQuery("^SELECT (.+) FROM `articles` WHERE (.+)$").
					WillReturnRows(sqlmock.NewRows([]string{"id", "title", "body", "favorites_count"}).
						AddRow(1, "Sample Title", "Sample Body", 10))
				mock.ExpectQuery("^SELECT (.+) FROM `users` WHERE (.+)$").
					WithArgs(1).
					WillReturnRows(sqlmock.NewRows([]string{"id", "username", "bio"}).
						AddRow(1, "testuser", "testbio"))
				mock.ExpectQuery("^SELECT (.+) FROM `favorite_articles` WHERE (.+)$").
					WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
			},
			req: &pb.GetArticleRequest{Slug: "1"},
			expectedResp: &pb.ArticleResponse{Article: &pb.Article{
				Slug:           "1",
				Title:          "Sample Title",
				Body:           "Sample Body",
				FavoritesCount: 10,
				Favorited:      true,
				Author: &pb.Profile{
					Username:  "testuser",
					Bio:       "testbio",
					Following: false,
				},
			}},
			expectedError: codes.OK,
			contextFunc:   func() context.Context { return createMockContextWithUserID(t, 1) },
		},

		{
			name: "Invalid Article Slug (Non-integer Conversion)",
			setupMocks: func(req *pb.GetArticleRequest) {

			},
			req:           &pb.GetArticleRequest{Slug: "abc"},
			expectedResp:  nil,
			expectedError: codes.InvalidArgument,
			contextFunc:   context.Background,
		},
		{
			name: "Article Not Found",
			setupMocks: func(req *pb.GetArticleRequest) {
				mock.ExpectQuery("^SELECT (.+) FROM `articles` WHERE (.+)$").
					WillReturnError(gorm.ErrRecordNotFound)
			},
			req:           &pb.GetArticleRequest{Slug: "2"},
			expectedResp:  nil,
			expectedError: codes.InvalidArgument,
			contextFunc:   context.Background,
		},
		{
			name: "Unauthenticated User Accessing an Article",
			setupMocks: func(req *pb.GetArticleRequest) {
				mock.ExpectQuery("^SELECT (.+) FROM `articles` WHERE (.+)$").
					WillReturnRows(sqlmock.NewRows([]string{"id", "title", "body", "favorites_count"}).
						AddRow(1, "Sample Title", "Sample Body", 10))
			},
			req: &pb.GetArticleRequest{Slug: "1"},
			expectedResp: &pb.ArticleResponse{Article: &pb.Article{
				Slug:           "1",
				Title:          "Sample Title",
				Body:           "Sample Body",
				FavoritesCount: 10,
				Favorited:      false,
				Author:         &pb.Profile{Following: false},
			}},
			expectedError: codes.OK,
			contextFunc:   context.Background,
		},
		{
			name: "Authenticated User Not Found",
			setupMocks: func(req *pb.GetArticleRequest) {
				mock.ExpectQuery("^SELECT (.+) FROM `articles` WHERE (.+)$").
					WillReturnRows(sqlmock.NewRows([]string{"id", "title", "body", "favorites_count"}).
						AddRow(1, "Sample Title", "Sample Body", 10))
				mock.ExpectQuery("^SELECT (.+) FROM `users` WHERE (.+)$").
					WillReturnError(gorm.ErrRecordNotFound)
			},
			req:           &pb.GetArticleRequest{Slug: "1"},
			expectedResp:  nil,
			expectedError: codes.NotFound,
			contextFunc:   func() context.Context { return createMockContextWithUserID(t, 1) },
		},
		{
			name: "Favorited Status Retrieval Error",
			setupMocks: func(req *pb.GetArticleRequest) {
				mock.ExpectQuery("^SELECT (.+) FROM `articles` WHERE (.+)$").
					WillReturnRows(sqlmock.NewRows([]string{"id", "title", "body", "favorites_count"}).
						AddRow(1, "Sample Title", "Sample Body", 10))
				mock.ExpectQuery("^SELECT (.+) FROM `users` WHERE (.+)$").
					WithArgs(1).
					WillReturnRows(sqlmock.NewRows([]string{"id", "username", "bio"}).
						AddRow(1, "testuser", "testbio"))
				mock.ExpectQuery("^SELECT (.+) FROM `favorite_articles` WHERE (.+)$").
					WillReturnError(gorm.ErrConnectionNotFound)
			},
			req:           &pb.GetArticleRequest{Slug: "1"},
			expectedResp:  nil,
			expectedError: codes.Aborted,
			contextFunc:   func() context.Context { return createMockContextWithUserID(t, 1) },
		},
		{
			name: "Following Status Retrieval Error",
			setupMocks: func(req *pb.GetArticleRequest) {
				mock.ExpectQuery("^SELECT (.+) FROM `articles` WHERE (.+)$").
					WillReturnRows(sqlmock.NewRows([]string{"id", "title", "body", "favorites_count"}).
						AddRow(1, "Sample Title", "Sample Body", 10))
				mock.ExpectQuery("^SELECT (.+) FROM `users` WHERE (.+)$").
					WithArgs(1).
					WillReturnRows(sqlmock.NewRows([]string{"id", "username", "bio"}).
						AddRow(1, "testuser", "testbio"))
				mock.ExpectQuery("^SELECT (.+) FROM `follows` WHERE (.+)$").
					WillReturnError(gorm.ErrConnectionNotFound)
			},
			req:           &pb.GetArticleRequest{Slug: "1"},
			expectedResp:  nil,
			expectedError: codes.InvalidArgument,
			contextFunc:   func() context.Context { return createMockContextWithUserID(t, 1) },
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			ctx := tc.contextFunc()
			tc.setupMocks(tc.req)

			resp, err := h.GetArticle(ctx, tc.req)
			if tc.expectedError != codes.OK {
				assert.Error(t, err)
				errStatus, _ := status.FromError(err)
				assert.Equal(t, tc.expectedError, errStatus.Code(), "unexpected error code")
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedResp, resp, "unexpected response")
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}
func createMockContextWithUserID(t *testing.T, userID uint) context.Context {
	ctx := context.Background()
	mockAuth := new(auth.MockAuth)
	mockAuth.On("GetUserID", ctx).Return(userID, nil)

	return ctx
}
func createMockStores(t *testing.T) (*store.UserStore, *store.ArticleStore, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}

	us := &store.UserStore{db: db}
	as := &store.ArticleStore{db: db}
	return us, as, mock
}
