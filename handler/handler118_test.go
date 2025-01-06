package handler

import (
	"testing"
	"github.com/raahii/golang-grpc-realworld-example/store"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)


/*
ROOST_METHOD_HASH=New_5541bf24ba
ROOST_METHOD_SIG_HASH=New_7d9b4d5982


 */
func TestNew(t *testing.T) {
	type testCase struct {
		name                  string
		logger                *zerolog.Logger
		userStore             *store.UserStore
		articleStore          *store.ArticleStore
		expectNilLogger       bool
		expectNilUserStore    bool
		expectNilArticleStore bool
	}

	tests := []testCase{
		{
			name:                  "Create Handler with Valid Logger and Stores",
			logger:                &zerolog.Logger{},
			userStore:             &store.UserStore{},
			articleStore:          &store.ArticleStore{},
			expectNilLogger:       false,
			expectNilUserStore:    false,
			expectNilArticleStore: false,
		},
		{
			name:                  "Create Handler with Nil Logger",
			logger:                nil,
			userStore:             &store.UserStore{},
			articleStore:          &store.ArticleStore{},
			expectNilLogger:       true,
			expectNilUserStore:    false,
			expectNilArticleStore: false,
		},
		{
			name:                  "Create Handler with Nil UserStore",
			logger:                &zerolog.Logger{},
			userStore:             nil,
			articleStore:          &store.ArticleStore{},
			expectNilLogger:       false,
			expectNilUserStore:    true,
			expectNilArticleStore: false,
		},
		{
			name:                  "Create Handler with Nil ArticleStore",
			logger:                &zerolog.Logger{},
			userStore:             &store.UserStore{},
			articleStore:          nil,
			expectNilLogger:       false,
			expectNilUserStore:    false,
			expectNilArticleStore: true,
		},
		{
			name:                  "Full Nil Input Test",
			logger:                nil,
			userStore:             nil,
			articleStore:          nil,
			expectNilLogger:       true,
			expectNilUserStore:    true,
			expectNilArticleStore: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			handler := New(tc.logger, tc.userStore, tc.articleStore)
			assert.NotNil(t, handler, "Handler should not be nil")

			if tc.expectNilLogger {
				assert.Nil(t, handler.logger, "Logger should be nil")
			} else {
				assert.NotNil(t, handler.logger, "Logger should not be nil")
			}

			if tc.expectNilUserStore {
				assert.Nil(t, handler.us, "UserStore should be nil")
			} else {
				assert.NotNil(t, handler.us, "UserStore should not be nil")
			}

			if tc.expectNilArticleStore {
				assert.Nil(t, handler.as, "ArticleStore should be nil")
			} else {
				assert.NotNil(t, handler.as, "ArticleStore should not be nil")
			}

			t.Logf("Test scenario '%s' passed successfully.", tc.name)
		})
	}
}

