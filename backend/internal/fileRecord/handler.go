package fileRecord

import (
	"FileLogix/database"
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
)

func File(w http.ResponseWriter, r *http.Request) {
	var body assignCabinetRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if body.DocumentID == uuid.Nil || body.CabinetID == uuid.Nil {
		http.Error(w, "missing document_id or cabinet_id", http.StatusBadRequest)
		return
	}

	const q = `
		UPDATE documents
		SET    cabinet_id = $1
		WHERE  id = $2
		  AND  deleted_at IS NULL
	`
	result, err := database.DB.Exec(q, body.CabinetID, body.DocumentID)
	if err != nil {
		http.Error(w, "failed to assign cabinet", http.StatusInternalServerError)
		return
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		http.Error(w, "document not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func CabinetMeta(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "missing cabinet id", http.StatusBadRequest)
		return
	}

	var meta CabinetMetaData
	const q = `
		SELECT id, name, description
		FROM   cabinets
		WHERE  id = $1
		  AND  deleted_at IS NULL
		  AND  in_use = TRUE
	`
	err := database.DB.QueryRow(q, id).Scan(&meta.ID, &meta.Name, &meta.Description)
	if err != nil {
		http.Error(w, "cabinet not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(meta)
}
