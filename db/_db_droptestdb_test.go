package db

import (
	"errors"
	"sync"
	"testing"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jinzhu/gorm"
)

type ExpectedClose struct {
	commonExpectation
}



type T struct {
	common
	isParallel bool
	isEnvSet   bool
	context    *testContext // For running tests and subtests.
}

func TestDropTestDB(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		mockSetup      func(sqlmock.Sqlmock) (*gorm.DB, error)
		expectedError  error
		fConcurrent    bool
		fAlreadyClosed bool
	}{
		{
			name: "Successfully Close a Gorm Database Connection",
			mockSetup: func(mock sqlmock.Sqlmock) (*gorm.DB, error) {
				db, err := gorm.Open("sqlmock", "")
				if err != nil {
					return nil, err
				}
				mock.ExpectClose().WillReturnError(nil)
				return db, nil
			},
			expectedError: nil,
		},
		{
			name: "Handle nil Database Connection Gracefully",
			mockSetup: func(mock sqlmock.Sqlmock) (*gorm.DB, error) {
				return nil, nil
			},
			expectedError: nil,
		},
		{
			name: "Simulate Error on Closing Database Connection",
			mockSetup: func(mock sqlmock.Sqlmock) (*gorm.DB, error) {
				db, err := gorm.Open("sqlmock", "")
				if err != nil {
					return nil, err
				}
				mock.ExpectClose().WillReturnError(errors.New("close error"))
				return db, nil
			},
			expectedError: errors.New("close error"),
		},
		{
			name: "Concurrent Calls to DropTestDB",
			mockSetup: func(mock sqlmock.Sqlmock) (*gorm.DB, error) {
				db, err := gorm.Open("sqlmock", "")
				if err != nil {
					return nil, err
				}
				mock.ExpectClose().WillReturnError(nil)
				return db, nil
			},
			expectedError: nil,
			fConcurrent:   true,
		},
		{
			name: "Check Effect of Dropping a Closed Database",
			mockSetup: func(mock sqlmock.Sqlmock) (*gorm.DB, error) {
				db, err := gorm.Open("sqlmock", "")
				if err != nil {
					return nil, err
				}
				mock.ExpectClose().WillReturnError(nil)
				db.Close()
				return db, nil
			},
			expectedError:  nil,
			fAlreadyClosed: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Logf("Running test case: %s", tt.name)

			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("failed to open sqlmock database: %s", err)
			}
			defer db.Close()

			gormDB, err := tt.mockSetup(mock)
			if err != nil {
				t.Fatalf("failed to set up mock gorm.DB: %s", err)
			}

			if tt.fConcurrent {
				var wg sync.WaitGroup
				for i := 0; i < 5; i++ {
					wg.Add(1)
					go func() {
						defer wg.Done()
						err := DropTestDB(gormDB)
						if err != nil && err.Error() != tt.expectedError.Error() {
							t.Errorf("expected error '%v', got '%v'", tt.expectedError, err)
						}
					}()
				}
				wg.Wait()
			} else {
				err = DropTestDB(gormDB)
				if err != nil && err.Error() != tt.expectedError.Error() {
					t.Errorf("expected error '%v', got '%v'", tt.expectedError, err)
				}

				if err := mock.ExpectationsWereMet(); err != nil {
					t.Errorf("unmet SQL expectations: %s", err)
				}
			}

			t.Logf("Test case `%s` passed!", tt.name)
		})
	}
}
