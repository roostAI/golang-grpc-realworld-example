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
	"github.com/raahii/golang-grpc-realworld-example/proto"
	"github.com/raahii/golang-grpc-realworld-example/store"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"github.com/rs/zerolog"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
	"gorm.io/driver/postgres"
	"gorm.io/gorm/logger"
	"fmt"
)

const authUserIDKey = "userID"
/*
ROOST_METHOD_HASH=DeleteArticle_0347183038
ROOST_METHOD_SIG_HASH=DeleteArticle_b2585946c3


 */
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


/*
ROOST_METHOD_HASH=GetArticle_8db60d3055
ROOST_METHOD_SIG_HASH=GetArticle_ea0095c9f8


 */
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


/*
ROOST_METHOD_HASH=FavoriteArticle_29edacd2dc
ROOST_METHOD_SIG_HASH=FavoriteArticle_eb25e62ccd


 */
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


/*
ROOST_METHOD_HASH=GetFeedArticles_87ea56b889
ROOST_METHOD_SIG_HASH=GetFeedArticles_2be3462049


 */
func TestHandlerGetFeedArticles(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserStore := NewMockUserStore(ctrl)
	mockArticleStore := NewMockArticleStore(ctrl)
	mockLogger := zerolog.New(nil)

	handler := &Handler{logger: &mockLogger, us: mockUserStore, as: mockArticleStore}
	ctx := context.Background()

	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	scenarios := []struct {
		name        string
		setup       func()
		verify      func(*pb.ArticlesResponse, error)
		description string
	}{
		{
			name: "Successfully Retrieve Feed Articles",
			setup: func() {
				userID := uint(1)
				req := &pb.GetFeedArticlesRequest{Limit: 10, Offset: 0}

				auth.GetUserID = func(ctx context.Context) (uint, error) {
					return userID, nil
				}

				currentUser := &model.User{ID: userID}
				followedIDs := []uint{2, 3}
				articles := []model.Article{
					{ID: 1, Title: "Article 1", Author: model.User{ID: 2}},
					{ID: 2, Title: "Article 2", Author: model.User{ID: 3}},
				}

				mockUserStore.EXPECT().GetByID(userID).Return(currentUser, nil)
				mockUserStore.EXPECT().GetFollowingUserIDs(currentUser).Return(followedIDs, nil)
				mockArticleStore.EXPECT().GetFeedArticles(followedIDs, req.GetLimit(), req.GetOffset()).Return(articles, nil)
				mockArticleStore.EXPECT().IsFavorited(&articles[0], currentUser).Return(true, nil)
				mockArticleStore.EXPECT().IsFavorited(&articles[1], currentUser).Return(false, nil)
				mockUserStore.EXPECT().IsFollowing(currentUser, &articles[0].Author).Return(true, nil)
				mockUserStore.EXPECT().IsFollowing(currentUser, &articles[1].Author).Return(false, nil)
			},
			verify: func(resp *pb.ArticlesResponse, err error) {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				assert.Equal(t, 2, len(resp.Articles))
				assert.Equal(t, "Article 1", resp.Articles[0].Title)
			},
			description: "Ensures correct articles retrieval with expected fields.",
		},
		{
			name: "Unauthenticated User Request",
			setup: func() {
				auth.GetUserID = func(ctx context.Context) (uint, error) {
					return 0, fmt.Errorf("unauthenticated")
				}
			},
			verify: func(resp *pb.ArticlesResponse, err error) {
				assert.Nil(t, resp)
				assert.Equal(t, status.Errorf(codes.Unauthenticated, "unauthenticated"), err)
			},
			description: "Ensures unauthenticated users are denied access.",
		},
	}

	for _, s := range scenarios {
		t.Run(s.name, func(t *testing.T) {
			s.setup()
			req := &pb.GetFeedArticlesRequest{Limit: 10, Offset: 0}
			resp, err := handler.GetFeedArticles(ctx, req)
			s.verify(resp, err)
			t.Logf("Scenario '%s': %s", s.name, s.description)
		})
	}
}


