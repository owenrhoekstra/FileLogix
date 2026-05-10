package viewRecord

import (
	"context"
	"strconv"
	"time"
)

func fetchRecords(ctx context.Context, p fetchParams, canViewSensitive bool, canDelete bool) ([]record, error) {
	showDeleted := p.SortBy == "deleted" && canDelete

	allowed := map[string]string{
		"added":    "d.date_filed",
		"modified": "d.last_modified",
		"deleted":  "d.deleted_at",
	}

	col, ok := allowed[p.SortBy]
	if !ok {
		col = "d.date_filed"
	}

	// ---- limit ----
	limit := 25
	if p.Limit != "" {
		if parsed, err := strconv.Atoi(p.Limit); err == nil {
			limit = parsed
		}
	}
	if limit > 100 {
		limit = 100
	}
	if limit <= 0 {
		limit = 25
	}

	// ---- offset ----
	offset := 0
	if p.Offset != "" {
		if parsed, err := strconv.Atoi(p.Offset); err == nil {
			offset = parsed
		}
		if offset < 0 {
			offset = 0
		}
	}

	// ---- date parsing ----
	parseDate := func(s string) *time.Time {
		if s == "" {
			return nil
		}
		t, err := time.Parse(time.RFC3339, s)
		if err != nil {
			return nil
		}
		return &t
	}

	qp := queryParams{
		SortCol: col,
		Limit:   limit,
		Offset:  offset,

		Name:  p.Name,
		Types: p.Types,

		DocDateFrom:   parseDate(p.DocDateFrom),
		DocDateTo:     parseDate(p.DocDateTo),
		FiledDateFrom: parseDate(p.FiledDateFrom),
		FiledDateTo:   parseDate(p.FiledDateTo),

		IncludeSensitive: canViewSensitive,
		ShowDeleted:      showDeleted,
	}

	return fetch(ctx, qp)
}
