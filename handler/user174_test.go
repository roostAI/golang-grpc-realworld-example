package handler

import (
	"context"
	"testing"
	"fmt"
	"github.com/raahii/golang-grpc-realworld-example/auth"
	"github.com/raahii/golang-grpc-realworld-example/model"
	"github.com/raahii/golang-grpc-realworld-example/proto"
	"github.com/raahii/golang-grpc-realworld-example/store"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"errors"
	"os"
	"github.com/DATA-DOG/go-sqlmock"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"github.com/golang/mock/gomock"
)


/*
ROOST_METHOD_HASH=CurrentUser_e3fa631d55
ROOST_METHOD_SIG_HASH=CurrentUser_29413339e9


 */
func TestHandlerCurrentUser(t *testing.T) {
	originalGenerateToken := auth.GenerateToken
	defer func() { auth.GenerateToken = originalGenerateToken }()

	originalGetUserID := auth.GetUserID
	defer func() { auth.GetUserID = originalGetUserID }()

	tests := []struct {
		name              string
		setup             func(mockUserStore *MockUserStore, logger *zerolog.Logger)
		expectedErrorCode codes.Code
		expectedUser      *proto.User
	}{
		{
			name: "Successful Retrieval of Current User",
			setup: func(mockUserStore *MockUserStore, logger *zerolog.Logger) {
				auth.GenerateToken = mockGenerateToken
				auth.GetUserID = func(ctx context.Context) (uint, error) {
					return 1, nil
				}
				mockUser := &model.User{
					Email:    "user@example.com",
					Username: "testuser",
				}
				mockUserStore.On("GetByID", uint(1)).Return(mockUser, nil)
			},
			expectedErrorCode: codes.OK,
			expectedUser:      &proto.User{Email: "user@example.com", Username: "testuser", Token: "valid_token"},
		},
		{
			name: "Unauthenticated Request",
			setup: func(mockUserStore *MockUserStore, logger *zerolog.Logger) {
				auth.GetUserID = func(ctx context.Context) (uint, error) {
					return 0, fmt.Errorf("token not found")
				}
			},
			expectedErrorCode: codes.Unauthenticated,
			expectedUser:      nil,
		},
		{
			name: "User Not Found",
			setup: func(mockUserStore *MockUserStore, logger *zerolog.Logger) {
				auth.GetUserID = func(ctx context.Context) (uint, error) { return 1, nil }
				mockUserStore.On("GetByID", uint(1)).Return(nil, fmt.Errorf("user not found"))
			},
			expectedErrorCode: codes.NotFound,
			expectedUser:      nil,
		},
		{
			name: "Token Generation Failure",
			setup: func(mockUserStore *MockUserStore, logger *zerolog.Logger) {
				auth.GenerateToken = mockGenerateToken
				auth.GetUserID = func(ctx context.Context) (uint, error) {
					return 2, nil
				}
				mockUser := &model.User{
					Email:    "user2@example.com",
					Username: "testuser2",
				}
				mockUserStore.On("GetByID", uint(2)).Return(mockUser, nil)
			},
			expectedErrorCode: codes.Aborted,
			expectedUser:      nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockUserStore := &MockUserStore{}
			var logger zerolog.Logger

			test.setup(mockUserStore, &logger)

			h := &Handler{
				logger: &logger,
				us:     mockUserStore,
			}

			resp, err := h.CurrentUser(context.Background(), &proto.Empty{})

			if test.expectedErrorCode == codes.OK {
				assert.NoError(t, err)
				assert.Equal(t, test.expectedUser, resp.User)
			} else {
				st, _ := status.FromError(err)
				assert.Nil(t, resp)
				assert.Equal(t, test.expectedErrorCode, st.Code())
			}

			mockUserStore.AssertExpectations(t)
		})
	}
}

func mockGenerateToken(userID uint) (string, error) {
	switch userID {
	case 1:
		return "valid_token", nil
	default:
		return "", fmt.Errorf("failed to generate token")
	}
}


/*
ROOST_METHOD_HASH=LoginUser_079a321a92
ROOST_METHOD_SIG_HASH=LoginUser_e7df23a6bd


 */
