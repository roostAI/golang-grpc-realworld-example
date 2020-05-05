package handler

import (
	"context"
	"fmt"
	"strconv"

	"github.com/k0kubun/pp"
	"github.com/raahii/golang-grpc-realworld-example/auth"
	"github.com/raahii/golang-grpc-realworld-example/model"
	pb "github.com/raahii/golang-grpc-realworld-example/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// CreateArticle creates a article
func (h *Handler) CreateArticle(ctx context.Context, req *pb.CreateAritcleRequest) (*pb.ArticleResponse, error) {
	h.logger.Info().Msgf("Create artcile | req: %+v\n", req)

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

	ra := req.GetArticle()
	tags := make([]model.Tag, 0, len(ra.GetTagList()))
	for _, t := range ra.GetTagList() {
		tags = append(tags, model.Tag{Name: t})
	}

	article := model.Article{
		Title:       ra.GetTitle(),
		Description: ra.GetDescription(),
		Body:        ra.GetBody(),
		Author:      *currentUser,
		Tags:        tags,
	}

	err = article.Validate()
	if err != nil {
		msg := "validation error"
		h.logger.Error().Err(err).Msg(msg)
		return nil, status.Error(codes.InvalidArgument, msg)
	}

	err = h.as.Create(&article)
	if err != nil {
		msg := "Failed to create user."
		h.logger.Error().Err(err).Msg(msg)
		return nil, status.Error(codes.Canceled, msg)
	}

	// get whether the article is current user's favorite
	favorited := false
	pa := article.ProtoArticle(favorited)

	// get whether current user follows article author
	following, err := h.us.IsFollowing(currentUser, &article.Author)
	if err != nil {
		msg := "failed to get following status"
		h.logger.Error().Err(err).Msg(msg)
		return nil, status.Error(codes.NotFound, "internal server error")
	}
	pa.Author = article.Author.ProtoProfile(following)

	return &pb.ArticleResponse{Article: pa}, nil
}

// GetArticle gets a article
func (h *Handler) GetArticle(ctx context.Context, req *pb.GetArticleRequest) (*pb.ArticleResponse, error) {
	h.logger.Info().Msgf("Get artcile | req: %+v\n", req)

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
		pp.Println(err)
		return nil, status.Error(codes.InvalidArgument, "invalid article id")
	}

	// get current user if exists
	userID, err := auth.GetUserID(ctx)
	if err != nil {
		pa := article.ProtoArticle(false)
		pa.Author = article.Author.ProtoProfile(false)
		return &pb.ArticleResponse{Article: pa}, nil
	}

	currentUser, err := h.us.GetByID(userID)
	if err != nil {
		msg := fmt.Sprintf("token is valid but the user not found")
		h.logger.Error().Err(err).Msg(msg)
		return nil, status.Error(codes.NotFound, msg)
	}

	// get whether the article is current user's favorite
	favorited := false
	pa := article.ProtoArticle(favorited)

	// get whether current user follows article author
	following, err := h.us.IsFollowing(currentUser, &article.Author)
	if err != nil {
		msg := "failed to get following status"
		h.logger.Error().Err(err).Msg(msg)
		return nil, status.Error(codes.NotFound, "internal server error")
	}
	pa.Author = article.Author.ProtoProfile(following)

	return &pb.ArticleResponse{Article: pa}, nil
}

// GetArticles gets recent articles globally
func (h *Handler) GetArticles(ctx context.Context, req *pb.GetArticlesRequest) (*pb.ArticlesResponse, error) {
	h.logger.Info().Msgf("Get artciles | req: %+v\n", req)

	limitQuery := req.GetLimit()
	if limitQuery == 0 {
		limitQuery = 20
	}

	as, err := h.as.GetArticles(req.GetTag(), req.GetAuthor(), limitQuery, req.GetOffset())
	if err != nil {
		h.logger.Error().Err(err).Msg("failed to search articles in the database")
		return nil, status.Error(codes.Aborted, "internal server error")
	}

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

	pas := make([]*pb.Article, 0, len(as))
	for _, a := range as {
		// get whether the article is current user's favorite
		favorited := false
		pa := a.ProtoArticle(favorited)

		// get whether current user follows article author
		following, err := h.us.IsFollowing(currentUser, &a.Author)
		if err != nil {
			msg := "failed to get following status"
			h.logger.Error().Err(err).Msg(msg)
			return nil, status.Error(codes.NotFound, "internal server error")
		}
		pa.Author = a.Author.ProtoProfile(following)

		pas = append(pas, pa)
	}

	return &pb.ArticlesResponse{Articles: pas}, nil
}

