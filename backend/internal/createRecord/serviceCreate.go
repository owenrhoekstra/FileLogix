package createRecord

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"

	"FileLogix/middleware"
	"FileLogix/rabbitmq"
	"FileLogix/utilities/logger"

	"github.com/google/uuid"
)

const storageRoot = "/srv/FileLogix/files"

func CreateRecord(ctx context.Context, input CreateRecordInput) (uuid.UUID, error) {
	requestID := ctx.Value(middleware.RequestIDKey).(uuid.UUID)
	userID := ctx.Value(middleware.UserIDKey).(uuid.UUID)

	documentID := uuid.New()

	storagePath := buildStoragePath(documentID)
	if err := os.MkdirAll(storagePath, 0755); err != nil {
		logger.Errorf(requestID, userID, "CreateRecord: failed to create storage directory: %v", err)
		return uuid.Nil, fmt.Errorf("failed to create storage directory: %w", err)
	}

	if err := InsertDocument(documentID, input); err != nil {
		logger.Errorf(requestID, userID, "CreateRecord: failed to insert document: %v", err)
		return uuid.Nil, fmt.Errorf("failed to insert document: %w", err)
	}

	for i, fh := range input.Files {
		if err := saveFile(requestID, userID, documentID, fh, i, storagePath); err != nil {
			logger.Errorf(requestID, userID, "CreateRecord: failed to save file %d: %v", i, err)
			return uuid.Nil, fmt.Errorf("failed to save file %d: %w", i, err)
		}
	}

	event, err := json.Marshal(map[string]any{
		"document_id": documentID,
	})
	if err != nil {
		logger.Errorf(requestID, userID, "CreateRecord: failed to marshal ocr.pending event: %v", err)
	} else if err := rabbitmq.Publish("ocr.pending", event); err != nil {
		logger.Errorf(requestID, userID, "CreateRecord: failed to publish ocr.pending event: %v", err)
		// intentionally not returned — record creation succeeds regardless
	}

	logger.Infof(requestID, userID, "CreateRecord: document %v created successfully", documentID)

	return documentID, nil
}

func buildStoragePath(documentID uuid.UUID) string {
	hash := sha256.Sum256([]byte(documentID.String()))
	hashHex := fmt.Sprintf("%x", hash)
	bucket1 := hashHex[0:2]
	bucket2 := hashHex[2:4]
	return filepath.Join(storageRoot, bucket1, bucket2, documentID.String())
}

func saveFile(requestID, userID, documentID uuid.UUID, fh *multipart.FileHeader, pageNumber int, storagePath string) error {
	ext := filepath.Ext(fh.Filename)
	if ext == "" {
		ext = ".webp"
	}

	filename := fmt.Sprintf("%d%s", pageNumber, ext)
	fullPath := filepath.Join(storagePath, filename)

	src, err := fh.Open()
	if err != nil {
		logger.Errorf(requestID, userID, "saveFile: failed to open uploaded file %d: %v", pageNumber, err)
		return fmt.Errorf("failed to open uploaded file: %w", err)
	}
	defer src.Close()

	dst, err := os.Create(fullPath)
	if err != nil {
		logger.Errorf(requestID, userID, "saveFile: failed to create destination file %d: %v", pageNumber, err)
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		logger.Errorf(requestID, userID, "saveFile: failed to write file %d: %v", pageNumber, err)
		return fmt.Errorf("failed to write file: %w", err)
	}

	if err := InsertFile(documentID, fullPath, fh.Size, fh.Header.Get("Content-Type"), pageNumber); err != nil {
		logger.Errorf(requestID, userID, "saveFile: failed to insert file record %d: %v", pageNumber, err)
		return fmt.Errorf("failed to insert file record: %w", err)
	}

	return nil
}
