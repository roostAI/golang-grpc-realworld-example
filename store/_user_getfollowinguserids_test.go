package store

import (
	"errors"
	"testing"
	"github.com/DATA-DOG/go-sqlmock"
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

type Rows struct {
	converter driver.ValueConverter
	cols      []string
	def       []*Column
	rows      [][]driver.Value
	pos       int
	nextErr   map[int]error
	closeErr  error
}




type T struct {
	common
	isParallel bool
	isEnvSet   bool
	context    *testContext // For running tests and subtests.
}
func TestUserStoreGetFollowingUserIDs(t *testing.T) {

	setupMockDB := func() (*UserStore, sqlmock.Sqlmock) {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
		}

		return &UserStore{db: &gorm.DB{CommonDB: db}}, mock
	}

	testCases := []struct {
		name          string
		prepareMock   func(mock sqlmock.Sqlmock)
		user          model.User
		expectedIDs   []uint
		expectedError bool
	}{
		{
			name: "Successfully Retrieve Following User IDs",
			prepareMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"to_user_id"}).
					AddRow(uint(1)).
					AddRow(uint(2)).
					AddRow(uint(3))

				mock.ExpectQuery("^SELECT (.+) FROM follows WHERE (.+)$").
					WithArgs(uint(42)).
					WillReturnRows(rows)
			},
			user:        model.User{Model: gorm.Model{ID: uint(42)}},
			expectedIDs: []uint{1, 2, 3},
		},
		{
			name: "User Is Not Following Anyone",
			prepareMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("^SELECT (.+) FROM follows WHERE (.+)$").
					WithArgs(uint(42)).
					WillReturnRows(sqlmock.NewRows([]string{"to_user_id"}))
			},
			user:        model.User{Model: gorm.Model{ID: uint(42)}},
			expectedIDs: []uint{},
		},
		{
			name: "Database Error While Retrieving Followers",
			prepareMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("^SELECT (.+) FROM follows WHERE (.+)$").
					WithArgs(uint(42)).
					WillReturnError(errors.New("db error"))
			},
			user:          model.User{Model: gorm.Model{ID: uint(42)}},
			expectedIDs:   []uint{},
			expectedError: true,
		},
		{
			name: "Large Number of Followings",
			prepareMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"to_user_id"})
				for i := 0; i < 1000; i++ {
					rows.AddRow(uint(i + 1))
				}
				mock.ExpectQuery("^SELECT (.+) FROM follows WHERE (.+)$").
					WithArgs(uint(42)).
					WillReturnRows(rows)
			},
			user:        model.User{Model: gorm.Model{ID: uint(42)}},
			expectedIDs: createSequentialSlice(1000),
		},
		{
			name: "User ID Does Not Exist",
			prepareMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("^SELECT (.+) FROM follows WHERE (.+)$").
					WithArgs(uint(100)).
					WillReturnRows(sqlmock.NewRows([]string{"to_user_id"}))
			},
			user:        model.User{Model: gorm.Model{ID: uint(100)}},
			expectedIDs: []uint{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			store, mock := setupMockDB()
			defer store.db.Close()

			tc.prepareMock(mock)
			ids, err := store.GetFollowingUserIDs(&tc.user)

			if tc.expectedError {
				if err == nil {
					t.Errorf("expected error but got nil")
				}
			} else {
				if err != nil {
					t.Errorf("did not expect error but got %v", err)
				}
				if !equalSlices(ids, tc.expectedIDs) {
					t.Errorf("expected %v but got %v", tc.expectedIDs, ids)
				}
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}
func createSequentialSlice(n int) []uint {
	slice := make([]uint, n)
	for i := 0; i < n; i++ {
		slice[i] = uint(i + 1)
	}
	return slice
}
func equalSlices(a, b []uint) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