/*
ROOST_METHOD_HASH=UnfavoriteArticle_47bfda8100
ROOST_METHOD_SIG_HASH=UnfavoriteArticle_9043d547fd


 */
func TestHandlerUnfavoriteArticle(t *testing.T) {
	t.Run("Scenario 1: Successful Unfavoriting of an Article", func(t *testing.T) {

		mockDB, mock, err := sqlmock.New()
		assert.NoError(t, err)
		defer mockDB.Close()

		articleStore := &store.ArticleStore{Db: mockDB}
		userStore := &store.UserStore{Db: mockDB}

		logger := zerolog.New(os.Stdout)
		mockHandler := &Handler{logger: &logger, us: userStore, as: articleStore}

		authCtx := context.WithValue(context.Background(), authUserIDKey, uint(1))
		req := &proto.UnfavoriteArticleRequest{Slug: "1"}

		mock.ExpectQuery("SELECT (.+) FROM `articles` WHERE").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
		mock.ExpectQuery("SELECT (.+) FROM `users` WHERE").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
		mock.ExpectExec("DELETE FROM `favorites` WHERE").WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectExec("UPDATE `articles` SET").WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectQuery("SELECT (.+) FROM `users` WHERE").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

		response, err := mockHandler.UnfavoriteArticle(authCtx, req)

		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.False(t, response.Article.Favorited)

		if err = mock.ExpectationsWereMet(); err != nil {
			t.Errorf("unmet database expectations: %s", err)
		}
		t.Log("Test scenario for successful unfavoriting of an article passed.")
	})

	t.Run("Scenario 2: Unauthenticated User", func(t *testing.T) {

		mockDB, _, err := sqlmock.New()
		assert.NoError(t, err)
		defer mockDB.Close()

		articleStore := &store.ArticleStore{Db: mockDB}
		userStore := &store.UserStore{Db: mockDB}

		logger := zerolog.New(os.Stdout)
		mockHandler := &Handler{logger: &logger, us: userStore, as: articleStore}

		req := &proto.UnfavoriteArticleRequest{Slug: "1"}

		_, err = mockHandler.UnfavoriteArticle(context.Background(), req)

		assert.Error(t, err)
		assert.Equal(t, codes.Unauthenticated, status.Code(err))
		t.Log("Test scenario for unauthenticated user passed.")
	})

	t.Run("Scenario 3: Article Not Found", func(t *testing.T) {

		mockDB, mock, err := sqlmock.New()
		assert.NoError(t, err)
		defer mockDB.Close()

		articleStore := &store.ArticleStore{Db: mockDB}
		userStore := &store.UserStore{Db: mockDB}

		logger := zerolog.New(os.Stdout)
		mockHandler := &Handler{logger: &logger, us: userStore, as: articleStore}

		authCtx := context.WithValue(context.Background(), authUserIDKey, uint(1))
		req := &proto.UnfavoriteArticleRequest{Slug: "9999"}

		mock.ExpectQuery("SELECT (.+) FROM `articles` WHERE").WillReturnError(errors.New("record not found"))

		_, err = mockHandler.UnfavoriteArticle(authCtx, req)

		assert.Error(t, err)
		assert.Equal(t, codes.InvalidArgument, status.Code(err))
		t.Log("Test scenario for article not found passed.")
	})

	t.Run("Scenario 4: Invalid Slug Format", func(t *testing.T) {

		mockDB, _, err := sqlmock.New()
		assert.NoError(t, err)
		defer mockDB.Close()

		articleStore := &store.ArticleStore{Db: mockDB}
		userStore := &store.UserStore{Db: mockDB}

		logger := zerolog.New(os.Stdout)
		mockHandler := &Handler{logger: &logger, us: userStore, as: articleStore}

		authCtx := context.WithValue(context.Background(), authUserIDKey, uint(1))
		req := &proto.UnfavoriteArticleRequest{Slug: "invalid_slug"}

		_, err = mockHandler.UnfavoriteArticle(authCtx, req)

		assert.Error(t, err)
		assert.Equal(t, codes.InvalidArgument, status.Code(err))
		t.Log("Test scenario for invalid slug format passed.")
	})

	t.Run("Scenario 5: User Not Found", func(t *testing.T) {

		mockDB, mock, err := sqlmock.New()
		assert.NoError(t, err)
		defer mockDB.Close()

		articleStore := &store.ArticleStore{Db: mockDB}
		userStore := &store.UserStore{Db: mockDB}

		logger := zerolog.New(os.Stdout)
		mockHandler := &Handler{logger: &logger, us: userStore, as: articleStore}

		authCtx := context.WithValue(context.Background(), authUserIDKey, uint(9999))
		req := &proto.UnfavoriteArticleRequest{Slug: "1"}

		mock.ExpectQuery("SELECT (.+) FROM `users` WHERE").WillReturnError(errors.New("record not found"))

		_, err = mockHandler.UnfavoriteArticle(authCtx, req)

		assert.Error(t, err)
		assert.Equal(t, codes.NotFound, status.Code(err))
		t.Log("Test scenario for user not found passed.")
	})

	t.Run("Scenario 6: Failure to Remove Favorite", func(t *testing.T) {

		mockDB, mock, err := sqlmock.New()
		assert.NoError(t, err)
		defer mockDB.Close()

		articleStore := &store.ArticleStore{Db: mockDB}
		userStore := &store.UserStore{Db: mockDB}

		logger := zerolog.New(os.Stdout)
		mockHandler := &Handler{logger: &logger, us: userStore, as: articleStore}

		authCtx := context.WithValue(context.Background(), authUserIDKey, uint(1))
		req := &proto.UnfavoriteArticleRequest{Slug: "1"}

		mock.ExpectQuery("SELECT (.+) FROM `articles` WHERE").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
		mock.ExpectQuery("SELECT (.+) FROM `users` WHERE").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
		mock.ExpectExec("DELETE FROM `favorites` WHERE").WillReturnError(errors.New("database failure"))

		_, err = mockHandler.UnfavoriteArticle(authCtx, req)

		assert.Error(t, err)
		assert.Equal(t, codes.InvalidArgument, status.Code(err))
		t.Log("Test scenario for failure to remove favorite passed.")
	})

	t.Run("Scenario 7: Failure to Determine Following Status", func(t *testing.T) {

		mockDB, mock, err := sqlmock.New()
		assert.NoError(t, err)
		defer mockDB.Close()

		articleStore := &store.ArticleStore{Db: mockDB}
		userStore := &store.UserStore{Db: mockDB}

		logger := zerolog.New(os.Stdout)
		mockHandler := &Handler{logger: &logger, us: userStore, as: articleStore}

		authCtx := context.WithValue(context.Background(), authUserIDKey, uint(1))
		req := &proto.UnfavoriteArticleRequest{Slug: "1"}

		mock.ExpectQuery("SELECT (.+) FROM `articles` WHERE").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
		mock.ExpectQuery("SELECT (.+) FROM `users` WHERE").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
		mock.ExpectExec("DELETE FROM `favorites` WHERE").WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectExec("UPDATE `articles` SET").WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectQuery("SELECT (.+) FROM `users` WHERE").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
		mock.ExpectQuery("SELECT (.+) FROM `follows` WHERE").WillReturnError(errors.New("database failure"))

		_, err = mockHandler.UnfavoriteArticle(authCtx, req)

		assert.Error(t, err)
		assert.Equal(t, codes.NotFound, status.Code(err))
		t.Log("Test scenario for failure to determine following status passed.")
	})
}


