package comments_service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/qandoni/debatesApp/internal/core/domain"
	core_enum "github.com/qandoni/debatesApp/internal/core/enum"
	core_errors "github.com/qandoni/debatesApp/internal/core/errors"
	core_postgres "github.com/qandoni/debatesApp/internal/core/repository/postgres"
)

type mockCommentsRepository struct {
	setAuthorLikeFn        func(ctx context.Context, commentID int, liked bool) (domain.Comment, error)
	getWithoutPaginationFn func(ctx context.Context, postID int) ([]domain.Comment, error)
	getArgumentsRepliesFn  func(ctx context.Context, postID int, limit, offset *int) ([]domain.CommentWithRating, error)
	updateCommentFn        func(ctx context.Context, comment domain.Comment) (domain.Comment, error)
	hasArgumentFn          func(ctx context.Context, userID, postID int) (bool, error)
	getByPostIDFn          func(ctx context.Context, postID int, limit, offset *int) ([]domain.Comment, error)
	getByIDFn              func(ctx context.Context, commentID int) (domain.Comment, error)
	createCommentFn        func(ctx context.Context, comment domain.Comment) (domain.Comment, error)
}

func (m *mockCommentsRepository) SetAuthorLike(ctx context.Context, commentID int, liked bool) (domain.Comment, error) {
	return m.setAuthorLikeFn(ctx, commentID, liked)
}

func (m *mockCommentsRepository) GetByPostIDWithoutPagination(ctx context.Context, postID int) ([]domain.Comment, error) {
	return m.getWithoutPaginationFn(ctx, postID)
}

func (m *mockCommentsRepository) GetArgumentsWithReplies(ctx context.Context, postID int, limit, offset *int) ([]domain.CommentWithRating, error) {
	return m.getArgumentsRepliesFn(ctx, postID, limit, offset)
}

func (m *mockCommentsRepository) UpdateComment(ctx context.Context, comment domain.Comment) (domain.Comment, error) {
	return m.updateCommentFn(ctx, comment)
}

func (m *mockCommentsRepository) HasUserArgumentInPost(ctx context.Context, userID, postID int) (bool, error) {
	return m.hasArgumentFn(ctx, userID, postID)
}

func (m *mockCommentsRepository) GetByPostID(ctx context.Context, postID int, limit, offset *int) ([]domain.Comment, error) {
	return m.getByPostIDFn(ctx, postID, limit, offset)
}

func (m *mockCommentsRepository) GetByID(ctx context.Context, commentID int) (domain.Comment, error) {
	return m.getByIDFn(ctx, commentID)
}

func (m *mockCommentsRepository) CreateComment(ctx context.Context, comment domain.Comment) (domain.Comment, error) {
	return m.createCommentFn(ctx, comment)
}

type mockPostsRepository struct {
	createPostFn func(ctx context.Context, post domain.Post) (domain.Post, error)
	deletePostFn func(ctx context.Context, userID, postID int) error
	getPostFn    func(ctx context.Context, postID int) (domain.Post, error)
	patchPostFn  func(ctx context.Context, userID, postID int, patch domain.PostPatch) (domain.Post, error)
	getPostsFn   func(ctx context.Context, limit, offset *int) ([]domain.Post, error)
}

func (m *mockPostsRepository) CreatePost(ctx context.Context, post domain.Post) (domain.Post, error) {
	return m.createPostFn(ctx, post)
}

func (m *mockPostsRepository) DeletePost(ctx context.Context, userID, postID int) error {
	return m.deletePostFn(ctx, userID, postID)
}

func (m *mockPostsRepository) GetPost(ctx context.Context, postID int) (domain.Post, error) {
	return m.getPostFn(ctx, postID)
}

func (m *mockPostsRepository) PatchPost(ctx context.Context, userID, postID int, patch domain.PostPatch) (domain.Post, error) {
	return m.patchPostFn(ctx, userID, postID, patch)
}

func (m *mockPostsRepository) GetPosts(ctx context.Context, limit, offset *int) ([]domain.Post, error) {
	return m.getPostsFn(ctx, limit, offset)
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

type mockTxManager struct {
	withinFn func(ctx context.Context, fn func(ctx context.Context) error) error
}

func (m *mockTxManager) WithinTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	return m.withinFn(ctx, fn)
}

