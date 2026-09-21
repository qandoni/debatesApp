package comment_ratings_service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/qandoni/debatesApp/internal/core/domain"
	core_errors "github.com/qandoni/debatesApp/internal/core/errors"
)

type mockCommentRatingsRepository struct {
	createFn func(ctx context.Context, rating domain.CommentRating) (domain.CommentRating, error)
	getFn    func(ctx context.Context, commentID, userID int) (domain.CommentRating, error)
	updateFn func(ctx context.Context, rating domain.CommentRating) (domain.CommentRating, error)
}

func (m *mockCommentRatingsRepository) CreateCommentRating(ctx context.Context, rating domain.CommentRating) (domain.CommentRating, error) {
	return m.createFn(ctx, rating)
}

func (m *mockCommentRatingsRepository) GetByCommentAndUser(ctx context.Context, commentID, userID int) (domain.CommentRating, error) {
	return m.getFn(ctx, commentID, userID)
}

func (m *mockCommentRatingsRepository) UpdateCommentRating(ctx context.Context, rating domain.CommentRating) (domain.CommentRating, error) {
	return m.updateFn(ctx, rating)
}

type mockCommentsRepository struct {
	getByIDFn func(ctx context.Context, commentID int) (domain.Comment, error)
}

func (m *mockCommentsRepository) GetByID(ctx context.Context, commentID int) (domain.Comment, error) {
	return m.getByIDFn(ctx, commentID)
}

type mockDebatesRepository struct {
	getByPostFn func(ctx context.Context, postID int) (domain.Debate, error)
}

func (m *mockDebatesRepository) GetByPostID(ctx context.Context, postID int) (domain.Debate, error) {
	return m.getByPostFn(ctx, postID)
}

func newRatingsService(
	ratings CommentRatingsRepository,
	comments CommentsRepository,
	debates DebatesRepository,
) *CommentRatingsService {
	return NewCommentRatingsService(ratings, comments, debates)
}

func openDebate() domain.Debate {
	return domain.Debate{ID: 10, PostID: 100, Status: "OPEN"}
}

func finishedDebate() domain.Debate {
	return domain.Debate{ID: 10, PostID: 100, Status: "FINISHED"}
}

func rootArgument() domain.Comment {
	sideID := 2
	return domain.Comment{ID: 1, PostID: 100, DebateSideID: &sideID}
}

func TestRate_InvalidScore(t *testing.T) {
	svc := newRatingsService(&mockCommentRatingsRepository{}, &mockCommentsRepository{}, &mockDebatesRepository{})

	for _, score := range []int{0, 6, -1, 100} {
		_, err := svc.Rate(context.Background(), 5, 1, score)
		if err == nil {
			t.Fatalf("expected error for score %d, got nil", score)
		}
		if !errors.Is(err, core_errors.ErrInvalidArgument) {
			t.Fatalf("expected ErrInvalidArgument for score %d, got: %v", score, err)
		}
	}
}

func TestRate_NotRootComment(t *testing.T) {
	parentID := 3
	commentsRepo := &mockCommentsRepository{
		getByIDFn: func(ctx context.Context, commentID int) (domain.Comment, error) {
			return domain.Comment{ID: 1, PostID: 100, ParentCommentID: &parentID}, nil
		},
	}

	svc := newRatingsService(&mockCommentRatingsRepository{}, commentsRepo, &mockDebatesRepository{})
	_, err := svc.Rate(context.Background(), 5, 1, 4)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, core_errors.ErrInvalidArgument) {
		t.Fatalf("expected ErrInvalidArgument, got: %v", err)
	}
}

func TestRate_NotArgument(t *testing.T) {
	commentsRepo := &mockCommentsRepository{
		getByIDFn: func(ctx context.Context, commentID int) (domain.Comment, error) {
			return domain.Comment{ID: 1, PostID: 100}, nil
		},
	}

	svc := newRatingsService(&mockCommentRatingsRepository{}, commentsRepo, &mockDebatesRepository{})
	_, err := svc.Rate(context.Background(), 5, 1, 4)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, core_errors.ErrInvalidArgument) {
		t.Fatalf("expected ErrInvalidArgument, got: %v", err)
	}
}

