package createRecord

import (
	"crypto/sha256"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"

	"github.com/google/uuid"
)

const storageRoot = "/srv/FileLogix/files"

func CreateRecord(input CreateRecordInput) (uuid.UUID, error) {
	documentID := uuid.New()

	storagePath := buildStoragePath(documentID)
	if err := os.MkdirAll(storagePath, 0755); err != nil {
		return uuid.Nil, fmt.Errorf("failed to create storage directory: %w", err)
	}

	if err := InsertDocument(documentID, input); err != nil {
		return uuid.Nil, fmt.Errorf("failed to insert document: %w", err)
	}

	for i, fh := range input.Files {
		if err := saveFile(documentID, fh, i, storagePath); err != nil {
			return uuid.Nil, fmt.Errorf("failed to save file %d: %w", i, err)
		}
	}

	return documentID, nil
}

func buildStoragePath(documentID uuid.UUID) string {
	hash := sha256.Sum256([]byte(documentID.String()))
	hashHex := fmt.Sprintf("%x", hash)
	bucket1 := hashHex[0:2]
	bucket2 := hashHex[2:4]
	return filepath.Join(storageRoot, bucket1, bucket2, documentID.String())
}

func saveFile(documentID uuid.UUID, fh *multipart.FileHeader, pageNumber int, storagePath string) error {
	ext := filepath.Ext(fh.Filename)
	if ext == "" {
		ext = ".webp"
	}

	filename := fmt.Sprintf("%d%s", pageNumber, ext)
	fullPath := filepath.Join(storagePath, filename)

	src, err := fh.Open()
	if err != nil {
		return fmt.Errorf("failed to open uploaded file: %w", err)
	}
	defer src.Close()

	dst, err := os.Create(fullPath)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	if err := InsertFile(documentID, fullPath, fh.Size, fh.Header.Get("Content-Type"), pageNumber); err != nil {
		return fmt.Errorf("failed to insert file record: %w", err)
	}

	return nil
}
