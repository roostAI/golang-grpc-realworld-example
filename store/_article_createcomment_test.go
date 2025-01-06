package store

import (
	"errors"
	"testing"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jinzhu/gorm"
	"github.com/raahii/golang-grpc-realworld-example/model"
	"sync"
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
func TestArticleStoreCreateComment(t *testing.T) {
	tests := []struct {
		name       string
		comment    *model.Comment
		setupMocks func(mock sqlmock.Sqlmock)
		expectErr  bool
	}{
		{
			name: "Successfully Create a Valid Comment",
			comment: &model.Comment{
				Body:      "This is a comment",
				UserID:    1,
				ArticleID: 1,
			},
			setupMocks: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec(`INSERT INTO "comments"`).
					WithArgs(sqlmock.AnyArg(), "This is a comment", 1, 1).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
			expectErr: false,
		},
		{
			name: "Attempt to Create a Comment with Missing Required Fields",
			comment: &model.Comment{
				Body:   "",
				UserID: 0,
			},
			setupMocks: func(mock sqlmock.Sqlmock) {

			},
			expectErr: true,
		},
		{
			name: "Handle Database Connection Error When Creating Comment",
			comment: &model.Comment{
				Body:      "Test comment",
				UserID:    1,
				ArticleID: 1,
			},
			setupMocks: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec(`INSERT INTO "comments"`).
					WithArgs(sqlmock.AnyArg(), "Test comment", 1, 1).
					WillReturnError(errors.New("connection error"))
				mock.ExpectRollback()
			},
			expectErr: true,
		},
		{
			name: "Create a Comment with a Non-Existing ArticleID",
			comment: &model.Comment{
				Body:      "Another comment",
				UserID:    1,
				ArticleID: 999,
			},
			setupMocks: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec(`INSERT INTO "comments"`).
					WithArgs(sqlmock.AnyArg(), "Another comment", 1, 999).
					WillReturnError(errors.New("foreign key constraint fails"))
				mock.ExpectRollback()
			},
			expectErr: true,
		},
		{
			name: "Simulate High Concurrency with Simultaneous Comment Creations",
			comment: &model.Comment{
				Body:      "Concurrent comment",
				UserID:    1,
				ArticleID: 1,
			},
			setupMocks: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec(`INSERT INTO "comments"`).
					WithArgs(sqlmock.AnyArg(), "Concurrent comment", 1, 1).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("Failed to open sqlmock database: %s", err)
			}
			defer db.Close()

			tt.setupMocks(mock)

			sqlDB, err := gorm.Open("postgres", db)
			if err != nil {
				t.Fatalf("Failed to initialize gorm database: %s", err)
			}

			store := &ArticleStore{db: sqlDB}

			if tt.name == "Simulate High Concurrency with Simultaneous Comment Creations" {
				var wg sync.WaitGroup
				for i := 0; i < 10; i++ {
					wg.Add(1)
					go func() {
						defer wg.Done()
						err := store.CreateComment(tt.comment)
						if (err != nil) != tt.expectErr {
							t.Errorf("%s: expected error: %v, got: %v", tt.name, tt.expectErr, err)
						}
					}()
				}
				wg.Wait()
			} else {
				err := store.CreateComment(tt.comment)
				if (err != nil) != tt.expectErr {
					t.Errorf("%s: expected error: %v, got: %v", tt.name, tt.expectErr, err)
				}
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("There were unfulfilled expectations: %s", err)
			}
		})
		t.Logf("Successfully tested %s", tt.name)
	}
}
