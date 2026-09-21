package debate_votes_service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/qandoni/debatesApp/internal/core/domain"
	core_enum "github.com/qandoni/debatesApp/internal/core/enum"
	core_errors "github.com/qandoni/debatesApp/internal/core/errors"
)

type mockDebateVotesRepository struct {
	createFn          func(ctx context.Context, vote domain.DebateVote) (domain.DebateVote, error)
	updateFn          func(ctx context.Context, debateID, userID, debateSideID int, updatedAt time.Time) (domain.DebateVote, error)
	getByDebateUserFn func(ctx context.Context, debateID, userID int) (domain.DebateVote, error)
}

func (m *mockDebateVotesRepository) Create(ctx context.Context, vote domain.DebateVote) (domain.DebateVote, error) {
	return m.createFn(ctx, vote)
}

func (m *mockDebateVotesRepository) Update(ctx context.Context, debateID, userID, debateSideID int, updatedAt time.Time) (domain.DebateVote, error) {
	return m.updateFn(ctx, debateID, userID, debateSideID, updatedAt)
}

func (m *mockDebateVotesRepository) GetByDebateAndUser(ctx context.Context, debateID, userID int) (domain.DebateVote, error) {
	return m.getByDebateUserFn(ctx, debateID, userID)
}

type mockDebatesRepository struct {
	getByIDFn      func(ctx context.Context, debateID int) (domain.Debate, error)
	finishFn       func(ctx context.Context, debateID int) error
	getAuthorFn    func(ctx context.Context, debateID int) (int, error)
	getByPostFn    func(ctx context.Context, postID int) (domain.Debate, error)
	createDebateFn func(ctx context.Context, debate domain.Debate) (domain.Debate, error)
}

func (m *mockDebatesRepository) GetByID(ctx context.Context, debateID int) (domain.Debate, error) {
	return m.getByIDFn(ctx, debateID)
}

func (m *mockDebatesRepository) FinishDebate(ctx context.Context, debateID int) error {
	return m.finishFn(ctx, debateID)
}

func (m *mockDebatesRepository) GetAuthorID(ctx context.Context, debateID int) (int, error) {
	return m.getAuthorFn(ctx, debateID)
}

func (m *mockDebatesRepository) GetByPostID(ctx context.Context, postID int) (domain.Debate, error) {
	return m.getByPostFn(ctx, postID)
}

func (m *mockDebatesRepository) CreateDebate(ctx context.Context, debate domain.Debate) (domain.Debate, error) {
	return m.createDebateFn(ctx, debate)
}

type mockDebateSidesRepository struct {
	getByDebateIDFn func(ctx context.Context, debateID int) ([]domain.DebateSide, error)
	createSideFn    func(ctx context.Context, side domain.DebateSide) (domain.DebateSide, error)
}

func (m *mockDebateSidesRepository) GetByDebateID(ctx context.Context, debateID int) ([]domain.DebateSide, error) {
	return m.getByDebateIDFn(ctx, debateID)
}

func (m *mockDebateSidesRepository) CreateDebateSide(ctx context.Context, side domain.DebateSide) (domain.DebateSide, error) {
	return m.createSideFn(ctx, side)
}

type mockCommentsRepository struct {
	hasArgumentFn func(ctx context.Context, userID, postID int) (bool, error)
}

func (m *mockCommentsRepository) HasUserArgumentInPost(ctx context.Context, userID, postID int) (bool, error) {
	return m.hasArgumentFn(ctx, userID, postID)
}

func openDebate(debateID, postID int) domain.Debate {
	return domain.NewDebate(
		debateID,
		postID,
		core_enum.DebateStatusOpen,
		nil,
		time.Now(),
		nil,
		nil,
	)
}

func finishedDebate(debateID, postID int) domain.Debate {
	return domain.NewDebate(
		debateID,
		postID,
		core_enum.DebateStatusFinished,
		nil,
		time.Now(),
		nil,
		nil,
	)
}

func newTestService(
	debateVotes DebateVotesRepository,
	debates DebatesRepository,
	sides DebateSidesRepository,
	comments CommentsRepository,
) *DebateVotesService {
	return NewDebateVotesService(debateVotes, debates, sides, comments)
}

