package images_service

import (
	"bytes"
	"context"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"github.com/qandoni/debatesApp/internal/core/domain"
	core_logger "github.com/qandoni/debatesApp/internal/core/logger"
	core_postgres "github.com/qandoni/debatesApp/internal/core/repository/postgres"
	"github.com/qandoni/debatesApp/internal/features/storage"
	"go.uber.org/zap"
)

type mockPostsRepository struct {
	getPostFn func(ctx context.Context, postID int) (domain.Post, error)
}

func (m *mockPostsRepository) GetPost(ctx context.Context, postID int) (domain.Post, error) {
	return m.getPostFn(ctx, postID)
}

type mockPostImagesRepository struct {
	createFn              func(ctx context.Context, image domain.PostImage) (domain.PostImage, error)
	getByPostFn           func(ctx context.Context, postID int) ([]domain.PostImage, error)
	getByPostsFn          func(ctx context.Context, postIDs []int) (map[int][]domain.PostImage, error)
	deleteByPostFn        func(ctx context.Context, postID int) error
	markActiveFn          func(ctx context.Context, imageIDs []int) error
	deletePendingFn       func(ctx context.Context, imageIDs []int) error
	getPendingFn          func(ctx context.Context) ([]domain.PostImage, error)
	getPendingOlderThanFn func(ctx context.Context, staleAfter time.Duration) ([]domain.PostImage, error)
	getDeletedFn          func(ctx context.Context) ([]domain.PostImage, error)
}

func (m *mockPostImagesRepository) CreatePostImage(ctx context.Context, image domain.PostImage) (domain.PostImage, error) {
	return m.createFn(ctx, image)
}
func (m *mockPostImagesRepository) GetByPostID(ctx context.Context, postID int) ([]domain.PostImage, error) {
	return m.getByPostFn(ctx, postID)
}
func (m *mockPostImagesRepository) GetByPostIDs(ctx context.Context, postIDs []int) (map[int][]domain.PostImage, error) {
	return m.getByPostsFn(ctx, postIDs)
}
func (m *mockPostImagesRepository) DeleteByPostID(ctx context.Context, postID int) error {
	return m.deleteByPostFn(ctx, postID)
}
func (m *mockPostImagesRepository) MarkActiveByIDs(ctx context.Context, imageIDs []int) error {
	return m.markActiveFn(ctx, imageIDs)
}
func (m *mockPostImagesRepository) DeletePendingByIDs(ctx context.Context, imageIDs []int) error {
	return m.deletePendingFn(ctx, imageIDs)
}
func (m *mockPostImagesRepository) GetPending(ctx context.Context) ([]domain.PostImage, error) {
	return m.getPendingFn(ctx)
}
func (m *mockPostImagesRepository) GetPendingOlderThan(ctx context.Context, staleAfter time.Duration) ([]domain.PostImage, error) {
	return m.getPendingOlderThanFn(ctx, staleAfter)
}
func (m *mockPostImagesRepository) GetImagesOfDeletedPosts(ctx context.Context) ([]domain.PostImage, error) {
	return m.getDeletedFn(ctx)
}

type mockTxManager struct {
	withinFn func(ctx context.Context, fn func(ctx context.Context) error) error
}

func (m *mockTxManager) WithinTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	if m.withinFn == nil {
		return fn(ctx)
	}
	return m.withinFn(ctx, fn)
}

// fakeStorage имитирует ImageStorage: записывает вызовы и умеет падать.
type fakeStorage struct {
	uploaded       []string
	deleted        []string
	uploadErrs     map[string]error
	uploadFailFrom int // начиная с этого вызова Upload падает (1-based; 0 = никогда)
	uploadCalls    int
	deleteErrs     map[string]error
	onUpload       func(objectName string)
	exists         map[string]bool
	existsErr      error
}

func (f *fakeStorage) ObjectExists(ctx context.Context, objectName string) (bool, error) {
	if f.existsErr != nil {
		return false, f.existsErr
	}
	return f.exists[objectName], nil
}

func (f *fakeStorage) Upload(ctx context.Context, objectName string, reader io.Reader, size int64, contentType string) error {
	f.uploadCalls++
	if f.uploadFailFrom > 0 && f.uploadCalls >= f.uploadFailFrom {
		return errors.New("minio down")
	}
	if err := f.uploadErrs[objectName]; err != nil {
		return err
	}
	f.uploaded = append(f.uploaded, objectName)
	if f.onUpload != nil {
		f.onUpload(objectName)
	}
	return nil
}

func (f *fakeStorage) Delete(ctx context.Context, objectName string) error {
	if err := f.deleteErrs[objectName]; err != nil {
		return err
	}
	f.deleted = append(f.deleted, objectName)
	return nil
}

func (f *fakeStorage) GetURL(ctx context.Context, objectName string) (string, error) {
	return "http://minio.local/" + objectName, nil
}

func newImagesTestService(
	postsRepo PostsRepository,
	imagesRepo PostImagesRepository,
	tx core_postgres.TransactionManager,
	fake *fakeStorage,
) *ImagesService {
	manager := storage.NewManager(fake, &core_logger.Logger{Logger: zap.NewNop()})
	return NewImagesService(manager, postsRepo, imagesRepo, tx)
}

