package handler

import (
	"context"
	"testing"
	"github.com/golang/mock/gomock"
	"github.com/raahii/golang-grpc-realworld-example/auth"
	"github.com/raahii/golang-grpc-realworld-example/model"
	"github.com/raahii/golang-grpc-realworld-example/proto"
	"github.com/raahii/golang-grpc-realworld-example/store"
	"github.com/rs/zerolog"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"errors"
	"fmt"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
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
	context    *testContext
}
/*
ROOST_METHOD_HASH=CurrentUser_e3fa631d55
ROOST_METHOD_SIG_HASH=CurrentUser_29413339e9


 */
func TestHandlerCurrentUser(t *testing.T) {
	type testCase struct {
		name          string
		setupMocks    func(as *store.MockUserStore, ctrl *gomock.Controller)
		ctx           context.Context
		expectedError error
		expectedResp  *pb.UserResponse
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStore := store.NewMockUserStore(ctrl)
	mockLogger := &zerolog.Logger{}

	handler := &Handler{
		logger: mockLogger,
		us:     mockStore,
	}

	tests := []testCase{
		{
			name: "Successfully Retrieve Current User",
			setupMocks: func(us *store.MockUserStore, ctrl *gomock.Controller) {

				mockUser := &model.User{ID: 1, Email: "test@example.com", Username: "testuser"}
				us.EXPECT().GetByID(uint(1)).Return(mockUser, nil)

				auth.MockGetUserID(ctrl, uint(1), nil)
				auth.MockGenerateToken(ctrl, "valid-token", nil)
			},
			ctx:           context.Background(),
			expectedResp:  &pb.UserResponse{User: &pb.User{Token: "valid-token", Email: "test@example.com", Username: "testuser"}},
			expectedError: nil,
		},
		{
			name: "Context Without User ID Results in Unauthenticated Error",
			setupMocks: func(us *store.MockUserStore, ctrl *gomock.Controller) {

				auth.MockGetUserID(ctrl, uint(0), status.Errorf(codes.Unauthenticated, "unauthenticated"))
			},
			ctx:           context.Background(),
			expectedResp:  nil,
			expectedError: status.Errorf(codes.Unauthenticated, "unauthenticated"),
		},
		{
			name: "User Not Found in Database Returns Not Found Error",
			setupMocks: func(us *store.MockUserStore, ctrl *gomock.Controller) {

				us.EXPECT().GetByID(uint(1)).Return(nil, status.Error(codes.NotFound, "user not found"))
				auth.MockGetUserID(ctrl, uint(1), nil)
			},
			ctx:           context.Background(),
			expectedResp:  nil,
			expectedError: status.Error(codes.NotFound, "user not found"),
		},
		{
			name: "Fail to Generate Token Returns Internal Server Error",
			setupMocks: func(us *store.MockUserStore, ctrl *gomock.Controller) {

				mockUser := &model.User{ID: 1, Email: "test@example.com", Username: "testuser"}
				us.EXPECT().GetByID(uint(1)).Return(mockUser, nil)
				auth.MockGetUserID(ctrl, uint(1), nil)
				auth.MockGenerateToken(ctrl, "", status.Error(codes.Aborted, "internal server error"))
			},
			ctx:           context.Background(),
			expectedResp:  nil,
			expectedError: status.Error(codes.Aborted, "internal server error"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tc.setupMocks(mockStore, ctrl)

			resp, err := handler.CurrentUser(tc.ctx, &pb.Empty{})

			if err != nil && tc.expectedError == nil {
				t.Errorf("Expected no error, got %v", err)
			}

			if err == nil && tc.expectedError != nil {
				t.Errorf("Expected error %v, got none", tc.expectedError)
			}

			if err != nil && tc.expectedError != nil && err.Error() != tc.expectedError.Error() {
				t.Errorf("Expected error to be %v, got %v", tc.expectedError, err)
			}

			if resp != nil && !equalUserResponse(resp, tc.expectedResp) {
				t.Errorf("Expected response to be %v, got %v", tc.expectedResp, resp)
			}

			t.Log("Test ", tc.name, " passed")
		})
	}
}

func equalUserResponse(a, b *pb.UserResponse) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return a.User.Email == b.User.Email && a.User.Token == b.User.Token && a.User.Username == b.User.Username
}