// GetFeedArticles gets recent articles from users current user follow
func (h *Handler) GetFeedArticles(ctx context.Context, req *pb.GetFeedArticlesRequest) (*pb.ArticlesResponse, error) {
	h.logger.Info().Msgf("Get feed article | req: %+v\n", req)

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

	userIDs, err := h.us.GetFollowingUserIDs(currentUser)
	if err != nil {
		msg := fmt.Sprintf("failed to get following user ids of user %d", currentUser.ID)
		h.logger.Error().Err(err).Msg(msg)
		return nil, status.Error(codes.NotFound, "internal server error")
	}

	limitQuery := req.GetLimit()
	if limitQuery == 0 {
		limitQuery = 20
	}

	as, err := h.as.GetFeedArticles(userIDs, limitQuery, req.GetOffset())
	if err != nil {
		msg := "failed to get articles by user ids"
		h.logger.Error().Err(err).Msg(msg)
		return nil, status.Error(codes.NotFound, "internal server error")
	}

	pas := make([]*pb.Article, 0, len(as))
	for _, a := range as {
		// get whether the article is current user's favorite
		favorited := false
		pa := a.ProtoArticle(favorited)

		// get whether current user follows article author
		following, err := h.us.IsFollowing(currentUser, &a.Author)
		if err != nil {
			msg := "failed to get following status"
			h.logger.Error().Err(err).Msg(msg)
			return nil, status.Error(codes.NotFound, "internal server error")
		}
		pa.Author = a.Author.ProtoProfile(following)

		pas = append(pas, pa)
	}

	return &pb.ArticlesResponse{Articles: pas}, nil
}

