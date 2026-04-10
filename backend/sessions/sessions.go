package sessions

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"time"

	"file-tracker-backend/database"
)

type Session struct {
	ID        string
	UserID    []byte
	ExpiresAt time.Time
}

// generate secure token
func newSessionToken() string {
	b := make([]byte, 32)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

// CREATE session
func CreateSession(userID []byte) (string, error) {
	token := newSessionToken()
	expires := time.Now().Add(24 * time.Hour) // 1 day login

	_, err := database.DB.Exec(`
		INSERT INTO sessions (token, user_id, expires_at)
		VALUES ($1, $2, $3)
	`, token, userID, expires)

	if err != nil {
		return "", err
	}

	return token, nil
}

// VALIDATE session
func GetSession(token string) (*Session, error) {
	row := database.DB.QueryRow(`
		SELECT user_id, expires_at
		FROM sessions
		WHERE token = $1
		AND expires_at > NOW()
	`, token)

	var s Session
	s.ID = token

	err := row.Scan(&s.UserID, &s.ExpiresAt)
	if err != nil {
		return nil, err
	}

	return &s, nil
}

// DELETE session (logout)
func DeleteSession(token string) {
	_, _ = database.DB.Exec(`
		DELETE FROM sessions WHERE token = $1
	`, token)
}

// SET COOKIE
func SetSessionCookie(w http.ResponseWriter, token string) {
	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   false, // true in prod
		MaxAge:   86400, // 1 day
	})
}

// GET COOKIE
func GetSessionFromRequest(r *http.Request) (string, error) {
	cookie, err := r.Cookie("session")
	if err != nil {
		return "", err
	}
	return cookie.Value, nil
}
