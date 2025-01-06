package store

import (
	"testing"
	"fmt"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jinzhu/gorm"
	"github.com/raahii/golang-grpc-realworld-example/model"
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
func TestArticleStoreAddFavorite(t *testing.T) {
	tests := []struct {
		name        string
		article     *model.Article
		user        *model.User
		mock        func(mock sqlmock.Sqlmock)
		expectedErr error
		finalCount  int
	}{
		{
			name: "Scenario 1: Successfully Add a Favorite to an Article",
			article: &model.Article{
				Model:          gorm.Model{ID: 1},
				FavoritesCount: 0,
			},
			user: &model.User{
				Model: gorm.Model{ID: 1},
			},
			mock: func(mock sqlmock.Sqlmock) {

				mock.ExpectBegin()

				mock.ExpectExec("INSERT INTO `favorite_articles`").
					WillReturnResult(sqlmock.NewResult(1, 1))

				mock.ExpectExec("UPDATE `articles` SET `favorites_count`=favorites_count +").
					WillReturnResult(sqlmock.NewResult(1, 1))

				mock.ExpectCommit()
			},
			expectedErr: nil,
			finalCount:  1,
		},
		{
			name: "Scenario 2: Add Favorite When Article Already Favorited by the User",
			article: &model.Article{
				Model:          gorm.Model{ID: 1},
				FavoritesCount: 1,
				FavoritedUsers: []model.User{
					{Model: gorm.Model{ID: 1}},
				},
			},
			user: &model.User{
				Model: gorm.Model{ID: 1},
			},
			mock: func(mock sqlmock.Sqlmock) {

				mock.ExpectBegin()

				mock.ExpectExec("INSERT INTO `favorite_articles`").
					WillReturnError(fmt.Errorf("duplicate entry"))

				mock.ExpectRollback()
			},
			expectedErr: fmt.Errorf("duplicate entry"),
			finalCount:  1,
		},
		{
			name:    "Scenario 3: Add Favorite with Nonexistent Article",
			article: nil,
			user: &model.User{
				Model: gorm.Model{ID: 1},
			},
			mock: func(mock sqlmock.Sqlmock) {

			},
			expectedErr: fmt.Errorf("null-pointer dereference"),
			finalCount:  0,
		},
		{
			name: "Scenario 4: Database Transaction Rollback on Append Error",
			article: &model.Article{
				Model:          gorm.Model{ID: 1},
				FavoritesCount: 0,
			},
			user: &model.User{
				Model: gorm.Model{ID: 1},
			},
			mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()

				mock.ExpectExec("INSERT INTO `favorite_articles`").
					WillReturnError(fmt.Errorf("append error"))
				mock.ExpectRollback()
			},
			expectedErr: fmt.Errorf("append error"),
			finalCount:  0,
		},
		{
			name: "Scenario 5: Database Transaction Rollback on Favorites Count Update Error",
			article: &model.Article{
				Model:          gorm.Model{ID: 1},
				FavoritesCount: 0,
			},
			user: &model.User{
				Model: gorm.Model{ID: 1},
			},
			mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()

				mock.ExpectExec("INSERT INTO `favorite_articles`").
					WillReturnResult(sqlmock.NewResult(1, 1))

				mock.ExpectExec("UPDATE `articles` SET `favorites_count`=favorites_count +").
					WillReturnError(fmt.Errorf("update error"))
				mock.ExpectRollback()
			},
			expectedErr: fmt.Errorf("update error"),
			finalCount:  0,
		},
		{
			name: "Scenario 6: Concurrent Favoriting by Multiple Users",
			article: &model.Article{
				Model:          gorm.Model{ID: 1},
				FavoritesCount: 0,
			},
			user: &model.User{
				Model: gorm.Model{ID: 2},
			},
			mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()

				mock.ExpectExec("INSERT INTO `favorite_articles`").
					WillReturnResult(sqlmock.NewResult(1, 1))

				mock.ExpectExec("UPDATE `articles` SET `favorites_count`=favorites_count +").
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
			expectedErr: nil,
			finalCount:  1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("failed to open sqlmock database: %s", err)
			}
			defer db.Close()

			gormDB, err := gorm.Open("mysql", db)
			if err != nil {
				t.Fatalf("failed to open gorm DB: %s", err)
			}

			store := &ArticleStore{db: gormDB}

			tt.mock(mock)

			err = store.AddFavorite(tt.article, tt.user)
			if err != nil && tt.expectedErr == nil {
				t.Errorf("expected no error, but got: %v", err)
			}
			if err == nil && tt.expectedErr != nil {
				t.Errorf("expected error: %v, but got none", tt.expectedErr)
			}
			if err != nil && tt.expectedErr != nil && err.Error() != tt.expectedErr.Error() {
				t.Errorf("expected error: %v, but got: %v", tt.expectedErr, err)
			}

			if tt.article != nil && int(tt.article.FavoritesCount) != tt.finalCount {
				t.Errorf("expected favorites count: %v, but got: %v", tt.finalCount, tt.article.FavoritesCount)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}
