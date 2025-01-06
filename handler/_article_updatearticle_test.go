package handler

import (
	"context"
	"errors"
	"testing"
	"github.com/raahii/golang-grpc-realworld-example/model"
	pb "github.com/raahii/golang-grpc-realworld-example/proto"
	"github.com/raahii/golang-grpc-realworld-example/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"github.com/rs/zerolog"
	"os"
)

type MockUserStore struct {
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
func (m *MockUserStore) GetByID(id uint) (*model.User, error) {
	args := m.Called(id)
	user, _ := args.Get(0).(*model.User)
	return user, args.Error(1)
}
func (m *MockUserStore) IsFollowing(a *model.User, b *model.User) (bool, error) {
	args := m.Called(a, b)
	return args.Bool(0), args.Error(1)
}
func TestHandlerUpdateArticle(t *testing.T) {
	mockUserStore := new(MockUserStore)
	mockArticleStore := new(MockArticleStore)
	logger := zerolog.New(os.Stdout)

	handler := Handler{
		logger: &logger,
		us:     (*store.UserStore)(mockUserStore),
		as:     (*store.ArticleStore)(mockArticleStore),
	}

	testCases := []struct {
		scenario         string
		setupContext     func() context.Context
		setupMocks       func()
		req              *pb.UpdateArticleRequest
		expectedError    error
		expectedResponse *pb.ArticleResponse
	}{
		{
			scenario: "Successful Update of an Article",
			setupContext: func() context.Context {
				return context.WithValue(context.Background(), "userID", uint(1))
			},
			setupMocks: func() {
				mockUserStore.On("GetByID", uint(1)).Return(&model.User{ID: 1}, nil)
				mockArticleStore.On("GetByID", uint(123)).Return(&model.Article{
					ID:     123,
					Author: model.User{ID: 1},
				}, nil)
				mockArticleStore.On("Update", mock.Anything).Return(nil)
				mockUserStore.On("IsFollowing", mock.Anything, mock.Anything).Return(true, nil)
			},
			req: &pb.UpdateArticleRequest{
				Article: &pb.UpdateArticleRequest_Article{
					Slug:        "123",
					Title:       "Updated Title",
					Description: "Updated Description",
					Body:        "Updated Body",
				},
			},
			expectedError: nil,
			expectedResponse: &pb.ArticleResponse{
				Article: &pb.Article{
					Slug:           "123",
					Title:          "Updated Title",
					Description:    "Updated Description",
					Body:           "Updated Body",
					Favorited:      true,
					FavoritesCount: 0,
				},
			},
		},
		{
			scenario: "Unauthenticated User Error",
			setupContext: func() context.Context {
				return context.Background()
			},
			setupMocks: func() {
				mockUserStore.On("GetByID", uint(0)).Return(nil, errors.New("user not found"))
			},
			req: &pb.UpdateArticleRequest{
				Article: &pb.UpdateArticleRequest_Article{
					Slug:        "123",
					Title:       "Updated Title",
					Description: "Updated Description",
					Body:        "Updated Body",
				},
			},
			expectedError: status.Errorf(codes.Unauthenticated, "unauthenticated"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.scenario, func(t *testing.T) {
			tc.setupMocks()
			resp, err := handler.UpdateArticle(tc.setupContext(), tc.req)

			assert.Equal(t, tc.expectedError, err)
			if err == nil {
				assert.Equal(t, tc.expectedResponse.Article.Title, resp.Article.Title)
				assert.Equal(t, tc.expectedResponse.Article.Description, resp.Article.Description)
				assert.Equal(t, tc.expectedResponse.Article.Body, resp.Article.Body)
			}

			t.Logf("Scenario: %s - Execution completed.", tc.scenario)
		})
	}

	mockArticleStore.AssertExpectations(t)
	mockUserStore.AssertExpectations(t)
}
func (m *MockArticleStore) Update(article *model.Article) error {
	args := m.Called(article)
	return args.Error(0)
}
