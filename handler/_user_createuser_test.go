package handler

import (
	"context"
	"errors"
	"testing"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/golang/mock/gomock"
	"github.com/raahii/golang-grpc-realworld-example/auth"
	"github.com/raahii/golang-grpc-realworld-example/model"
	pb "github.com/raahii/golang-grpc-realworld-example/proto"
	"github.com/raahii/golang-grpc-realworld-example/store"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type MockUserStore struct {
	mock *gomock.Controller
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
	context    *testContext // For running tests and subtests.
}
func (m *MockUserStore) Create(u *model.User) error {
	return nil
}
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
