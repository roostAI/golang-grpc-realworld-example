package handler

import (
	"context"
	"errors"
	"os"
	"testing"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/raahii/golang-grpc-realworld-example/auth"
	"github.com/raahii/golang-grpc-realworld-example/model"
	pb "github.com/raahii/golang-grpc-realworld-example/proto"
	"github.com/raahii/golang-grpc-realworld-example/store"
	"github.com/rs/zerolog"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)








type T struct {
	common
	isParallel bool
	isEnvSet   bool
	context    *testContext // For running tests and subtests.
}
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
