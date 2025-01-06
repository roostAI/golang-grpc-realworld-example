package handler

import (
	"context"
	"errors"
	"testing"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/raahii/golang-grpc-realworld-example/auth"
	"github.com/raahii/golang-grpc-realworld-example/model"
	pb "github.com/raahii/golang-grpc-realworld-example/proto"
	"github.com/raahii/golang-grpc-realworld-example/store"
	"github.com/rs/zerolog"
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
func (mw *mockWriter) Write(p []byte) (n int, err error) {
	return len(p), nil
}
