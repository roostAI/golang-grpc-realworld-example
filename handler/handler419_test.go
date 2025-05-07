package handler

import (
	store "github.com/raahii/golang-grpc-realworld-example/store"
	zerolog "github.com/rs/zerolog"
	os "os"
	testing "testing"
)








/*
ROOST_METHOD_HASH=New_437eff3b29
ROOST_METHOD_SIG_HASH=New_6e92a7c68a

FUNCTION_DEF=func New(l *zerolog.Logger, us *store.UserStore, as *store.ArticleStore) *Handler // New returns a new handler with logger and database


*/
func TestNew(t *testing.T) {
	testCases := []struct {
		name         string
		logger       zerolog.Logger
		userStore    *store.UserStore
		articleStore *store.ArticleStore
		wantNil      bool
	}{
		{
			name:         "Testing the New Function with Valid Arguments",
			logger:       zerolog.New(os.Stderr).With().Logger(),
			userStore:    &store.UserStore{},
			articleStore: &store.ArticleStore{},
			wantNil:      false,
		},
		{
			name:         "Testing the New Function with Nil Arguments",
			logger:       zerolog.Logger{},
			userStore:    nil,
			articleStore: nil,
			wantNil:      true,
		},
		{
			name:         "New Function Handling Partly Nil Arguments",
			logger:       zerolog.New(os.Stderr).With().Logger(),
			userStore:    &store.UserStore{},
			articleStore: nil,
			wantNil:      false,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Logf("Panic encountered so failing test. %v\n", r)
					t.Fail()
				}
			}()

			h := New(&tt.logger, tt.userStore, tt.articleStore)

			if tt.userStore != nil && h.us != tt.userStore {
				t.Errorf("Expected user store to be equal to what was passed. Got %+v, want %+v", h.us, tt.userStore)
			}
			if tt.articleStore != nil && h.as != tt.articleStore {
				t.Errorf("Expected article store to be equal to what was passed. Got %+v, want %+v", h.as, tt.articleStore)
			}
			if tt.wantNil && (h.logger != nil || h.us != nil || h.as != nil) {
				t.Errorf("Expected handler fields to be nil. Got logger=%+v, userStore=%+v, articleStore=%+v.", h.logger, h.us, h.as)
			}
		})
	}
}

