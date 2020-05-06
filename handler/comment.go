package handler

import (
	"context"
	"fmt"
	"strconv"

	"github.com/raahii/golang-grpc-realworld-example/auth"
	"github.com/raahii/golang-grpc-realworld-example/model"
	pb "github.com/raahii/golang-grpc-realworld-example/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// CreateComment create a comment for an article
func (h *Handler) CreateComment(ctx context.Context, req *pb.CreateCommentRequest) (*pb.CommentResponse, error) {
	h.logger.Info().Msgf("Create comment | req: %+v\n", req)

	// get current user
	userID, err := auth.GetUserID(ctx)
	if err != nil {
		h.logger.Error().Err(err).Msg("unauthenticated")
		return nil, status.Errorf(codes.Unauthenticated, "unauthenticated")
	}

	currentUser, err := h.us.GetByID(userID)
	if err != nil {
		h.logger.Error().Err(err).Msg("current user not found")
		return nil, status.Error(codes.NotFound, "user not found")
	}

	// get article
	articleID, err := strconv.Atoi(req.GetSlug())
	if err != nil {
		msg := fmt.Sprintf("cannot convert slug (%s) into integer", req.GetSlug())
		h.logger.Error().Err(err).Msg(msg)
		return nil, status.Error(codes.InvalidArgument, "invalid article id")
	}

	article, err := h.as.GetByID(uint(articleID))
	if err != nil {
		msg := fmt.Sprintf("requested article (slug=%d) not found", articleID)
		h.logger.Error().Err(err).Msg(msg)
		return nil, status.Error(codes.InvalidArgument, "invalid article id")
	}

	// new comment
	comment := model.Comment{
		Body:      req.GetComment().GetBody(),
		Author:    *currentUser,
		ArticleID: article.ID,
	}

	err = comment.Validate()
	if err != nil {
		err = fmt.Errorf("validation error: %w", err)
		h.logger.Error().Err(err).Msg("validation error")
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// create comment
	err = h.as.CreateComment(&comment)
	if err != nil {
		msg := "failed to create comment."
		h.logger.Error().Err(err).Msg(msg)
		return nil, status.Error(codes.Aborted, msg)
	}

	// map model.Comment to pb.Comment
	pc := comment.ProtoComment()
	pc.Author = currentUser.ProtoProfile(false)

	return &pb.CommentResponse{Comment: pc}, nil
}
