package images_service

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/qandoni/debatesApp/internal/core/domain"
)

func TestRecover_RollsBackPendingUploads(t *testing.T) {
	pending := []domain.PostImage{
		{ID: 1, PostID: 10, ImageURL: "posts/10/a.png"},
		{ID: 2, PostID: 10, ImageURL: "posts/10/b.png"},
	}
	deletedRowsIDs := []int(nil)
	rowsByPost := []int(nil)

	imagesRepo := &mockPostImagesRepository{
		getPendingFn: func(ctx context.Context) ([]domain.PostImage, error) {
			return pending, nil
		},
		deletePendingFn: func(ctx context.Context, imageIDs []int) error {
			deletedRowsIDs = imageIDs
			return nil
		},
		getDeletedFn: func(ctx context.Context) ([]domain.PostImage, error) {
			return nil, nil
		},
		deleteByPostFn: func(ctx context.Context, postID int) error {
			rowsByPost = append(rowsByPost, postID)
			return nil
		},
	}
	fake := &fakeStorage{uploadErrs: map[string]error{}}
	svc := newImagesTestService(&mockPostsRepository{}, imagesRepo, &mockTxManager{}, fake)

	if err := svc.Recover(context.Background()); err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	// откат незавершённой загрузки: файлы батча удалены по API, затем строки
	want := []string{"posts/10/a.png", "posts/10/b.png"}
	if !reflect.DeepEqual(fake.deleted, want) {
		t.Fatalf("expected files deleted %v, got: %v", want, fake.deleted)
	}
	if !reflect.DeepEqual(deletedRowsIDs, []int{1, 2}) {
		t.Fatalf("expected pending rows [1 2] deleted, got: %v", deletedRowsIDs)
	}
	if len(rowsByPost) != 0 {
		t.Fatalf("expected no deleted-posts cleanup, got: %v", rowsByPost)
	}
}

func TestRecover_FinishesInterruptedDeletion(t *testing.T) {
	rowsByPost := []int(nil)

	imagesRepo := &mockPostImagesRepository{
		getPendingFn: func(ctx context.Context) ([]domain.PostImage, error) {
			return nil, nil
		},
		getDeletedFn: func(ctx context.Context) ([]domain.PostImage, error) {
			return []domain.PostImage{{ID: 5, PostID: 7, ImageURL: "posts/7/a.png"}}, nil
		},
		deleteByPostFn: func(ctx context.Context, postID int) error {
			rowsByPost = append(rowsByPost, postID)
			return nil
		},
	}
	fake := &fakeStorage{uploadErrs: map[string]error{}}
	svc := newImagesTestService(&mockPostsRepository{}, imagesRepo, &mockTxManager{}, fake)

	if err := svc.Recover(context.Background()); err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	// доделка прерванного удаления: файл удалён, затем строки поста
	if !reflect.DeepEqual(fake.deleted, []string{"posts/7/a.png"}) {
		t.Fatalf("expected file deleted, got: %v", fake.deleted)
	}
	if !reflect.DeepEqual(rowsByPost, []int{7}) {
		t.Fatalf("expected rows of post 7 deleted, got: %v", rowsByPost)
	}
}

func TestRecover_KeepsPendingRowsWhenFilesUndeletable(t *testing.T) {
	pendingRowsDeleted := false

	imagesRepo := &mockPostImagesRepository{
		getPendingFn: func(ctx context.Context) ([]domain.PostImage, error) {
			return []domain.PostImage{{ID: 1, PostID: 10, ImageURL: "posts/10/a.png"}}, nil
		},
		deletePendingFn: func(ctx context.Context, imageIDs []int) error {
			pendingRowsDeleted = true
			return nil
		},
		getDeletedFn: func(ctx context.Context) ([]domain.PostImage, error) {
			return nil, nil
		},
	}
	fake := &fakeStorage{
		uploadErrs: map[string]error{},
		deleteErrs: map[string]error{"posts/10/a.png": errors.New("minio down")},
	}
	svc := newImagesTestService(&mockPostsRepository{}, imagesRepo, &mockTxManager{}, fake)

	if err := svc.Recover(context.Background()); err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	// файлы не удалены → строки остаются маркером для следующего старта
	if pendingRowsDeleted {
		t.Fatal("pending rows must be kept while files are undeletable")
	}
}

func TestRecover_KeepsDeletedPostRowsWhenFilesUndeletable(t *testing.T) {
	rowsByPost := []int(nil)

	imagesRepo := &mockPostImagesRepository{
		getPendingFn: func(ctx context.Context) ([]domain.PostImage, error) {
			return nil, nil
		},
		getDeletedFn: func(ctx context.Context) ([]domain.PostImage, error) {
			return []domain.PostImage{{ID: 5, PostID: 7, ImageURL: "posts/7/a.png"}}, nil
		},
		deleteByPostFn: func(ctx context.Context, postID int) error {
			rowsByPost = append(rowsByPost, postID)
			return nil
		},
	}
	fake := &fakeStorage{
		uploadErrs: map[string]error{},
		deleteErrs: map[string]error{"posts/7/a.png": errors.New("minio down")},
	}
	svc := newImagesTestService(&mockPostsRepository{}, imagesRepo, &mockTxManager{}, fake)

	if err := svc.Recover(context.Background()); err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if len(rowsByPost) != 0 {
		t.Fatalf("rows must be kept while files are undeletable, got: %v", rowsByPost)
	}
}

// RecoverStale использует выборку "старше порога" (а не все PENDING), чтобы
// периодический проход не трогал живые незавершённые загрузки.
func TestRecoverStale_UsesStalePendingQuery(t *testing.T) {
	var gotStaleAfter time.Duration
	pendingRowsDeleted := []int(nil)

	imagesRepo := &mockPostImagesRepository{
		getPendingOlderThanFn: func(ctx context.Context, staleAfter time.Duration) ([]domain.PostImage, error) {
			gotStaleAfter = staleAfter
			return []domain.PostImage{{ID: 1, PostID: 10, ImageURL: "posts/10/a.png"}}, nil
		},
		deletePendingFn: func(ctx context.Context, imageIDs []int) error {
			pendingRowsDeleted = imageIDs
			return nil
		},
		getDeletedFn: func(ctx context.Context) ([]domain.PostImage, error) {
			return nil, nil
		},
	}
	fake := &fakeStorage{uploadErrs: map[string]error{}}
	svc := newImagesTestService(&mockPostsRepository{}, imagesRepo, &mockTxManager{}, fake)

	if err := svc.RecoverStale(context.Background(), 5*time.Minute); err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if gotStaleAfter != 5*time.Minute {
		t.Fatalf("expected staleAfter 5m, got: %v", gotStaleAfter)
	}
	if !reflect.DeepEqual(fake.deleted, []string{"posts/10/a.png"}) {
		t.Fatalf("expected stale pending file deleted, got: %v", fake.deleted)
	}
	if !reflect.DeepEqual(pendingRowsDeleted, []int{1}) {
		t.Fatalf("expected stale pending row [1] deleted, got: %v", pendingRowsDeleted)
	}
}
