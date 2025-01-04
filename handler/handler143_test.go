package handler

import (
	"testing"
	"github.com/raahii/golang-grpc-realworld-example/store"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

type T struct {
	common
	isParallel bool
	isEnvSet   bool
	context    *testContext
}
/*
ROOST_METHOD_HASH=New_5541bf24ba
ROOST_METHOD_SIG_HASH=New_7d9b4d5982


 */
func TestNew(t *testing.T) {

	tests := []struct {
		name                 string
		logger               *zerolog.Logger
		userStore            *store.UserStore
		articleStore         *store.ArticleStore
		expectedLogger       *zerolog.Logger
		expectedUserStore    *store.UserStore
		expectedArticleStore *store.ArticleStore
	}{
		{
			name:                 "Valid Inputs",
			logger:               &zerolog.Logger{},
			userStore:            &store.UserStore{DB: &gorm.DB{}},
			articleStore:         &store.ArticleStore{DB: &gorm.DB{}},
			expectedLogger:       &zerolog.Logger{},
			expectedUserStore:    &store.UserStore{DB: &gorm.DB{}},
			expectedArticleStore: &store.ArticleStore{DB: &gorm.DB{}},
		},
		{
			name:                 "Nil Logger",
			logger:               nil,
			userStore:            &store.UserStore{DB: &gorm.DB{}},
			articleStore:         &store.ArticleStore{DB: &gorm.DB{}},
			expectedLogger:       nil,
			expectedUserStore:    &store.UserStore{DB: &gorm.DB{}},
			expectedArticleStore: &store.ArticleStore{DB: &gorm.DB{}},
		},
		{
			name:                 "Nil UserStore",
			logger:               &zerolog.Logger{},
			userStore:            nil,
			articleStore:         &store.ArticleStore{DB: &gorm.DB{}},
			expectedLogger:       &zerolog.Logger{},
			expectedUserStore:    nil,
			expectedArticleStore: &store.ArticleStore{DB: &gorm.DB{}},
		},
		{
			name:                 "Nil ArticleStore",
			logger:               &zerolog.Logger{},
			userStore:            &store.UserStore{DB: &gorm.DB{}},
			articleStore:         nil,
			expectedLogger:       &zerolog.Logger{},
			expectedUserStore:    &store.UserStore{DB: &gorm.DB{}},
			expectedArticleStore: nil,
		},
		{
			name:                 "All Inputs Nil",
			logger:               nil,
			userStore:            nil,
			articleStore:         nil,
			expectedLogger:       nil,
			expectedUserStore:    nil,
			expectedArticleStore: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			h := New(tt.logger, tt.userStore, tt.articleStore)

			assert.Equal(t, tt.expectedLogger, h.logger, "Expected logger to match")
			assert.Equal(t, tt.expectedUserStore, h.us, "Expected user store to match")
			assert.Equal(t, tt.expectedArticleStore, h.as, "Expected article store to match")

			if h.logger == nil {
				t.Log("Logger was nil as expected")
			} else {
				t.Log("Logger was set as expected")
			}

			if h.us == nil {
				t.Log("UserStore was nil as expected")
			} else {
				t.Log("UserStore was set as expected")
			}

			if h.as == nil {
				t.Log("ArticleStore was nil as expected")
			} else {
				t.Log("ArticleStore was set as expected")
			}
		})
	}
}