func TestVote_Success(t *testing.T) {
	voteRepo := &mockDebateVotesRepository{
		createFn: func(ctx context.Context, vote domain.DebateVote) (domain.DebateVote, error) {
			if vote.DebateID != 10 || vote.UserID != 5 || vote.DebateSideID != 2 {
				t.Fatalf("unexpected vote: %+v", vote)
			}
			return vote, nil
		},
	}
	debatesRepo := &mockDebatesRepository{
		getByIDFn: func(ctx context.Context, debateID int) (domain.Debate, error) {
			return openDebate(10, 100), nil
		},
	}
	sidesRepo := &mockDebateSidesRepository{
		getByDebateIDFn: func(ctx context.Context, debateID int) ([]domain.DebateSide, error) {
			return []domain.DebateSide{{ID: 1}, {ID: 2}, {ID: 3}}, nil
		},
	}

	svc := newTestService(voteRepo, debatesRepo, sidesRepo, &mockCommentsRepository{})
	vote, err := svc.Vote(context.Background(), 5, 10, 2)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if vote.DebateSideID != 2 {
		t.Fatalf("expected side 2, got %d", vote.DebateSideID)
	}
}

func TestVote_DebateNotFound(t *testing.T) {
	voteRepo := &mockDebateVotesRepository{
		createFn: func(ctx context.Context, vote domain.DebateVote) (domain.DebateVote, error) {
			t.Fatal("Create should not be called")
			return domain.DebateVote{}, nil
		},
	}
	debatesRepo := &mockDebatesRepository{
		getByIDFn: func(ctx context.Context, debateID int) (domain.Debate, error) {
			return domain.Debate{}, core_errors.ErrNotFound
		},
	}

	svc := newTestService(voteRepo, debatesRepo, &mockDebateSidesRepository{}, &mockCommentsRepository{})
	_, err := svc.Vote(context.Background(), 5, 10, 2)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestVote_DebateClosed(t *testing.T) {
	voteRepo := &mockDebateVotesRepository{
		createFn: func(ctx context.Context, vote domain.DebateVote) (domain.DebateVote, error) {
			t.Fatal("Create should not be called")
			return domain.DebateVote{}, nil
		},
	}
	debatesRepo := &mockDebatesRepository{
		getByIDFn: func(ctx context.Context, debateID int) (domain.Debate, error) {
			return finishedDebate(10, 100), nil
		},
	}

	svc := newTestService(voteRepo, debatesRepo, &mockDebateSidesRepository{}, &mockCommentsRepository{})
	_, err := svc.Vote(context.Background(), 5, 10, 2)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestVote_SideNotBelongToDebate(t *testing.T) {
	voteRepo := &mockDebateVotesRepository{
		createFn: func(ctx context.Context, vote domain.DebateVote) (domain.DebateVote, error) {
			t.Fatal("Create should not be called")
			return domain.DebateVote{}, nil
		},
	}
	debatesRepo := &mockDebatesRepository{
		getByIDFn: func(ctx context.Context, debateID int) (domain.Debate, error) {
			return openDebate(10, 100), nil
		},
	}
	sidesRepo := &mockDebateSidesRepository{
		getByDebateIDFn: func(ctx context.Context, debateID int) ([]domain.DebateSide, error) {
			return []domain.DebateSide{{ID: 1}, {ID: 2}}, nil
		},
	}

	svc := newTestService(voteRepo, debatesRepo, sidesRepo, &mockCommentsRepository{})
	_, err := svc.Vote(context.Background(), 5, 10, 99)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, core_errors.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got: %v", err)
	}
}

func TestChangeVote_Success(t *testing.T) {
	voteRepo := &mockDebateVotesRepository{
		updateFn: func(ctx context.Context, debateID, userID, debateSideID int, updatedAt time.Time) (domain.DebateVote, error) {
			if debateID != 10 || userID != 5 || debateSideID != 3 {
				t.Fatalf("unexpected update args: %d %d %d", debateID, userID, debateSideID)
			}
			return domain.NewDebateVote(1, 2, 10, 5, 3, time.Now(), &updatedAt, true), nil
		},
	}
	debatesRepo := &mockDebatesRepository{
		getByIDFn: func(ctx context.Context, debateID int) (domain.Debate, error) {
			return openDebate(10, 100), nil
		},
	}
	sidesRepo := &mockDebateSidesRepository{
		getByDebateIDFn: func(ctx context.Context, debateID int) ([]domain.DebateSide, error) {
			return []domain.DebateSide{{ID: 1}, {ID: 3}}, nil
		},
	}
	commentsRepo := &mockCommentsRepository{
		hasArgumentFn: func(ctx context.Context, userID, postID int) (bool, error) {
			return false, nil
		},
	}

	svc := newTestService(voteRepo, debatesRepo, sidesRepo, commentsRepo)
	vote, err := svc.ChangeVote(context.Background(), 5, 10, 3)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if vote.DebateSideID != 3 {
		t.Fatalf("expected side 3, got %d", vote.DebateSideID)
	}
}

func TestChangeVote_ForbiddenAfterArgument(t *testing.T) {
	voteRepo := &mockDebateVotesRepository{
		updateFn: func(ctx context.Context, debateID, userID, debateSideID int, updatedAt time.Time) (domain.DebateVote, error) {
			t.Fatal("Update should not be called")
			return domain.DebateVote{}, nil
		},
	}
	debatesRepo := &mockDebatesRepository{
		getByIDFn: func(ctx context.Context, debateID int) (domain.Debate, error) {
			return openDebate(10, 100), nil
		},
	}
	sidesRepo := &mockDebateSidesRepository{
		getByDebateIDFn: func(ctx context.Context, debateID int) ([]domain.DebateSide, error) {
			return []domain.DebateSide{{ID: 1}, {ID: 3}}, nil
		},
	}
	commentsRepo := &mockCommentsRepository{
		hasArgumentFn: func(ctx context.Context, userID, postID int) (bool, error) {
			return true, nil
		},
	}

	svc := newTestService(voteRepo, debatesRepo, sidesRepo, commentsRepo)
	_, err := svc.ChangeVote(context.Background(), 5, 10, 3)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, core_errors.ErrAccessForbidden) {
		t.Fatalf("expected ErrAccessForbidden, got: %v", err)
	}
}