func newCommentsService(
	comments CommentsRepository,
	posts PostsRepository,
	debates DebatesRepository,
	sides DebateSidesRepository,
	votes DebateVotesRepository,
	tx core_postgres.TransactionManager,
) *CommentsService {
	return NewCommentsService(comments, posts, debates, sides, votes, tx)
}

func TestCreateComment_RegularPostNoParent(t *testing.T) {
	postsRepo := &mockPostsRepository{
		getPostFn: func(ctx context.Context, postID int) (domain.Post, error) {
			return domain.Post{ID: 100, IsDebate: false}, nil
		},
	}
	debatesRepo := &mockDebatesRepository{
		getByPostFn: func(ctx context.Context, postID int) (domain.Debate, error) {
			return domain.Debate{}, core_errors.ErrNotFound
		},
	}
	commentsRepo := &mockCommentsRepository{
		createCommentFn: func(ctx context.Context, comment domain.Comment) (domain.Comment, error) {
			if comment.PostID != 100 || comment.ParentCommentID != nil || comment.DebateSideID != nil {
				t.Fatalf("unexpected comment: %+v", comment)
			}
			comment.ID = 1
			return comment, nil
		},
	}

	svc := newCommentsService(commentsRepo, postsRepo, debatesRepo, &mockDebateSidesRepository{}, &mockDebateVotesRepository{}, &mockTxManager{})
	comment, err := svc.CreateComment(context.Background(), 5, 100, nil, "hello")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if comment.ID != 1 {
		t.Fatalf("expected ID 1, got %d", comment.ID)
	}
}

func TestCreateComment_DebatePostWithoutParent(t *testing.T) {
	postsRepo := &mockPostsRepository{
		getPostFn: func(ctx context.Context, postID int) (domain.Post, error) {
			return domain.Post{ID: 100, IsDebate: true}, nil
		},
	}
	debatesRepo := &mockDebatesRepository{
		getByPostFn: func(ctx context.Context, postID int) (domain.Debate, error) {
			return domain.Debate{ID: 10, PostID: 100}, nil
		},
	}

	svc := newCommentsService(&mockCommentsRepository{}, postsRepo, debatesRepo, &mockDebateSidesRepository{}, &mockDebateVotesRepository{}, &mockTxManager{})
	_, err := svc.CreateComment(context.Background(), 5, 100, nil, "hello")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, core_errors.ErrConflict) {
		t.Fatalf("expected ErrConflict, got: %v", err)
	}
}

func TestCreateComment_ParentFromDifferentPost(t *testing.T) {
	postsRepo := &mockPostsRepository{
		getPostFn: func(ctx context.Context, postID int) (domain.Post, error) {
			return domain.Post{ID: 100}, nil
		},
	}
	parentID := 7
	commentsRepo := &mockCommentsRepository{
		getByIDFn: func(ctx context.Context, commentID int) (domain.Comment, error) {
			return domain.Comment{ID: 7, PostID: 999}, nil
		},
	}

	svc := newCommentsService(commentsRepo, postsRepo, &mockDebatesRepository{}, &mockDebateSidesRepository{}, &mockDebateVotesRepository{}, &mockTxManager{})
	_, err := svc.CreateComment(context.Background(), 5, 100, &parentID, "hello")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, core_errors.ErrConflict) {
		t.Fatalf("expected ErrConflict, got: %v", err)
	}
}

func TestCreateComment_ParentInheritsDebateSide(t *testing.T) {
	postsRepo := &mockPostsRepository{
		getPostFn: func(ctx context.Context, postID int) (domain.Post, error) {
			return domain.Post{ID: 100, IsDebate: true}, nil
		},
	}
	parentID := 7
	sideID := 3
	commentsRepo := &mockCommentsRepository{
		getByIDFn: func(ctx context.Context, commentID int) (domain.Comment, error) {
			return domain.Comment{ID: 7, PostID: 100, DebateSideID: &sideID}, nil
		},
		createCommentFn: func(ctx context.Context, comment domain.Comment) (domain.Comment, error) {
			if comment.DebateSideID == nil || *comment.DebateSideID != 3 {
				t.Fatalf("expected inherited side 3, got %+v", comment.DebateSideID)
			}
			return comment, nil
		},
	}

	svc := newCommentsService(commentsRepo, postsRepo, &mockDebatesRepository{}, &mockDebateSidesRepository{}, &mockDebateVotesRepository{}, &mockTxManager{})
	_, err := svc.CreateComment(context.Background(), 5, 100, &parentID, "hello")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
}

