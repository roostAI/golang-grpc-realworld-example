package handler

import (
	"context"
	"testing"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/golang/mock/gomock"
	"github.com/raahii/golang-grpc-realworld-example/auth"
	"github.com/raahii/golang-grpc-realworld-example/proto"
	"github.com/raahii/golang-grpc-realworld-example/store"
	"github.com/rs/zerolog"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
	"errors"
	"os"
	"github.com/raahii/golang-grpc-realworld-example/model"
	"github.com/stretchr/testify/assert"
)

type ExpectedQuery struct {
	queryBasedExpectation
	rows             driver.Rows
	delay            time.Duration
	rowsMustBeClosed bool
	rowsWereClosed   bool
}
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
	context    *testContext
}
type ExpectedExec struct {
	queryBasedExpectation
	result driver.Result
	delay  time.Duration
}
/*
ROOST_METHOD_HASH=ShowProfile_3cf6e3a9fd
ROOST_METHOD_SIG_HASH=ShowProfile_4679c3d9a4


 */
func TestHandlerShowProfile(t *testing.T) {

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	var logger zerolog.Logger

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	gormDB, err := gorm.Open("postgres", db)
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a gorm database connection", err)
	}

	userStore := &store.UserStore{DB: gormDB}
	handler := &Handler{logger: &logger, us: userStore}

	type TestCase struct {
		Name          string
		SetupContext  func() context.Context
		SetupMocks    func()
		ExpectedError codes.Code
	}

	tests := []TestCase{
		{
			Name: "Successful Profile Retrieval",
			SetupContext: func() context.Context {
				return auth.ContextWithUserID(context.Background(), uint(1))
			},
			SetupMocks: func() {

				mock.ExpectQuery("^SELECT .+ FROM users WHERE id = \\$1$").
					WithArgs(1).
					WillReturnRows(sqlmock.NewRows([]string{"id", "username", "bio", "image"}).
						AddRow(1, "current_user", "bio", "image"))

				mock.ExpectQuery("^SELECT .+ FROM users WHERE username = \\$1$").
					WithArgs("requested_user").
					WillReturnRows(sqlmock.NewRows([]string{"id", "username", "bio", "image"}).
						AddRow(2, "requested_user", "bio", "image"))

				mock.ExpectQuery("^SELECT count\\(\\*\\) FROM follows WHERE from_user_id = \\$1 AND to_user_id = \\$2$").
					WithArgs(1, 2).
					WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
			},
			ExpectedError: codes.OK,
		},
		{
			Name: "Unauthenticated Request Error",
			SetupContext: func() context.Context {

				return context.Background()
			},
			SetupMocks:    func() {},
			ExpectedError: codes.Unauthenticated,
		},
		{
			Name: "Current User Not Found Error",
			SetupContext: func() context.Context {
				return auth.ContextWithUserID(context.Background(), uint(1))
			},
			SetupMocks: func() {

				mock.ExpectQuery("^SELECT .+ FROM users WHERE id = \\$1$").
					WithArgs(1).
					WillReturnError(status.Error(codes.NotFound, "user not found"))
			},
			ExpectedError: codes.NotFound,
		},
		{
			Name: "Request User Not Found Error",
			SetupContext: func() context.Context {
				return auth.ContextWithUserID(context.Background(), uint(1))
			},
			SetupMocks: func() {
				mock.ExpectQuery("^SELECT .+ FROM users WHERE id = \\$1$").
					WithArgs(1).
					WillReturnRows(sqlmock.NewRows([]string{"id", "username", "bio", "image"}).
						AddRow(1, "current_user", "bio", "image"))

				mock.ExpectQuery("^SELECT .+ FROM users WHERE username = \\$1$").
					WithArgs("non_existent_user").
					WillReturnError(status.Error(codes.NotFound, "user was not found"))
			},
			ExpectedError: codes.NotFound,
		},
		{
			Name: "Error in Following Status",
			SetupContext: func() context.Context {
				return auth.ContextWithUserID(context.Background(), uint(1))
			},
			SetupMocks: func() {
				mock.ExpectQuery("^SELECT .+ FROM users WHERE id = \\$1$").
					WithArgs(1).
					WillReturnRows(sqlmock.NewRows([]string{"id", "username", "bio", "image"}).
						AddRow(1, "current_user", "bio", "image"))

				mock.ExpectQuery("^SELECT .+ FROM users WHERE username = \\$1$").
					WithArgs("requested_user").
					WillReturnRows(sqlmock.NewRows([]string{"id", "username", "bio", "image"}).
						AddRow(2, "requested_user", "bio", "image"))

				mock.ExpectQuery("^SELECT count\\(\\*\\) FROM follows WHERE from_user_id = \\$1 AND to_user_id = \\$2$").
					WithArgs(1, 2).
					WillReturnError(status.Error(codes.Internal, "internal error"))
			},
			ExpectedError: codes.Internal,
		},
	}

	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			tc.SetupMocks()

			ctx := tc.SetupContext()
			req := &proto.ShowProfileRequest{Username: "requested_user"}

			resp, err := handler.ShowProfile(ctx, req)

			if tc.ExpectedError == codes.OK {

				if err != nil {
					t.Fatalf("expected no error, got %v", err)
				}
				if resp == nil || resp.Profile == nil || resp.Profile.Username != "requested_user" {
					t.Fatalf("unexpected profile response: %v", resp)
				}
			} else {
				if status.Code(err) != tc.ExpectedError {
					t.Fatalf("expected error code %v, got %v", tc.ExpectedError, status.Code(err))
				}
				if resp != nil {
					t.Fatalf("expected no response, got %v", resp)
				}
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatalf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}


