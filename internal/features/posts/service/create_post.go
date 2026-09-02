package posts_service

import (
	"context"
	"fmt"

	"github.com/qandoni/debatesApp/internal/core/domain"
	posts_contracts "github.com/qandoni/debatesApp/internal/features/posts/contracts"
)

func (s *PostsService) CreatePost(
	ctx context.Context,
	input posts_contracts.CreatePostInput,
) (domain.Post, error) {
	if input.IsDebate && input.Debate == nil {
		return domain.Post{}, fmt.Errorf("debate data is required")
	}
	if !input.IsDebate && input.Debate != nil {
		return domain.Post{}, fmt.Errorf("debate data must be empty")
	}

	if input.IsDebate {
		if len(input.Debate.Sides) < 2 || len(input.Debate.Sides) > 5 {
			return domain.Post{}, fmt.Errorf(
				"debate must have between 2 and 5 sides",
			)
		}
	}

	var post domain.Post
	err := s.txManager.WithinTransaction(ctx, func(txCtx context.Context) error {
		newPost := domain.NewPostUninitialized(
			input.AuthorID,
			input.Content,
			input.IsDebate,
		)
		var err error

		post, err = s.postsRepository.CreatePost(txCtx, newPost)
		if err != nil {
			return fmt.Errorf("create post: %w", err)
		}
		if !input.IsDebate {
			return nil
		}

		debate := domain.NewDebateUnitialized(
			post.ID,
			input.Debate.EndAt,
		)
		debate, err = s.debatesRepository.CreateDebate(txCtx, debate)
		if err != nil {
			return fmt.Errorf("create debate: %w", err)
		}

		debate.Sides = make([]domain.DebateSide, 0, len(input.Debate.Sides))

		for i, sideInput := range input.Debate.Sides {
			side := domain.NewDebateSide(
				0,
				debate.ID,
				sideInput.Name,
				sideInput.Description,
				i+1,
			)

			createdSide, err := s.debateSidesRepository.CreateDebateSide(
				txCtx,
				side,
			)
			if err != nil {
				return fmt.Errorf("create debate side: %w", err)
			}

			debate.Sides = append(debate.Sides, createdSide)
		}

		post.Debate = &debate
		return nil
	})
	if err != nil {
		return domain.Post{}, fmt.Errorf("create post transaction: %w", err)
	}
	return post, nil
}
