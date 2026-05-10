package viewRecord

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"FileLogix/database"
	"FileLogix/internal"
	"FileLogix/middleware"
	"FileLogix/utilities/logger"

	"github.com/google/uuid"
)

func fetch(ctx context.Context, qp queryParams) ([]record, error) {
	requestID := ctx.Value(middleware.RequestIDKey).(uuid.UUID)
	userID := ctx.Value(middleware.UserIDKey).(uuid.UUID)

	query := `
        SELECT d.id, d.name, d.types, d.sensitive, d.date_of_doc, d.date_filed,
               COALESCE(c.name, ''), COALESCE(c.description, ''),
               COALESCE((
                   SELECT f.path FROM files f
                   WHERE f.document_id = d.id
                   ORDER BY f.page_number ASC
                   LIMIT 1
               ), '')
        FROM documents d
        LEFT JOIN cabinets c ON d.cabinet_id = c.id
        WHERE 1=1
    `

	if qp.ShowDeleted {
		query += " AND d.deleted = true"
	} else {
		query += " AND d.deleted = false"
	}

	args := []any{}
	argIdx := 1

	if qp.Name != "" {
		query += fmt.Sprintf(" AND d.name ILIKE $%d", argIdx)
		args = append(args, "%"+qp.Name+"%")
		argIdx++
	}

	if len(qp.Types) > 0 {
		jsonTypes := make([]string, len(qp.Types))
		for i, t := range qp.Types {
			jsonTypes[i] = fmt.Sprintf(`"%s"`, t)
		}
		query += fmt.Sprintf(" AND d.types @> $%d::jsonb", argIdx)
		args = append(args, fmt.Sprintf("[%s]", strings.Join(jsonTypes, ",")))
		argIdx++
	}

	if qp.DocDateFrom != nil {
		query += fmt.Sprintf(" AND d.date_of_doc >= $%d", argIdx)
		args = append(args, *qp.DocDateFrom)
		argIdx++
	}
	if qp.DocDateTo != nil {
		query += fmt.Sprintf(" AND d.date_of_doc <= $%d", argIdx)
		args = append(args, *qp.DocDateTo)
		argIdx++
	}

	if qp.FiledDateFrom != nil {
		query += fmt.Sprintf(" AND d.date_filed >= $%d", argIdx)
		args = append(args, *qp.FiledDateFrom)
		argIdx++
	}
	if qp.FiledDateTo != nil {
		query += fmt.Sprintf(" AND d.date_filed <= $%d", argIdx)
		args = append(args, *qp.FiledDateTo)
		argIdx++
	}

	if !qp.IncludeSensitive {
		query += " AND d.sensitive = false"
	}

	query += fmt.Sprintf(" ORDER BY %s DESC LIMIT $%d OFFSET $%d", qp.SortCol, argIdx, argIdx+1)
	args = append(args, qp.Limit, qp.Offset)

	rows, err := database.DB.QueryContext(ctx, query, args...)
	if err != nil {
		logger.Errorf(requestID, userID, "fetch: query error: %v", err)
		return nil, fmt.Errorf("querying records: %w", err)
	}
	defer rows.Close()

	var results []record
	for rows.Next() {
		var rec record
		var typesJSON []byte
		var docDate time.Time
		var filedDate time.Time
		var thumbnailPath string

		if err := rows.Scan(
			&rec.ID, &rec.Name, &typesJSON, &rec.Sensitive,
			&docDate, &filedDate, &rec.Location, &rec.Description,
			&thumbnailPath,
		); err != nil {
			logger.Errorf(requestID, userID, "fetch: scan error: %v", err)
			return nil, fmt.Errorf("scanning record: %w", err)
		}

		if err := json.Unmarshal(typesJSON, &rec.Types); err != nil {
			logger.Errorf(requestID, userID, "fetch: unmarshal types error: %v", err)
			return nil, fmt.Errorf("unmarshaling types: %w", err)
		}

		rec.DateOfDoc = docDate.Format("2006-01-02")
		rec.DateFiled = filedDate.Format(time.RFC3339)

		if thumbnailPath != "" {
			rel := strings.TrimPrefix(thumbnailPath, internal.StorageRoot)
			rec.Thumbnail = internal.URLPrefix + rel
		}

		results = append(results, rec)
	}

	if err := rows.Err(); err != nil {
		logger.Errorf(requestID, userID, "fetch: rows iteration error: %v", err)
		return nil, fmt.Errorf("iterating rows: %w", err)
	}

	return results, nil
}
