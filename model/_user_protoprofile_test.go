package model

import (
	"testing"
	pb "github.com/raahii/golang-grpc-realworld-example/proto"
)





type T struct {
	common
	isParallel bool
	isEnvSet   bool
	context    *testContext // For running tests and subtests.
}
func TestUserProtoProfile(t *testing.T) {
	tests := []struct {
		name            string
		user            User
		following       bool
		expectedProfile *pb.Profile
	}{
		{
			name: "Convert User to Profile with Following Status True",
			user: User{
				Username: "testuser",
				Bio:      "test bio",
				Image:    "testimage.png",
			},
			following: true,
			expectedProfile: &pb.Profile{
				Username:  "testuser",
				Bio:       "test bio",
				Image:     "testimage.png",
				Following: true,
			},
		},
		{
			name: "Convert User to Profile with Following Status False",
			user: User{
				Username: "anotheruser",
				Bio:      "another bio",
				Image:    "anotherimage.png",
			},
			following: false,
			expectedProfile: &pb.Profile{
				Username:  "anotheruser",
				Bio:       "another bio",
				Image:     "anotherimage.png",
				Following: false,
			},
		},
		{
			name: "Conversion of User with Empty Values",
			user: User{
				Username: "",
				Bio:      "",
				Image:    "",
			},
			following: false,
			expectedProfile: &pb.Profile{
				Username:  "",
				Bio:       "",
				Image:     "",
				Following: false,
			},
		},
		{
			name: "Conversion of a User with Special Characters",
			user: User{
				Username: "special!@#$%",
				Bio:      "bio*&^%",
				Image:    "special.png",
			},
			following: true,
			expectedProfile: &pb.Profile{
				Username:  "special!@#$%",
				Bio:       "bio*&^%",
				Image:     "special.png",
				Following: true,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			actualProfile := tc.user.ProtoProfile(tc.following)

			if actualProfile.Username != tc.expectedProfile.Username ||
				actualProfile.Bio != tc.expectedProfile.Bio ||
				actualProfile.Image != tc.expectedProfile.Image ||
				actualProfile.Following != tc.expectedProfile.Following {
				t.Logf("Expected: %+v, Got: %+v", tc.expectedProfile, actualProfile)
				t.Error("Profile conversion failed.")
			} else {
				t.Logf("Test '%s' passed.\n", tc.name)
			}
		})
	}
}
