package handler

import (
	"context"
	"os"
	"testing"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/raahii/golang-grpc-realworld-example/auth"
	"github.com/raahii/golang-grpc-realworld-example/model"
	pb "github.com/raahii/golang-grpc-realworld-example/proto"
	"github.com/raahii/golang-grpc-realworld-example/store"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)


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





type ConsoleWriter struct {
	// Out is the output destination.
	Out io.Writer

	// NoColor disables the colorized output.
	NoColor bool

	// TimeFormat specifies the format for timestamp in output.
	TimeFormat string

	// PartsOrder defines the order of parts in output.
	PartsOrder []string

	FormatTimestamp     Formatter
	FormatLevel         Formatter
	FormatCaller        Formatter
	FormatMessage       Formatter
	FormatFieldName     Formatter
	FormatFieldValue    Formatter
	FormatErrFieldName  Formatter
	FormatErrFieldValue Formatter
}


type T struct {
	common
	isParallel bool
	isEnvSet   bool
	context    *testContext // For running tests and subtests.
}
func TestHandlerUnfollowUser(t *testing.T) {

	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn: db,
	}), &gorm.Config{})
	assert.NoError(t, err)

	mockUserStore := &store.UserStore{DB: gormDB}

	consoleWriter := zerolog.ConsoleWriter{Out: os.Stdout}
	mockLogger := zerolog.New(consoleWriter)

	h := &Handler{
		logger: &mockLogger,
		us:     mockUserStore,
	}

	type testCase struct {
		name       string
		setupMocks func()
		ctx        context.Context
		req        *pb.UnfollowRequest
		wantErr    codes.Code
	}

	tests := []testCase{
		{
			name: "Successfully Unfollow a User",
			setupMocks: func() {
				auth.GetUserID = func(ctx context.Context) (uint, error) {
					return 1, nil
				}
				mock.ExpectQuery("^SELECT (.+) FROM \"users\" WHERE id = ? LIMIT 1$").
					WithArgs(1).
					WillReturnRows(sqlmock.NewRows([]string{"id", "username"}).AddRow(1, "currentUser"))

				mock.ExpectQuery("^SELECT (.+) FROM \"users\" WHERE username = ? LIMIT 1$").
					WithArgs("anotherUser").
					WillReturnRows(sqlmock.NewRows([]string{"id", "username"}).AddRow(2, "anotherUser"))

				mock.ExpectQuery("^SELECT (.+) FROM \"follows\"").
					WithArgs(1, 2).
					WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

				mock.ExpectExec("^DELETE FROM \"follows\"").
					WithArgs(1, 2).
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
			ctx:     context.Background(),
			req:     &pb.UnfollowRequest{Username: "anotherUser"},
			wantErr: codes.OK,
		},
		{
			name: "Unauthenticated User",
			setupMocks: func() {
				auth.GetUserID = func(ctx context.Context) (uint, error) {
					return 0, status.Error(codes.Unauthenticated, "unauthenticated")
				}
			},
			ctx:     context.Background(),
			req:     &pb.UnfollowRequest{Username: "anotherUser"},
			wantErr: codes.Unauthenticated,
		},
		{
			name: "Current User Not Found",
			setupMocks: func() {
				auth.GetUserID = func(ctx context.Context) (uint, error) {
					return 1, nil
				}
				mock.ExpectQuery("^SELECT (.+) FROM \"users\" WHERE id = ? LIMIT 1$").
					WithArgs(1).
					WillReturnError(status.Error(codes.NotFound, "user not found"))
			},
			ctx:     context.Background(),
			req:     &pb.UnfollowRequest{Username: "anotherUser"},
			wantErr: codes.NotFound,
		},
		{
			name: "Attempt to Unfollow Self",
			setupMocks: func() {
				auth.GetUserID = func(ctx context.Context) (uint, error) {
					return 1, nil
				}
				mock.ExpectQuery("^SELECT (.+) FROM \"users\" WHERE id = ? LIMIT 1$").
					WithArgs(1).
					WillReturnRows(sqlmock.NewRows([]string{"id", "username"}).AddRow(1, "currentUser"))
			},
			ctx:     context.Background(),
			req:     &pb.UnfollowRequest{Username: "currentUser"},
			wantErr: codes.InvalidArgument,
		},
		{
			name: "Request User Not Found",
			setupMocks: func() {
				auth.GetUserID = func(ctx context.Context) (uint, error) {
					return 1, nil
				}
				mock.ExpectQuery("^SELECT (.+) FROM \"users\" WHERE id = ? LIMIT 1$").
					WithArgs(1).
					WillReturnRows(sqlmock.NewRows([]string{"id", "username"}).AddRow(1, "currentUser"))

				mock.ExpectQuery("^SELECT (.+) FROM \"users\" WHERE username = ? LIMIT 1$").
					WithArgs("nonexistentUser").
					WillReturnError(status.Error(codes.NotFound, "user not found"))
			},
			ctx:     context.Background(),
			req:     &pb.UnfollowRequest{Username: "nonexistentUser"},
			wantErr: codes.NotFound,
		},
		{
			name: "Current User Not Following Request User",
			setupMocks: func() {
				auth.GetUserID = func(ctx context.Context) (uint, error) {
					return 1, nil
				}
				mock.ExpectQuery("^SELECT (.+) FROM \"users\" WHERE id = ? LIMIT 1$").
					WithArgs(1).
					WillReturnRows(sqlmock.NewRows([]string{"id", "username"}).AddRow(1, "currentUser"))

				mock.ExpectQuery("^SELECT (.+) FROM \"users\" WHERE username = ? LIMIT 1$").
					WithArgs("anotherUser").
					WillReturnRows(sqlmock.NewRows([]string{"id", "username"}).AddRow(2, "anotherUser"))

				mock.ExpectQuery("^SELECT (.+) FROM \"follows\"").
					WithArgs(1, 2).
					WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
			},
			ctx:     context.Background(),
			req:     &pb.UnfollowRequest{Username: "anotherUser"},
			wantErr: codes.Unauthenticated,
		},
		{
			name: "Unfollow Operation Fails",
			setupMocks: func() {
				auth.GetUserID = func(ctx context.Context) (uint, error) {
					return 1, nil
				}
				mock.ExpectQuery("^SELECT (.+) FROM \"users\" WHERE id = ? LIMIT 1$").
					WithArgs(1).
					WillReturnRows(sqlmock.NewRows([]string{"id", "username"}).AddRow(1, "currentUser"))

				mock.ExpectQuery("^SELECT (.+) FROM \"users\" WHERE username = ? LIMIT 1$").
					WithArgs("anotherUser").
					WillReturnRows(sqlmock.NewRows([]string{"id", "username"}).AddRow(2, "anotherUser"))

				mock.ExpectQuery("^SELECT (.+) FROM \"follows\"").
					WithArgs(1, 2).
					WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

				mock.ExpectExec("^DELETE FROM \"follows\"").
					WithArgs(1, 2).
					WillReturnError(status.Error(codes.Aborted, "unfollow operation failed"))
			},
			ctx:     context.Background(),
			req:     &pb.UnfollowRequest{Username: "anotherUser"},
			wantErr: codes.Aborted,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()

			resp, err := h.UnfollowUser(tt.ctx, tt.req)
			if tt.wantErr != codes.OK {
				assert.Error(t, err)
				assert.Nil(t, resp)
				if st, ok := status.FromError(err); ok {
					assert.Equal(t, tt.wantErr, st.Code())
				} else {
					t.Fatalf("expected grpc status error, got: %v", err)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}