// pngHeader собирает валидный multipart-заголовок для файлa image/png.
func pngHeader(t *testing.T) *multipart.FileHeader {
	t.Helper()

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	fw, err := writer.CreateFormFile("file", "a.png")
	if err != nil {
		t.Fatalf("create form file: %v", err)
	}
	if _, err := fw.Write([]byte("\x89PNG\r\n\x1a\nfake-payload")); err != nil {
		t.Fatalf("write payload: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close writer: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/", &buf)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	if err := req.ParseMultipartForm(1 << 20); err != nil {
		t.Fatalf("parse multipart form: %v", err)
	}
	return req.MultipartForm.File["file"][0]
}

func TestCreatePostImages_RowsCommittedBeforeUpload(t *testing.T) {
	var order []string
	createdNames := make([]string, 0, 2)

	postsRepo := &mockPostsRepository{
		getPostFn: func(ctx context.Context, postID int) (domain.Post, error) {
			return domain.Post{ID: 1, AuthorID: 5}, nil
		},
	}
	imagesRepo := &mockPostImagesRepository{
		createFn: func(ctx context.Context, image domain.PostImage) (domain.PostImage, error) {
			order = append(order, "db-insert")
			image.ID = len(order)
			createdNames = append(createdNames, image.ImageURL)
			return image, nil
		},
		markActiveFn: func(ctx context.Context, imageIDs []int) error {
			order = append(order, "mark-active")
			return nil
		},
	}
	fake := &fakeStorage{
		uploadErrs: map[string]error{},
		onUpload: func(objectName string) {
			order = append(order, "upload")
		},
	}
	svc := newImagesTestService(postsRepo, imagesRepo, &mockTxManager{}, fake)

	files := []*multipart.FileHeader{pngHeader(t), pngHeader(t)}
	images, err := svc.CreatePostImages(context.Background(), 5, 1, files)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if len(images) != 2 || images[0].ID == 0 || images[1].ID == 0 {
		t.Fatalf("unexpected images: %+v", images)
	}

	// ключевой порядок: строки (PENDING) → файлы → подтверждение ACTIVE
	want := []string{"db-insert", "db-insert", "upload", "upload", "mark-active"}
	if !reflect.DeepEqual(order, want) {
		t.Fatalf("expected order %v, got: %v", want, order)
	}
	if !reflect.DeepEqual(fake.uploaded, createdNames) {
		t.Fatalf("expected uploaded %v, got: %v", createdNames, fake.uploaded)
	}
}

func TestCreatePostImages_UploadFailureCompensates(t *testing.T) {
	createdNames := make([]string, 0, 2)
	createdIDs := make([]int, 0, 2)
	marked := false
	compensatedIDs := []int(nil)

	postsRepo := &mockPostsRepository{
		getPostFn: func(ctx context.Context, postID int) (domain.Post, error) {
			return domain.Post{ID: 1, AuthorID: 5}, nil
		},
	}
	imagesRepo := &mockPostImagesRepository{
		createFn: func(ctx context.Context, image domain.PostImage) (domain.PostImage, error) {
			image.ID = len(createdNames) + 1
			createdNames = append(createdNames, image.ImageURL)
			createdIDs = append(createdIDs, image.ID)
			return image, nil
		},
		deletePendingFn: func(ctx context.Context, imageIDs []int) error {
			compensatedIDs = imageIDs
			return nil
		},
		markActiveFn: func(ctx context.Context, imageIDs []int) error {
			marked = true
			return nil
		},
	}
	fake := &fakeStorage{uploadErrs: map[string]error{}, uploadFailFrom: 2}
	svc := newImagesTestService(postsRepo, imagesRepo, &mockTxManager{}, fake)

	files := []*multipart.FileHeader{pngHeader(t), pngHeader(t)}
	if _, err := svc.CreatePostImages(context.Background(), 5, 1, files); err == nil {
		t.Fatal("expected upload error, got nil")
	}

	// компенсация: удалены ВСЕ имена батча (незалитые — идемпотентный no-op)
	if !reflect.DeepEqual(fake.deleted, createdNames) {
		t.Fatalf("expected compensated deletion %v, got: %v", createdNames, fake.deleted)
	}
	if !reflect.DeepEqual(compensatedIDs, createdIDs) {
		t.Fatalf("expected pending rows %v deleted, got: %v", createdIDs, compensatedIDs)
	}
	if marked {
		t.Fatal("batch must not be activated after upload failure")
	}
}

func TestCreatePostImages_CommitFailureNeverUploads(t *testing.T) {
	txCalls := 0
	pendingDeleted := false

	postsRepo := &mockPostsRepository{
		getPostFn: func(ctx context.Context, postID int) (domain.Post, error) {
			return domain.Post{ID: 1, AuthorID: 5}, nil
		},
	}
	imagesRepo := &mockPostImagesRepository{
		createFn: func(ctx context.Context, image domain.PostImage) (domain.PostImage, error) {
			image.ID = 1
			return image, nil
		},
		deletePendingFn: func(ctx context.Context, imageIDs []int) error {
			pendingDeleted = true
			return nil
		},
	}
	tx := &mockTxManager{
		withinFn: func(ctx context.Context, fn func(ctx context.Context) error) error {
			txCalls++
			if err := fn(ctx); err != nil {
				return err
			}
			if txCalls == 1 {
				return errors.New("commit failed")
			}
			return nil
		},
	}
	fake := &fakeStorage{uploadErrs: map[string]error{}}
	svc := newImagesTestService(postsRepo, imagesRepo, tx, fake)

	files := []*multipart.FileHeader{pngHeader(t)}
	if _, err := svc.CreatePostImages(context.Background(), 5, 1, files); err == nil {
		t.Fatal("expected commit error, got nil")
	}

	// коммит упал → в MinIO не было НИ ОДНОЙ заливки: порядок коммит → файлы
	if len(fake.uploaded) != 0 || fake.uploadCalls != 0 {
		t.Fatalf("expected no uploads after commit failure, got: %v (calls=%d)", fake.uploaded, fake.uploadCalls)
	}
	// PENDING-строки остались — журнал для recovery (компенсация не нужна:
	// файлов нет)
	if pendingDeleted {
		t.Fatal("pending rows must be kept for recovery when commit fails")
	}
}