func TestRate_DebateFinished(t *testing.T) {
	commentsRepo := &mockCommentsRepository{
		getByIDFn: func(ctx context.Context, commentID int) (domain.Comment, error) {
			return rootArgument(), nil
		},
	}
	debatesRepo := &mockDebatesRepository{
		getByPostFn: func(ctx context.Context, postID int) (domain.Debate, error) {
			return finishedDebate(), nil
		},
	}

	svc := newRatingsService(&mockCommentRatingsRepository{}, commentsRepo, debatesRepo)
	_, err := svc.Rate(context.Background(), 5, 1, 4)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, core_errors.ErrAccessForbidden) {
		t.Fatalf("expected ErrAccessForbidden, got: %v", err)
	}
}

func TestRate_FirstRatingCreates(t *testing.T) {
	commentsRepo := &mockCommentsRepository{
		getByIDFn: func(ctx context.Context, commentID int) (domain.Comment, error) {
			return rootArgument(), nil
		},
	}
	debatesRepo := &mockDebatesRepository{
		getByPostFn: func(ctx context.Context, postID int) (domain.Debate, error) {
			return openDebate(), nil
		},
	}
	ratingsRepo := &mockCommentRatingsRepository{
		getFn: func(ctx context.Context, commentID, userID int) (domain.CommentRating, error) {
			return domain.CommentRating{}, core_errors.ErrNotFound
		},
		createFn: func(ctx context.Context, rating domain.CommentRating) (domain.CommentRating, error) {
			if rating.Rating != 4 || rating.CommentID != 1 || rating.UserID != 5 {
				t.Fatalf("unexpected rating: %+v", rating)
			}
			rating.ID = 1
			return rating, nil
		},
	}

	svc := newRatingsService(ratingsRepo, commentsRepo, debatesRepo)
	created, err := svc.Rate(context.Background(), 5, 1, 4)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if created.ID != 1 {
		t.Fatalf("expected ID 1, got %d", created.ID)
	}
}

func TestRate_ChangeRatingUpdates(t *testing.T) {
	commentsRepo := &mockCommentsRepository{
		getByIDFn: func(ctx context.Context, commentID int) (domain.Comment, error) {
			return rootArgument(), nil
		},
	}
	debatesRepo := &mockDebatesRepository{
		getByPostFn: func(ctx context.Context, postID int) (domain.Debate, error) {
			return openDebate(), nil
		},
	}
	ratingsRepo := &mockCommentRatingsRepository{
		getFn: func(ctx context.Context, commentID, userID int) (domain.CommentRating, error) {
			return domain.CommentRating{ID: 1, CommentID: 1, UserID: 5, Rating: 2}, nil
		},
		updateFn: func(ctx context.Context, rating domain.CommentRating) (domain.CommentRating, error) {
			if rating.Rating != 5 {
				t.Fatalf("expected rating 5, got %d", rating.Rating)
			}
			if rating.UpdatedAt == nil {
				t.Fatal("expected UpdatedAt to be set")
			}
			return rating, nil
		},
	}

	svc := newRatingsService(ratingsRepo, commentsRepo, debatesRepo)
	updated, err := svc.Rate(context.Background(), 5, 1, 5)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if updated.Rating != 5 {
		t.Fatalf("expected rating 5, got %d", updated.Rating)
	}
}

func TestRate_SameRatingReturnsExisting(t *testing.T) {
	commentsRepo := &mockCommentsRepository{
		getByIDFn: func(ctx context.Context, commentID int) (domain.Comment, error) {
			return rootArgument(), nil
		},
	}
	debatesRepo := &mockDebatesRepository{
		getByPostFn: func(ctx context.Context, postID int) (domain.Debate, error) {
			return openDebate(), nil
		},
	}
	ratingsRepo := &mockCommentRatingsRepository{
		getFn: func(ctx context.Context, commentID, userID int) (domain.CommentRating, error) {
			return domain.CommentRating{ID: 1, CommentID: 1, UserID: 5, Rating: 3}, nil
		},
		updateFn: func(ctx context.Context, rating domain.CommentRating) (domain.CommentRating, error) {
			t.Fatal("Update should not be called for same rating")
			return domain.CommentRating{}, nil
		},
	}

	svc := newRatingsService(ratingsRepo, commentsRepo, debatesRepo)
	existing, err := svc.Rate(context.Background(), 5, 1, 3)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if existing.Rating != 3 {
		t.Fatalf("expected rating 3, got %d", existing.Rating)
	}
}

var _ = time.Now