func TestChangeVote_DebateClosed(t *testing.T) {
	voteRepo := &mockDebateVotesRepository{
		updateFn: func(ctx context.Context, debateID, userID, debateSideID int, updatedAt time.Time) (domain.DebateVote, error) {
			t.Fatal("Update should not be called")
			return domain.DebateVote{}, nil
		},
	}
	debatesRepo := &mockDebatesRepository{
		getByIDFn: func(ctx context.Context, debateID int) (domain.Debate, error) {
			return finishedDebate(10, 100), nil
		},
	}

	svc := newTestService(voteRepo, debatesRepo, &mockDebateSidesRepository{}, &mockCommentsRepository{})
	_, err := svc.ChangeVote(context.Background(), 5, 10, 3)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestFinishDebate_Success(t *testing.T) {
	debatesRepo := &mockDebatesRepository{
		getByIDFn: func(ctx context.Context, debateID int) (domain.Debate, error) {
			return openDebate(10, 100), nil
		},
		getAuthorFn: func(ctx context.Context, debateID int) (int, error) {
			return 5, nil
		},
		finishFn: func(ctx context.Context, debateID int) error {
			if debateID != 10 {
				t.Fatalf("expected debate 10, got %d", debateID)
			}
			return nil
		},
	}

	svc := newTestService(&mockDebateVotesRepository{}, debatesRepo, &mockDebateSidesRepository{}, &mockCommentsRepository{})
	err := svc.FinishDebate(context.Background(), 5, 10)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
}

func TestFinishDebate_NotAuthor(t *testing.T) {
	debatesRepo := &mockDebatesRepository{
		getByIDFn: func(ctx context.Context, debateID int) (domain.Debate, error) {
			return openDebate(10, 100), nil
		},
		getAuthorFn: func(ctx context.Context, debateID int) (int, error) {
			return 5, nil
		},
		finishFn: func(ctx context.Context, debateID int) error {
			t.Fatal("FinishDebate should not be called")
			return nil
		},
	}

	svc := newTestService(&mockDebateVotesRepository{}, debatesRepo, &mockDebateSidesRepository{}, &mockCommentsRepository{})
	err := svc.FinishDebate(context.Background(), 99, 10)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, core_errors.ErrAccessForbidden) {
		t.Fatalf("expected ErrAccessForbidden, got: %v", err)
	}
}

func TestFinishDebate_AlreadyFinished(t *testing.T) {
	debatesRepo := &mockDebatesRepository{
		getByIDFn: func(ctx context.Context, debateID int) (domain.Debate, error) {
			return finishedDebate(10, 100), nil
		},
		getAuthorFn: func(ctx context.Context, debateID int) (int, error) {
			return 5, nil
		},
		finishFn: func(ctx context.Context, debateID int) error {
			t.Fatal("FinishDebate should not be called")
			return nil
		},
	}

	svc := newTestService(&mockDebateVotesRepository{}, debatesRepo, &mockDebateSidesRepository{}, &mockCommentsRepository{})
	err := svc.FinishDebate(context.Background(), 5, 10)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
