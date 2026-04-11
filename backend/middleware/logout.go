package middleware

import "net/http"

func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	token, err := GetSessionFromRequest(r)
	if err == nil {
		DeleteSession(token)
	}
	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   true,
	})
	w.WriteHeader(http.StatusOK)
}
