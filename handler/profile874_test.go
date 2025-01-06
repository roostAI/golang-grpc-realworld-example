package handler

import (
	"context"
	"testing"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/raahii/golang-grpc-realworld-example/auth"
	"github.com/raahii/golang-grpc-realworld-example/proto"
	"github.com/raahii/golang-grpc-realworld-example/store"
	"github.com/rs/zerolog"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
	"errors"
	"github.com/raahii/golang-grpc-realworld-example/model"
	"os"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/postgres"
)


/*
ROOST_METHOD_HASH=ShowProfile_3cf6e3a9fd
ROOST_METHOD_SIG_HASH=ShowProfile_4679c3d9a4


 */
func TestHandlerShowProfile(t *testing.T) {
	t.Parallel()
	db, mock, _ := sqlmock.New()
	gormDB, _ := gorm.Open("postgres", db)
	defer db.Close()

	userStore := &store.UserStore{Db: gormDB}
	articleStore := &store.ArticleStore{Db: gormDB}

	logger := &zerolog.Logger{}
	handler := &Handler{
		logger: logger,
		us:     userStore,
		as:     articleStore,
	}

	validUserID := uint(1)
	otherUserID := uint(2)
	validUsername := "validusername"

	t.Run("Scenario 1: Valid profile retrieval", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), "userID", validUserID)
		req := &proto.ShowProfileRequest{Username: validUsername}

		mock.ExpectQuery("SELECT (.+) FROM users WHERE id = ?").
			WithArgs(validUserID).
			WillReturnRows(sqlmock.NewRows([]string{"ID"}).AddRow(validUserID))

		mock.ExpectQuery("SELECT (.+) FROM users WHERE username = ?").
			WithArgs(validUsername).
			WillReturnRows(sqlmock.NewRows([]string{"ID"}).AddRow(otherUserID))

		mock.ExpectQuery("SELECT COUNT").
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

		resp, err := handler.ShowProfile(ctx, req)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if resp.Profile == nil || resp.Profile.Username != validUsername {
			t.Fatalf("Expected profile with username %s, got %v", validUsername, resp.Profile)
		}
		t.Logf("Successfully retrieved profile: %+v", resp.Profile)
	})

	t.Run("Scenario 2: Unauthenticated request", func(t *testing.T) {
		ctx := context.Background()

		req := &proto.ShowProfileRequest{Username: validUsername}

		_, err := handler.ShowProfile(ctx, req)
		if err == nil || status.Code(err) != codes.Unauthenticated {
			t.Fatalf("Expected Unauthenticated error, got %v", err)
		}
		t.Log("Correctly identified unauthenticated request")
	})

	t.Run("Scenario 3: Current user not found", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), "userID", validUserID)
		req := &proto.ShowProfileRequest{Username: validUsername}

		mock.ExpectQuery("SELECT (.+) FROM users WHERE id = ?").
			WithArgs(validUserID).
			WillReturnError(status.Error(codes.NotFound, "user not found"))

		_, err := handler.ShowProfile(ctx, req)
		if err == nil || status.Code(err) != codes.NotFound {
			t.Fatalf("Expected NotFound error, got %v", err)
		}
		t.Log("Correctly handled current user not found")
	})

	t.Run("Scenario 4: Requested user not found", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), "userID", validUserID)
		req := &proto.ShowProfileRequest{Username: validUsername}

		mock.ExpectQuery("SELECT (.+) FROM users WHERE id = ?").
			WithArgs(validUserID).
			WillReturnRows(sqlmock.NewRows([]string{"ID"}).AddRow(validUserID))

		mock.ExpectQuery("SELECT (.+) FROM users WHERE username = ?").
			WithArgs(validUsername).
			WillReturnError(status.Error(codes.NotFound, "user was not found"))

		_, err := handler.ShowProfile(ctx, req)
		if err == nil || status.Code(err) != codes.NotFound {
			t.Fatalf("Expected NotFound error for requested user, got %v", err)
		}
		t.Log("Correctly handled requested user not found")
	})

	t.Run("Scenario 5: Error retrieving following status", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), "userID", validUserID)
		req := &proto.ShowProfileRequest{Username: validUsername}

		mock.ExpectQuery("SELECT (.+) FROM users WHERE id = ?").
			WithArgs(validUserID).
			WillReturnRows(sqlmock.NewRows([]string{"ID"}).AddRow(validUserID))

		mock.ExpectQuery("SELECT (.+) FROM users WHERE username = ?").
			WithArgs(validUsername).
			WillReturnRows(sqlmock.NewRows([]string{"ID"}).AddRow(otherUserID))

		mock.ExpectQuery("SELECT COUNT").
			WillReturnError(status.Error(codes.Internal, "internal server error"))

		_, err := handler.ShowProfile(ctx, req)
		if err == nil || status.Code(err) != codes.Internal {
			t.Fatalf("Expected Internal server error, got %v", err)
		}
		t.Log("Correctly handled error in retrieving following status")
	})

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unmet expectations: %s", err)
	}
}


