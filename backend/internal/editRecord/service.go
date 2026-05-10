package editRecord

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

func editRecord(ctx context.Context, id uuid.UUID, name string, sensitive bool, types []string, dateOfDoc time.Time) error {
	// Validate name
	if name == "" {
		return fmt.Errorf("name cannot be empty")
	}
	if len(name) > 50 {
		return fmt.Errorf("name cannot exceed 50 characters")
	}

	// Validate types
	if len(types) > 3 {
		return fmt.Errorf("cannot have more than 3 types")
	}

	// Marshal types to JSONB
	typesJSON, err := json.Marshal(types)
	if err != nil {
		return fmt.Errorf("marshaling types: %w", err)
	}

	return updateRecord(ctx, id, name, sensitive, typesJSON, dateOfDoc)
}

func restoreRecord(ctx context.Context, id uuid.UUID) error {
	return undeleteRecord(ctx, id)
}
