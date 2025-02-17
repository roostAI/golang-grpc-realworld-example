package store

import (
	"reflect"
	"testing"
	"time"

	"github.com/jinzhu/gorm"
)

/*
ROOST_METHOD_HASH=NewUserStore_fb599438e5
ROOST_METHOD_SIG_HASH=NewUserStore_c0075221af

FUNCTION_DEF=func NewUserStore(db *gorm.DB) *UserStore // NewUserStore returns a new UserStore
*/
func TestNewUserStore(t *testing.T) {
	type args struct {
		db *gorm.DB
	}
	tests := []struct {
		name string
		args args
		want *UserStore
	}{
		{
			name: "Create a New UserStore with Valid DB Connection",
			args: args{
				db: &gorm.DB{},
			},
			want: &UserStore{db: &gorm.DB{}},
		},
		{
			name: "Create a New UserStore with Nil DB Connection",
			args: args{
				db: nil,
			},
			want: &UserStore{db: nil},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewUserStore(tt.args.db)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewUserStore() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewUserStoreDBReferenceIntegrity(t *testing.T) {
	db := &gorm.DB{
		Value: "unique_identifier",
	}
	store := NewUserStore(db)

	if store.db != db {
		t.Error("NewUserStore() did not maintain DB reference integrity")
	}

	if store.db.Value != "unique_identifier" {
		t.Error("NewUserStore() did not preserve DB properties")
	}
}

func TestNewUserStoreImmutability(t *testing.T) {
	db := &gorm.DB{}
	store1 := NewUserStore(db)
	store2 := NewUserStore(db)

	if store1 == store2 {
		t.Error("NewUserStore() returned the same instance for different calls")
	}

	if store1.db != store2.db {
		t.Error("NewUserStore() did not use the same DB reference for different instances")
	}
}

func TestNewUserStorePerformance(t *testing.T) {
	db := &gorm.DB{}
	iterations := 1000

	start := time.Now()
	for i := 0; i < iterations; i++ {
		NewUserStore(db)
	}
	duration := time.Since(start)

	t.Logf("Time taken to create %d UserStore instances: %v", iterations, duration)
	if duration > time.Second {
		t.Errorf("NewUserStore() took too long to create %d instances", iterations)
	}
}
