package handler

import (
	"testing"
	"context"
	"fmt"
	"errors"
	"os"
	"strconv"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/raahii/golang-grpc-realworld-example/auth"
	"github.com/raahii/golang-grpc-realworld-example/model"
	"github.com/raahii/golang-grpc-realworld-example/proto"
	"github.com/raahii/golang-grpc-realworld-example/store"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const authUserIDKey = "userID"
type ExpectedExec struct {
	queryBasedExpectation
	result driver.Result
	delay  time.Duration
}

type ExpectedQuery struct {
	queryBasedExpectation
	rows             driver.Rows
	delay            time.Duration
	rowsMustBeClosed bool
	rowsWereClosed   bool
}






type T struct {
	common
	isParallel bool
	isEnvSet   bool
	context    *testContext // For running tests and subtests.
}
func TestHandlerUnfavoriteArticle(t *testing.T) {
	t.Run("Scenario 1: Successful Unfavoriting of an Article", func(t *testing.T) {

		mockDB, mock, err := sqlmock.New()
		assert.NoError(t, err)
		defer mockDB.Close()

		articleStore := &store.ArticleStore{Db: mockDB}
		userStore := &store.UserStore{Db: mockDB}

		logger := zerolog.New(os.Stdout)
		mockHandler := &Handler{logger: &logger, us: userStore, as: articleStore}

		authCtx := context.WithValue(context.Background(), authUserIDKey, uint(1))
		req := &proto.UnfavoriteArticleRequest{Slug: "1"}

		mock.ExpectQuery("SELECT (.+) FROM `articles` WHERE").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
		mock.ExpectQuery("SELECT (.+) FROM `users` WHERE").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
		mock.ExpectExec("DELETE FROM `favorites` WHERE").WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectExec("UPDATE `articles` SET").WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectQuery("SELECT (.+) FROM `users` WHERE").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

		response, err := mockHandler.UnfavoriteArticle(authCtx, req)

		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.False(t, response.Article.Favorited)

		if err = mock.ExpectationsWereMet(); err != nil {
			t.Errorf("unmet database expectations: %s", err)
		}
		t.Log("Test scenario for successful unfavoriting of an article passed.")
	})

	t.Run("Scenario 2: Unauthenticated User", func(t *testing.T) {

		mockDB, _, err := sqlmock.New()
		assert.NoError(t, err)
		defer mockDB.Close()

		articleStore := &store.ArticleStore{Db: mockDB}
		userStore := &store.UserStore{Db: mockDB}

		logger := zerolog.New(os.Stdout)
		mockHandler := &Handler{logger: &logger, us: userStore, as: articleStore}

		req := &proto.UnfavoriteArticleRequest{Slug: "1"}

		_, err = mockHandler.UnfavoriteArticle(context.Background(), req)

		assert.Error(t, err)
		assert.Equal(t, codes.Unauthenticated, status.Code(err))
		t.Log("Test scenario for unauthenticated user passed.")
	})

	t.Run("Scenario 3: Article Not Found", func(t *testing.T) {

		mockDB, mock, err := sqlmock.New()
		assert.NoError(t, err)
		defer mockDB.Close()

		articleStore := &store.ArticleStore{Db: mockDB}
		userStore := &store.UserStore{Db: mockDB}

		logger := zerolog.New(os.Stdout)
		mockHandler := &Handler{logger: &logger, us: userStore, as: articleStore}

		authCtx := context.WithValue(context.Background(), authUserIDKey, uint(1))
		req := &proto.UnfavoriteArticleRequest{Slug: "9999"}

		mock.ExpectQuery("SELECT (.+) FROM `articles` WHERE").WillReturnError(errors.New("record not found"))

		_, err = mockHandler.UnfavoriteArticle(authCtx, req)

		assert.Error(t, err)
		assert.Equal(t, codes.InvalidArgument, status.Code(err))
		t.Log("Test scenario for article not found passed.")
	})

	t.Run("Scenario 4: Invalid Slug Format", func(t *testing.T) {

		mockDB, _, err := sqlmock.New()
		assert.NoError(t, err)
		defer mockDB.Close()

		articleStore := &store.ArticleStore{Db: mockDB}
		userStore := &store.UserStore{Db: mockDB}

		logger := zerolog.New(os.Stdout)
		mockHandler := &Handler{logger: &logger, us: userStore, as: articleStore}

		authCtx := context.WithValue(context.Background(), authUserIDKey, uint(1))
		req := &proto.UnfavoriteArticleRequest{Slug: "invalid_slug"}

		_, err = mockHandler.UnfavoriteArticle(authCtx, req)

		assert.Error(t, err)
		assert.Equal(t, codes.InvalidArgument, status.Code(err))
		t.Log("Test scenario for invalid slug format passed.")
	})

	t.Run("Scenario 5: User Not Found", func(t *testing.T) {

		mockDB, mock, err := sqlmock.New()
		assert.NoError(t, err)
		defer mockDB.Close()

		articleStore := &store.ArticleStore{Db: mockDB}
		userStore := &store.UserStore{Db: mockDB}

		logger := zerolog.New(os.Stdout)
		mockHandler := &Handler{logger: &logger, us: userStore, as: articleStore}

		authCtx := context.WithValue(context.Background(), authUserIDKey, uint(9999))
		req := &proto.UnfavoriteArticleRequest{Slug: "1"}

		mock.ExpectQuery("SELECT (.+) FROM `users` WHERE").WillReturnError(errors.New("record not found"))

		_, err = mockHandler.UnfavoriteArticle(authCtx, req)

		assert.Error(t, err)
		assert.Equal(t, codes.NotFound, status.Code(err))
		t.Log("Test scenario for user not found passed.")
	})

	t.Run("Scenario 6: Failure to Remove Favorite", func(t *testing.T) {

		mockDB, mock, err := sqlmock.New()
		assert.NoError(t, err)
		defer mockDB.Close()

		articleStore := &store.ArticleStore{Db: mockDB}
		userStore := &store.UserStore{Db: mockDB}

		logger := zerolog.New(os.Stdout)
		mockHandler := &Handler{logger: &logger, us: userStore, as: articleStore}

		authCtx := context.WithValue(context.Background(), authUserIDKey, uint(1))
		req := &proto.UnfavoriteArticleRequest{Slug: "1"}

		mock.ExpectQuery("SELECT (.+) FROM `articles` WHERE").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
		mock.ExpectQuery("SELECT (.+) FROM `users` WHERE").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
		mock.ExpectExec("DELETE FROM `favorites` WHERE").WillReturnError(errors.New("database failure"))

		_, err = mockHandler.UnfavoriteArticle(authCtx, req)

		assert.Error(t, err)
		assert.Equal(t, codes.InvalidArgument, status.Code(err))
		t.Log("Test scenario for failure to remove favorite passed.")
	})

	t.Run("Scenario 7: Failure to Determine Following Status", func(t *testing.T) {

		mockDB, mock, err := sqlmock.New()
		assert.NoError(t, err)
		defer mockDB.Close()

		articleStore := &store.ArticleStore{Db: mockDB}
		userStore := &store.UserStore{Db: mockDB}

		logger := zerolog.New(os.Stdout)
		mockHandler := &Handler{logger: &logger, us: userStore, as: articleStore}

		authCtx := context.WithValue(context.Background(), authUserIDKey, uint(1))
		req := &proto.UnfavoriteArticleRequest{Slug: "1"}

		mock.ExpectQuery("SELECT (.+) FROM `articles` WHERE").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
		mock.ExpectQuery("SELECT (.+) FROM `users` WHERE").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
		mock.ExpectExec("DELETE FROM `favorites` WHERE").WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectExec("UPDATE `articles` SET").WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectQuery("SELECT (.+) FROM `users` WHERE").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
		mock.ExpectQuery("SELECT (.+) FROM `follows` WHERE").WillReturnError(errors.New("database failure"))

		_, err = mockHandler.UnfavoriteArticle(authCtx, req)

		assert.Error(t, err)
		assert.Equal(t, codes.NotFound, status.Code(err))
		t.Log("Test scenario for failure to determine following status passed.")
	})
}