/*
ROOST_METHOD_HASH=CreateArticle_64372fa1a8
ROOST_METHOD_SIG_HASH=CreateArticle_ce1c125740


 */
func TestHandlerCreateArticle(t *testing.T) {
	type args struct {
		ctx context.Context
		req *pb.CreateAritcleRequest
	}
	tests := []struct {
		name       string
		args       args
		setupMock  func(us *store.UserStore, as *store.ArticleStore)
		wantErr    bool
		wantErrMsg string
	}{

		{
			name: "successful article creation",
			args: args{
				ctx: context.TODO(),
				req: &pb.CreateAritcleRequest{
					Article: &pb.CreateAritcleRequest_Article{
						Title:       "Test Title",
						Description: "Test Description",
						Body:        "Test Body",
						TagList:     []string{"test"},
					},
				},
			},
			setupMock: func(us *store.UserStore, as *store.ArticleStore) {
				user := &model.User{Model: model.Model{ID: 1}, Username: "testuser"}
				us.On("GetByID", uint(1)).Return(user, nil)

				as.On("Create", &model.Article{}).
					Return(nil)

				us.On("IsFollowing", user, &model.User{Model: model.Model{ID: 1}}).
					Return(false, nil)
			},
			wantErr: false,
		},

		{
			name: "unauthenticated user",
			args: args{
				ctx: context.TODO(),
				req: &pb.CreateAritcleRequest{},
			},
			setupMock: func(_, _ *store.UserStore, _ *store.ArticleStore) {
				auth.GetUserID = func(ctx context.Context) (uint, error) {
					return 0, errors.New("unauthenticated")
				}
			},
			wantErr:    true,
			wantErrMsg: "rpc error: code = Unauthenticated desc = unauthenticated",
		},

		{
			name: "missing article fields",
			args: args{
				ctx: context.TODO(),
				req: &pb.CreateAritcleRequest{
					Article: &pb.CreateAritcleRequest_Article{},
				},
			},
			setupMock: func(us *store.UserStore, _ *store.ArticleStore) {
				user := &model.User{Model: model.Model{ID: 1}}
				us.On("GetByID", uint(1)).Return(user, nil)
			},
			wantErr:    true,
			wantErrMsg: "rpc error: code = InvalidArgument desc = validation error",
		},

		{
			name: "database failure on user retrieval",
			args: args{
				ctx: context.TODO(),
				req: &pb.CreateAritcleRequest{
					Article: &pb.CreateAritcleRequest_Article{},
				},
			},
			setupMock: func(us *store.UserStore, _ *store.ArticleStore) {
				us.On("GetByID", uint(1)).Return(nil, errors.New("database down"))
			},
			wantErr:    true,
			wantErrMsg: "rpc error: code = NotFound desc = user not found",
		},

		{
			name: "tag processing check",
			args: args{
				ctx: context.TODO(),
				req: &pb.CreateAritcleRequest{
					Article: &pb.CreateAritcleRequest_Article{
						Title:   "Title",
						TagList: []string{"tag1", "tag2"},
					},
				},
			},
			setupMock: func(us *store.UserStore, as *store.ArticleStore) {
				user := &model.User{Model: model.Model{ID: 1}}
				us.On("GetByID", uint(1)).Return(user, nil)
				as.On("Create", &model.Article{}).Return(nil)
			},
			wantErr: false,
		},

		{
			name: "fail to create article in database",
			args: args{
				ctx: context.TODO(),
				req: &pb.CreateAritcleRequest{
					Article: &pb.CreateAritcleRequest_Article{
						Title: "Test Title",
					},
				},
			},
			setupMock: func(us *store.UserStore, as *store.ArticleStore) {
				user := &model.User{Model: model.Model{ID: 1}}
				us.On("GetByID", uint(1)).Return(user, nil)
				as.On("Create", &model.Article{}).Return(errors.New("db failure"))
			},
			wantErr:    true,
			wantErrMsg: "rpc error: code = Canceled desc = Failed to create user.",
		},

		{
			name: "author following status resolution",
			args: args{
				ctx: context.TODO(),
				req: &pb.CreateAritcleRequest{
					Article: &pb.CreateAritcleRequest_Article{
						Title: "Title",
					},
				},
			},
			setupMock: func(us *store.UserStore, as *store.ArticleStore) {
				user := &model.User{Model: model.Model{ID: 1}, Username: "author"}
				us.On("GetByID", uint(1)).Return(user, nil)
				as.On("Create", &model.Article{}).Return(nil)
				us.On("IsFollowing", user, user).Return(true, nil)
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			db, mock, _ := sqlmock.New()

			us := store.NewUserStore(db)
			as := store.NewArticleStore(db)
			logger := zerolog.New(&mockWriter{})

			if tt.setupMock != nil {
				tt.setupMock(us, as)
			}

			h := &Handler{
				us:     us,
				as:     as,
				logger: &logger,
			}

			_, err := h.CreateArticle(tt.args.ctx, tt.args.req)

			if (err != nil) != tt.wantErr {
				t.Errorf("CreateArticle() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err != nil && err.Error() != tt.wantErrMsg {
				t.Errorf("CreateArticle() error message = %v, wantErrMsg %v", err.Error(), tt.wantErrMsg)
			}

			mock.ExpectationsWereMet()
		})
	}
}


/*
ROOST_METHOD_HASH=GetArticles_f87b10d80e
ROOST_METHOD_SIG_HASH=GetArticles_5d9fe7bf44


 */
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


/*
ROOST_METHOD_HASH=UpdateArticle_c5b82e271b
ROOST_METHOD_SIG_HASH=UpdateArticle_f36cc09d87


 */
func TestHandlerUpdateArticle(t *testing.T) {
	mockUserStore := new(MockUserStore)
	mockArticleStore := new(MockArticleStore)
	logger := zerolog.New(os.Stdout)

	handler := Handler{
		logger: &logger,
		us:     (*store.UserStore)(mockUserStore),
		as:     (*store.ArticleStore)(mockArticleStore),
	}

	testCases := []struct {
		scenario         string
		setupContext     func() context.Context
		setupMocks       func()
		req              *pb.UpdateArticleRequest
		expectedError    error
		expectedResponse *pb.ArticleResponse
	}{
		{
			scenario: "Successful Update of an Article",
			setupContext: func() context.Context {
				return context.WithValue(context.Background(), "userID", uint(1))
			},
			setupMocks: func() {
				mockUserStore.On("GetByID", uint(1)).Return(&model.User{ID: 1}, nil)
				mockArticleStore.On("GetByID", uint(123)).Return(&model.Article{
					ID:     123,
					Author: model.User{ID: 1},
				}, nil)
				mockArticleStore.On("Update", mock.Anything).Return(nil)
				mockUserStore.On("IsFollowing", mock.Anything, mock.Anything).Return(true, nil)
			},
			req: &pb.UpdateArticleRequest{
				Article: &pb.UpdateArticleRequest_Article{
					Slug:        "123",
					Title:       "Updated Title",
					Description: "Updated Description",
					Body:        "Updated Body",
				},
			},
			expectedError: nil,
			expectedResponse: &pb.ArticleResponse{
				Article: &pb.Article{
					Slug:           "123",
					Title:          "Updated Title",
					Description:    "Updated Description",
					Body:           "Updated Body",
					Favorited:      true,
					FavoritesCount: 0,
				},
			},
		},
		{
			scenario: "Unauthenticated User Error",
			setupContext: func() context.Context {
				return context.Background()
			},
			setupMocks: func() {
				mockUserStore.On("GetByID", uint(0)).Return(nil, errors.New("user not found"))
			},
			req: &pb.UpdateArticleRequest{
				Article: &pb.UpdateArticleRequest_Article{
					Slug:        "123",
					Title:       "Updated Title",
					Description: "Updated Description",
					Body:        "Updated Body",
				},
			},
			expectedError: status.Errorf(codes.Unauthenticated, "unauthenticated"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.scenario, func(t *testing.T) {
			tc.setupMocks()
			resp, err := handler.UpdateArticle(tc.setupContext(), tc.req)

			assert.Equal(t, tc.expectedError, err)
			if err == nil {
				assert.Equal(t, tc.expectedResponse.Article.Title, resp.Article.Title)
				assert.Equal(t, tc.expectedResponse.Article.Description, resp.Article.Description)
				assert.Equal(t, tc.expectedResponse.Article.Body, resp.Article.Body)
			}

			t.Logf("Scenario: %s - Execution completed.", tc.scenario)
		})
	}

	mockArticleStore.AssertExpectations(t)
	mockUserStore.AssertExpectations(t)
}

