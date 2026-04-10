package authentication

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"time"

	"file-tracker-backend/database"

	"github.com/go-webauthn/webauthn/webauthn"
)

type DBSessionStore struct{}

func newDBSessionStore() *DBSessionStore {
	return &DBSessionStore{}
}

// generate UUID-like session ID (fast, no dependency bloat)
func newSessionID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

// SAVE session
func (s *DBSessionStore) set(email string, data *webauthn.SessionData) string {
	sessionID := newSessionID()

	encoded, _ := json.Marshal(data)

	_, _ = database.DB.Exec(`
		INSERT INTO webauthn_sessions (id, email, challenge, data, expires_at)
		VALUES ($1, $2, $3, $4, $5)
	`,
		sessionID,
		email,
		data.Challenge,
		encoded,
		time.Now().Add(3*time.Minute),
	)

	return sessionID
}

// GET session
func (s *DBSessionStore) get(sessionID string) (*webauthn.SessionData, bool) {
	row := database.DB.QueryRow(`
		SELECT data
		FROM webauthn_sessions
		WHERE id = $1
		AND expires_at > NOW()
	`, sessionID)

	var raw []byte
	if err := row.Scan(&raw); err != nil {
		return nil, false
	}

	var data webauthn.SessionData
	if err := json.Unmarshal(raw, &data); err != nil {
		return nil, false
	}

	return &data, true
}

// optional cleanup
func cleanupExpiredSessions() {
	_, _ = database.DB.Exec(`
		DELETE FROM webauthn_sessions
		WHERE expires_at < NOW()
	`)
}

var regSessions = newDBSessionStore()
var loginSessions = newDBSessionStore()
