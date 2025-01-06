package handler

import (
	"context"
	"errors"
	"os"
	"strconv"
	"testing"
	"github.com/golang/mock/gomock"
	"github.com/raahii/golang-grpc-realworld-example/auth"
	"github.com/raahii/golang-grpc-realworld-example/model"
	pb "github.com/raahii/golang-grpc-realworld-example/proto"
	"github.com/raahii/golang-grpc-realworld-example/store"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"github.com/rs/zerolog"
)


type Controller struct {
	mu            sync.Mutex
	t             TestReporter
	expectedCalls *callSet
	finished      bool
}






type T struct {
	common
	isParallel bool
	isEnvSet   bool
	context    *testContext // For running tests and subtests.
}
func TestHandlerDeleteArticle(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserStore := store.NewMockUserStore(ctrl)
	mockArticleStore := store.NewMockArticleStore(ctrl)

	logger := zerolog.New(os.Stdout)
	handler := &Handler{
		logger: &logger,
		us:     mockUserStore,
		as:     mockArticleStore,
	}

	tests := []struct {
		name        string
		setup       func() context.Context
		request     *pb.DeleteArticleRequest
		mockSetup   func()
		expectedErr error
	}{
		{
			name: "Unauthorized Access",
			setup: func() context.Context {
				authGetUserID := auth.GetUserID
				auth.GetUserID = func(ctx context.Context) (uint, error) {
					return 0, errors.New("unauthenticated")
				}
				defer func() { auth.GetUserID = authGetUserID }()
				return context.TODO()
			},
			request:     &pb.DeleteArticleRequest{Slug: "1"},
			mockSetup:   func() {},
			expectedErr: status.Error(codes.Unauthenticated, "unauthenticated"),
		},
		{
			name: "User Not Found",
			setup: func() context.Context {
				authGetUserID := auth.GetUserID
				auth.GetUserID = func(ctx context.Context) (uint, error) {
					return 123, nil
				}
				defer func() { auth.GetUserID = authGetUserID }()
				return context.TODO()
			},
			request: &pb.DeleteArticleRequest{Slug: "1"},
			mockSetup: func() {
				mockUserStore.EXPECT().GetByID(uint(123)).Return(nil, errors.New("user not found"))
			},
			expectedErr: status.Error(codes.NotFound, "not user found"),
		},
		{
			name: "Invalid Slug Format",
			setup: func() context.Context {
				authGetUserID := auth.GetUserID
				auth.GetUserID = func(ctx context.Context) (uint, error) {
					return 123, nil
				}
				defer func() { auth.GetUserID = authGetUserID }()
				return context.TODO()
			},
			request:     &pb.DeleteArticleRequest{Slug: "invalidSlug"},
			mockSetup:   func() {},
			expectedErr: status.Error(codes.InvalidArgument, "invalid article id"),
		},
		{
			name: "Article Not Found",
			setup: func() context.Context {
				authGetUserID := auth.GetUserID
				auth.GetUserID = func(ctx context.Context) (uint, error) {
					return 123, nil
				}
				defer func() { auth.GetUserID = authGetUserID }()
				return context.TODO()
			},
			request: &pb.DeleteArticleRequest{Slug: "1"},
			mockSetup: func() {
				mockUserStore.EXPECT().GetByID(uint(123)).Return(&model.User{ID: 123}, nil)
				mockArticleStore.EXPECT().GetByID(uint(1)).Return(nil, errors.New("article not found"))
			},
			expectedErr: status.Error(codes.InvalidArgument, "invalid article id"),
		},
		{
			name: "Article Ownership Violation",
			setup: func() context.Context {
				authGetUserID := auth.GetUserID
				auth.GetUserID = func(ctx context.Context) (uint, error) {
					return 123, nil
				}
				defer func() { auth.GetUserID = authGetUserID }()
				return context.TODO()
			},
			request: &pb.DeleteArticleRequest{Slug: "1"},
			mockSetup: func() {
				mockUserStore.EXPECT().GetByID(uint(123)).Return(&model.User{ID: 123}, nil)
				mockArticleStore.EXPECT().GetByID(uint(1)).Return(&model.Article{ID: 1, Author: &model.User{ID: 456}}, nil)
			},
			expectedErr: status.Error(codes.Unauthenticated, "forbidden"),
		},
		{
			name: "Successful Article Deletion",
			setup: func() context.Context {
				authGetUserID := auth.GetUserID
				auth.GetUserID = func(ctx context.Context) (uint, error) {
					return 123, nil
				}
				defer func() { auth.GetUserID = authGetUserID }()
				return context.TODO()
			},
			request: &pb.DeleteArticleRequest{Slug: "1"},
			mockSetup: func() {
				mockUserStore.EXPECT().GetByID(uint(123)).Return(&model.User{ID: 123}, nil)
				mockArticleStore.EXPECT().GetByID(uint(1)).Return(&model.Article{ID: 1, Author: &model.User{ID: 123}}, nil)
				mockArticleStore.EXPECT().Delete(gomock.Any()).Return(nil)
			},
			expectedErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			_, err := handler.DeleteArticle(tt.setup(), tt.request)

			if status.Code(err) != status.Code(tt.expectedErr) {
				t.Errorf("expected error: %v, got: %v", tt.expectedErr, err)
			}
		})
	}
}