/*
ROOST_METHOD_HASH=LoginUser_079a321a92
ROOST_METHOD_SIG_HASH=LoginUser_e7df23a6bd


 */
func TestHandlerLoginUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := zerolog.New(zerolog.ConsoleWriter{Out: errors.New(""), NoColor: true})

	mockUserStore := model.NewMockUserStore(ctrl)
	mockAuth := auth.NewMockAuth(ctrl)

	h := &Handler{
		logger: &mockLogger,
		us:     mockUserStore,
	}

	type testCase struct {
		name     string
		setup    func()
		req      *pb.LoginUserRequest
		expected *pb.UserResponse
		err      error
	}

	tests := []testCase{
		{
			name: "Scenario 1: Successful Login with Valid Credentials",
			setup: func() {
				mockUser := &model.User{
					Email:    "valid@example.com",
					Password: "$2a$12$somethinghashed",
					Username: "username",
					Bio:      "bio",
					Image:    "imageurl",
				}

				mockUserStore.EXPECT().GetByEmail("valid@example.com").Return(mockUser, nil)

				token := "generatedToken"
				mockAuth.EXPECT().GenerateToken(mockUser.ID).Return(token, nil)
			},
			req: &pb.LoginUserRequest{
				User: &pb.LoginUserRequest_User{
					Email:    "valid@example.com",
					Password: "plainPassword",
				},
			},
			expected: &pb.UserResponse{
				User: &pb.User{
					Email:    "valid@example.com",
					Token:    "generatedToken",
					Username: "username",
					Bio:      "bio",
					Image:    "imageurl",
				},
			},
			err: nil,
		},
		{
			name: "Scenario 2: Error on Invalid Email",
			setup: func() {
				mockUserStore.EXPECT().GetByEmail("invalid@example.com").Return(nil, sqlmock.ErrNoRows)
			},
			req: &pb.LoginUserRequest{
				User: &pb.LoginUserRequest_User{
					Email:    "invalid@example.com",
					Password: "somepassword",
				},
			},
			expected: nil,
			err:      status.Error(codes.InvalidArgument, "invalid email or password"),
		},
		{
			name: "Scenario 3: Error on Invalid Password for Existing Email",
			setup: func() {
				mockUser := &model.User{
					Email:    "valid@example.com",
					Password: "$2a$12$somethinghashed",
				}
				mockUserStore.EXPECT().GetByEmail("valid@example.com").Return(mockUser, nil)
			},
			req: &pb.LoginUserRequest{
				User: &pb.LoginUserRequest_User{
					Email:    "valid@example.com",
					Password: "wrongpassword",
				},
			},
			expected: nil,
			err:      status.Error(codes.InvalidArgument, "invalid email or password"),
		},
		{
			name: "Scenario 4: Token Generation Failure",
			setup: func() {
				mockUser := &model.User{
					Email:    "valid@example.com",
					Password: "$2a$12$somethinghashed",
				}
				mockUserStore.EXPECT().GetByEmail("valid@example.com").Return(mockUser, nil)

				mockAuth.EXPECT().GenerateToken(mockUser.ID).Return("", status.Error(codes.Internal, "token generation failed"))
			},
			req: &pb.LoginUserRequest{
				User: &pb.LoginUserRequest_User{
					Email:    "valid@example.com",
					Password: "plainPassword",
				},
			},
			expected: nil,
			err:      status.Error(codes.Aborted, "internal server error"),
		},
	}

	for _, test := range tests {
		t.Logf("Running test case: %s", test.name)
		test.setup()

		ctx := context.Background()
		response, err := h.LoginUser(ctx, test.req)

		if test.err != nil {
			assert.Nil(t, response)
			assert.EqualError(t, err, test.err.Error())
			t.Logf("Expected error: %v, got: %v", test.err, err)
		} else {
			assert.NoError(t, err)
			assert.Equal(t, test.expected, response)
			t.Logf("Expected response: %v, got: %v", test.expected, response)
		}
	}
}

