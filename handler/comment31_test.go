package handler

import (
	"context"
	"testing"
	"strconv"
	"fmt"
	"github.com/raahii/golang-grpc-realworld-example/auth"
	"github.com/raahii/golang-grpc-realworld-example/model"
	"github.com/raahii/golang-grpc-realworld-example/proto"
	"github.com/raahii/golang-grpc-realworld-example/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"errors"
	"github.com/DATA-DOG/go-sqlmock"
	"gorm.io/gorm"
	"github.com/golang/mock/gomock"
)


var mockArticleStore = new(MockArticleStore)

var mockUserStore = new(MockUserStore)
/*
ROOST_METHOD_HASH=DeleteComment_452af2f984
ROOST_METHOD_SIG_HASH=DeleteComment_27615e7d69


 */
func TestHandlerDeleteComment(t *testing.T) {
	h := &Handler{
		logger: &log.Logger,
		us:     &store.UserStore{db: nil},
		as:     &store.ArticleStore{db: nil},
	}

	tests := []struct {
		name          string
		setup         func() context.Context
		request       *pb.DeleteCommentRequest
		expectedError error
	}{
		{
			name: "Unauthenticated User Tries to Delete a Comment",
			setup: func() context.Context {
				ctx := context.Background()
				return ctx
			},
			request:       &pb.DeleteCommentRequest{Slug: "123", Id: "456"},
			expectedError: status.Error(codes.Unauthenticated, "unauthenticated"),
		},
		{
			name: "User Not Found in the Database",
			setup: func() context.Context {
				ctx := context.WithValue(context.Background(), auth.UserIDKey, uint(1))
				mockUserStore.On("GetByID", uint(1)).Return((*model.User)(nil), status.Error(codes.NotFound, "user not found"))
				return ctx
			},
			request:       &pb.DeleteCommentRequest{Slug: "123", Id: "456"},
			expectedError: status.Error(codes.NotFound, "user not found"),
		},
		{
			name: "Invalid Comment ID Format",
			setup: func() context.Context {
				ctx := context.WithValue(context.Background(), auth.UserIDKey, uint(1))
				mockUserStore.On("GetByID", uint(1)).Return(&model.User{ID: 1}, nil)
				return ctx
			},
			request:       &pb.DeleteCommentRequest{Slug: "123", Id: "invalid"},
			expectedError: status.Error(codes.InvalidArgument, "invalid article id"),
		},
		{
			name: "Comment Not Found in the Article",
			setup: func() context.Context {
				ctx := context.WithValue(context.Background(), auth.UserIDKey, uint(1))
				mockUserStore.On("GetByID", uint(1)).Return(&model.User{ID: 1}, nil)
				mockArticleStore.On("GetCommentByID", uint(456)).Return(&model.Comment{ArticleID: 888}, nil)
				return ctx
			},
			request:       &pb.DeleteCommentRequest{Slug: "123", Id: "456"},
			expectedError: status.Error(codes.InvalidArgument, "the comment is not in the article"),
		},
		{
			name: "User Lacking Permission to Delete Comment",
			setup: func() context.Context {
				ctx := context.WithValue(context.Background(), auth.UserIDKey, uint(2))
				mockUserStore.On("GetByID", uint(2)).Return(&model.User{ID: 2}, nil)
				mockArticleStore.On("GetCommentByID", uint(456)).Return(&model.Comment{UserID: 3, ArticleID: 123}, nil)
				return ctx
			},
			request:       &pb.DeleteCommentRequest{Slug: "123", Id: "456"},
			expectedError: status.Error(codes.InvalidArgument, "forbidden"),
		},
		{
			name: "Successful Comment Deletion",
			setup: func() context.Context {
				ctx := context.WithValue(context.Background(), auth.UserIDKey, uint(1))
				mockUserStore.On("GetByID", uint(1)).Return(&model.User{ID: 1}, nil)
				mockArticleStore.On("GetCommentByID", uint(456)).Return(&model.Comment{UserID: 1, ArticleID: 123}, nil)
				mockArticleStore.On("DeleteComment", mock.Anything).Return(nil)
				return ctx
			},
			request:       &pb.DeleteCommentRequest{Slug: "123", Id: "456"},
			expectedError: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx := test.setup()
			_, err := h.DeleteComment(ctx, test.request)

			if test.expectedError != nil {
				assert.EqualError(t, err, test.expectedError.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}


/*
ROOST_METHOD_HASH=GetComments_265127fb6a
ROOST_METHOD_SIG_HASH=GetComments_20efd5abae


 */
func TestHandlerGetComments(t *testing.T) {

	var logger zerolog.Logger
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to open sqlmock database: %s", err)
	}
	defer db.Close()


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

		
			resp, err := handler.GetComments(context.Background(), tc.req)

		
			if err != nil && status.Code(err) != tc.expectErr {
				t.Errorf("Expected error code %v, got %v", tc.expectErr, status.Code(err))
			}

		
			if tc.expectErr == codes.OK && (resp == nil || len(resp.Comments) > 0 && resp.Comments[0].Body != "Test comment") {
				t.Errorf("Expected valid comments response, got %v", resp)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("There were unfulfilled expectations: %s", err)
			}
		})
	}
}


/*
ROOST_METHOD_HASH=CreateComment_c4ccd62dc5
ROOST_METHOD_SIG_HASH=CreateComment_19a3ee5a3b


 */
func TestHandlerCreateComment(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()


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

