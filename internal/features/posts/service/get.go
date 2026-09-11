package posts_service

import (
	"context"
	"errors"
	"fmt"

	"github.com/qandoni/debatesApp/internal/core/domain"
	core_errors "github.com/qandoni/debatesApp/internal/core/errors"
)

func (s *PostsService) GetPost(
	ctx context.Context,
	postID int,
) (domain.Post, error) {
	post, err := s.postsRepository.GetPost(ctx, postID)
	if err != nil {
		return domain.Post{}, fmt.Errorf("get post from repository: %w", err)
	}

	images, err := s.imagesService.GetByPostID(ctx, postID)
	if err != nil {
		return domain.Post{}, fmt.Errorf("get post images: %w", err)
	}

	post.Images = images

	return post, nil
}

func (s *PostsService) GetPosts(
	ctx context.Context,
	limit *int,
	offset *int,
) ([]domain.Post, error) {
	if limit != nil && *limit < 0 {
		return nil, fmt.Errorf(
			"limit must be non-negative: %w",
			core_errors.ErrInvalidArgument,
		)
	}

	if offset != nil && *offset < 0 {
		return nil, fmt.Errorf(
			"offset must be non-negative: %w",
			core_errors.ErrInvalidArgument,
		)
	}

	posts, err := s.postsRepository.GetPosts(ctx, limit, offset)
	if err != nil {
		return []domain.Post{}, fmt.Errorf(
			"get posts from repository: %w",
			err,
		)
	}

	postIDs := make([]int, 0, len(posts))

	for _, post := range posts {
		postIDs = append(postIDs, post.ID)
	}

	imagesByPostID, err := s.imagesService.GetByPostIDs(ctx, postIDs)
	if err != nil {
		return []domain.Post{}, fmt.Errorf(
			"get images for posts: %w",
			err,
		)
	}

	for i := range posts {
		posts[i].Images = imagesByPostID[posts[i].ID]

		debate, err := s.debatesRepository.GetByPostID(
			ctx,
			posts[i].ID,
		)
		if err != nil {
			if errors.Is(err, core_errors.ErrNotFound) {
				continue
			}

			return []domain.Post{}, fmt.Errorf(
				"get debate for post with id '%d': %w",
				posts[i].ID,
				err,
			)
		}

		sides, err := s.debateSidesRepository.GetByDebateID(
			ctx,
			debate.ID,
		)
		if err != nil {
			return []domain.Post{}, fmt.Errorf(
				"get debate sides for post with id '%d': %w",
				posts[i].ID,
				err,
			)
		}

		debate.Sides = sides
		posts[i].Debate = &debate
	}

	return posts, nil
}
