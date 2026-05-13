package documentTypes

import (
	"context"
	"encoding/json"
	"fmt"

	"FileLogix/database"
	"FileLogix/middleware"
	"FileLogix/utilities/logger"

	"github.com/google/uuid"
)

type documentTypeEntry struct {
	DocumentLabel      string `json:"documentLabel"`
	DocumentLabelValue string `json:"documentLabelValue"`
}

func insertDocumentType(ctx context.Context, entry documentTypeEntry) error {
	requestID := ctx.Value(middleware.RequestIDKey).(uuid.UUID)
	userID := ctx.Value(middleware.UserIDKey).(uuid.UUID)

	entryJSON, err := json.Marshal(entry)
	if err != nil {
		logger.Errorf(requestID, userID, "insertDocumentType: marshal error: %v", err)
		return err
	}

	_, err = database.DB.ExecContext(ctx, `
		UPDATE settings
		SET value = value || $1::jsonb
		WHERE key = 'document_types'
	`, fmt.Sprintf("[%s]", entryJSON))
	if err != nil {
		logger.Errorf(requestID, userID, "insertDocumentType: exec error: %v", err)
		return err
	}
	return nil
}
