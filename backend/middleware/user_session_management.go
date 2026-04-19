package middleware

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"

	"FileLogix/database"
)

type Session struct {
	ID          string
	UserID      []byte
	ExpiresAt   time.Time
	LastSeen    time.Time
	RoleName    string
	Permissions map[string]bool
}

const (
	sessionTTL  = 6 * time.Hour
	idleTimeout = 15 * time.Minute
)

var (
	ErrSessionExpired = errors.New("expired")
	ErrSessionIdle    = errors.New("idle timeout")
)

func newSessionToken() string {
	b := make([]byte, 32)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

func CreateSession(userID []byte, roleName string, permissions map[string]bool) (string, error) {
	token := newSessionToken()

	s := Session{
		ID:          token,
		UserID:      userID,
		ExpiresAt:   time.Now().Add(sessionTTL),
		LastSeen:    time.Now(),
		RoleName:    roleName,
		Permissions: permissions,
	}

	data, err := json.Marshal(s)
	if err != nil {
		return "", err
	}

	err = database.RDB.Set(context.Background(), "session:"+token, data, sessionTTL).Err()
	if err != nil {
		return "", err
	}

	return token, nil
}

func GetSession(token string) (*Session, error) {
	data, err := database.RDB.Get(context.Background(), "session:"+token).Bytes()
	if err != nil {
		return nil, err
	}

	var s Session
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, err
	}

	if time.Now().After(s.ExpiresAt) {
		if err := DeleteSession(token); err != nil {
			log.Printf("failed to delete session %s: %v", token, err)
		}
		return nil, ErrSessionExpired
	}

	if time.Since(s.LastSeen) > idleTimeout {
		if err := DeleteSession(token); err != nil {
			log.Printf("failed to delete session %s: %v", token, err)
		}
		return nil, ErrSessionIdle
	}

	return &s, nil
}

func DeleteSession(token string) error {
	return database.RDB.Del(context.Background(), "session:"+token).Err()
}

func SetSessionCookie(w http.ResponseWriter, token string) {
	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		MaxAge:   int(sessionTTL.Seconds()),
		SameSite: http.SameSiteLaxMode,
	})
}

func GetSessionFromRequest(r *http.Request) (string, error) {
	cookie, err := r.Cookie("session")
	if err != nil {
		return "", err
	}
	return cookie.Value, nil
}

func TouchSession(s *Session) error {
	s.LastSeen = time.Now()

	data, err := json.Marshal(s)
	if err != nil {
		return err
	}

	key := "session:" + s.ID

	exists, err := database.RDB.Exists(context.Background(), key).Result()
	if err != nil || exists == 0 {
		return nil
	}

	remaining := time.Until(s.ExpiresAt)
	if remaining <= 0 {
		_ = DeleteSession(s.ID)
		return nil
	}

	return database.RDB.Set(context.Background(), key, data, remaining).Err()
}
