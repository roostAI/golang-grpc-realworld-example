package handler

import (
	"context"
	"testing"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/raahii/golang-grpc-realworld-example/auth"
	"github.com/raahii/golang-grpc-realworld-example/model"
	pb "github.com/raahii/golang-grpc-realworld-example/proto"
	"github.com/raahii/golang-grpc-realworld-example/store"
	"github.com/rs/zerolog"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)


type ExpectedExec struct {
	queryBasedExpectation
	result driver.Result
	delay  time.Duration
}

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
func TestHandlerFavoriteArticle(t *testing.T) {

	db, mock, _ := sqlmock.New()
	defer db.Close()

	gormDB, _ := gorm.Open(postgres.New(postgres.Config{
		Conn: db,
	}), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})

	mockLogger := zerolog.New(nil)
	userStore := &store.UserStore{Db: gormDB}
	articleStore := &store.ArticleStore{Db: gormDB}

	handler := &Handler{
		logger: &mockLogger,
		us:     userStore,
		as:     articleStore,
	}

	tests := []struct {
		name      string
		setup     func()
		req       *pb.FavoriteArticleRequest
		expectErr codes.Code
	}{
		{
			name: "User not authenticated",
			setup: func() {
				auth.GetUserID = func(ctx context.Context) (uint, error) {
					return 0, status.Errorf(codes.Unauthenticated, "unauthenticated")
				}
			},
			req:       &pb.FavoriteArticleRequest{Slug: "1"},
			expectErr: codes.Unauthenticated,
		},
		{
			name: "User authenticated but not found",
			setup: func() {
				auth.GetUserID = func(ctx context.Context) (uint, error) {
					return 1, nil
				}
				mock.ExpectQuery("SELECT * FROM \"users\"").WillReturnError(gorm.ErrRecordNotFound)
			},
			req:       &pb.FavoriteArticleRequest{Slug: "1"},
			expectErr: codes.NotFound,
		},
		{
			name: "Slug cannot be converted to integer",
			setup: func() {
				auth.GetUserID = func(ctx context.Context) (uint, error) {
					return 1, nil
				}
			},
			req:       &pb.FavoriteArticleRequest{Slug: "abc"},
			expectErr: codes.InvalidArgument,
		},
		{
			name: "Article not found in the database",
			setup: func() {
				auth.GetUserID = func(ctx context.Context) (uint, error) {
					return 1, nil
				}
				mock.ExpectQuery("SELECT * FROM \"articles\"").WillReturnError(gorm.ErrRecordNotFound)
			},
			req:       &pb.FavoriteArticleRequest{Slug: "1"},
			expectErr: codes.InvalidArgument,
		},
		{
			name: "Favoriting an article successfully",
			setup: func() {
				auth.GetUserID = func(ctx context.Context) (uint, error) {
					return 1, nil
				}
				mock.ExpectQuery("SELECT * FROM \"articles\"").
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
				mock.ExpectExec("INSERT INTO \"favorited_users\"").
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectExec("UPDATE \"articles\" SET \"favorites_count\"").
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectQuery("SELECT count(*) FROM follows").
					WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
			},
			req:       &pb.FavoriteArticleRequest{Slug: "1"},
			expectErr: codes.OK,
		},
		{
			name: "Error adding article to favorites",
			setup: func() {
				auth.GetUserID = func(ctx context.Context) (uint, error) {
					return 1, nil
				}
				mock.ExpectQuery("SELECT * FROM \"articles\"").
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
				mock.ExpectExec("INSERT INTO \"favorited_users\"").
					WillReturnError(gorm.ErrInvalidTransaction)
			},
			req:       &pb.FavoriteArticleRequest{Slug: "1"},
			expectErr: codes.InvalidArgument,
		},
		{
			name: "Error determining following status",
			setup: func() {
				auth.GetUserID = func(ctx context.Context) (uint, error) {
					return 1, nil
				}
				mock.ExpectQuery("SELECT * FROM \"articles\"").
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
				mock.ExpectExec("INSERT INTO \"favorited_users\"").
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectExec("UPDATE \"articles\" SET \"favorites_count\"").
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectQuery("SELECT count(*) FROM follows").
					WillReturnError(gorm.ErrInvalidTransaction)
			},
			req:       &pb.FavoriteArticleRequest{Slug: "1"},
			expectErr: codes.NotFound,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tc.setup()
			resp, err := handler.FavoriteArticle(context.Background(), tc.req)

			if err != nil {
				s, _ := status.FromError(err)
				if s.Code() != tc.expectErr {
					t.Errorf("expected error code %v, got %v", tc.expectErr, s.Code())
				}
			} else if tc.expectErr != codes.OK {
				t.Errorf("expected error code %v, but got success", tc.expectErr)
			} else if resp.Article == nil {
				t.Error("expected article response, got nil")
			}
		})
	}
}
