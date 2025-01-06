package handler

import (
	"context"
	"fmt"
	"testing"
	"github.com/golang/mock/gomock"
	gomock "github.com/golang/mock/gomock"
	"github.com/raahii/golang-grpc-realworld-example/auth"
	"github.com/raahii/golang-grpc-realworld-example/model"
	pb "github.com/raahii/golang-grpc-realworld-example/proto"
	"github.com/raahii/golang-grpc-realworld-example/store"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)



type Controller struct {
	mu            sync.Mutex
	t             TestReporter
	expectedCalls *callSet
	finished      bool
}



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
func TestHandlerGetFeedArticles(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserStore := NewMockUserStore(ctrl)
	mockArticleStore := NewMockArticleStore(ctrl)
	mockLogger := zerolog.New(nil)

	handler := &Handler{logger: &mockLogger, us: mockUserStore, as: mockArticleStore}
	ctx := context.Background()

	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	scenarios := []struct {
		name        string
		setup       func()
		verify      func(*pb.ArticlesResponse, error)
		description string
	}{
		{
			name: "Successfully Retrieve Feed Articles",
			setup: func() {
				userID := uint(1)
				req := &pb.GetFeedArticlesRequest{Limit: 10, Offset: 0}

				auth.GetUserID = func(ctx context.Context) (uint, error) {
					return userID, nil
				}

				currentUser := &model.User{ID: userID}
				followedIDs := []uint{2, 3}
				articles := []model.Article{
					{ID: 1, Title: "Article 1", Author: model.User{ID: 2}},
					{ID: 2, Title: "Article 2", Author: model.User{ID: 3}},
				}

				mockUserStore.EXPECT().GetByID(userID).Return(currentUser, nil)
				mockUserStore.EXPECT().GetFollowingUserIDs(currentUser).Return(followedIDs, nil)
				mockArticleStore.EXPECT().GetFeedArticles(followedIDs, req.GetLimit(), req.GetOffset()).Return(articles, nil)
				mockArticleStore.EXPECT().IsFavorited(&articles[0], currentUser).Return(true, nil)
				mockArticleStore.EXPECT().IsFavorited(&articles[1], currentUser).Return(false, nil)
				mockUserStore.EXPECT().IsFollowing(currentUser, &articles[0].Author).Return(true, nil)
				mockUserStore.EXPECT().IsFollowing(currentUser, &articles[1].Author).Return(false, nil)
			},
			verify: func(resp *pb.ArticlesResponse, err error) {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				assert.Equal(t, 2, len(resp.Articles))
				assert.Equal(t, "Article 1", resp.Articles[0].Title)
			},
			description: "Ensures correct articles retrieval with expected fields.",
		},
		{
			name: "Unauthenticated User Request",
			setup: func() {
				auth.GetUserID = func(ctx context.Context) (uint, error) {
					return 0, fmt.Errorf("unauthenticated")
				}
			},
			verify: func(resp *pb.ArticlesResponse, err error) {
				assert.Nil(t, resp)
				assert.Equal(t, status.Errorf(codes.Unauthenticated, "unauthenticated"), err)
			},
			description: "Ensures unauthenticated users are denied access.",
		},
	}

	for _, s := range scenarios {
		t.Run(s.name, func(t *testing.T) {
			s.setup()
			req := &pb.GetFeedArticlesRequest{Limit: 10, Offset: 0}
			resp, err := handler.GetFeedArticles(ctx, req)
			s.verify(resp, err)
			t.Logf("Scenario '%s': %s", s.name, s.description)
		})
	}
}
