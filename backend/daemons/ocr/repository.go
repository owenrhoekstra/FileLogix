package ocr

import (
	"database/sql"
	"errors"

	"github.com/google/uuid"

	"FileLogix/database"
)

type fileRecord struct {
	ID         uuid.UUID
	DocumentID uuid.UUID
	Path       string
}

func resetStuck() error {
	_, err := database.DB.Exec(`
		UPDATE files
		SET ocr_status = 'pending'
		WHERE ocr_status = 'processing'
	`)
	return err
}

func claimNext() (*fileRecord, error) {
	row := database.DB.QueryRow(`
		UPDATE files
		SET ocr_status = 'processing'
		WHERE id = (
			SELECT id FROM files
			WHERE ocr_status = 'pending'
			AND deleted = FALSE
			ORDER BY created_at
			LIMIT 1
			FOR UPDATE SKIP LOCKED
		)
		RETURNING id, document_id, path
	`)

	var f fileRecord
	err := row.Scan(&f.ID, &f.DocumentID, &f.Path)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &f, nil
}

func markFailed(fileID uuid.UUID) error {
	_, err := database.DB.Exec(`
        UPDATE files 
        SET ocr_retry_count = ocr_retry_count + 1,
            ocr_status = CASE 
                WHEN ocr_retry_count + 1 >= 3 THEN 'failed'::pipeline_status
                ELSE 'pending'::pipeline_status
            END
        WHERE id = $1
    `, fileID)
	return err
}

func saveResult(fileID uuid.UUID, documentID uuid.UUID, text string) error {
	tx, err := database.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.Exec(`
		UPDATE files SET ocr_status = 'complete' WHERE id = $1
	`, fileID)
	if err != nil {
		return err
	}

	_, err = tx.Exec(`
		UPDATE documents SET ocr_text = $1 WHERE id = $2
	`, text, documentID)
	if err != nil {
		return err
	}

	return tx.Commit()
}
