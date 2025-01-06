package handler

import (
	"context"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/golang/mock/gomock"
	"github.com/raahii/golang-grpc-realworld-example/auth"
	"github.com/raahii/golang-grpc-realworld-example/model"
	"github.com/raahii/golang-grpc-realworld-example/proto"
	"github.com/raahii/golang-grpc-realworld-example/store"
	"github.com/rs/zerolog"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestHandlerCreateComment(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Mock dependencies
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	userStoreMock := store.NewMockUserStore(ctrl)
	articleStoreMock := store.NewMockArticleStore(ctrl)
	logger := zerolog.NewMockLogger(ctrl)

	handler := &Handler{
		logger: logger,
		us:     userStoreMock,
		as:     articleStoreMock,
	}

	mockAuth := auth.NewMockAuth(ctrl)

	testCases := []struct {
		name                string
		setup               func(req *proto.CreateCommentRequest, ctx context.Context)
		req                 *proto.CreateCommentRequest
		expectedError       error
		expectedCommentBody string
	}{
		{
			name: "Successfully Creating a Comment",
			setup: func(req *proto.CreateCommentRequest, ctx context.Context) {
				mockAuth.EXPECT().GetUserID(ctx).Return(uint(1), nil)
				userStoreMock.EXPECT().GetByID(uint(1)).Return(&model.User{ID: 1}, nil)
				articleStoreMock.EXPECT().GetByID(uint(1)).Return(&model.Article{ID: 1}, nil)
				articleStoreMock.EXPECT().CreateComment(gomock.Any()).Return(nil)
			},
			req: &proto.CreateCommentRequest{
				Slug: "1",
				Comment: &proto.CreateCommentRequest_Comment{
					Body: "This is a comment",
				},
			},
			expectedError:       nil,
			expectedCommentBody: "This is a comment",
		},
		{
			name: "Authentication Failure",
			setup: func(req *proto.CreateCommentRequest, ctx context.Context) {
				mockAuth.EXPECT().GetUserID(ctx).Return(uint(0), errors.New("unauthenticated"))
			},
			req: &proto.CreateCommentRequest{
				Slug: "1",
				Comment: &proto.CreateCommentRequest_Comment{
					Body: "Auth test body",
				},
			},
			expectedError: status.Errorf(codes.Unauthenticated, "unauthenticated"),
		},
		{
			name: "User Not Found",
			setup: func(req *proto.CreateCommentRequest, ctx context.Context) {
				mockAuth.EXPECT().GetUserID(ctx).Return(uint(1), nil)
				userStoreMock.EXPECT().GetByID(uint(1)).Return(nil, errors.New("user not found"))
			},
			req: &proto.CreateCommentRequest{
				Slug: "1",
				Comment: &proto.CreateCommentRequest_Comment{
					Body: "User not found test",
				},
			},
			expectedError: status.Errorf(codes.NotFound, "user not found"),
		},
		{
			name: "Invalid Article ID in Slug",
			setup: func(req *proto.CreateCommentRequest, ctx context.Context) {
			},
			req: &proto.CreateCommentRequest{
				Slug: "invalid_slug",
				Comment: &proto.CreateCommentRequest_Comment{
					Body: "Invalid slug test",
				},
			},
			expectedError: status.Errorf(codes.InvalidArgument, "invalid article id"),
		},
		{
			name: "Article Not Found",
			setup: func(req *proto.CreateCommentRequest, ctx context.Context) {
				mockAuth.EXPECT().GetUserID(ctx).Return(uint(1), nil)
				userStoreMock.EXPECT().GetByID(uint(1)).Return(&model.User{ID: 1}, nil)
				articleID := uint(1)
				articleStoreMock.EXPECT().GetByID(articleID).Return(nil, errors.New("article not found"))
			},
			req: &proto.CreateCommentRequest{
				Slug: "1",
				Comment: &proto.CreateCommentRequest_Comment{
					Body: "Article not found test",
				},
			},
			expectedError: status.Errorf(codes.InvalidArgument, "invalid article id"),
		},
		{
			name: "Validation Error on Comment Content",
			setup: func(req *proto.CreateCommentRequest, ctx context.Context) {
				mockAuth.EXPECT().GetUserID(ctx).Return(uint(1), nil)
				userStoreMock.EXPECT().GetByID(uint(1)).Return(&model.User{ID: 1}, nil)
				articleID := uint(1)
				articleStoreMock.EXPECT().GetByID(articleID).Return(&model.Article{ID: 1}, nil)
			},
			req: &proto.CreateCommentRequest{
				Slug: "1",
				Comment: &proto.CreateCommentRequest_Comment{
					Body: "",
				},
			},
			expectedError: status.Errorf(codes.InvalidArgument, "validation error: body"),
		},
		{
			name: "Database Failure when Creating Comment",
			setup: func(req *proto.CreateCommentRequest, ctx context.Context) {
				mockAuth.EXPECT().GetUserID(ctx).Return(uint(1), nil)
				userStoreMock.EXPECT().GetByID(uint(1)).Return(&model.User{ID: 1}, nil)
				articleStoreMock.EXPECT().GetByID(uint(1)).Return(&model.Article{ID: 1}, nil)
				articleStoreMock.EXPECT().CreateComment(gomock.Any()).Return(errors.New("db error"))
			},
			req: &proto.CreateCommentRequest{
				Slug: "1",
				Comment: &proto.CreateCommentRequest_Comment{
					Body: "DB error test",
				},
			},
			expectedError: status.Errorf(codes.Aborted, "failed to create comment."),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			tc.setup(tc.req, ctx)
			resp, err := handler.CreateComment(ctx, tc.req)

			if tc.expectedError != nil {
				if err == nil || err.Error() != tc.expectedError.Error() {
					t.Fatalf("expected error: %v, got: %v", tc.expectedError, err)
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}

				if resp.Comment.Body != tc.expectedCommentBody {
					t.Fatalf("expected comment body: %s, got: %s", tc.expectedCommentBody, resp.Comment.Body)
				}
			}
		})
	}
}
