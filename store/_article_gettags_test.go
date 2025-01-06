package store

import (
	"testing"
	"fmt"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jinzhu/gorm"
	"github.com/raahii/golang-grpc-realworld-example/model"
	"github.com/stretchr/testify/assert"
	_ "github.com/go-sql-driver/mysql"
	"sync"
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

func TestArticleStoreGetTags(t *testing.T) {
	tests := []struct {
		name           string
		mockSetup      func(sqlmock.Sqlmock)
		expectedTags   []model.Tag
		expectedError  error
		concurrentTest bool
	}{
		{
			name: "Retrieving Tags Successfully",
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "name"}).
					AddRow(1, "go").
					AddRow(2, "programming")
				mock.ExpectQuery("^SELECT (.+) FROM `tags`").WillReturnRows(rows)
			},
			expectedTags: []model.Tag{
				{Model: gorm.Model{ID: 1}, Name: "go"},
				{Model: gorm.Model{ID: 2}, Name: "programming"},
			},
			expectedError:  nil,
			concurrentTest: false,
		},
		{
			name: "Database Error While Retrieving Tags",
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("^SELECT (.+) FROM `tags`").WillReturnError(gorm.ErrInvalidSQL)
			},
			expectedTags:   nil,
			expectedError:  gorm.ErrInvalidSQL,
			concurrentTest: false,
		},
		{
			name: "No Tags Available in the Database",
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "name"})
				mock.ExpectQuery("^SELECT (.+) FROM `tags`").WillReturnRows(rows)
			},
			expectedTags:   []model.Tag{},
			expectedError:  nil,
			concurrentTest: false,
		},
		{
			name: "Database Connection Issue",
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("^SELECT (.+) FROM `tags`").WillReturnError(gorm.ErrInvalidTransaction)
			},
			expectedTags:   nil,
			expectedError:  gorm.ErrInvalidTransaction,
			concurrentTest: false,
		},
		{
			name: "Concurrent Access to the Tags Table",
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "name"}).
					AddRow(1, "go").
					AddRow(2, "programming")
				mock.ExpectQuery("^SELECT (.+) FROM `tags`").WillReturnRows(rows)
			},
			expectedTags: []model.Tag{
				{Model: gorm.Model{ID: 1}, Name: "go"},
				{Model: gorm.Model{ID: 2}, Name: "programming"},
			},
			expectedError:  nil,
			concurrentTest: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			db, mock, err := sqlmock.New()
			assert.NoError(t, err)
			defer db.Close()

			gormDB, err := gorm.Open("mysql", db)
			assert.NoError(t, err)
			store := &ArticleStore{db: gormDB}

			tt.mockSetup(mock)

			if tt.concurrentTest {
				var wg sync.WaitGroup
				results := make(chan struct {
					tags []model.Tag
					err  error
				}, 5)

				for i := 0; i < 5; i++ {
					wg.Add(1)
					go func() {
						defer wg.Done()
						tags, err := store.GetTags()
						results <- struct {
							tags []model.Tag
							err  error
						}{tags, err}
					}()
				}

				wg.Wait()
				close(results)

				for res := range results {
					assert.Equal(t, tt.expectedTags, res.tags)
					assert.Equal(t, tt.expectedError, res.err)
				}
			} else {

				tags, err := store.GetTags()

				assert.Equal(t, tt.expectedTags, tags)
				assert.Equal(t, tt.expectedError, err)

				if tt.expectedError == nil {
					t.Logf("Passed: %s", tt.name)
				} else {
					t.Logf("Error caught as expected in: %s - %v", tt.name, err)
				}
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("Expectations were not met: %v", err)
			}
		})
	}
}
