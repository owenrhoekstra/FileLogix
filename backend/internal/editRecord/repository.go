package editRecord

import (
	"context"
	"fmt"
	"time"

	"FileLogix/database"
	"FileLogix/middleware"
	"FileLogix/utilities/logger"

	"github.com/google/uuid"
)

func updateRecord(ctx context.Context, id uuid.UUID, name string, sensitive bool, typesJSON []byte, dateOfDoc time.Time) error {
	requestID := ctx.Value(middleware.RequestIDKey).(uuid.UUID)
	userID := ctx.Value(middleware.UserIDKey).(uuid.UUID)

	_, err := database.DB.ExecContext(ctx, `
		UPDATE documents
		SET name        = $1,
		    sensitive   = $2,
		    types       = $3::jsonb,
		    date_of_doc = $4
		WHERE id = $5
		  AND deleted = false
	`, name, sensitive, typesJSON, dateOfDoc, id)
	if err != nil {
		logger.Errorf(requestID, userID, "updateRecord: exec error: %v", err)
		return fmt.Errorf("updating record: %w", err)
	}

	return nil
}

func undeleteRecord(ctx context.Context, id uuid.UUID) error {
	requestID := ctx.Value(middleware.RequestIDKey).(uuid.UUID)
	userID := ctx.Value(middleware.UserIDKey).(uuid.UUID)

	_, err := database.DB.ExecContext(ctx, `
		UPDATE documents
		SET deleted    = false,
		    deleted_at = NULL,
		    deleted_by = NULL
		WHERE id = $1
		  AND deleted = true
	`, id)
	if err != nil {
		logger.Errorf(requestID, userID, "undeleteRecord: exec error: %v", err)
		return fmt.Errorf("restoring record: %w", err)
	}

	return nil
}
