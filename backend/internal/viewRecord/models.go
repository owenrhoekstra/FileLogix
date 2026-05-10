package viewRecord

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

type fetchParams struct {
	SortBy string
	Limit  string
	Offset string

	Name  string
	Types []string

	DocDateFrom   string
	DocDateTo     string
	FiledDateFrom string
	FiledDateTo   string
}

type record struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Types       []string `json:"types"`
	Sensitive   bool     `json:"sensitive"`
	DateOfDoc   string   `json:"dateOfDoc"`
	DateFiled   string   `json:"dateFiled"`
	Location    string   `json:"location"`
	Description string   `json:"description,omitempty"`
	Thumbnail   string   `json:"thumbnail,omitempty"`
}

type queryParams struct {
	SortCol string
	Limit   int
	Offset  int

	Name  string
	Types []string

	DocDateFrom   *time.Time
	DocDateTo     *time.Time
	FiledDateFrom *time.Time
	FiledDateTo   *time.Time

	IncludeSensitive bool
	ShowDeleted      bool
}

type documentRow struct {
	ID                 uuid.UUID
	Name               string
	Types              []byte
	DateOfDoc          string
	DateFiled          string
	Sensitive          bool
	Deleted            bool
	CabinetName        sql.NullString
	CabinetDescription sql.NullString
}

type fileRow struct {
	Path       string
	PageNumber int
}

type documentDetail struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Types       []string `json:"types"`
	DateOfDoc   string   `json:"dateOfDoc"`
	DateFiled   string   `json:"dateFiled"`
	Location    string   `json:"location"`
	Description string   `json:"description"`
	Sensitive   bool     `json:"sensitive"`
	Pages       []string `json:"pages"`
	Deleted     bool     `json:"deleted"`
}
