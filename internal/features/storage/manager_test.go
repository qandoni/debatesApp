package storage

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"

	core_logger "github.com/qandoni/debatesApp/internal/core/logger"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
)

type fakeImageStorage struct {
	uploaded   []string
	deleted    []string
	deleteErrs map[string]error
	uploadErr  error
	exists     map[string]bool
}

func (f *fakeImageStorage) ObjectExists(ctx context.Context, objectName string) (bool, error) {
	return f.exists[objectName], nil
}

func (f *fakeImageStorage) Upload(ctx context.Context, objectName string, reader io.Reader, size int64, contentType string) error {
	if f.uploadErr != nil {
		return f.uploadErr
	}
	f.uploaded = append(f.uploaded, objectName)
	return nil
}

func (f *fakeImageStorage) Delete(ctx context.Context, objectName string) error {
	if err := f.deleteErrs[objectName]; err != nil {
		return err
	}
	f.deleted = append(f.deleted, objectName)
	return nil
}

func (f *fakeImageStorage) GetURL(ctx context.Context, objectName string) (string, error) {
	return "http://minio.local/" + objectName, nil
}

func newTestManager(fake *fakeImageStorage) (*Manager, *observer.ObservedLogs) {
	core, logs := observer.New(zap.InfoLevel)
	return NewManager(fake, &core_logger.Logger{Logger: zap.New(core)}), logs
}

func TestManager_UploadLogsCreatedObject(t *testing.T) {
	fake := &fakeImageStorage{}
	m, logs := newTestManager(fake)

	if err := m.Upload(context.Background(), "posts/1/a.png", nil, 42, "image/png"); err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if len(fake.uploaded) != 1 || fake.uploaded[0] != "posts/1/a.png" {
		t.Fatalf("unexpected uploads: %v", fake.uploaded)
	}

	entries := logs.FilterMessage("minio object created").All()
	if len(entries) != 1 {
		t.Fatalf("expected 1 created-log entry, got %d", len(entries))
	}
	if got := entries[0].ContextMap()["object"]; got != "posts/1/a.png" {
		t.Fatalf("expected logged object posts/1/a.png, got: %v", got)
	}
}

func TestManager_DeleteAllCollectsOrphans(t *testing.T) {
	fake := &fakeImageStorage{
		deleteErrs: map[string]error{"b": errors.New("minio down")},
	}
	m, logs := newTestManager(fake)

	orphaned := m.DeleteAll(context.Background(), []string{"a", "b"})

	if len(orphaned) != 1 || orphaned[0] != "b" {
		t.Fatalf("expected orphan [b], got: %v", orphaned)
	}
	if len(fake.deleted) != 1 || fake.deleted[0] != "a" {
		t.Fatalf("expected [a] deleted, got: %v", fake.deleted)
	}
	entries := logs.FilterMessage("minio object orphaned").All()
	if len(entries) != 1 {
		t.Fatalf("expected 1 orphan-log entry, got %d", len(entries))
	}
	if got := entries[0].ContextMap()["object"]; got != "b" {
		t.Fatalf("expected orphan object b, got: %v", got)
	}
}

func TestManagerTx_RollbackDeletesUploadedObjects(t *testing.T) {
	fake := &fakeImageStorage{}
	m, _ := newTestManager(fake)
	ctx := context.Background()

	tx := m.Begin()
	if err := tx.Upload(ctx, "posts/1/1.png", nil, 1, "image/png"); err != nil {
		t.Fatalf("upload 1: %v", err)
	}
	if err := tx.Upload(ctx, "posts/1/2.png", nil, 1, "image/png"); err != nil {
		t.Fatalf("upload 2: %v", err)
	}

	tx.Rollback(ctx)

	if len(fake.deleted) != 2 {
		t.Fatalf("expected 2 deleted objects, got: %v", fake.deleted)
	}
	// идемпотентен: повторный Rollback ничего не удаляет
	tx.Rollback(ctx)
	if len(fake.deleted) != 2 {
		t.Fatalf("expected still 2 deleted objects, got: %v", fake.deleted)
	}
}

func TestManagerTx_RollbackAfterPartialUploadFailure(t *testing.T) {
	fake := &fakeImageStorage{}
	m, _ := newTestManager(fake)
	ctx := context.Background()

	tx := m.Begin()
	if err := tx.Upload(ctx, "posts/1/1.png", nil, 1, "image/png"); err != nil {
		t.Fatalf("upload 1: %v", err)
	}
	fake.uploadErr = errors.New("minio down")
	if err := tx.Upload(ctx, "posts/1/2.png", nil, 1, "image/png"); err == nil {
		t.Fatal("expected upload 2 to fail")
	}

	tx.Rollback(ctx)

	if len(fake.deleted) != 1 || fake.deleted[0] != "posts/1/1.png" {
		t.Fatalf("expected uploaded-then-rolled-back object deleted, got: %v", fake.deleted)
	}
}

func TestManagerTx_CommitKeepsUploadedObjects(t *testing.T) {
	fake := &fakeImageStorage{}
	m, _ := newTestManager(fake)
	ctx := context.Background()

	tx := m.Begin()
	if err := tx.Upload(ctx, "posts/1/1.png", nil, 1, "image/png"); err != nil {
		t.Fatalf("upload: %v", err)
	}
	tx.Commit()
	tx.Rollback(ctx) // после Commit откат ничего не делает

	if len(fake.deleted) != 0 {
		t.Fatalf("expected no deletions after commit, got: %v", fake.deleted)
	}
}

func TestManager_NewObjectName(t *testing.T) {
	m, _ := newTestManager(&fakeImageStorage{})

	a := m.NewObjectName("posts/1", ".png")
	b := m.NewObjectName("posts/1", ".png")

	if !strings.HasPrefix(a, "posts/1/") || !strings.HasSuffix(a, ".png") {
		t.Fatalf("unexpected name format: %q", a)
	}
	if a == b {
		t.Fatalf("expected unique names, got twice %q", a)
	}
}