func TestCreateArgument_Success(t *testing.T) {
	postsRepo := &mockPostsRepository{
		getPostFn: func(ctx context.Context, postID int) (domain.Post, error) {
			return domain.Post{ID: 100}, nil
		},
	}
	debatesRepo := &mockDebatesRepository{
		getByPostFn: func(ctx context.Context, postID int) (domain.Debate, error) {
			return domain.Debate{ID: 10, PostID: 100, Status: core_enum.DebateStatusOpen}, nil
		},
	}
	sidesRepo := &mockDebateSidesRepository{
		getByDebateIDFn: func(ctx context.Context, debateID int) ([]domain.DebateSide, error) {
			return []domain.DebateSide{{ID: 2}}, nil
		},
	}
	votesRepo := &mockDebateVotesRepository{
		getByDebateUserFn: func(ctx context.Context, debateID, userID int) (domain.DebateVote, error) {
			return domain.DebateVote{DebateID: 10, UserID: 5, DebateSideID: 2}, nil
		},
	}
	commentsRepo := &mockCommentsRepository{
		createCommentFn: func(ctx context.Context, comment domain.Comment) (domain.Comment, error) {
			if comment.ParentCommentID != nil || comment.DebateSideID == nil || *comment.DebateSideID != 2 {
				t.Fatalf("unexpected argument: %+v", comment)
			}
			comment.ID = 42
			return comment, nil
		},
	}

	svc := newCommentsService(commentsRepo, postsRepo, debatesRepo, sidesRepo, votesRepo, &mockTxManager{})
	arg, err := svc.CreateArgument(context.Background(), 5, 100, 2, "my argument")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if arg.ID != 42 {
		t.Fatalf("expected ID 42, got %d", arg.ID)
	}
}

func TestCreateArgument_DebateClosed(t *testing.T) {
	postsRepo := &mockPostsRepository{
		getPostFn: func(ctx context.Context, postID int) (domain.Post, error) {
			return domain.Post{ID: 100}, nil
		},
	}
	debatesRepo := &mockDebatesRepository{
		getByPostFn: func(ctx context.Context, postID int) (domain.Debate, error) {
			return domain.Debate{ID: 10, PostID: 100, Status: core_enum.DebateStatusFinished}, nil
		},
	}

	svc := newCommentsService(&mockCommentsRepository{}, postsRepo, debatesRepo, &mockDebateSidesRepository{}, &mockDebateVotesRepository{}, &mockTxManager{})
	_, err := svc.CreateArgument(context.Background(), 5, 100, 2, "my argument")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, core_errors.ErrConflict) {
		t.Fatalf("expected ErrConflict, got: %v", err)
	}
}

func TestCreateArgument_VotedForOtherSide(t *testing.T) {
	postsRepo := &mockPostsRepository{
		getPostFn: func(ctx context.Context, postID int) (domain.Post, error) {
			return domain.Post{ID: 100}, nil
		},
	}
	debatesRepo := &mockDebatesRepository{
		getByPostFn: func(ctx context.Context, postID int) (domain.Debate, error) {
			return domain.Debate{ID: 10, PostID: 100, Status: core_enum.DebateStatusOpen}, nil
		},
	}
	sidesRepo := &mockDebateSidesRepository{
		getByDebateIDFn: func(ctx context.Context, debateID int) ([]domain.DebateSide, error) {
			return []domain.DebateSide{{ID: 1}, {ID: 2}}, nil
		},
	}
	votesRepo := &mockDebateVotesRepository{
		getByDebateUserFn: func(ctx context.Context, debateID, userID int) (domain.DebateVote, error) {
			return domain.DebateVote{DebateID: 10, UserID: 5, DebateSideID: 1}, nil
		},
	}

	svc := newCommentsService(&mockCommentsRepository{}, postsRepo, debatesRepo, sidesRepo, votesRepo, &mockTxManager{})
	_, err := svc.CreateArgument(context.Background(), 5, 100, 2, "my argument")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, core_errors.ErrConflict) {
		t.Fatalf("expected ErrConflict, got: %v", err)
	}
}