func TestHandlerLoginUser(t *testing.T) {
	mockDB, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock db: %v", err)
	}
	sqlDB, _ := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})

	logger := zerolog.New(zerolog.ConsoleWriter{Out: os.Stdout})
	us := &store.UserStore{DB: sqlDB}

	type testCase struct {
		name          string
		req           *pb.LoginUserRequest
		wantErr       bool
		expectedError error
	}

	testCases := []testCase{
		{
			name: "Successful Login with Valid Credentials",
			req: &pb.LoginUserRequest{
				User: &pb.LoginUserRequest_User{
					Email:    "valid@example.com",
					Password: "correctpassword",
				},
			},
			wantErr: false,
		},
		{
			name: "Invalid Email Error",
			req: &pb.LoginUserRequest{
				User: &pb.LoginUserRequest_User{
					Email:    "invalid@example.com",
					Password: "somepassword",
				},
			},
			wantErr: true,
			expectedError: status.Error(codes.InvalidArgument,
				"invalid email or password"),
		},
		{
			name: "Invalid Password Error",
			req: &pb.LoginUserRequest{
				User: &pb.LoginUserRequest_User{
					Email:    "valid@example.com",
					Password: "wrongpassword",
				},
			},
			wantErr: true,
			expectedError: status.Error(codes.InvalidArgument,
				"invalid email or password"),
		},
		{
			name: "Internal Server Error on Token Generation",
			req: &pb.LoginUserRequest{
				User: &pb.LoginUserRequest_User{
					Email:    "valid@example.com",
					Password: "correctpassword",
				},
			},
			wantErr:       true,
			expectedError: status.Error(codes.Aborted, "internal server error"),
		},
		{
			name:    "Empty Login Request",
			req:     nil,
			wantErr: true,
			expectedError: status.Error(codes.InvalidArgument,
				"invalid email or password"),
		},
	}

	originalGetByEmail := us.GetByEmail
	defer func() { us.GetByEmail = originalGetByEmail }()
	us.GetByEmail = func(email string) (*model.User, error) {
		if email == "valid@example.com" {
			return &model.User{
				Email:    email,
				Password: "$2y$12$somethingencrypted",
			}, nil
		}
		return nil, errors.New("user not found")
	}

	originalCheckPassword := model.User{}.CheckPassword
	defer func() { model.User{}.CheckPassword = originalCheckPassword }()
	model.User{}.CheckPassword = func(u *model.User, plain string) bool {
		return plain == "correctpassword"
	}

	originalGenerateToken := auth.GenerateToken
	defer func() { auth.GenerateToken = originalGenerateToken }()
	auth.GenerateToken = func(id uint) (string, error) {
		if id == 1 {
			return "", errors.New("token generation failure")
		}
		return "sometoken", nil
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			h := &Handler{
				logger: &logger,
				us:     us,
			}

			_, err := h.LoginUser(context.Background(), tc.req)
			if (err != nil) != tc.wantErr {
				t.Errorf("unexpected error: %v, wantErr: %v", err, tc.wantErr)
			}

			if tc.expectedError != nil && err.Error() != tc.expectedError.Error() {
				t.Errorf("expected error: %v, got: %v", tc.expectedError, err)
			}
		})
	}
}


/*
ROOST_METHOD_HASH=CreateUser_f2f8a1c84a
ROOST_METHOD_SIG_HASH=CreateUser_a3af3934da


 */
func TestHandlerCreateUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	_, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer mock.Close()

	mockUserStore := &MockUserStore{mock: ctrl}

	logger := zerolog.Nop()
	handler := &Handler{

		logger: &logger,
		us:     mockUserStore,
		as:     &store.ArticleStore{},
	}
	validUserReq := &pb.CreateUserRequest{
		User: &pb.CreateUserRequest_User{
			Username: "validusername",
			Email:    "valid@example.com",
			Password: "ValidPassword123!",
		},
	}

	tests := []struct {
		name      string
		req       *pb.CreateUserRequest
		setup     func()
		wantErr   bool
		errorCode codes.Code
	}{
		{
			name: "Create a User Successfully",
			req:  validUserReq,
			setup: func() {
				mockUserStore.EXPECT().Create(gomock.Any()).Return(nil)

				auth.GenerateToken = func(id uint) (string, error) {
					return "valid-token", nil
				}
			},
			wantErr: false,
		},
		{
			name: "Validation Error on User Creation",
			req: &pb.CreateUserRequest{
				User: &pb.CreateUserRequest_User{
					Email:    "invalid-email",
					Password: "short",
				},
			},

			setup:     func() {},
			wantErr:   true,
			errorCode: codes.InvalidArgument,
		},
		{
			name: "Password Hashing Fails",
			req:  validUserReq,
			setup: func() {
				auth.GenerateToken = func(id uint) (string, error) {
					return "", errors.New("hashing error")
				}
			},
			wantErr:   true,
			errorCode: codes.Aborted,
		},
		{
			name: "Database Error during User Creation",
			req:  validUserReq,
			setup: func() {
				mockUserStore.EXPECT().Create(gomock.Any()).Return(errors.New("db error"))
			},
			wantErr:   true,
			errorCode: codes.Canceled,
		},
		{
			name: "Token Generation Failure",
			req:  validUserReq,
			setup: func() {
				mockUserStore.EXPECT().Create(gomock.Any()).Return(nil)
				auth.GenerateToken = func(id uint) (string, error) {
					return "", errors.New("token generation error")
				}
			},
			wantErr:   true,
			errorCode: codes.Aborted,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()

			resp, err := handler.CreateUser(context.Background(), tt.req)

			if tt.wantErr {
				assert.Nil(t, resp)
				assert.Error(t, err)
				assert.Equal(t, tt.errorCode, status.Code(err))
			} else {
				assert.NotNil(t, resp)
				assert.Equal(t, "validusername", resp.User.Username)
				assert.NoError(t, err)
			}
		})
	}
}


