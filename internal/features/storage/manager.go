package storage

import (
	"context"
	"fmt"
	"io"

	"github.com/google/uuid"
	core_logger "github.com/qandoni/debatesApp/internal/core/logger"
	"go.uber.org/zap"
)

type Manager struct {
	storage ImageStorage
	log     *core_logger.Logger
}

func NewManager(storage ImageStorage, log *core_logger.Logger) *Manager {
	return &Manager{storage: storage, log: log}
}

func (m *Manager) NewObjectName(prefix, extension string) string {
	return fmt.Sprintf("%s/%s%s", prefix, uuid.New().String(), extension)
}

func (m *Manager) Upload(
	ctx context.Context,
	objectName string,
	reader io.Reader,
	size int64,
	contentType string,
) error {
	if err := m.storage.Upload(ctx, objectName, reader, size, contentType); err != nil {
		return err
	}
	m.log.Info("minio object created",
		zap.String("object", objectName),
		zap.Int64("size", size),
		zap.String("content_type", contentType),
	)
	return nil
}

func (m *Manager) GetURL(ctx context.Context, objectName string) (string, error) {
	return m.storage.GetURL(ctx, objectName)
}

func (m *Manager) ObjectExists(ctx context.Context, objectName string) (bool, error) {
	return m.storage.ObjectExists(ctx, objectName)
}

func (m *Manager) Delete(ctx context.Context, objectName string) error {
	if err := m.storage.Delete(ctx, objectName); err != nil {
		return err
	}
	m.log.Info("minio object deleted", zap.String("object", objectName))
	return nil
}

func (m *Manager) DeleteAll(ctx context.Context, objectNames []string) []string {
	orphaned := make([]string, 0)
	for _, objectName := range objectNames {
		if err := m.Delete(ctx, objectName); err != nil {
			m.log.Error("minio object orphaned",
				zap.String("object", objectName),
				zap.Error(err),
			)
			orphaned = append(orphaned, objectName)
		}
	}
	return orphaned
}

func (m *Manager) Begin() *Tx {
	return &Tx{manager: m}
}

type Tx struct {
	manager  *Manager
	uploaded []string
}

func (tx *Tx) Upload(
	ctx context.Context,
	objectName string,
	reader io.Reader,
	size int64,
	contentType string,
) error {
	if err := tx.manager.Upload(ctx, objectName, reader, size, contentType); err != nil {
		return err
	}
	tx.uploaded = append(tx.uploaded, objectName)
	return nil
}

func (tx *Tx) Commit() {
	tx.uploaded = nil
}

func (tx *Tx) Rollback(ctx context.Context) {
	tx.manager.DeleteAll(ctx, tx.uploaded)
	tx.uploaded = nil
}
