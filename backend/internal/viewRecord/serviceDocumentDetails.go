package viewRecord

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"FileLogix/internal"
	"FileLogix/middleware"
	"FileLogix/utilities/logger"

	"github.com/google/uuid"
)

func buildDocumentDetail(ctx context.Context, id uuid.UUID, canDelete bool) (*documentDetail, error) {
	requestID := ctx.Value(middleware.RequestIDKey).(uuid.UUID)
	userID := ctx.Value(middleware.UserIDKey).(uuid.UUID)

	doc, err := queryDocument(ctx, id, canDelete)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		logger.Errorf(requestID, userID, "buildDocumentDetail: queryDocument error: %v", err)
		return nil, fmt.Errorf("query document: %w", err)
	}

	var types []string
	if err := json.Unmarshal(doc.Types, &types); err != nil {
		logger.Errorf(requestID, userID, "buildDocumentDetail: unmarshal types error: %v", err)
		return nil, fmt.Errorf("unmarshal types: %w", err)
	}

	files, err := queryFiles(ctx, id)
	if err != nil {
		logger.Errorf(requestID, userID, "buildDocumentDetail: queryFiles error: %v", err)
		return nil, fmt.Errorf("query files: %w", err)
	}

	pages := make([]string, len(files))
	for i, f := range files {
		rel := strings.TrimPrefix(f.Path, internal.StorageRoot)
		pages[i] = internal.URLPrefix + rel
	}

	location := ""
	if doc.CabinetName.Valid {
		location = doc.CabinetName.String
	}

	description := ""
	if doc.CabinetDescription.Valid {
		description = doc.CabinetDescription.String
	}

	return &documentDetail{
		ID:          doc.ID.String(),
		Name:        doc.Name,
		Types:       types,
		DateOfDoc:   doc.DateOfDoc,
		DateFiled:   doc.DateFiled,
		Location:    location,
		Description: description,
		Sensitive:   doc.Sensitive,
		Deleted:     doc.Deleted,
		Pages:       pages,
	}, nil
}
