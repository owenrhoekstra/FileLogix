package authentication

import (
	"encoding/json"
	"log"
	"net/http"
)

func CheckEmailHandler(w http.ResponseWriter, r *http.Request) {
	var req EmailRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	// Not allowed — return same shape as allowed but hasPasskey false
	// This prevents email enumeration of the whitelist
	if !isAllowed(req.Email) {
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"hasPasskey": false,
		})
		return
	}

	u, err := getUser(req.Email)
	if err != nil {
		log.Println("CheckEmailHandler: getUser error for", req.Email, ":", err)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"hasPasskey": false,
		})
		return
	}

	hasPasskey := false
	if u != nil {
		creds, err := getCredentialsByUserID(u.ID)
		if err == nil && len(creds) > 0 {
			hasPasskey = true
			log.Println("CheckEmailHandler:", req.Email, "has", len(creds), "credential(s)")
		}
	}

	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"hasPasskey": hasPasskey,
	})
}
