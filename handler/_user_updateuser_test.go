package handler

import (
	"context"
	"errors"
	"testing"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/raahii/golang-grpc-realworld-example/auth"
	"github.com/raahii/golang-grpc-realworld-example/model"
	"github.com/raahii/golang-grpc-realworld-example/proto"
	"github.com/raahii/golang-grpc-realworld-example/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)











type Call struct {
	Parent *Mock

	// The name of the method that was or will be called.
	Method string

	// Holds the arguments of the method.
	Arguments Arguments

	// Holds the arguments that should be returned when
	// this method is called.
	ReturnArguments Arguments

	// Holds the caller info for the On() call
	callerInfo []string

	// The number of times to return the return arguments when setting
	// expectations. 0 means to always return the value.
	Repeatability int

	// Amount of times this call has been called
	totalCalls int

	// Call to this method can be optional
	optional bool

	// Holds a channel that will be used to block the Return until it either
	// receives a message or is closed. nil means it returns immediately.
	WaitFor <-chan time.Time

	waitTime time.Duration

	// Holds a handler used to manipulate arguments content that are passed by
	// reference. It's useful when mocking methods such as unmarshalers or
	// decoders.
	RunFn func(Arguments)

	// PanicMsg holds msg to be used to mock panic on the function call
	//  if the PanicMsg is set to a non nil string the function call will panic
	// irrespective of other settings
	PanicMsg *string

	// Calls which must be satisfied before this call can be
	requires []*Call
}

type Mock struct {
	// Represents the calls that are expected of
	// an object.
	ExpectedCalls []*Call

	// Holds the calls that were made to this mocked object.
	Calls []Call

	// test is An optional variable that holds the test struct, to be used when an
	// invalid mock call was made.
	test TestingT

	// TestData holds any data that might be useful for testing.  Testify ignores
	// this data completely allowing you to do whatever you like with it.
	testData objx.Map

	mutex sync.Mutex
}


type T struct {
	common
	isParallel bool
	isEnvSet   bool
	context    *testContext // For running tests and subtests.
}
func (m *mockStore) GetByID(id uint) (*model.User, error) {
	args := m.Called(id)
	return args.Get(0).(*model.User), args.Error(1)
}
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
func (m *mockStore) Update(user *model.User) error {
	args := m.Called(user)
	return args.Error(0)
}