func TestCreateArgument_SideNotFound(t *testing.T) {
	postsRepo := &mockPostsRepository{
		getPostFn: func(ctx context.Context, postID int) (domain.Post, error) {
			return domain.Post{ID: 100}, nil
		},
	}
	debatesRepo := &mockDebatesRepository{
		getByPostFn: func(ctx context.Context, postID int) (domain.Debate, error) {
			return domain.Debate{ID: 10, PostID: 100, Status: core_enum.DebateStatusOpen}, nil
		},
	}
	sidesRepo := &mockDebateSidesRepository{
		getByDebateIDFn: func(ctx context.Context, debateID int) ([]domain.DebateSide, error) {
			return []domain.DebateSide{{ID: 1}}, nil
		},
	}

	svc := newCommentsService(&mockCommentsRepository{}, postsRepo, debatesRepo, sidesRepo, &mockDebateVotesRepository{}, &mockTxManager{})
	_, err := svc.CreateArgument(context.Background(), 5, 100, 99, "my argument")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, core_errors.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got: %v", err)
	}
}

func TestUpdateComment_NotAuthor(t *testing.T) {
	commentsRepo := &mockCommentsRepository{
		getByIDFn: func(ctx context.Context, commentID int) (domain.Comment, error) {
			return domain.Comment{ID: 1, AuthorID: 5, PostID: 100}, nil
		},
	}

	svc := newCommentsService(commentsRepo, &mockPostsRepository{}, &mockDebatesRepository{}, &mockDebateSidesRepository{}, &mockDebateVotesRepository{}, &mockTxManager{})
	_, err := svc.UpdateComment(context.Background(), 99, 1, "new content")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, core_errors.ErrAccessForbidden) {
		t.Fatalf("expected ErrAccessForbidden, got: %v", err)
	}
}

func TestUpdateComment_ArgumentInClosedDebate(t *testing.T) {
	sideID := 2
	commentsRepo := &mockCommentsRepository{
		getByIDFn: func(ctx context.Context, commentID int) (domain.Comment, error) {
			return domain.Comment{ID: 1, AuthorID: 5, PostID: 100, DebateSideID: &sideID}, nil
		},
	}
	debatesRepo := &mockDebatesRepository{
		getByPostFn: func(ctx context.Context, postID int) (domain.Debate, error) {
			return domain.Debate{ID: 10, PostID: 100, Status: core_enum.DebateStatusFinished}, nil
		},
	}

	svc := newCommentsService(commentsRepo, &mockPostsRepository{}, debatesRepo, &mockDebateSidesRepository{}, &mockDebateVotesRepository{}, &mockTxManager{})
	_, err := svc.UpdateComment(context.Background(), 5, 1, "new content")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, core_errors.ErrConflict) {
		t.Fatalf("expected ErrConflict, got: %v", err)
	}
}

func TestUpdateComment_Success(t *testing.T) {
	commentsRepo := &mockCommentsRepository{
		getByIDFn: func(ctx context.Context, commentID int) (domain.Comment, error) {
			return domain.Comment{ID: 1, AuthorID: 5, PostID: 100}, nil
		},
		updateCommentFn: func(ctx context.Context, comment domain.Comment) (domain.Comment, error) {
			if comment.Content != "new content" {
				t.Fatalf("expected updated content, got: %s", comment.Content)
			}
			return comment, nil
		},
	}

	svc := newCommentsService(commentsRepo, &mockPostsRepository{}, &mockDebatesRepository{}, &mockDebateSidesRepository{}, &mockDebateVotesRepository{}, &mockTxManager{})
	updated, err := svc.UpdateComment(context.Background(), 5, 1, "new content")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if updated.Content != "new content" {
		t.Fatalf("expected 'new content', got: %s", updated.Content)
	}
}

func TestSetAuthorLike_NotArgument(t *testing.T) {
	commentsRepo := &mockCommentsRepository{
		getByIDFn: func(ctx context.Context, commentID int) (domain.Comment, error) {
			return domain.Comment{ID: 1, PostID: 100}, nil
		},
	}

	svc := newCommentsService(commentsRepo, &mockPostsRepository{}, &mockDebatesRepository{}, &mockDebateSidesRepository{}, &mockDebateVotesRepository{}, &mockTxManager{})
	_, err := svc.SetAuthorLike(context.Background(), 5, 1, true)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, core_errors.ErrInvalidArgument) {
		t.Fatalf("expected ErrInvalidArgument, got: %v", err)
	}
}

