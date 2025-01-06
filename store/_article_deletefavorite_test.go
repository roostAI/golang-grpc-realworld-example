package store

import (
	"errors"
	"testing"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jinzhu/gorm"
	"github.com/raahii/golang-grpc-realworld-example/model"
	_ "github.com/jinzhu/gorm/dialects/postgres"
)

type ExpectedBegin struct {
	commonExpectation
	delay time.Duration
}

type ExpectedCommit struct {
	commonExpectation
}

type ExpectedExec struct {
	queryBasedExpectation
	result driver.Result
	delay  time.Duration
}

type ExpectedRollback struct {
	commonExpectation
}




type T struct {
	common
	isParallel bool
	isEnvSet   bool
	context    *testContext // For running tests and subtests.
}
func TestArticleStoreDeleteFavorite(t *testing.T) {

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("An error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	gormDB, err := gorm.Open("postgres", db)
	if err != nil {
		t.Fatalf("An error '%s' was not expected when opening a gorm database", err)
	}

	aStore := &ArticleStore{db: gormDB}

	type testCase struct {
		name           string
		article        *model.Article
		user           *model.User
		mockSetup      func()
		expectedCount  int32
		expectError    bool
		finalFavorites []model.User
	}

	tests := []testCase{
		{
			name: "Scenario 1: Successfully Remove Favorite User from Article",
			article: &model.Article{
				FavoritesCount: 1,
				FavoritedUsers: []model.User{{Model: gorm.Model{ID: 1}}},
			},
			user: &model.User{Model: gorm.Model{ID: 1}},
			mockSetup: func() {
				mock.ExpectBegin()
				mock.ExpectExec("DELETE FROM \"favorite_articles\" WHERE").WithArgs(1, 1).WillReturnResult(sqlmock.NewResult(0, 1))
				mock.ExpectExec("UPDATE \"articles\" SET \"favorites_count\" = \"favorites_count\" - ? WHERE \"articles\".\"id\" = ?").
					WithArgs(1, 1).WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
			expectedCount:  0,
			expectError:    false,
			finalFavorites: []model.User{},
		},
		{
			name: "Scenario 2: Removing Non-existing Favorite User",
			article: &model.Article{
				FavoritesCount: 0,
				FavoritedUsers: []model.User{},
			},
			user: &model.User{Model: gorm.Model{ID: 1}},
			mockSetup: func() {
				mock.ExpectBegin()
				mock.ExpectExec("DELETE FROM \"favorite_articles\" WHERE").WithArgs(1, 1).WillReturnResult(sqlmock.NewResult(0, 0))
				mock.ExpectExec("UPDATE \"articles\" SET \"favorites_count\" = \"favorites_count\" - ? WHERE \"articles\".\"id\" = ?").
					WithArgs(1, 1).WillReturnError(errors.New("no rows were updated"))
				mock.ExpectRollback()
			},
			expectedCount:  0,
			expectError:    true,
			finalFavorites: []model.User{},
		},
		{
			name: "Scenario 3: Database Error During User Removal",
			article: &model.Article{
				FavoritesCount: 1,
				FavoritedUsers: []model.User{{Model: gorm.Model{ID: 1}}},
			},
			user: &model.User{Model: gorm.Model{ID: 1}},
			mockSetup: func() {
				mock.ExpectBegin()
				mock.ExpectExec("DELETE FROM \"favorite_articles\" WHERE").WithArgs(1, 1).WillReturnError(errors.New("db error"))
				mock.ExpectRollback()
			},
			expectedCount:  1,
			expectError:    true,
			finalFavorites: []model.User{{Model: gorm.Model{ID: 1}}},
		},
		{
			name: "Scenario 4: Database Error During Count Update",
			article: &model.Article{
				FavoritesCount: 1,
				FavoritedUsers: []model.User{{Model: gorm.Model{ID: 1}}},
			},
			user: &model.User{Model: gorm.Model{ID: 1}},
			mockSetup: func() {
				mock.ExpectBegin()
				mock.ExpectExec("DELETE FROM \"favorite_articles\" WHERE").WithArgs(1, 1).WillReturnResult(sqlmock.NewResult(0, 1))
				mock.ExpectExec("UPDATE \"articles\" SET \"favorites_count\" = \"favorites_count\" - ? WHERE \"articles\".\"id\" = ?").
					WithArgs(1, 1).WillReturnError(errors.New("count update error"))
				mock.ExpectRollback()
			},
			expectedCount:  1,
			expectError:    true,
			finalFavorites: []model.User{{Model: gorm.Model{ID: 1}}},
		},
		{
			name: "Scenario 5: User and Article Interaction With No Prior Relation",
			article: &model.Article{
				FavoritesCount: 0,
				FavoritedUsers: []model.User{},
			},
			user: &model.User{Model: gorm.Model{ID: 2}},
			mockSetup: func() {
				mock.ExpectBegin()
				mock.ExpectExec("DELETE FROM \"favorite_articles\" WHERE").WithArgs(1, 2).WillReturnResult(sqlmock.NewResult(0, 0))
				mock.ExpectExec("UPDATE \"articles\" SET \"favorites_count\" = \"favorites_count\" - ? WHERE \"articles\".\"id\" = ?").
					WithArgs(1, 1).WillReturnError(errors.New("no rows were updated"))
				mock.ExpectRollback()
			},
			expectedCount:  0,
			expectError:    true,
			finalFavorites: []model.User{},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tc.mockSetup()

			err := aStore.DeleteFavorite(tc.article, tc.user)
			if tc.expectError && err == nil {
				t.Errorf("Expected error but got nil")
			}
			if !tc.expectError && err != nil {
				t.Errorf("Did not expect error but got: %v", err)
			}

			if tc.article.FavoritesCount != tc.expectedCount {
				t.Errorf("Expected FavoritesCount to be %d but got %d", tc.expectedCount, tc.article.FavoritesCount)
			}

			if len(tc.article.FavoritedUsers) != len(tc.finalFavorites) {
				t.Errorf("Expected FavoritedUsers length to be %d but got %d", len(tc.finalFavorites), len(tc.article.FavoritedUsers))
			}

			err = mock.ExpectationsWereMet()
			if err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}