// UpdateArticle updates an article
func (h *Handler) UpdateArticle(ctx context.Context, req *pb.UpdateArticleRequest) (*pb.ArticleResponse, error) {
	h.logger.Info().Msgf("Update artcile | req: %+v\n", req)

	userID, err := auth.GetUserID(ctx)
	if err != nil {
		msg := "unauthenticated"
		h.logger.Error().Err(err).Msg(msg)
		return nil, status.Errorf(codes.Unauthenticated, msg)
	}

	currentUser, err := h.us.GetByID(userID)
	if err != nil {
		msg := "not user found"
		err = fmt.Errorf("token is valid but the user not found: %w", err)
		h.logger.Error().Err(err).Msg(msg)
		return nil, status.Error(codes.NotFound, msg)
	}

	slug := req.GetArticle().GetSlug()
	articleID, err := strconv.Atoi(slug)
	if err != nil {
		msg := fmt.Sprintf("cannot convert slug (%s) into integer", slug)
		h.logger.Error().Err(err).Msg(msg)
		return nil, status.Error(codes.InvalidArgument, "invalid article id")
	}

	article, err := h.as.GetByID(uint(articleID))
	if err != nil {
		msg := fmt.Sprintf("requested article (slug=%d) not found", articleID)
		h.logger.Error().Err(err).Msg(msg)
		pp.Println(err)
		return nil, status.Error(codes.InvalidArgument, "invalid article id")
	}

	if article.Author.ID != currentUser.ID {
		msg := fmt.Sprintf("user(id=%d) attempted to update other user's article(id=%d)",
			currentUser.ID, article.ID)
		h.logger.Error().Err(err).Msg(msg)
		return nil, status.Errorf(codes.Unauthenticated, "forbidden")
	}

	article.Overwrite(
		req.GetArticle().GetTitle(),
		req.GetArticle().GetDescription(),
		req.GetArticle().GetBody(),
	)

	err = article.Validate()
	if err != nil {
		err = fmt.Errorf("validation error: %w", err)
		h.logger.Error().Err(err).Msg("validation error")
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if err := h.as.Update(article); err != nil {
		h.logger.Error().Err(err).Msg("failed to update article")
		return nil, status.Error(codes.InvalidArgument, "internal server error")
	}

	// get whether the article is current user's favorite
	favorited := false
	pa := article.ProtoArticle(favorited)

	// get whether current user follows article author
	following, err := h.us.IsFollowing(currentUser, &article.Author)
	if err != nil {
		msg := "failed to get following status"
		h.logger.Error().Err(err).Msg(msg)
		return nil, status.Error(codes.NotFound, "internal server error")
	}
	pa.Author = article.Author.ProtoProfile(following)

	return &pb.ArticleResponse{Article: pa}, nil
}

// DeleteArticle deletes an article
func (h *Handler) DeleteArticle(ctx context.Context, req *pb.DeleteArticleRequest) (*pb.Empty, error) {
	h.logger.Info().Msgf("Delete artcile | req: %+v\n", req)

	userID, err := auth.GetUserID(ctx)
	if err != nil {
		msg := "unauthenticated"
		h.logger.Error().Err(err).Msg(msg)
		return nil, status.Errorf(codes.Unauthenticated, msg)
	}

	currentUser, err := h.us.GetByID(userID)
	if err != nil {
		msg := "not user found"
		err = fmt.Errorf("token is valid but the user not found: %w", err)
		h.logger.Error().Err(err).Msg(msg)
		return nil, status.Error(codes.NotFound, msg)
	}

	slug := req.GetSlug()
	articleID, err := strconv.Atoi(slug)
	if err != nil {
		msg := fmt.Sprintf("cannot convert slug (%s) into integer", slug)
		h.logger.Error().Err(err).Msg(msg)
		return nil, status.Error(codes.InvalidArgument, "invalid article id")
	}

	article, err := h.as.GetByID(uint(articleID))
	if err != nil {
		msg := fmt.Sprintf("requested article (slug=%d) not found", articleID)
		h.logger.Error().Err(err).Msg(msg)
		pp.Println(err)
		return nil, status.Error(codes.InvalidArgument, "invalid article id")
	}

	if article.Author.ID != currentUser.ID {
		msg := fmt.Sprintf("user(id=%d) attempted to update other user's article(id=%d)",
			currentUser.ID, article.ID)
		h.logger.Error().Err(err).Msg(msg)
		return nil, status.Errorf(codes.Unauthenticated, "forbidden")
	}

	if err := h.as.Delete(article); err != nil {
		msg := "failed to delete article"
		h.logger.Error().Err(err).Msg(msg)
		return nil, status.Errorf(codes.Unauthenticated, msg)
	}

	return &pb.Empty{}, nil
}

// FavoriteArticle add an article to user favorites
func (h *Handler) FavoriteArticle(ctx context.Context, req *pb.FavoriteArticleRequest) (*pb.ArticleResponse, error) {
	h.logger.Info().Msgf("Delete artcile | req: %+v\n", req)

	userID, err := auth.GetUserID(ctx)
	if err != nil {
		msg := "unauthenticated"
		h.logger.Error().Err(err).Msg(msg)
		return nil, status.Errorf(codes.Unauthenticated, msg)
	}

	currentUser, err := h.us.GetByID(userID)
	if err != nil {
		msg := "not user found"
		err = fmt.Errorf("token is valid but the user not found: %w", err)
		h.logger.Error().Err(err).Msg(msg)
		return nil, status.Error(codes.NotFound, msg)
	}

	slug := req.GetSlug()
	articleID, err := strconv.Atoi(slug)
	if err != nil {
		msg := fmt.Sprintf("cannot convert slug (%s) into integer", slug)
		h.logger.Error().Err(err).Msg(msg)
		return nil, status.Error(codes.InvalidArgument, "invalid article id")
	}

	article, err := h.as.GetByID(uint(articleID))
	if err != nil {
		msg := fmt.Sprintf("requested article (slug=%d) not found", articleID)
		h.logger.Error().Err(err).Msg(msg)
		pp.Println(err)
		return nil, status.Error(codes.InvalidArgument, "invalid article id")
	}

	err = h.as.AddFavorite(article, currentUser)
	if err != nil {
		h.logger.Error().Err(err).Msg("failed to add favorite")
		pp.Println(err)
		return nil, status.Error(codes.Aborted, "internal server error")
	}

	// get whether current user follows article author
	pa := article.ProtoArticle(true)
	following, err := h.us.IsFollowing(currentUser, &article.Author)
	if err != nil {
		msg := "failed to get following status"
		h.logger.Error().Err(err).Msg(msg)
		return nil, status.Error(codes.NotFound, "internal server error")
	}
	pa.Author = article.Author.ProtoProfile(following)

	return &pb.ArticleResponse{Article: pa}, nil
}