func TestSetAuthorLike_ChildComment(t *testing.T) {
	sideID := 2
	parentID := 3
	commentsRepo := &mockCommentsRepository{
		getByIDFn: func(ctx context.Context, commentID int) (domain.Comment, error) {
			return domain.Comment{ID: 1, PostID: 100, DebateSideID: &sideID, ParentCommentID: &parentID}, nil
		},
	}

	svc := newCommentsService(commentsRepo, &mockPostsRepository{}, &mockDebatesRepository{}, &mockDebateSidesRepository{}, &mockDebateVotesRepository{}, &mockTxManager{})
	_, err := svc.SetAuthorLike(context.Background(), 5, 1, true)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, core_errors.ErrInvalidArgument) {
		t.Fatalf("expected ErrInvalidArgument, got: %v", err)
	}
}

func TestSetAuthorLike_NotDebateAuthor(t *testing.T) {
	sideID := 2
	commentsRepo := &mockCommentsRepository{
		getByIDFn: func(ctx context.Context, commentID int) (domain.Comment, error) {
			return domain.Comment{ID: 1, PostID: 100, DebateSideID: &sideID}, nil
		},
	}
	debatesRepo := &mockDebatesRepository{
		getByPostFn: func(ctx context.Context, postID int) (domain.Debate, error) {
			return domain.Debate{ID: 10, PostID: 100, Status: core_enum.DebateStatusOpen}, nil
		},
		getAuthorFn: func(ctx context.Context, debateID int) (int, error) {
			return 5, nil
		},
	}

	svc := newCommentsService(commentsRepo, &mockPostsRepository{}, debatesRepo, &mockDebateSidesRepository{}, &mockDebateVotesRepository{}, &mockTxManager{})
	_, err := svc.SetAuthorLike(context.Background(), 99, 1, true)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, core_errors.ErrAccessForbidden) {
		t.Fatalf("expected ErrAccessForbidden, got: %v", err)
	}
}

func TestSetAuthorLike_AfterDebateFinished(t *testing.T) {
	sideID := 2
	commentsRepo := &mockCommentsRepository{
		getByIDFn: func(ctx context.Context, commentID int) (domain.Comment, error) {
			return domain.Comment{ID: 1, PostID: 100, DebateSideID: &sideID}, nil
		},
	}
	debatesRepo := &mockDebatesRepository{
		getByPostFn: func(ctx context.Context, postID int) (domain.Debate, error) {
			return domain.Debate{ID: 10, PostID: 100, Status: core_enum.DebateStatusFinished}, nil
		},
	}

	svc := newCommentsService(commentsRepo, &mockPostsRepository{}, debatesRepo, &mockDebateSidesRepository{}, &mockDebateVotesRepository{}, &mockTxManager{})
	_, err := svc.SetAuthorLike(context.Background(), 5, 1, true)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, core_errors.ErrConflict) {
		t.Fatalf("expected ErrConflict, got: %v", err)
	}
}

func TestSetAuthorLike_Success(t *testing.T) {
	sideID := 2
	commentsRepo := &mockCommentsRepository{
		getByIDFn: func(ctx context.Context, commentID int) (domain.Comment, error) {
			return domain.Comment{ID: 1, PostID: 100, DebateSideID: &sideID}, nil
		},
		setAuthorLikeFn: func(ctx context.Context, commentID int, liked bool) (domain.Comment, error) {
			if !liked {
				t.Fatalf("expected liked=true, got false")
			}
			return domain.Comment{ID: 1, PostID: 100, DebateSideID: &sideID, AuthorLiked: true}, nil
		},
	}
	debatesRepo := &mockDebatesRepository{
		getByPostFn: func(ctx context.Context, postID int) (domain.Debate, error) {
			return domain.Debate{ID: 10, PostID: 100, Status: core_enum.DebateStatusOpen}, nil
		},
		getAuthorFn: func(ctx context.Context, debateID int) (int, error) {
			return 5, nil
		},
	}

	svc := newCommentsService(commentsRepo, &mockPostsRepository{}, debatesRepo, &mockDebateSidesRepository{}, &mockDebateVotesRepository{}, &mockTxManager{})
	updated, err := svc.SetAuthorLike(context.Background(), 5, 1, true)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if !updated.AuthorLiked {
		t.Fatal("expected AuthorLiked=true")
	}
}