/*
ROOST_METHOD_HASH=UpdateUser_6fa4ecf979
ROOST_METHOD_SIG_HASH=UpdateUser_883937d25b


 */
func TestHandlerUpdateUser(t *testing.T) {
	type mockBehavior func(us *mockStore, userID uint, user *model.User)

	testCases := []struct {
		name         string
		prepareMocks mockBehavior
		contextAuth  func(ctx context.Context) (uint, error)
		req          *proto.UpdateUserRequest
		want         *proto.UserResponse
		wantErr      codes.Code
	}{
		{
			name: "Successful User Update",
			prepareMocks: func(us *mockStore, userID uint, user *model.User) {
				us.On("GetByID", userID).Return(user, nil).Once()
				us.On("Update", user).Return(nil).Once()
			},
			contextAuth: func(ctx context.Context) (uint, error) {
				return 1, nil
			},
			req: &proto.UpdateUserRequest{
				User: &proto.UpdateUserRequest_User{
					Username: "newusername",
					Email:    "newemail@example.com",
					Bio:      "newbio",
					Image:    "newimage.png",
					Password: "newpassword",
				},
			},
			want: &proto.UserResponse{
				User: &proto.User{
					Email:    "newemail@example.com",
					Username: "newusername",
					Bio:      "newbio",
					Image:    "newimage.png",
					Token:    "dummy-token",
				},
			},
			wantErr: codes.OK,
		},
		{
			name:         "Unauthenticated User",
			prepareMocks: func(us *mockStore, userID uint, user *model.User) {},
			contextAuth: func(ctx context.Context) (uint, error) {
				return 0, errors.New("unauthenticated")
			},
			req:     &proto.UpdateUserRequest{},
			want:    nil,
			wantErr: codes.Unauthenticated,
		},
		{
			name: "User Not Found",
			prepareMocks: func(us *mockStore, userID uint, user *model.User) {
				us.On("GetByID", userID).Return(nil, errors.New("user not found")).Once()
			},
			contextAuth: func(ctx context.Context) (uint, error) {
				return 1, nil
			},
			req:     &proto.UpdateUserRequest{},
			want:    nil,
			wantErr: codes.NotFound,
		},
		{
			name: "Validation Error on User Update",
			prepareMocks: func(us *mockStore, userID uint, user *model.User) {
				us.On("GetByID", userID).Return(user, nil).Once()
			},
			contextAuth: func(ctx context.Context) (uint, error) {
				return 1, nil
			},
			req: &proto.UpdateUserRequest{
				User: &proto.UpdateUserRequest_User{
					Email: "invalid-email",
				},
			},
			want:    nil,
			wantErr: codes.InvalidArgument,
		},
		{
			name: "Password Hash Failure",
			prepareMocks: func(us *mockStore, userID uint, user *model.User) {
				us.On("GetByID", userID).Return(user, nil).Once()
			},
			contextAuth: func(ctx context.Context) (uint, error) {
				return 1, nil
			},
			req: &proto.UpdateUserRequest{
				User: &proto.UpdateUserRequest_User{
					Password: "newpassword",
				},
			},
			want:    nil,
			wantErr: codes.Aborted,
		},
		{
			name: "Store Update Failure",
			prepareMocks: func(us *mockStore, userID uint, user *model.User) {
				us.On("GetByID", userID).Return(user, nil).Once()
				us.On("Update", user).Return(errors.New("update failed")).Once()
			},
			contextAuth: func(ctx context.Context) (uint, error) {
				return 1, nil
			},
			req:     &proto.UpdateUserRequest{},
			want:    nil,
			wantErr: codes.InvalidArgument,
		},
		{
			name: "Token Generation Failure",
			prepareMocks: func(us *mockStore, userID uint, user *model.User) {
				us.On("GetByID", userID).Return(user, nil).Once()
				us.On("Update", user).Return(nil).Once()
			},
			contextAuth: func(ctx context.Context) (uint, error) {
				return 1, nil
			},
			req:     &proto.UpdateUserRequest{},
			want:    nil,
			wantErr: codes.Aborted,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			db, mock, _ := sqlmock.New()
			defer db.Close()

			userStore := new(mockStore)
			handler := Handler{
				us: userStore,
			}

			tokenFunc := auth.GenerateToken
			defer func() {
				auth.GenerateToken = tokenFunc
			}()

			auth.GenerateToken = func(id uint) (string, error) {
				if tc.wantErr == codes.Aborted {
					return "", errors.New("token failure")
				}
				return "dummy-token", nil
			}

			tc.prepareMocks(userStore, 1, &model.User{})

			ctx := context.Background()
			userIDFunc := auth.GetUserID
			defer func() {
				auth.GetUserID = userIDFunc
			}()
			auth.GetUserID = tc.contextAuth

			got, err := handler.UpdateUser(ctx, tc.req)

			if tc.wantErr == codes.OK {
				assert.Nil(t, err)
				assert.Equal(t, tc.want, got)
			} else {
				assert.NotNil(t, err)
				st, _ := status.FromError(err)
				assert.Equal(t, tc.wantErr, st.Code())
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

