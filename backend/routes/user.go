package routes

import (
	"FileLogix/database"
	"FileLogix/middleware"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/google/uuid"
)

func UserRoutes() http.Handler {
	mux := http.NewServeMux()

	mux.Handle("/setup", middleware.RequireAuth(setupHandler))

	return mux
}

func setupHandler(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(uuid.UUID)

	switch r.Method {
	case http.MethodPost:
		var req struct {
			FirstName string `json:"first_name"`
			LastName  string `json:"last_name"`
			Phone     string `json:"phone"`
			Initials  string `json:"initials"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request", http.StatusBadRequest)
			return
		}

		req.FirstName = strings.TrimSpace(req.FirstName)
		req.LastName = strings.TrimSpace(req.LastName)
		req.Phone = strings.TrimSpace(req.Phone)
		req.Initials = strings.TrimSpace(req.Initials)

		if req.FirstName == "" || req.LastName == "" || req.Phone == "" || req.Initials == "" {
			http.Error(w, "all fields are required", http.StatusBadRequest)
			return
		}

		_, err := database.DB.Exec(`
			UPDATE users
			SET first_name        = $1,
			    last_name         = $2,
			    phone             = $3,
			    initials          = $4,
			    metadata_complete = TRUE
			WHERE id = $5
		`, req.FirstName, req.LastName, req.Phone, req.Initials, userID)
		if err != nil {
			http.Error(w, "failed to save profile", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})

	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}
