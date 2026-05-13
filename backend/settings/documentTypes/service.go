package documentTypes

import (
	"context"
	"strings"
)

func addDocumentTypeService(ctx context.Context, label string) error {
	value := documentTypeEntry{
		DocumentLabel:      label,
		DocumentLabelValue: sanitizeLabelValue(label),
	}
	return insertDocumentType(ctx, value)
}

func sanitizeLabelValue(label string) string {
	return strings.ToLower(strings.ReplaceAll(strings.TrimSpace(label), " ", ""))
}
