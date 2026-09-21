package posts_service

import (
	"context"
	"errors"
	"mime/multipart"
	"testing"
	"time"

	"github.com/qandoni/debatesApp/internal/core/domain"
	core_errors "github.com/qandoni/debatesApp/internal/core/errors"
	core_postgres "github.com/qandoni/debatesApp/internal/core/repository/postgres"
	posts_contracts "github.com/qandoni/debatesApp/internal/features/posts/contracts"
)

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

type mockImagesService struct {
	getByPostIDFn       func(ctx context.Context, postID int) ([]domain.PostImage, error)
	createPostImagesFn  func(ctx context.Context, userID, postID int, files []*multipart.FileHeader) ([]domain.PostImage, error)
	deleteRecordsFn     func(ctx context.Context, postID int) error
	getByPostIDsFn      func(ctx context.Context, postIDs []int) (map[int][]domain.PostImage, error)
	deleteFromStorageFn func(ctx context.Context, images []domain.PostImage) error
}

func (m *mockImagesService) GetByPostID(ctx context.Context, postID int) ([]domain.PostImage, error) {
	return m.getByPostIDFn(ctx, postID)
}

func (m *mockImagesService) CreatePostImages(ctx context.Context, userID, postID int, files []*multipart.FileHeader) ([]domain.PostImage, error) {
	return m.createPostImagesFn(ctx, userID, postID, files)
}

func (m *mockImagesService) DeleteRecordsByPostID(ctx context.Context, postID int) error {
	return m.deleteRecordsFn(ctx, postID)
}

func (m *mockImagesService) GetByPostIDs(ctx context.Context, postIDs []int) (map[int][]domain.PostImage, error) {
	return m.getByPostIDsFn(ctx, postIDs)
}

func (m *mockImagesService) DeleteFromStorage(ctx context.Context, images []domain.PostImage) error {
	return m.deleteFromStorageFn(ctx, images)
}

type mockTxManager struct {
	withinFn func(ctx context.Context, fn func(ctx context.Context) error) error
}

func (m *mockTxManager) WithinTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	return m.withinFn(ctx, fn)
}

func newPostsService(
	posts PostsRepository,
	images ImagesService,
	debates DebatesRepository,
	sides DebateSidesRepository,
	tx core_postgres.TransactionManager,
) *PostsService {
	return NewPostsService(posts, images, debates, sides, tx)
}

func debateInput(sidesCount int) *posts_contracts.CreateDebateInput {
	sides := make([]posts_contracts.CreateDebateSideInput, 0, sidesCount)
	for i := 0; i < sidesCount; i++ {
		sides = append(sides, posts_contracts.CreateDebateSideInput{Name: "side"})
	}
	return &posts_contracts.CreateDebateInput{Sides: sides}
}

