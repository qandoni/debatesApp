package images_service

import (
	"context"
	"fmt"
	"time"

	"github.com/qandoni/debatesApp/internal/core/domain"
)

func (s *ImagesService) Recover(ctx context.Context) error {
	return s.recover(ctx, s.postImagesRepository.GetPending)
}

func (s *ImagesService) RecoverStale(ctx context.Context, staleAfter time.Duration) error {
	return s.recover(ctx, func(ctx context.Context) ([]domain.PostImage, error) {
		return s.postImagesRepository.GetPendingOlderThan(ctx, staleAfter)
	})
}

func (s *ImagesService) recover(
	ctx context.Context,
	getPending func(ctx context.Context) ([]domain.PostImage, error),
) error {
	pending, err := getPending(ctx)
	if err != nil {
		return fmt.Errorf("get pending images: %w", err)
	}
	if len(pending) > 0 {
		toActivate := make([]int, 0, len(pending))
		toDelete := make([]int, 0, len(pending))
		for _, image := range pending {
			exists, err := s.minio.ObjectExists(ctx, image.ImageURL)
			if err != nil {
				return fmt.Errorf("check object %q: %w", image.ImageURL, err)
			}
			if exists {
				toActivate = append(toActivate, image.ID)
			} else {
				toDelete = append(toDelete, image.ID)
			}
		}
		if err := s.txManager.WithinTransaction(ctx, func(txCtx context.Context) error {
			if len(toActivate) > 0 {
				if err := s.postImagesRepository.MarkActiveByIDs(txCtx, toActivate); err != nil {
					return err
				}
			}
			if len(toDelete) > 0 {
				if err := s.postImagesRepository.DeletePendingByIDs(txCtx, toDelete); err != nil {
					return err
				}
			}
			return nil
		}); err != nil {
			return fmt.Errorf("reconcile pending images: %w", err)
		}
	}

	deleted, err := s.postImagesRepository.GetImagesOfDeletedPosts(ctx)
	if err != nil {
		return fmt.Errorf("get images of deleted posts: %w", err)
	}
	if len(deleted) == 0 {
		return nil
	}

	if orphans := s.minio.DeleteAll(ctx, imageNames(deleted)); len(orphans) > 0 {
		return nil
	}
	for _, postID := range uniquePostIDs(deleted) {
		if err := s.txManager.WithinTransaction(ctx, func(txCtx context.Context) error {
			return s.postImagesRepository.DeleteByPostID(txCtx, postID)
		}); err != nil {
			return fmt.Errorf("delete image rows of post %d: %w", postID, err)
		}
	}
	return nil
}

func uniquePostIDs(images []domain.PostImage) []int {
	seen := make(map[int]struct{}, len(images))
	postIDs := make([]int, 0, len(images))
	for _, image := range images {
		if _, ok := seen[image.PostID]; ok {
			continue
		}
		seen[image.PostID] = struct{}{}
		postIDs = append(postIDs, image.PostID)
	}
	return postIDs
}
