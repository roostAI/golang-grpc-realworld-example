package handler

import (
	"context"
	"testing"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/raahii/golang-grpc-realworld-example/proto"
	"github.com/raahii/golang-grpc-realworld-example/store"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)


/*
ROOST_METHOD_HASH=GetTags_42221e4328
ROOST_METHOD_SIG_HASH=GetTags_52f72598a3


 */
func TestHandlerGetTags(t *testing.T) {
	testCases := []struct {
		name          string
		setupMocks    func(mock sqlmock.Sqlmock)
		expectedTags  []string
		expectedError error
	}{
		{
			name: "Retrieve Tags Successfully",
			setupMocks: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("^SELECT \\* FROM \"tags\"$").
					WillReturnRows(sqlmock.NewRows([]string{"name"}).AddRow("tag1").AddRow("tag2"))
			},
			expectedTags:  []string{"tag1", "tag2"},
			expectedError: nil,
		},
		{
			name: "ArticleStore Returns Error",
			setupMocks: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("^SELECT \\* FROM \"tags\"$").
					WillReturnError(sqlmock.ErrCancelled)
			},
			expectedTags:  nil,
			expectedError: status.Error(codes.Aborted, "internal server error"),
		},
		{
			name: "Empty Tag List",
			setupMocks: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("^SELECT \\* FROM \"tags\"$").WillReturnRows(sqlmock.NewRows([]string{"name"}))
			},
			expectedTags:  []string{},
			expectedError: nil,
		},
	}

	logger := zerolog.New(nil)

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			db, mock, err := sqlmock.New()
			require.NoError(t, err)
			defer db.Close()

			dialector := postgres.New(postgres.Config{
				Conn:       db,
				DriverName: "postgres",
			})

			gormDB, err := gorm.Open(dialector, &gorm.Config{})
			require.NoError(t, err)

			articleStore := &store.ArticleStore{DB: gormDB}

			tc.setupMocks(mock)

			handler := &Handler{
				logger: &logger,
				as:     articleStore,
			}

			ctx := context.Background()
			resp, err := handler.GetTags(ctx, &proto.Empty{})

			assert.Equal(t, tc.expectedError, err)
			if resp != nil {
				assert.ElementsMatch(t, tc.expectedTags, resp.GetTags())
			} else {
				assert.Nil(t, resp)
			}

			t.Logf("Test case '%s' completed", tc.name)
		})
	}
}

