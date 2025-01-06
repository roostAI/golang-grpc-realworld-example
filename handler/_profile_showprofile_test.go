package handler

import (
	"context"
	"testing"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/raahii/golang-grpc-realworld-example/auth"
	"github.com/raahii/golang-grpc-realworld-example/proto"
	"github.com/raahii/golang-grpc-realworld-example/store"
	"github.com/rs/zerolog"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)


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
func TestHandlerShowProfile(t *testing.T) {
	t.Parallel()
	db, mock, _ := sqlmock.New()
	gormDB, _ := gorm.Open("postgres", db)
	defer db.Close()

	userStore := &store.UserStore{Db: gormDB}
	articleStore := &store.ArticleStore{Db: gormDB}

	logger := &zerolog.Logger{}
	handler := &Handler{
		logger: logger,
		us:     userStore,
		as:     articleStore,
	}

	validUserID := uint(1)
	otherUserID := uint(2)
	validUsername := "validusername"

	t.Run("Scenario 1: Valid profile retrieval", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), "userID", validUserID)
		req := &proto.ShowProfileRequest{Username: validUsername}

		mock.ExpectQuery("SELECT (.+) FROM users WHERE id = ?").
			WithArgs(validUserID).
			WillReturnRows(sqlmock.NewRows([]string{"ID"}).AddRow(validUserID))

		mock.ExpectQuery("SELECT (.+) FROM users WHERE username = ?").
			WithArgs(validUsername).
			WillReturnRows(sqlmock.NewRows([]string{"ID"}).AddRow(otherUserID))

		mock.ExpectQuery("SELECT COUNT").
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

		resp, err := handler.ShowProfile(ctx, req)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if resp.Profile == nil || resp.Profile.Username != validUsername {
			t.Fatalf("Expected profile with username %s, got %v", validUsername, resp.Profile)
		}
		t.Logf("Successfully retrieved profile: %+v", resp.Profile)
	})

	t.Run("Scenario 2: Unauthenticated request", func(t *testing.T) {
		ctx := context.Background()

		req := &proto.ShowProfileRequest{Username: validUsername}

		_, err := handler.ShowProfile(ctx, req)
		if err == nil || status.Code(err) != codes.Unauthenticated {
			t.Fatalf("Expected Unauthenticated error, got %v", err)
		}
		t.Log("Correctly identified unauthenticated request")
	})

	t.Run("Scenario 3: Current user not found", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), "userID", validUserID)
		req := &proto.ShowProfileRequest{Username: validUsername}

		mock.ExpectQuery("SELECT (.+) FROM users WHERE id = ?").
			WithArgs(validUserID).
			WillReturnError(status.Error(codes.NotFound, "user not found"))

		_, err := handler.ShowProfile(ctx, req)
		if err == nil || status.Code(err) != codes.NotFound {
			t.Fatalf("Expected NotFound error, got %v", err)
		}
		t.Log("Correctly handled current user not found")
	})

	t.Run("Scenario 4: Requested user not found", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), "userID", validUserID)
		req := &proto.ShowProfileRequest{Username: validUsername}

		mock.ExpectQuery("SELECT (.+) FROM users WHERE id = ?").
			WithArgs(validUserID).
			WillReturnRows(sqlmock.NewRows([]string{"ID"}).AddRow(validUserID))

		mock.ExpectQuery("SELECT (.+) FROM users WHERE username = ?").
			WithArgs(validUsername).
			WillReturnError(status.Error(codes.NotFound, "user was not found"))

		_, err := handler.ShowProfile(ctx, req)
		if err == nil || status.Code(err) != codes.NotFound {
			t.Fatalf("Expected NotFound error for requested user, got %v", err)
		}
		t.Log("Correctly handled requested user not found")
	})

	t.Run("Scenario 5: Error retrieving following status", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), "userID", validUserID)
		req := &proto.ShowProfileRequest{Username: validUsername}

		mock.ExpectQuery("SELECT (.+) FROM users WHERE id = ?").
			WithArgs(validUserID).
			WillReturnRows(sqlmock.NewRows([]string{"ID"}).AddRow(validUserID))

		mock.ExpectQuery("SELECT (.+) FROM users WHERE username = ?").
			WithArgs(validUsername).
			WillReturnRows(sqlmock.NewRows([]string{"ID"}).AddRow(otherUserID))

		mock.ExpectQuery("SELECT COUNT").
			WillReturnError(status.Error(codes.Internal, "internal server error"))

		_, err := handler.ShowProfile(ctx, req)
		if err == nil || status.Code(err) != codes.Internal {
			t.Fatalf("Expected Internal server error, got %v", err)
		}
		t.Log("Correctly handled error in retrieving following status")
	})

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unmet expectations: %s", err)
	}
}
