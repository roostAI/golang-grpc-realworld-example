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
)

type MockUserStore struct {
	mock.Mock
}








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
func (m *MockUserStore) GetByID(id uint) (*model.User, error) {
	args := m.Called(id)
	return args.Get(0).(*model.User), args.Error(1)
}
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
