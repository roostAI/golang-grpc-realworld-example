package store

import (
	"errors"
	"testing"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jinzhu/gorm"
	"github.com/raahii/golang-grpc-realworld-example/model"
	"github.com/stretchr/testify/assert"
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
func TestArticleStoreGetCommentByID(t *testing.T) {

	t.Run("Fetch Existing Comment", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		assert.NoError(t, err)
		defer db.Close()

		gormDB, err := gorm.Open("postgres", db)
		assert.NoError(t, err)

		comment := model.Comment{
			Model:     gorm.Model{ID: 1},
			Body:      "Test Comment",
			UserID:    1,
			ArticleID: 1,
		}

		mock.ExpectQuery(`SELECT \* FROM "comments" WHERE "comments"\."deleted_at" IS NULL AND "comments"\."id" = \$1`).
			WithArgs(1).
			WillReturnRows(sqlmock.NewRows([]string{"id", "body", "user_id", "article_id"}).
				AddRow(comment.ID, comment.Body, comment.UserID, comment.ArticleID))

		store := ArticleStore{db: gormDB}
		result, err := store.GetCommentByID(1)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, comment.ID, result.ID)

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Log("Expectations were not met:", err)
		}
	})

	t.Run("Fetch Non-Existing Comment", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		assert.NoError(t, err)
		defer db.Close()

		gormDB, err := gorm.Open("postgres", db)
		assert.NoError(t, err)

		mock.ExpectQuery(`SELECT \* FROM "comments" WHERE "comments"\."deleted_at" IS NULL AND "comments"\."id" = \$1`).
			WithArgs(2).WillReturnError(gorm.ErrRecordNotFound)

		store := ArticleStore{db: gormDB}
		result, err := store.GetCommentByID(2)
		assert.Error(t, err)
		assert.Nil(t, result)

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Log("Expectations were not met:", err)
		}
	})

	t.Run("Database Connection Error", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		assert.NoError(t, err)
		defer db.Close()

		gormDB, err := gorm.Open("postgres", db)
		assert.NoError(t, err)

		mock.ExpectQuery(`SELECT \* FROM "comments" WHERE "comments"\."deleted_at" IS NULL AND "comments"\."id" = \$1`).
			WithArgs(1).WillReturnError(errors.New("database connection error"))

		store := ArticleStore{db: gormDB}
		result, err := store.GetCommentByID(1)
		assert.Error(t, err)
		assert.Nil(t, result)

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Log("Expectations were not met:", err)
		}
	})

	t.Run("Fetch Comment with ID Zero", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		assert.NoError(t, err)
		defer db.Close()

		gormDB, err := gorm.Open("postgres", db)
		assert.NoError(t, err)

		mock.ExpectQuery(`SELECT \* FROM "comments" WHERE "comments"\."deleted_at" IS NULL AND "comments"\."id" = \$1`).
			WithArgs(0).WillReturnError(gorm.ErrRecordNotFound)

		store := ArticleStore{db: gormDB}
		result, err := store.GetCommentByID(0)
		assert.Error(t, err)
		assert.Nil(t, result)

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Log("Expectations were not met:", err)
		}
	})

	t.Run("Specific Error for Non-Existing ID", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		assert.NoError(t, err)
		defer db.Close()

		gormDB, err := gorm.Open("postgres", db)
		assert.NoError(t, err)

		mock.ExpectQuery(`SELECT \* FROM "comments" WHERE "comments"\."deleted_at" IS NULL AND "comments"\."id" = \$1`).
			WithArgs(3).WillReturnError(gorm.ErrRecordNotFound)

		store := ArticleStore{db: gormDB}
		result, err := store.GetCommentByID(3)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, gorm.ErrRecordNotFound))
		assert.Nil(t, result)

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Log("Expectations were not met:", err)
		}
	})
}
