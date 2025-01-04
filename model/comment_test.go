package model

import (
	"fmt"
	"testing"
	"time"
	"regexp"
	"github.com/stretchr/testify/assert"
	"github.com/jinzhu/gorm"
	"github.com/raahii/golang-grpc-realworld-example/proto"
	"github.com/go-ozzo/ozzo-validation"
)

const ISO8601 = "2006-01-02T15:04:05Z"type T struct {
	common
	isParallel bool
	isEnvSet   bool
	context    *testContext
}
/*
ROOST_METHOD_HASH=ProtoComment_f8354e88c8
ROOST_METHOD_SIG_HASH=ProtoComment_ac7368a67c


 */
func TestCommentProtoComment(t *testing.T) {

	tests := []struct {
		name       string
		comment    Comment
		expectedPB *pb.Comment
	}{
		{
			name: "Scenario 1: Successful Conversion of a Comment to a Proto Comment",
			comment: Comment{
				Model: gorm.Model{
					ID:        1,
					CreatedAt: time.Date(2023, time.March, 10, 15, 0, 0, 0, time.UTC),
					UpdatedAt: time.Date(2023, time.March, 11, 16, 0, 0, 0, time.UTC),
				},
				Body: "Test body",
			},
			expectedPB: &pb.Comment{
				Id:        "1",
				Body:      "Test body",
				CreatedAt: "2023-03-10T15:00:00Z",
				UpdatedAt: "2023-03-11T16:00:00Z",
			},
		},
		{
			name:    "Scenario 2: Handling Zero Values in Comment Fields",
			comment: Comment{},
			expectedPB: &pb.Comment{
				Id:        "0",
				Body:      "",
				CreatedAt: "0001-01-01T00:00:00Z",
				UpdatedAt: "0001-01-01T00:00:00Z",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			protoComment := test.comment.ProtoComment()

			assert.Equal(t, test.expectedPB.Id, protoComment.Id)

			assert.Equal(t, test.expectedPB.Body, protoComment.Body)

			matched, _ := regexp.MatchString(`\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}Z`, protoComment.CreatedAt)
			assert.True(t, matched, "CreatedAt is not in ISO8601 format")

			matched, _ = regexp.MatchString(`\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}Z`, protoComment.UpdatedAt)
			assert.True(t, matched, "UpdatedAt is not in ISO8601 format")

			if assert.Equal(t, test.expectedPB, protoComment) {
				t.Logf("Success: %s", test.name)
			} else {
				t.Logf("Failure: %s", test.name)
			}
		})
	}

	t.Run("Scenario 4: Concurrent Access Handling", func(t *testing.T) {
		comment := Comment{
			Model: gorm.Model{ID: 1, CreatedAt: time.Now(), UpdatedAt: time.Now()},
			Body:  "Concurrent Test Body",
		}

		const routineCount = 100
		results := make(chan *pb.Comment, routineCount)

		for i := 0; i < routineCount; i++ {
			go func() {
				results <- comment.ProtoComment()
			}()
		}

		for i := 0; i < routineCount; i++ {
			protoComment := <-results
			assert.Equal(t, "1", protoComment.Id)
			assert.Equal(t, "Concurrent Test Body", protoComment.Body)
		}

		t.Log("Concurrent access handled successfully, with consistent data across goroutines.")
	})
}


/*
ROOST_METHOD_HASH=Validate_1df97b5695
ROOST_METHOD_SIG_HASH=Validate_0591f679fe


 */
func TestCommentValidate(t *testing.T) {

	tests := []struct {
		name         string
		comment      Comment
		expectError  bool
		errorMessage string
	}{
		{"Valid Comment", Comment{Body: "This is a valid comment body."}, false, ""},
		{"Missing Body", Comment{Body: ""}, true, "Body: cannot be blank."},
		{"Whitespace Body", Comment{Body: "    "}, true, "Body: cannot be blank."},
		{"Large Body", Comment{Body: string(make([]byte, 10000))}, false, ""},
		{"Special Characters", Comment{Body: "!@#$%^&*() 😊"}, false, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Logf("Running Scenario: %s", tt.name)

			err := tt.comment.Validate()

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got nil for: %s", tt.name)
				} else if err.Error() != tt.errorMessage {
					t.Errorf("Expected error message '%s', got '%s'", tt.errorMessage, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %s for: %s", err.Error(), tt.name)
				}
			}

			t.Logf("Finished Scenario: %s", tt.name)
		})
	}
}

