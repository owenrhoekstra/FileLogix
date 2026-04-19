package authentication

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"time"

	"FileLogix/database"

	"github.com/go-webauthn/webauthn/webauthn"
)

const webAuthnSessionTTL = 3 * time.Minute

type RedisSessionStore struct{}

func newRedisSessionStore() *RedisSessionStore {
	return &RedisSessionStore{}
}

func newSessionID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

func (s *RedisSessionStore) set(email string, data *webauthn.SessionData) string {
	sessionID := newSessionID()

	payload := map[string]interface{}{
		"email": email,
		"data":  data,
	}
	encoded, _ := json.Marshal(payload)

	_ = database.RDB.Set(context.Background(), "webauthn:"+sessionID, encoded, webAuthnSessionTTL).Err()

	return sessionID
}

func (s *RedisSessionStore) get(sessionID string) (*webauthn.SessionData, bool) {
	raw, err := database.RDB.Get(context.Background(), "webauthn:"+sessionID).Bytes()
	if err != nil {
		return nil, false
	}

	var payload struct {
		Email string               `json:"email"`
		Data  webauthn.SessionData `json:"data"`
	}
	if err := json.Unmarshal(raw, &payload); err != nil {
		return nil, false
	}

	return &payload.Data, true
}

var regSessions = newRedisSessionStore()
var loginSessions = newRedisSessionStore()
