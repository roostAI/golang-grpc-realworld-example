package store

import (
	"testing"

	"github.com/jinzhu/gorm"
	"github.com/raahii/golang-grpc-realworld-example/model"
)

// mockDB implements minimal gorm.DB functionality needed for testing
type mockDB struct {
	modelValue interface{}
	updateErr  error
}

func (m *mockDB) Model(value interface{}) *gorm.DB {
	m.modelValue = value
	return &gorm.DB{} 
}

func (m *mockDB) Update(attrs ...interface{}) *gorm.DB {
	return &gorm.DB{Error: m.updateErr}
}

func TestUserStoreUpdate(t *testing.T) {
	tests := []struct {
		name    string
		user    *model.User
		dbError error
		wantErr bool
	}{
		{
			name: "successful update",
			user: &model.User{
				ID:       1,
				Username: "testuser",
				Email:    "test@example.com",
			},
			dbError: nil,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock DB
			mockDB := &mockDB{
				updateErr: tt.dbError,
			}

			// Initialize UserStore with mock DB
			store := &UserStore{
				db: &gorm.DB{}, // Using mock DB
			}

			// Call Update
			err := store.Update(tt.user)

			// Verify results
			if (err != nil) != tt.wantErr {
				t.Errorf("UserStore.Update() error = %v, wantErr %v", err, tt.wantErr)
			}

			// Verify model value was passed correctly
			if mockDB.modelValue != tt.user {
				t.Errorf("UserStore.Update() model value = %v, want %v", mockDB.modelValue, tt.user)
			}
		})
	}
}
