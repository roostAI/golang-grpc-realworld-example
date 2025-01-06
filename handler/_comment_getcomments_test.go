package handler

import (
	"context"
	"errors"
	"fmt"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/raahii/golang-grpc-realworld-example/auth"
	"github.com/raahii/golang-grpc-realworld-example/model"
	"github.com/raahii/golang-grpc-realworld-example/proto"
	"github.com/raahii/golang-grpc-realworld-example/store"
	"github.com/rs/zerolog"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"strconv"
	"testing"
	"gorm.io/gorm"
)

func TestHandlerGetComments(t *testing.T) {

	var logger zerolog.Logger
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to open sqlmock database: %s", err)
	}
	defer db.Close()

	// Correct the initialization of the stores using the correct gorm.DB field
	articleStore := &store.ArticleStore{db: &gorm.DB{}}
	userStore := &store.UserStore{db: &gorm.DB{}}
	handler := &Handler{logger: &logger, us: userStore, as: articleStore}

	testCases := []struct {
		name       string
		req        *proto.GetCommentsRequest
		setupMocks func()
		expectErr  codes.Code
	}{
		{
			name: "Valid Article Slug with Existing Comments",
			req:  &proto.GetCommentsRequest{Slug: "1"},
			setupMocks: func() {
				mock.ExpectQuery(`SELECT \* FROM "articles" WHERE "articles"."id" = \$1`).
					WithArgs(1).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

				mock.ExpectQuery(`SELECT \* FROM "comments" WHERE article_id = \$1`).
					WithArgs(1).
					WillReturnRows(sqlmock.NewRows([]string{"id", "body"}).AddRow(1, "Test comment"))
			},
			expectErr: codes.OK,
		},
		{
			name: "Invalid Article Slug Conversion",
			req:  &proto.GetCommentsRequest{Slug: "abc"},
			setupMocks: func() {
				// No setup necessary for invalid slug conversion
			},
			expectErr: codes.InvalidArgument,
		},
		{
			name: "Article Not Found",
			req:  &proto.GetCommentsRequest{Slug: "999"},
			setupMocks: func() {
				mock.ExpectQuery(`SELECT \* FROM "articles" WHERE "articles"."id" = \$1`).
					WithArgs(999).
					WillReturnError(errors.New("not found"))
			},
			expectErr: codes.InvalidArgument,
		},
		{
			name: "Comments Retrieval Failure",
			req:  &proto.GetCommentsRequest{Slug: "1"},
			setupMocks: func() {
				mock.ExpectQuery(`SELECT \* FROM "articles" WHERE "articles"."id" = \$1`).
					WithArgs(1).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

				mock.ExpectQuery(`SELECT \* FROM "comments" WHERE article_id = \$1`).
					WithArgs(1).
					WillReturnError(errors.New("retrieval error"))
			},
			expectErr: codes.Aborted,
		},
		{
			name: "Valid Article Slug with No Comments",
			req:  &proto.GetCommentsRequest{Slug: "1"},
			setupMocks: func() {
				mock.ExpectQuery(`SELECT \* FROM "articles" WHERE "articles"."id" = \$1`).
					WithArgs(1).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

				mock.ExpectQuery(`SELECT \* FROM "comments" WHERE article_id = \$1`).
					WithArgs(1).
					WillReturnRows(sqlmock.NewRows([]string{"id", "body"}))
			},
			expectErr: codes.OK,
		},
		{
			name: "Current User Not Found",
			req:  &proto.GetCommentsRequest{Slug: "1"},
			setupMocks: func() {
				mock.ExpectQuery(`SELECT \* FROM "articles" WHERE "articles"."id" = \$1`).
					WithArgs(1).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

				// Mock UserStore.GetByID to return an error
				auth.GetUserID = func(ctx context.Context) (uint, error) {
					return 1, nil
				}
				mock.ExpectQuery(`SELECT \* FROM "users" WHERE "users"."id" = \$1`).
					WithArgs(1).
					WillReturnError(errors.New("not found"))
			},
			expectErr: codes.NotFound,
		},
		{
			name: "Following Status Check Failure",
			req:  &proto.GetCommentsRequest{Slug: "1"},
			setupMocks: func() {
				mock.ExpectQuery(`SELECT \* FROM "articles" WHERE "articles"."id" = \$1`).
					WithArgs(1).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

				mock.ExpectQuery(`SELECT \* FROM "comments" WHERE article_id = \$1`).
					WithArgs(1).
					WillReturnRows(sqlmock.NewRows([]string{"id", "body"}).AddRow(1, "Test comment"))

				auth.GetUserID = func(ctx context.Context) (uint, error) {
					return 1, nil
				}
				mock.ExpectQuery(`SELECT \* FROM "users" WHERE "users"."id" = \$1`).
					WithArgs(1).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

				// Mock UserStore.IsFollowing to return an error
				userStore.IsFollowing = func(a *model.User, b *model.User) (bool, error) {
					return false, errors.New("following check failed")
				}
			},
			expectErr: codes.NotFound,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.setupMocks()

			// Run the main functionality
			resp, err := handler.GetComments(context.Background(), tc.req)

			// Validate error codes where applicable
			if err != nil && status.Code(err) != tc.expectErr {
				t.Errorf("Expected error code %v, got %v", tc.expectErr, status.Code(err))
			}

			// Validate for successful response
			if tc.expectErr == codes.OK && (resp == nil || len(resp.Comments) > 0 && resp.Comments[0].Body != "Test comment") {
				t.Errorf("Expected valid comments response, got %v", resp)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("There were unfulfilled expectations: %s", err)
			}
		})
	}
}
