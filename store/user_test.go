package store

import (
	"reflect"
	"sync"
	"testing"
	"github.com/jinzhu/gorm"
)








/*
ROOST_METHOD_HASH=NewUserStore_fb599438e5
ROOST_METHOD_SIG_HASH=NewUserStore_c0075221af

FUNCTION_DEF=func NewUserStore(db *gorm.DB) *UserStore // NewUserStore returns a new UserStore


*/
func TestNewUserStore(t *testing.T) {
	tests := []struct {
		name string
		db   *gorm.DB
		want *UserStore
	}{
		{
			name: "Create UserStore with valid gorm.DB instance",
			db:   &gorm.DB{},
			want: &UserStore{db: &gorm.DB{}},
		},
		{
			name: "Create UserStore with nil gorm.DB instance",
			db:   nil,
			want: &UserStore{db: nil},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewUserStore(tt.db)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewUserStore() = %v, want %v", got, tt.want)
			}
		})
	}

	t.Run("Verify UserStore uniqueness for different gorm.DB instances", func(t *testing.T) {
		db1 := &gorm.DB{}
		db2 := &gorm.DB{}

		store1 := NewUserStore(db1)
		store2 := NewUserStore(db2)

		if store1 == store2 {
			t.Error("NewUserStore() returned the same instance for different db connections")
		}

		if store1.db != db1 {
			t.Error("NewUserStore() did not set the correct db for store1")
		}

		if store2.db != db2 {
			t.Error("NewUserStore() did not set the correct db for store2")
		}
	})

	t.Run("Check UserStore creation with a configured gorm.DB instance", func(t *testing.T) {

		configuredDB := &gorm.DB{}

		store := NewUserStore(configuredDB)

		if store.db != configuredDB {
			t.Error("NewUserStore() did not preserve the configured db instance")
		}
	})

	t.Run("Verify thread-safety of NewUserStore", func(t *testing.T) {
		var wg sync.WaitGroup
		numGoroutines := 10

		stores := make([]*UserStore, numGoroutines)
		dbs := make([]*gorm.DB, numGoroutines)

		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			dbs[i] = &gorm.DB{}

			go func(index int) {
				defer wg.Done()
				stores[index] = NewUserStore(dbs[index])
			}(i)
		}

		wg.Wait()

		for i := 0; i < numGoroutines; i++ {
			if stores[i] == nil {
				t.Errorf("NewUserStore() failed to create UserStore in goroutine %d", i)
			}
			if stores[i].db != dbs[i] {
				t.Errorf("NewUserStore() set incorrect db for UserStore in goroutine %d", i)
			}
		}
	})
}