/*
ROOST_METHOD_HASH=FollowUser_36d65b5263
ROOST_METHOD_SIG_HASH=FollowUser_bf8ceb04bb


 */
func TestHandlerFollowUser(t *testing.T) {
	type args struct {
		ctx context.Context
		req *proto.FollowRequest
	}

	tests := []struct {
		name      string
		args      args
		setupMock func(userStore *UserStoreMock)
		wantErr   error
	}{
		{
			name: "Follow a Valid User Successfully",
			args: args{
				ctx: context.Background(),
				req: &proto.FollowRequest{Username: "validTargetUser"},
			},
			setupMock: func(userStore *UserStoreMock) {
				userStore.On("GetByID", uint(1)).Return(&model.User{Username: "currentUser"}, nil)
				userStore.On("GetByUsername", "validTargetUser").Return(&model.User{Username: "validTargetUser"}, nil)
				userStore.On("Follow", &model.User{Username: "currentUser"}, &model.User{Username: "validTargetUser"}).Return(nil)
			},
			wantErr: nil,
		},
		{
			name: "Attempt to Follow Yourself",
			args: args{
				ctx: context.Background(),
				req: &proto.FollowRequest{Username: "currentUser"},
			},
			setupMock: func(userStore *UserStoreMock) {
				userStore.On("GetByID", uint(1)).Return(&model.User{Username: "currentUser"}, nil)
			},
			wantErr: status.Error(codes.InvalidArgument, "cannot follow yourself"),
		},
		{
			name: "Unauthenticated User Attempt",
			args: args{
				ctx: context.Background(),
				req: &proto.FollowRequest{},
			},
			setupMock: func(userStore *UserStoreMock) {
				auth.GetUserID = func(ctx context.Context) (uint, error) {
					return 0, errors.New("unauthenticated")
				}
			},
			wantErr: status.Errorf(codes.Unauthenticated, "unauthenticated"),
		},
		{
			name: "Follow a Non-existent User",
			args: args{
				ctx: context.Background(),
				req: &proto.FollowRequest{Username: "nonExistentUser"},
			},
			setupMock: func(userStore *UserStoreMock) {
				userStore.On("GetByID", uint(1)).Return(&model.User{Username: "currentUser"}, nil)
				userStore.On("GetByUsername", "nonExistentUser").Return(nil, errors.New("user not found"))
			},
			wantErr: status.Error(codes.NotFound, "user was not found"),
		},
		{
			name: "Database Error on Current User Lookup",
			args: args{
				ctx: context.Background(),
				req: &proto.FollowRequest{Username: "validTargetUser"},
			},
			setupMock: func(userStore *UserStoreMock) {
				userStore.On("GetByID", uint(1)).Return(nil, errors.New("db error"))
			},
			wantErr: status.Error(codes.NotFound, "user not found"),
		},
		{
			name: "Database Error on Follow Operation",
			args: args{
				ctx: context.Background(),
				req: &proto.FollowRequest{Username: "validTargetUser"},
			},
			setupMock: func(userStore *UserStoreMock) {
				userStore.On("GetByID", uint(1)).Return(&model.User{Username: "currentUser"}, nil)
				userStore.On("GetByUsername", "validTargetUser").Return(&model.User{Username: "validTargetUser"}, nil)
				userStore.On("Follow", &model.User{Username: "currentUser"}, &model.User{Username: "validTargetUser"}).Return(errors.New("db error"))
			},
			wantErr: status.Error(codes.Aborted, "failed to follow user"),
		},
		{
			name: "Empty Username in FollowRequest",
			args: args{
				ctx: context.Background(),
				req: &proto.FollowRequest{Username: ""},
			},
			setupMock: func(userStore *UserStoreMock) {
				userStore.On("GetByID", uint(1)).Return(&model.User{Username: "currentUser"}, nil)
			},
			wantErr: status.Error(codes.InvalidArgument, "cannot follow yourself"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("failed to open sqlmock database connection: %v", err)
			}
			defer db.Close()

			userStore := &UserStoreMock{Sqlmock: mock}
			tt.setupMock(userStore)

			handler := &Handler{
				logger: newMockLogger(),
				us:     userStore,
			}

			_, err = handler.FollowUser(tt.args.ctx, tt.args.req)
			if err != nil && tt.wantErr == nil || err == nil && tt.wantErr != nil || (err != nil && err.Error() != tt.wantErr.Error()) {
				t.Errorf("FollowUser() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func newMockLogger() *zerolog.Logger {
	logger := zerolog.Nop()
	return &logger
}


/*
ROOST_METHOD_HASH=UnfollowUser_843a2807ea
ROOST_METHOD_SIG_HASH=UnfollowUser_a64840f937


 */
func TestHandlerUnfollowUser(t *testing.T) {

	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn: db,
	}), &gorm.Config{})
	assert.NoError(t, err)

	mockUserStore := &store.UserStore{DB: gormDB}

	consoleWriter := zerolog.ConsoleWriter{Out: os.Stdout}
	mockLogger := zerolog.New(consoleWriter)

	h := &Handler{
		logger: &mockLogger,
		us:     mockUserStore,
	}

	type testCase struct {
		name       string
		setupMocks func()
		ctx        context.Context
		req        *pb.UnfollowRequest
		wantErr    codes.Code
	}

	tests := []testCase{
		{
			name: "Successfully Unfollow a User",
			setupMocks: func() {
				auth.GetUserID = func(ctx context.Context) (uint, error) {
					return 1, nil
				}
				mock.ExpectQuery("^SELECT (.+) FROM \"users\" WHERE id = ? LIMIT 1$").
					WithArgs(1).
					WillReturnRows(sqlmock.NewRows([]string{"id", "username"}).AddRow(1, "currentUser"))

				mock.ExpectQuery("^SELECT (.+) FROM \"users\" WHERE username = ? LIMIT 1$").
					WithArgs("anotherUser").
					WillReturnRows(sqlmock.NewRows([]string{"id", "username"}).AddRow(2, "anotherUser"))

				mock.ExpectQuery("^SELECT (.+) FROM \"follows\"").
					WithArgs(1, 2).
					WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

				mock.ExpectExec("^DELETE FROM \"follows\"").
					WithArgs(1, 2).
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
			ctx:     context.Background(),
			req:     &pb.UnfollowRequest{Username: "anotherUser"},
			wantErr: codes.OK,
		},
		{
			name: "Unauthenticated User",
			setupMocks: func() {
				auth.GetUserID = func(ctx context.Context) (uint, error) {
					return 0, status.Error(codes.Unauthenticated, "unauthenticated")
				}
			},
			ctx:     context.Background(),
			req:     &pb.UnfollowRequest{Username: "anotherUser"},
			wantErr: codes.Unauthenticated,
		},
		{
			name: "Current User Not Found",
			setupMocks: func() {
				auth.GetUserID = func(ctx context.Context) (uint, error) {
					return 1, nil
				}
				mock.ExpectQuery("^SELECT (.+) FROM \"users\" WHERE id = ? LIMIT 1$").
					WithArgs(1).
					WillReturnError(status.Error(codes.NotFound, "user not found"))
			},
			ctx:     context.Background(),
			req:     &pb.UnfollowRequest{Username: "anotherUser"},
			wantErr: codes.NotFound,
		},
		{
			name: "Attempt to Unfollow Self",
			setupMocks: func() {
				auth.GetUserID = func(ctx context.Context) (uint, error) {
					return 1, nil
				}
				mock.ExpectQuery("^SELECT (.+) FROM \"users\" WHERE id = ? LIMIT 1$").
					WithArgs(1).
					WillReturnRows(sqlmock.NewRows([]string{"id", "username"}).AddRow(1, "currentUser"))
			},
			ctx:     context.Background(),
			req:     &pb.UnfollowRequest{Username: "currentUser"},
			wantErr: codes.InvalidArgument,
		},
		{
			name: "Request User Not Found",
			setupMocks: func() {
				auth.GetUserID = func(ctx context.Context) (uint, error) {
					return 1, nil
				}
				mock.ExpectQuery("^SELECT (.+) FROM \"users\" WHERE id = ? LIMIT 1$").
					WithArgs(1).
					WillReturnRows(sqlmock.NewRows([]string{"id", "username"}).AddRow(1, "currentUser"))

				mock.ExpectQuery("^SELECT (.+) FROM \"users\" WHERE username = ? LIMIT 1$").
					WithArgs("nonexistentUser").
					WillReturnError(status.Error(codes.NotFound, "user not found"))
			},
			ctx:     context.Background(),
			req:     &pb.UnfollowRequest{Username: "nonexistentUser"},
			wantErr: codes.NotFound,
		},
		{
			name: "Current User Not Following Request User",
			setupMocks: func() {
				auth.GetUserID = func(ctx context.Context) (uint, error) {
					return 1, nil
				}
				mock.ExpectQuery("^SELECT (.+) FROM \"users\" WHERE id = ? LIMIT 1$").
					WithArgs(1).
					WillReturnRows(sqlmock.NewRows([]string{"id", "username"}).AddRow(1, "currentUser"))

				mock.ExpectQuery("^SELECT (.+) FROM \"users\" WHERE username = ? LIMIT 1$").
					WithArgs("anotherUser").
					WillReturnRows(sqlmock.NewRows([]string{"id", "username"}).AddRow(2, "anotherUser"))

				mock.ExpectQuery("^SELECT (.+) FROM \"follows\"").
					WithArgs(1, 2).
					WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
			},
			ctx:     context.Background(),
			req:     &pb.UnfollowRequest{Username: "anotherUser"},
			wantErr: codes.Unauthenticated,
		},
		{
			name: "Unfollow Operation Fails",
			setupMocks: func() {
				auth.GetUserID = func(ctx context.Context) (uint, error) {
					return 1, nil
				}
				mock.ExpectQuery("^SELECT (.+) FROM \"users\" WHERE id = ? LIMIT 1$").
					WithArgs(1).
					WillReturnRows(sqlmock.NewRows([]string{"id", "username"}).AddRow(1, "currentUser"))

				mock.ExpectQuery("^SELECT (.+) FROM \"users\" WHERE username = ? LIMIT 1$").
					WithArgs("anotherUser").
					WillReturnRows(sqlmock.NewRows([]string{"id", "username"}).AddRow(2, "anotherUser"))

				mock.ExpectQuery("^SELECT (.+) FROM \"follows\"").
					WithArgs(1, 2).
					WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

				mock.ExpectExec("^DELETE FROM \"follows\"").
					WithArgs(1, 2).
					WillReturnError(status.Error(codes.Aborted, "unfollow operation failed"))
			},
			ctx:     context.Background(),
			req:     &pb.UnfollowRequest{Username: "anotherUser"},
			wantErr: codes.Aborted,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()

			resp, err := h.UnfollowUser(tt.ctx, tt.req)
			if tt.wantErr != codes.OK {
				assert.Error(t, err)
				assert.Nil(t, resp)
				if st, ok := status.FromError(err); ok {
					assert.Equal(t, tt.wantErr, st.Code())
				} else {
					t.Fatalf("expected grpc status error, got: %v", err)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

