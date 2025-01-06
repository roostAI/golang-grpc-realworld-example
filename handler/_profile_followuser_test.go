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
	"github.com/rs/zerolog"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)






type User struct {
	gorm.Model
	Username         string    `gorm:"unique_index;not null"`
	Email            string    `gorm:"unique_index;not null"`
	Password         string    `gorm:"not null"`
	Bio              string    `gorm:"not null"`
	Image            string    `gorm:"not null"`
	Follows          []User    `gorm:"many2many:follows;jointable_foreignkey:from_user_id;association_jointable_foreignkey:to_user_id"`
	FavoriteArticles []Article `gorm:"many2many:favorite_articles;"`
}





type T struct {
	common
	isParallel bool
	isEnvSet   bool
	context    *testContext // For running tests and subtests.
}
func (mock *UserStoreMock) Follow(a *model.User, b *model.User) error {
	args := mock.Called(a, b)
	return args.Error(0)
}
func (mock *UserStoreMock) GetByID(id uint) (*model.User, error) {
	args := mock.Called(id)
	user, ok := args.Get(0).(*model.User)
	if !ok {
		return nil, args.Error(1)
	}
	return user, args.Error(1)
}
func (mock *UserStoreMock) GetByUsername(username string) (*model.User, error) {
	args := mock.Called(username)
	user, ok := args.Get(0).(*model.User)
	if !ok {
		return nil, args.Error(1)
	}
	return user, args.Error(1)
}
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
