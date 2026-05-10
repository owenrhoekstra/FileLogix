package fileRecord

import "github.com/google/uuid"

type assignCabinetRequest struct {
	DocumentID uuid.UUID `json:"documentId"`
	CabinetID  uuid.UUID `json:"cabinetId"`
}

type CabinetMetaData struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}
