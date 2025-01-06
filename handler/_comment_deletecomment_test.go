package handler

import (
	"context"
	"testing"
	"strconv"
	"fmt"
	"github.com/raahii/golang-grpc-realworld-example/auth"
	"github.com/raahii/golang-grpc-realworld-example/model"
	pb "github.com/raahii/golang-grpc-realworld-example/proto"
	"github.com/raahii/golang-grpc-realworld-example/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var mockArticleStore = new(MockArticleStore)
var mockUserStore = new(MockUserStore)type MockUserStore struct {
	mock.Mock
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
func (m *MockArticleStore) DeleteComment(c *model.Comment) error {
	args := m.Called(c)
	return args.Error(0)
}
func (m *MockUserStore) GetByID(id uint) (*model.User, error) {
	args := m.Called(id)
	user, _ := args.Get(0).(*model.User)
	return user, args.Error(1)
}
func (m *MockArticleStore) GetCommentByID(id uint) (*model.Comment, error) {
	args := m.Called(id)
	comment, _ := args.Get(0).(*model.Comment)
	return comment, args.Error(1)
}
func TestHandlerDeleteComment(t *testing.T) {
	h := &Handler{
		logger: &log.Logger,
		us:     &store.UserStore{db: nil},
		as:     &store.ArticleStore{db: nil},
	}

	tests := []struct {
		name          string
		setup         func() context.Context
		request       *pb.DeleteCommentRequest
		expectedError error
	}{
		{
			name: "Unauthenticated User Tries to Delete a Comment",
			setup: func() context.Context {
				ctx := context.Background()
				return ctx
			},
			request:       &pb.DeleteCommentRequest{Slug: "123", Id: "456"},
			expectedError: status.Error(codes.Unauthenticated, "unauthenticated"),
		},
		{
			name: "User Not Found in the Database",
			setup: func() context.Context {
				ctx := context.WithValue(context.Background(), auth.UserIDKey, uint(1))
				mockUserStore.On("GetByID", uint(1)).Return((*model.User)(nil), status.Error(codes.NotFound, "user not found"))
				return ctx
			},
			request:       &pb.DeleteCommentRequest{Slug: "123", Id: "456"},
			expectedError: status.Error(codes.NotFound, "user not found"),
		},
		{
			name: "Invalid Comment ID Format",
			setup: func() context.Context {
				ctx := context.WithValue(context.Background(), auth.UserIDKey, uint(1))
				mockUserStore.On("GetByID", uint(1)).Return(&model.User{ID: 1}, nil)
				return ctx
			},
			request:       &pb.DeleteCommentRequest{Slug: "123", Id: "invalid"},
			expectedError: status.Error(codes.InvalidArgument, "invalid article id"),
		},
		{
			name: "Comment Not Found in the Article",
			setup: func() context.Context {
				ctx := context.WithValue(context.Background(), auth.UserIDKey, uint(1))
				mockUserStore.On("GetByID", uint(1)).Return(&model.User{ID: 1}, nil)
				mockArticleStore.On("GetCommentByID", uint(456)).Return(&model.Comment{ArticleID: 888}, nil)
				return ctx
			},
			request:       &pb.DeleteCommentRequest{Slug: "123", Id: "456"},
			expectedError: status.Error(codes.InvalidArgument, "the comment is not in the article"),
		},
		{
			name: "User Lacking Permission to Delete Comment",
			setup: func() context.Context {
				ctx := context.WithValue(context.Background(), auth.UserIDKey, uint(2))
				mockUserStore.On("GetByID", uint(2)).Return(&model.User{ID: 2}, nil)
				mockArticleStore.On("GetCommentByID", uint(456)).Return(&model.Comment{UserID: 3, ArticleID: 123}, nil)
				return ctx
			},
			request:       &pb.DeleteCommentRequest{Slug: "123", Id: "456"},
			expectedError: status.Error(codes.InvalidArgument, "forbidden"),
		},
		{
			name: "Successful Comment Deletion",
			setup: func() context.Context {
				ctx := context.WithValue(context.Background(), auth.UserIDKey, uint(1))
				mockUserStore.On("GetByID", uint(1)).Return(&model.User{ID: 1}, nil)
				mockArticleStore.On("GetCommentByID", uint(456)).Return(&model.Comment{UserID: 1, ArticleID: 123}, nil)
				mockArticleStore.On("DeleteComment", mock.Anything).Return(nil)
				return ctx
			},
			request:       &pb.DeleteCommentRequest{Slug: "123", Id: "456"},
			expectedError: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx := test.setup()
			_, err := h.DeleteComment(ctx, test.request)

			if test.expectedError != nil {
				assert.EqualError(t, err, test.expectedError.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