func TestCreatePost_IsDebateWithoutData(t *testing.T) {
	svc := newPostsService(&mockPostsRepository{}, &mockImagesService{}, &mockDebatesRepository{}, &mockDebateSidesRepository{}, &mockTxManager{})
	_, err := svc.CreatePost(context.Background(), posts_contracts.CreatePostInput{
		AuthorID: 5,
		Content:  "hello",
		IsDebate: true,
		Debate:   nil,
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestCreatePost_NotDebateWithData(t *testing.T) {
	svc := newPostsService(&mockPostsRepository{}, &mockImagesService{}, &mockDebatesRepository{}, &mockDebateSidesRepository{}, &mockTxManager{})
	_, err := svc.CreatePost(context.Background(), posts_contracts.CreatePostInput{
		AuthorID: 5,
		Content:  "hello",
		IsDebate: false,
		Debate:   debateInput(2),
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestCreatePost_DebateTooFewSides(t *testing.T) {
	svc := newPostsService(&mockPostsRepository{}, &mockImagesService{}, &mockDebatesRepository{}, &mockDebateSidesRepository{}, &mockTxManager{})
	_, err := svc.CreatePost(context.Background(), posts_contracts.CreatePostInput{
		AuthorID: 5,
		Content:  "hello",
		IsDebate: true,
		Debate:   debateInput(1),
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestCreatePost_DebateTooManySides(t *testing.T) {
	svc := newPostsService(&mockPostsRepository{}, &mockImagesService{}, &mockDebatesRepository{}, &mockDebateSidesRepository{}, &mockTxManager{})
	_, err := svc.CreatePost(context.Background(), posts_contracts.CreatePostInput{
		AuthorID: 5,
		Content:  "hello",
		IsDebate: true,
		Debate:   debateInput(6),
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestCreatePost_RegularPostSuccess(t *testing.T) {
	postsRepo := &mockPostsRepository{
		createPostFn: func(ctx context.Context, post domain.Post) (domain.Post, error) {
			if post.AuthorID != 5 || post.Content != "hello" || post.IsDebate {
				t.Fatalf("unexpected post: %+v", post)
			}
			post.ID = 1
			return post, nil
		},
	}
	tx := &mockTxManager{
		withinFn: func(ctx context.Context, fn func(ctx context.Context) error) error {
			return fn(ctx)
		},
	}

	svc := newPostsService(postsRepo, &mockImagesService{}, &mockDebatesRepository{}, &mockDebateSidesRepository{}, tx)
	post, err := svc.CreatePost(context.Background(), posts_contracts.CreatePostInput{
		AuthorID: 5,
		Content:  "hello",
		IsDebate: false,
	})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if post.ID != 1 {
		t.Fatalf("expected ID 1, got %d", post.ID)
	}
}

func TestCreatePost_DebateSuccess(t *testing.T) {
	postsRepo := &mockPostsRepository{
		createPostFn: func(ctx context.Context, post domain.Post) (domain.Post, error) {
			post.ID = 1
			return post, nil
		},
	}
	debatesRepo := &mockDebatesRepository{
		createDebateFn: func(ctx context.Context, debate domain.Debate) (domain.Debate, error) {
			if debate.PostID != 1 {
				t.Fatalf("expected post 1, got %d", debate.PostID)
			}
			debate.ID = 10
			return debate, nil
		},
	}
	sidesRepo := &mockDebateSidesRepository{
		createSideFn: func(ctx context.Context, side domain.DebateSide) (domain.DebateSide, error) {
			side.ID = side.DisplayOrder
			return side, nil
		},
	}
	tx := &mockTxManager{
		withinFn: func(ctx context.Context, fn func(ctx context.Context) error) error {
			return fn(ctx)
		},
	}

	svc := newPostsService(postsRepo, &mockImagesService{}, debatesRepo, sidesRepo, tx)
	post, err := svc.CreatePost(context.Background(), posts_contracts.CreatePostInput{
		AuthorID: 5,
		Content:  "debate post",
		IsDebate: true,
		Debate:   debateInput(3),
	})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if post.Debate == nil {
		t.Fatal("expected debate to be attached")
	}
	if len(post.Debate.Sides) != 3 {
		t.Fatalf("expected 3 sides, got %d", len(post.Debate.Sides))
	}
}

func TestDeletePost_NotAuthor(t *testing.T) {
	postsRepo := &mockPostsRepository{
		getPostFn: func(ctx context.Context, postID int) (domain.Post, error) {
			return domain.Post{ID: 1, AuthorID: 5}, nil
		},
	}

	svc := newPostsService(postsRepo, &mockImagesService{}, &mockDebatesRepository{}, &mockDebateSidesRepository{}, &mockTxManager{})
	err := svc.DeletePost(context.Background(), 99, 1)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, core_errors.ErrAccessForbidden) {
		t.Fatalf("expected ErrAccessForbidden, got: %v", err)
	}
}

func TestDeletePost_Success(t *testing.T) {
	postsRepo := &mockPostsRepository{
		getPostFn: func(ctx context.Context, postID int) (domain.Post, error) {
			return domain.Post{ID: 1, AuthorID: 5}, nil
		},
		deletePostFn: func(ctx context.Context, userID, postID int) error {
			if userID != 5 || postID != 1 {
				t.Fatalf("unexpected delete args: %d %d", userID, postID)
			}
			return nil
		},
	}
	imagesSvc := &mockImagesService{
		getByPostIDFn: func(ctx context.Context, postID int) ([]domain.PostImage, error) {
			return []domain.PostImage{{ID: 1}}, nil
		},
		deleteRecordsFn: func(ctx context.Context, postID int) error {
			return nil
		},
		deleteFromStorageFn: func(ctx context.Context, images []domain.PostImage) error {
			if len(images) != 1 {
				t.Fatalf("expected 1 image, got %d", len(images))
			}
			return nil
		},
	}
	tx := &mockTxManager{
		withinFn: func(ctx context.Context, fn func(ctx context.Context) error) error {
			return fn(ctx)
		},
	}

	svc := newPostsService(postsRepo, imagesSvc, &mockDebatesRepository{}, &mockDebateSidesRepository{}, tx)
	err := svc.DeletePost(context.Background(), 5, 1)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
}

var _ = time.Now