/*
ROOST_METHOD_HASH=FollowUser_36d65b5263
ROOST_METHOD_SIG_HASH=FollowUser_bf8ceb04bb


 */
func TestHandlerFollowUser(t *testing.T) {
	type testCase struct {
		name          string
		setupMocks    func(usMock sqlmock.Sqlmock)
		request       *pb.FollowRequest
		expectedError error
		expectedCode  codes.Code
	}

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("unexpected error opening sqlmock database connection: %s", err)
	}
	defer db.Close()

	userStore := &store.UserStore{db: db}
	logger := zerolog.New(nil).With().Logger()
	authFunc := auth.GetUserID

	handler := &Handler{
		logger: &logger,
		us:     userStore,
	}

	tests := []testCase{
		{
			name: "Successfully Follow Another User",
			setupMocks: func(usMock sqlmock.Sqlmock) {
				authFunc = func(ctx context.Context) (uint, error) {
					return 1, nil
				}

				usMock.ExpectQuery(`^SELECT \* FROM "users" WHERE id=\?`).
					WithArgs(1).
					WillReturnRows(sqlmock.NewRows([]string{"id", "username"}).AddRow(1, "currentuser"))

				usMock.ExpectQuery(`^SELECT \* FROM "users" WHERE username=\?`).
					WithArgs("targetuser").
					WillReturnRows(sqlmock.NewRows([]string{"id", "username"}).AddRow(2, "targetuser"))

				usMock.ExpectExec(`^INSERT INTO "user_follows" `).
					WithArgs(1, 2).
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
			request:       &pb.FollowRequest{Username: "targetuser"},
			expectedError: nil,
		},
		{
			name: "Fail Due to Self-Follow Attempt",
			setupMocks: func(usMock sqlmock.Sqlmock) {
				authFunc = func(ctx context.Context) (uint, error) {
					return 1, nil
				}

				usMock.ExpectQuery(`^SELECT \* FROM "users" WHERE id=\?`).
					WithArgs(1).
					WillReturnRows(sqlmock.NewRows([]string{"id", "username"}).AddRow(1, "yourself"))
			},
			request:       &pb.FollowRequest{Username: "yourself"},
			expectedError: status.Error(codes.InvalidArgument, "cannot follow yourself"),
			expectedCode:  codes.InvalidArgument,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()

			tc.setupMocks(mock)

			resp, err := handler.FollowUser(ctx, tc.request)

			if tc.expectedError != nil {
				if status.Code(err) != tc.expectedCode {
					t.Errorf("Expected error code: %v, got: %v", tc.expectedCode, status.Code(err))
				}
				if err == nil || err.Error() != tc.expectedError.Error() {
					t.Errorf("Expected error: %v, got: %v", tc.expectedError, err)
				}
			} else if resp.GetProfile().Username != tc.request.GetUsername() {
				t.Errorf("Expected username in profile: %v, got: %v", tc.request.GetUsername(), resp.GetProfile().Username)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatalf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}


/*
ROOST_METHOD_HASH=UnfollowUser_843a2807ea
ROOST_METHOD_SIG_HASH=UnfollowUser_a64840f937


 */
func TestHandlerUnfollowUser(t *testing.T) {
	tests := []struct {
		name          string
		setupMocks    func(us *store.UserStore, mock sqlmock.Sqlmock)
		contextUserID uint
		request       *pb.UnfollowRequest
		expectedResp  *pb.ProfileResponse
		expectedErr   error
	}{
		{
			name: "Successfully Unfollow a User",
			setupMocks: func(us *store.UserStore, mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT \\* FROM \"users\" WHERE id = \\$1").
					WithArgs(1).
					WillReturnRows(sqlmock.NewRows([]string{"id", "username"}).AddRow(1, "current_user"))
				mock.ExpectQuery("SELECT \\* FROM \"users\" WHERE username = \\$1").
					WithArgs("target_user").
					WillReturnRows(sqlmock.NewRows([]string{"id", "username"}).AddRow(2, "target_user"))
				mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM \"follows\"").
					WithArgs(1, 2).
					WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
				mock.ExpectExec("DELETE FROM \"follows\" WHERE").
					WithArgs(2).WillReturnResult(sqlmock.NewResult(1, 1))
			},
			contextUserID: 1,
			request:       &pb.UnfollowRequest{Username: "target_user"},
			expectedResp:  &pb.ProfileResponse{Profile: &pb.Profile{Username: "target_user", Following: false}},
			expectedErr:   nil,
		},
		{
			name: "Fail due to Unauthenticated User",
			setupMocks: func(us *store.UserStore, mock sqlmock.Sqlmock) {

			},
			contextUserID: 0,
			request:       &pb.UnfollowRequest{Username: "target_user"},
			expectedResp:  nil,
			expectedErr:   status.Error(codes.Unauthenticated, "unauthenticated"),
		},
		{
			name: "Current User Not Found",
			setupMocks: func(us *store.UserStore, mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT \\* FROM \"users\" WHERE id = \\$1").
					WithArgs(1).
					WillReturnError(errors.New("user not found"))
			},
			contextUserID: 1,
			request:       &pb.UnfollowRequest{Username: "target_user"},
			expectedResp:  nil,
			expectedErr:   status.Error(codes.NotFound, "user not found"),
		},
		{
			name: "User Tries to Unfollow Themselves",
			setupMocks: func(us *store.UserStore, mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT \\* FROM \"users\" WHERE id = \\$1").
					WithArgs(1).
					WillReturnRows(sqlmock.NewRows([]string{"id", "username"}).AddRow(1, "current_user"))
			},
			contextUserID: 1,
			request:       &pb.UnfollowRequest{Username: "current_user"},
			expectedResp:  nil,
			expectedErr:   status.Error(codes.InvalidArgument, "cannot follow yourself"),
		},
		{
			name: "Target User Not Found",
			setupMocks: func(us *store.UserStore, mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT \\* FROM \"users\" WHERE id = \\$1").
					WithArgs(1).
					WillReturnRows(sqlmock.NewRows([]string{"id", "username"}).AddRow(1, "current_user"))
				mock.ExpectQuery("SELECT \\* FROM \"users\" WHERE username = \\$1").
					WithArgs("target_user").
					WillReturnError(errors.New("user was not found"))
			},
			contextUserID: 1,
			request:       &pb.UnfollowRequest{Username: "target_user"},
			expectedResp:  nil,
			expectedErr:   status.Error(codes.NotFound, "user was not found"),
		},
		{
			name: "User Not Currently Following Target User",
			setupMocks: func(us *store.UserStore, mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT \\* FROM \"users\" WHERE id = \\$1").
					WithArgs(1).
					WillReturnRows(sqlmock.NewRows([]string{"id", "username"}).AddRow(1, "current_user"))
				mock.ExpectQuery("SELECT \\* FROM \"users\" WHERE username = \\$1").
					WithArgs("target_user").
					WillReturnRows(sqlmock.NewRows([]string{"id", "username"}).AddRow(2, "target_user"))
				mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM \"follows\"").
					WithArgs(1, 2).
					WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
			},
			contextUserID: 1,
			request:       &pb.UnfollowRequest{Username: "target_user"},
			expectedResp:  nil,
			expectedErr:   status.Error(codes.Unauthenticated, "you are not following the user"),
		},
		{
			name: "Error during Unfollow Operation",
			setupMocks: func(us *store.UserStore, mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT \\* FROM \"users\" WHERE id = \\$1").
					WithArgs(1).
					WillReturnRows(sqlmock.NewRows([]string{"id", "username"}).AddRow(1, "current_user"))
				mock.ExpectQuery("SELECT \\* FROM \"users\" WHERE username = \\$1").
					WithArgs("target_user").
					WillReturnRows(sqlmock.NewRows([]string{"id", "username"}).AddRow(2, "target_user"))
				mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM \"follows\"").
					WithArgs(1, 2).
					WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
				mock.ExpectExec("DELETE FROM \"follows\" WHERE").
					WithArgs(2).WillReturnError(errors.New("failed to unfollow user"))
			},
			contextUserID: 1,
			request:       &pb.UnfollowRequest{Username: "target_user"},
			expectedResp:  nil,
			expectedErr:   status.Error(codes.Aborted, "failed to unfollow user"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			assert.NoError(t, err)
			defer db.Close()
			us := store.UserStore{DB: db}
			logger := zerolog.New(os.Stdout)
			h := Handler{logger: &logger, us: &us}

			tt.setupMocks(&us, mock)

			ctx := context.Background()
			if tt.contextUserID != 0 {
				ctx = context.WithValue(ctx, auth.UserIDKey, tt.contextUserID)
			}

			resp, err := h.UnfollowUser(ctx, tt.request)
			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedErr, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedResp, resp)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("expectations were not met: %v", err)
			}
		})
	}
}

