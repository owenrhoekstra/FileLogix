package middleware

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"time"

	"file-tracker-backend/database"
)

type Session struct {
	ID        string
	UserID    []byte
	ExpiresAt time.Time
}

const sessionTTL = 6 * time.Hour

func newSessionToken() string {
	b := make([]byte, 32)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

func CreateSession(userID []byte) (string, error) {
	token := newSessionToken()

	s := Session{
		ID:        token,
		UserID:    userID,
		ExpiresAt: time.Now().Add(sessionTTL),
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

	return &s, nil
}

func DeleteSession(token string) {
	_ = database.RDB.Del(context.Background(), "session:"+token).Err()
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
