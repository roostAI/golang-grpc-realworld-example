package handler

import (
	"context"
	"errors"
	"testing"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/raahii/golang-grpc-realworld-example/proto"
	"github.com/raahii/golang-grpc-realworld-example/store"
	"github.com/stretchr/testify/assert"
	"github.com/rs/zerolog"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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
	context    *testContext
}
/*
ROOST_METHOD_HASH=GetTags_42221e4328
ROOST_METHOD_SIG_HASH=GetTags_52f72598a3


 */
func TestHandlerGetTags(t *testing.T) {
	type testCase struct {
		name       string
		setupMocks func(as *store.ArticleStore, logger *zerolog.Logger)
		expected   *proto.TagsResponse
		expectErr  bool
	}

	mockDB, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer mockDB.Close()

	articleStore := &store.ArticleStore{}
	logger := &zerolog.Logger{}

	tests := []testCase{
		{
			name: "Successfully Retrieve Tags",
			setupMocks: func(as *store.ArticleStore, logger *zerolog.Logger) {
				mock.ExpectQuery("SELECT * FROM tags").
					WillReturnRows(sqlmock.NewRows([]string{"name"}).AddRow("tag1").AddRow("tag2"))
			},
			expected:  &proto.TagsResponse{Tags: []string{"tag1", "tag2"}},
			expectErr: false,
		},
		{
			name: "Handle Internal Server Error",
			setupMocks: func(as *store.ArticleStore, logger *zerolog.Logger) {
				mock.ExpectQuery("SELECT * FROM tags").
					WillReturnError(errors.New("fail to get tags"))
			},
			expected:  nil,
			expectErr: true,
		},
		{
			name: "No Tags Present",
			setupMocks: func(as *store.ArticleStore, logger *zerolog.Logger) {
				mock.ExpectQuery("SELECT * FROM tags").
					WillReturnRows(sqlmock.NewRows([]string{"name"}))
			},
			expected:  &proto.TagsResponse{Tags: []string{}},
			expectErr: false,
		},
		{
			name: "Valid Context Passed",
			setupMocks: func(as *store.ArticleStore, logger *zerolog.Logger) {
				mock.ExpectQuery("SELECT * FROM tags").
					WillReturnRows(sqlmock.NewRows([]string{"name"}).AddRow("tag1").AddRow("tag2"))
			},
			expected:  &proto.TagsResponse{Tags: []string{"tag1", "tag2"}},
			expectErr: false,
		},
		{
			name: "Invalid Context",
			setupMocks: func(as *store.ArticleStore, logger *zerolog.Logger) {
				mock.ExpectQuery("SELECT * FROM tags").
					WillReturnError(errors.New("context canceled"))
			},
			expected:  nil,
			expectErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			h := &Handler{
				logger: logger,
				as:     articleStore,
			}

			tc.setupMocks(articleStore, logger)

			resp, err := h.GetTags(context.Background(), &proto.Empty{})

			if tc.expectErr {
				assert.Error(t, err)
				if errCode, ok := status.FromError(err); ok {
					assert.Equal(t, codes.Aborted, errCode.Code())
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				assert.Equal(t, tc.expected.Tags, resp.Tags)
			}
		})
	}
}

