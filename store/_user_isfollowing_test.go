package store

import (
	"testing"
	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/jinzhu/gorm"
	"github.com/raahii/golang-grpc-realworld-example/model"
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
	context    *testContext // For running tests and subtests.
}

func TestUserStoreIsFollowing(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	gormDB, err := gorm.Open("mysql", db)
	if err != nil {
		t.Fatalf("unexpected error when opening gorm: %v", err)
	}
	defer gormDB.Close()

	userStore := &UserStore{db: gormDB}

	type TestCase struct {
		name          string
		userA         *model.User
		userB         *model.User
		mockBehaviour func()
		expected      bool
		expectError   bool
	}

	testCases := []TestCase{
		{
			name: "Scenario 1: Check If User A Follows User B",
			userA: &model.User{
				Model: gorm.Model{ID: 1},
			},
			userB: &model.User{
				Model: gorm.Model{ID: 2},
			},
			mockBehaviour: func() {
				mock.ExpectQuery("^SELECT count(.+) FROM `follows` WHERE (.+)$").
					WithArgs(1, 2).
					WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
			},
			expected:    true,
			expectError: false,
		},
		{
			name: "Scenario 2: User A Does Not Follow User B",
			userA: &model.User{
				Model: gorm.Model{ID: 1},
			},
			userB: &model.User{
				Model: gorm.Model{ID: 2},
			},
			mockBehaviour: func() {
				mock.ExpectQuery("^SELECT count(.+) FROM `follows` WHERE (.+)$").
					WithArgs(1, 2).
					WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
			},
			expected:    false,
			expectError: false,
		},
		{
			name:          "Scenario 3: User A or User B is Nil",
			userA:         nil,
			userB:         nil,
			mockBehaviour: func() {},
			expected:      false,
			expectError:   false,
		},
		{
			name: "Scenario 4: Database Error Handling",
			userA: &model.User{
				Model: gorm.Model{ID: 1},
			},
			userB: &model.User{
				Model: gorm.Model{ID: 2},
			},
			mockBehaviour: func() {
				mock.ExpectQuery("^SELECT count(.+) FROM `follows` WHERE (.+)$").
					WithArgs(1, 2).
					WillReturnError(gorm.ErrInvalidSQL)
			},
			expected:    false,
			expectError: true,
		},
		{
			name: "Scenario 5: Check for Self-following",
			userA: &model.User{
				Model: gorm.Model{ID: 1},
			},
			userB: &model.User{
				Model: gorm.Model{ID: 1},
			},
			mockBehaviour: func() {
				mock.ExpectQuery("^SELECT count(.+) FROM `follows` WHERE (.+)$").
					WithArgs(1, 1).
					WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
			},
			expected:    false,
			expectError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.mockBehaviour()

			result, err := userStore.IsFollowing(tc.userA, tc.userB)

			if tc.expectError {
				if err == nil {
					t.Logf("expected error but got nil")
					t.Fail()
				}
			} else {
				if err != nil {
					t.Logf("did not expect error but got: %v", err)
					t.Fail()
				}
			}

			if result != tc.expected {
				t.Errorf("expected %v, got %v", tc.expected, result)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %v", err)
			}
		})
	}
}
