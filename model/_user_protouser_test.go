package model

import (
	"testing"
	"reflect"
	pb "github.com/raahii/golang-grpc-realworld-example/proto"
)





type T struct {
	common
	isParallel bool
	isEnvSet   bool
	context    *testContext // For running tests and subtests.
}
func TestUserProtoUser(t *testing.T) {

	tests := []struct {
		name  string
		user  User
		token string
		want  *pb.User
	}{
		{
			name: "Successful Conversion of User to Proto User with Valid Token",
			user: User{
				Username: "testuser",
				Email:    "test@example.com",
				Bio:      "This is a bio",
				Image:    "http://example.com/image.png",
			},
			token: "validtoken123",
			want: &pb.User{
				Email:    "test@example.com",
				Token:    "validtoken123",
				Username: "testuser",
				Bio:      "This is a bio",
				Image:    "http://example.com/image.png",
			},
		},
		{
			name: "Conversion of User with Empty Fields to Proto User",
			user: User{
				Username: "emptyuser",
				Email:    "empty@example.com",
				Bio:      "",
				Image:    "",
			},
			token: "emptiestoken",
			want: &pb.User{
				Email:    "empty@example.com",
				Token:    "emptiestoken",
				Username: "emptyuser",
				Bio:      "",
				Image:    "",
			},
		},
		{
			name: "Conversion of User with Special Characters in Fields",
			user: User{
				Username: "specialuser😊",
				Email:    "special@exämple.com",
				Bio:      "Bio with emojis 🤔✨",
				Image:    "http://example.com/image!@#.png",
			},
			token: "specialtoken",
			want: &pb.User{
				Email:    "special@exämple.com",
				Token:    "specialtoken",
				Username: "specialuser😊",
				Bio:      "Bio with emojis 🤔✨",
				Image:    "http://example.com/image!@#.png",
			},
		},
		{
			name: "Consistency Across Multiple ProtoUser Conversions",
			user: User{
				Username: "consistentuser",
				Email:    "consistent@example.com",
				Bio:      "Consistent bio",
				Image:    "http://example.com/consistent.png",
			},
			token: "consisttoken",
			want: &pb.User{
				Email:    "consistent@example.com",
				Token:    "consisttoken",
				Username: "consistentuser",
				Bio:      "Consistent bio",
				Image:    "http://example.com/consistent.png",
			},
		},
		{
			name: "Conversion of Proto User with Null Token Parameter",
			user: User{
				Username: "nulltokenuser",
				Email:    "nulltoken@example.com",
				Bio:      "Some bio",
				Image:    "http://example.com/nullimage.png",
			},
			token: "",
			want: &pb.User{
				Email:    "nulltoken@example.com",
				Token:    "",
				Username: "nulltokenuser",
				Bio:      "Some bio",
				Image:    "http://example.com/nullimage.png",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.user.ProtoUser(tt.token)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("%s failed; got %v, want %v", tt.name, got, tt.want)
				t.Logf("Expected ProtoUser to convert fields as specified.")
			} else {
				t.Logf("%s passed successfully.", tt.name)
			}
		})
	}

}
