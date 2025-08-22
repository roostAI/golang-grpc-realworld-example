package model

import (
	testing "testing"
	time "time"
	debug "runtime/debug"
	gorm "github.com/jinzhu/gorm"
	proto "github.com/raahii/golang-grpc-realworld-example/proto"
)








/*
ROOST_METHOD_HASH=Comment.ProtoComment_2ec41d63f7
ROOST_METHOD_SIG_HASH=Comment.ProtoComment_fcb2b2ac3a

FUNCTION_DEF=func (c *Comment) ProtoComment() *pb.Comment // ProtoComment generates proto comment model from article


*/
func TestCommentProtoComment(t *testing.T) {

	testCases := []struct {
		name     string
		comment  Comment
		expected *proto.Comment
	}{
		{
			name: "Test ProtoComment with valid Comment data",
			comment: Comment{
				gorm.Model{
					ID:        1,
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				},
				"Test Comment",
				1,
				User{},
				1,
				Article{},
			},
			expected: &proto.Comment{
				Id:        "1",
				Body:      "Test Comment",
				CreatedAt: time.Now().Format(ISO8601),
				UpdatedAt: time.Now().Format(ISO8601),
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Logf("Panic encountered so failing test. %v\n%s", r, string(debug.Stack()))
					t.Fail()
				}
			}()

			result := tc.comment.ProtoComment()

			if result.Id != tc.expected.Id || result.Body != tc.expected.Body || result.CreatedAt != tc.expected.CreatedAt || result.UpdatedAt != tc.expected.UpdatedAt {
				t.Errorf("ProtoComment() = %v; want %v", result, tc.expected)
			} else {
				t.Logf("Success: Expected = %v; Got = %v", tc.expected, result)
			}
		})
	}
}

