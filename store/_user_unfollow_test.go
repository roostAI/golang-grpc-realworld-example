package store

import (
	"testing"
	"github.com/stretchr/testify/assert"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/raahii/golang-grpc-realworld-example/model"
	gorm "github.com/jinzhu/gorm"
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
func TestUserStoreUnfollow(t *testing.T) {
	tests := []struct {
		name          string
		follower      model.User
		followee      model.User
		setupMockFunc func(mock sqlmock.Sqlmock)
		expectedError error
	}{
		{
			name: "Successfully Unfollow a User",
			follower: model.User{
				Model:   gorm.Model{ID: 1},
				Follows: []model.User{{Model: gorm.Model{ID: 2}}},
			},
			followee: model.User{
				Model: gorm.Model{ID: 2},
			},
			setupMockFunc: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec(`DELETE FROM "follows"`).
					WithArgs(1, 2).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
			expectedError: nil,
		},
		{
			name: "Attempt to Unfollow a Non-Followed User",
			follower: model.User{
				Model:   gorm.Model{ID: 1},
				Follows: []model.User{},
			},
			followee: model.User{
				Model: gorm.Model{ID: 2},
			},
			setupMockFunc: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec(`DELETE FROM "follows"`).
					WithArgs(1, 2).
					WillReturnResult(sqlmock.NewResult(1, 0))
				mock.ExpectCommit()
			},
			expectedError: nil,
		},
		{
			name: "Unfollow with a Non-Existent User in the Database",
			follower: model.User{
				Model: gorm.Model{ID: 1},
			},
			followee: model.User{
				Model: gorm.Model{ID: 999},
			},
			setupMockFunc: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec(`DELETE FROM "follows"`).
					WithArgs(1, 999).
					WillReturnError(gorm.ErrRecordNotFound)
				mock.ExpectRollback()
			},
			expectedError: gorm.ErrRecordNotFound,
		},
		{
			name: "Database Error During Unfollow Operation",
			follower: model.User{
				Model:   gorm.Model{ID: 1},
				Follows: []model.User{{Model: gorm.Model{ID: 2}}},
			},
			followee: model.User{
				Model: gorm.Model{ID: 2},
			},
			setupMockFunc: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec(`DELETE FROM "follows"`).
					WithArgs(1, 2).
					WillReturnError(gorm.ErrCantStartTransaction)
				mock.ExpectRollback()
			},
			expectedError: gorm.ErrCantStartTransaction,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			assert.NoError(t, err)
			defer db.Close()
			gormDB, err := gorm.Open("postgres", db)
			assert.NoError(t, err)

			tt.setupMockFunc(mock)

			userStore := &UserStore{db: gormDB}
			err = userStore.Unfollow(&tt.follower, &tt.followee)

			if tt.expectedError != nil {
				assert.EqualError(t, err, tt.expectedError.Error())
			} else {
				assert.NoError(t, err)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
