package store

import (
	fmt "fmt"
	testing "testing"

	gorm "github.com/jinzhu/gorm"
)

/*
ROOST_METHOD_HASH=NewUserStore_fb599438e5
ROOST_METHOD_SIG_HASH=NewUserStore_c0075221af

FUNCTION_DEF=func NewUserStore(db *gorm.DB) *UserStore // NewUserStore returns a new UserStore
*/
func TestNewUserStoreFailingDBInitialization(t *testing.T) {

	db := &gorm.DB{Error: fmt.Errorf("initialization failed")}
	store := NewUserStore(db)
	if store.db == nil {
		t.Error("NewUserStore() did not handle failing DB initialization gracefully")
	}
	if store.db.Error.Error() != "initialization failed" {
		t.Error("NewUserStore() did not preserve the failing DB error")
	}
}

func TestNewUserStoreWithDifferentDBConnections(t *testing.T) {
	db1 := &gorm.DB{Value: "db1"}
	db2 := &gorm.DB{Value: "db2"}
	store1 := NewUserStore(db1)
	store2 := NewUserStore(db2)
	if store1.db == store2.db {
		t.Error("NewUserStore() did not create different instances for different DB connections")
	}
	if store1.db.Value != "db1" || store2.db.Value != "db2" {
		t.Error("NewUserStore() did not preserve DB properties for different DB connections")
	}
}

func TestNewUserStoreWithSameDBInstanceDifferentConfigurations(t *testing.T) {
	db := &gorm.DB{Value: "original"}
	store1 := NewUserStore(db)
	db.Value = "modified"
	store2 := NewUserStore(db)
	if store1.db != store2.db {
		t.Error("NewUserStore() did not use the same DB reference for different configurations")
	}
	if store1.db.Value != "modified" || store2.db.Value != "modified" {
		t.Error("NewUserStore() did not preserve the modified DB properties")
	}
}
