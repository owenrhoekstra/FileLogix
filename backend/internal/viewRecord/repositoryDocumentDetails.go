package viewRecord

import (
	"context"

	"FileLogix/database"
	"FileLogix/middleware"
	"FileLogix/utilities/logger"

	"github.com/google/uuid"
)

func queryDocument(ctx context.Context, id uuid.UUID, canDelete bool) (*documentRow, error) {
	requestID := ctx.Value(middleware.RequestIDKey).(uuid.UUID)
	userID := ctx.Value(middleware.UserIDKey).(uuid.UUID)

	deletedFilter := ""
	if !canDelete {
		deletedFilter = " AND d.deleted = false"
	}

	row := &documentRow{}
	err := database.DB.QueryRowContext(ctx, `
        SELECT d.id, d.name, d.types, d.date_of_doc, d.date_filed, d.sensitive, d.deleted,
               c.name, c.description
        FROM documents d
        LEFT JOIN cabinets c ON c.id = d.cabinet_id
        WHERE d.id = $1`+deletedFilter, id).Scan(
		&row.ID, &row.Name, &row.Types, &row.DateOfDoc, &row.DateFiled,
		&row.Sensitive, &row.Deleted, &row.CabinetName, &row.CabinetDescription,
	)
	if err != nil {
		logger.Errorf(requestID, userID, "queryDocument: scan error: %v", err)
		return nil, err
	}
	return row, nil
}

func queryFiles(ctx context.Context, documentID uuid.UUID) ([]fileRow, error) {
	requestID := ctx.Value(middleware.RequestIDKey).(uuid.UUID)
	userID := ctx.Value(middleware.UserIDKey).(uuid.UUID)

	rows, err := database.DB.QueryContext(ctx, `
        SELECT path, page_number
        FROM files
        WHERE document_id = $1
        ORDER BY page_number ASC
    `, documentID)
	if err != nil {
		logger.Errorf(requestID, userID, "queryFiles: query error: %v", err)
		return nil, err
	}
	defer rows.Close()

	var files []fileRow
	for rows.Next() {
		var f fileRow
		if err := rows.Scan(&f.Path, &f.PageNumber); err != nil {
			logger.Errorf(requestID, userID, "queryFiles: scan error: %v", err)
			return nil, err
		}
		files = append(files, f)
	}
	return files, nil
}
