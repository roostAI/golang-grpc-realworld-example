package store

import (
	"errors"
	"testing"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jinzhu/gorm"
	"github.com/raahii/golang-grpc-realworld-example/model"
	"github.com/stretchr/testify/assert"
)

type ExpectedExec struct {
	queryBasedExpectation
	result driver.Result
	delay  time.Duration
}





type T struct {
	common
	isParallel bool
	isEnvSet   bool
	context    *testContext // For running tests and subtests.
}
func TestArticleStoreDeleteComment(t *testing.T) {
	type args struct {
		comment *model.Comment
	}
	tests := []struct {
		name     string
		mockFunc func(sqlmock.Sqlmock)
		args     args
		wantErr  bool
	}{
		{
			name: "Successful Deletion of an Existing Comment",
			mockFunc: func(mock sqlmock.Sqlmock) {

				mock.ExpectExec("DELETE FROM `comments` WHERE `comments`.`deleted_at` IS NULL AND \\(`comments`.`id` = \\?\\)").WithArgs(1).WillReturnResult(sqlmock.NewResult(1, 1))
			},
			args: args{
				comment: &model.Comment{Model: gorm.Model{ID: 1}, Body: "Test Comment", UserID: 1, ArticleID: 1},
			},
			wantErr: false,
		},
		{
			name: "Attempt to Delete a Non-Existent Comment",
			mockFunc: func(mock sqlmock.Sqlmock) {

				mock.ExpectExec("DELETE FROM `comments` WHERE `comments`.`deleted_at` IS NULL AND \\(`comments`.`id` = \\?\\)").WithArgs(999).WillReturnResult(sqlmock.NewResult(0, 0))
			},
			args: args{
				comment: &model.Comment{Model: gorm.Model{ID: 999}, Body: "Non-existent Comment", UserID: 999, ArticleID: 999},
			},
			wantErr: true,
		},
		{
			name: "Deletion with Database Error",
			mockFunc: func(mock sqlmock.Sqlmock) {

				mock.ExpectExec("DELETE FROM `comments` WHERE `comments`.`deleted_at` IS NULL AND \\(`comments`.`id` = \\?\\)").WithArgs(2).WillReturnError(errors.New("database error"))
			},
			args: args{
				comment: &model.Comment{Model: gorm.Model{ID: 2}, Body: "Error-prone Comment", UserID: 2, ArticleID: 2},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			assert.NoError(t, err)
			defer db.Close()

			gormDB, err := gorm.Open("sqlite3", db)
			assert.NoError(t, err)

			store := &ArticleStore{db: gormDB}

			tt.mockFunc(mock)

			err = store.DeleteComment(tt.args.comment)
			assert.Equal(t, tt.wantErr, err != nil)

			err = mock.ExpectationsWereMet()
			assert.NoError(t, err)
		})
	}
}
