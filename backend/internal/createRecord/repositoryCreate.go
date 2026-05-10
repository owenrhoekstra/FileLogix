package createRecord

import (
	"encoding/json"
	"fmt"

	"FileLogix/database"

	"github.com/google/uuid"
)

func InsertDocument(documentID uuid.UUID, input CreateRecordInput) error {
	typesJSON, err := json.Marshal(input.Types)
	if err != nil {
		return fmt.Errorf("failed to marshal types: %w", err)
	}

	_, err = database.DB.Exec(`
		INSERT INTO documents (
			id,
			name,
			types,
			date_of_doc,
			sensitive,
			uploaded_by
		) VALUES ($1, $2, $3, $4, $5, $6)
	`,
		documentID,
		input.Name,
		typesJSON,
		input.DateOfDoc,
		input.Sensitive,
		input.UploadedBy,
	)
	if err != nil {
		return fmt.Errorf("failed to insert document: %w", err)
	}

	return nil
}

func InsertFile(documentID uuid.UUID, path string, size int64, mimeType string, pageNumber int) error {
	fileID := uuid.New()

	_, err := database.DB.Exec(`
		INSERT INTO files (
			id,
			document_id,
			path,
			size,
			mime_type,
			status,
			page_number
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
	`,
		fileID,
		documentID,
		path,
		size,
		mimeType,
		FileStatusPending,
		pageNumber,
	)
	if err != nil {
		return fmt.Errorf("failed to insert file record: %w", err)
	}

	return nil
}
