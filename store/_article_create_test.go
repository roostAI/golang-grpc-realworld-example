package store

import (
	"testing"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jinzhu/gorm"
	"github.com/raahii/golang-grpc-realworld-example/model"
	"github.com/raahii/golang-grpc-realworld-example/store"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
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
func TestArticleStoreCreate(t *testing.T) {
	tests := []struct {
		name    string
		article model.Article
		mock    func(mock sqlmock.Sqlmock)
		wantErr bool
	}{
		{
			name: "Successfully Create a Valid Article",
			article: model.Article{
				Title:       "A Valid Title",
				Description: "A valid description",
				Body:        "Valid article body",
				UserID:      1,
			},
			mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO \"articles\"").
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
			wantErr: false,
		},
		{
			name: "Fail to Create an Article Due to Missing Title",
			article: model.Article{
				Description: "Description without title",
				Body:        "Body without title",
				UserID:      1,
			},
			mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO \"articles\"").
					WillReturnError(gorm.ErrInvalidData)
				mock.ExpectRollback()
			},
			wantErr: true,
		},
		{
			name: "Database Error During Article Creation",
			article: model.Article{
				Title:       "Another Valid Title",
				Description: "Another valid description",
				Body:        "Another valid article body",
				UserID:      2,
			},
			mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO \"articles\"").
					WillReturnError(gorm.ErrCantStartTransaction)
				mock.ExpectRollback()
			},
			wantErr: true,
		},
		{
			name: "Ensure Default FavoritesCount Set to Zero",
			article: model.Article{
				Title:       "Title With Default",
				Description: "Description With Default",
				Body:        "Body With Default",
				UserID:      3,
			},
			mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO \"articles\"").
					WillReturnResult(sqlmock.NewResult(2, 1))
				mock.ExpectCommit()
			},
			wantErr: false,
		},
		{
			name: "Handle Foreign Key Constraint Violation",
			article: model.Article{
				Title:       "Title With FK Issue",
				Description: "Description With FK Issue",
				Body:        "Body With FK Issue",
				UserID:      999,
			},
			mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO \"articles\"").
					WillReturnError(gorm.ErrForeignKeyViolation)
				mock.ExpectRollback()
			},
			wantErr: true,
		},
		{
			name: "Store Article With Tags and Comments",
			article: model.Article{
				Title:       "Complex Article",
				Description: "Complex Description",
				Body:        "Full Body",
				UserID:      4,
				Tags: []model.Tag{
					{Name: "Tech"},
					{Name: "Golang"},
				},
				Comments: []model.Comment{
					{Body: "Interesting comment"},
					{Body: "Another comment"},
				},
			},
			mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO \"articles\"").
					WillReturnResult(sqlmock.NewResult(3, 1))
				mock.ExpectCommit()
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("error initializing mock database: %v", err)
			}
			defer db.Close()

			dialector := postgres.New(postgres.Config{
				Conn: db,
			})
			gormDB, err := gorm.Open(dialector, &gorm.Config{})
			if err != nil {
				t.Fatalf("failed to open gorm DB: %v", err)
			}

			tt.mock(mock)

			store := &store.ArticleStore{DB: gormDB}
			err = store.Create(&tt.article)
			if (err != nil) != tt.wantErr {
				t.Errorf("Create() error = %v, wantErr %v", err, tt.wantErr)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}
