package createRecord

import (
	"mime/multipart"
	"time"

	"github.com/google/uuid"
)

type CreateRecordInput struct {
	Name       string
	DateOfDoc  string
	Sensitive  bool
	Types      []string
	Files      []*multipart.FileHeader
	UploadedBy uuid.UUID
}

type Document struct {
	ID           uuid.UUID
	Name         string
	Types        []string
	DateOfDoc    time.Time
	DateFiled    time.Time
	LastModified time.Time
	Sensitive    bool
	UploadedBy   uuid.UUID
	OcrText      *string
	Embedding    *[]float32
}

type File struct {
	ID         uuid.UUID
	DocumentID uuid.UUID
	Path       string
	Size       int64
	MimeType   string
	Status     FileStatus
	PageNumber int
	CreatedAt  time.Time
}

type FileStatus string

const (
	FileStatusPending    FileStatus = "pending"
	FileStatusProcessing FileStatus = "processing"
	FileStatusComplete   FileStatus = "complete"
	FileStatusFailed     FileStatus = "failed"
)
