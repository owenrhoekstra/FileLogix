package viewRecord

import (
	"FileLogix/database"
	"FileLogix/utilities/logger"

	"github.com/google/uuid"
)

func softDeleteDocument(requestID, documentID, deletedBy uuid.UUID) error {
	_, err := database.DB.Exec(`
		UPDATE documents
		SET deleted = true, deleted_at = now(), deleted_by = $1
		WHERE id = $2 AND deleted = false`,
		deletedBy, documentID,
	)
	if err != nil {
		logger.Errorf(requestID, deletedBy, "softDeleteDocument: db error for doc %s: %v", documentID, err)
	}
	return err
}
