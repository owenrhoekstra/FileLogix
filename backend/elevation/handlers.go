package elevation

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"FileLogix/database"
	"FileLogix/middleware"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/lib/pq"
)

var WebAuthn *webauthn.WebAuthn

const challengeTTL = 3 * time.Minute

type challengeStore struct{}

func (s *challengeStore) set(sessionToken string, data *webauthn.SessionData) error {
	encoded, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return database.RDB.Set(context.Background(), "elevation_challenge:"+sessionToken, encoded, challengeTTL).Err()
}

func (s *challengeStore) get(sessionToken string) (*webauthn.SessionData, bool) {
	raw, err := database.RDB.Get(context.Background(), "elevation_challenge:"+sessionToken).Bytes()
	if err != nil {
		return nil, false
	}
	var data webauthn.SessionData
	if err := json.Unmarshal(raw, &data); err != nil {
		return nil, false
	}
	return &data, true
}

func (s *challengeStore) delete(sessionToken string) {
	_ = database.RDB.Del(context.Background(), "elevation_challenge:"+sessionToken).Err()
}

var challenges = &challengeStore{}

type elevationUser struct {
	id          []byte
	credentials []webauthn.Credential
}

func (u *elevationUser) WebAuthnID() []byte                         { return u.id }
func (u *elevationUser) WebAuthnName() string                       { return "" }
func (u *elevationUser) WebAuthnDisplayName() string                { return "" }
func (u *elevationUser) WebAuthnCredentials() []webauthn.Credential { return u.credentials }

func loadUserCredentials(userID []byte) (*elevationUser, error) {
	rows, err := database.DB.Query(`
		SELECT credential_id, public_key, attestation_type, transports,
		       sign_count, backup_eligible, backup_state
		FROM credentials
		WHERE user_id = $1
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var creds []webauthn.Credential
	for rows.Next() {
		var cred webauthn.Credential
		var transports pq.StringArray

		if err := rows.Scan(
			&cred.ID,
			&cred.PublicKey,
			&cred.AttestationType,
			&transports,
			&cred.Authenticator.SignCount,
			&cred.Flags.BackupEligible,
			&cred.Flags.BackupState,
		); err != nil {
			log.Println("loadUserCredentials scan error:", err)
			continue
		}

		for _, t := range transports {
			cred.Transport = append(cred.Transport, protocol.AuthenticatorTransport(t))
		}

		creds = append(creds, cred)
	}

	return &elevationUser{id: userID, credentials: creds}, nil
}

func ChallengeHandler(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).([]byte)
	permissions := r.Context().Value(middleware.PermissionsKey).(map[string]bool)

	var req struct {
		Type ElevationType `json:"type"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || (req.Type != ActionElevation && req.Type != ViewElevation) {
		http.Error(w, "invalid elevation type", http.StatusBadRequest)
		return
	}

	switch req.Type {
	case ActionElevation:
		if !permissions["can_action_elevate"] {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
	case ViewElevation:
		if !permissions["can_view_elevate"] {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
	}

	u, err := loadUserCredentials(userID)
	if err != nil || len(u.credentials) == 0 {
		log.Println("loadUserCredentials error:", err)
		http.Error(w, "no credentials found", http.StatusInternalServerError)
		return
	}

	options, sessionData, err := WebAuthn.BeginLogin(u)
	if err != nil {
		log.Println("elevation BeginLogin error:", err)
		http.Error(w, "failed to begin challenge", http.StatusInternalServerError)
		return
	}

	token, _ := middleware.GetSessionFromRequest(r)
	if err := challenges.set(token, sessionData); err != nil {
		http.Error(w, "failed to store challenge", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"options": options,
		"type":    req.Type,
	})
}

func VerifyHandler(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).([]byte)

	elevType := ElevationType(r.Header.Get("X-Elevation-Type"))
	if elevType != ActionElevation && elevType != ViewElevation {
		http.Error(w, "invalid elevation type", http.StatusBadRequest)
		return
	}

	token, _ := middleware.GetSessionFromRequest(r)

	sessionData, ok := challenges.get(token)
	if !ok {
		http.Error(w, "no pending challenge", http.StatusBadRequest)
		return
	}
	challenges.delete(token)

	u, err := loadUserCredentials(userID)
	if err != nil {
		http.Error(w, "failed to load credentials", http.StatusInternalServerError)
		return
	}

	_, err = WebAuthn.FinishLogin(u, *sessionData, r)
	if err != nil {
		log.Println("elevation verify error:", err)
		http.Error(w, "authentication failed", http.StatusUnauthorized)
		return
	}

	if err := SetElevation(token, elevType); err != nil {
		http.Error(w, "failed to set elevation", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}
