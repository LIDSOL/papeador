package auth

import (
	"net/http"

	"lidsol.org/papeador/security"
)

func IsAuthenticated(r *http.Request) bool {
	cookieJWT, err := r.Cookie("jwt")
	cookieUsername, err := r.Cookie("username")

	ok, err := security.ValidateJWT(cookieJWT.Value, cookieUsername.Value)
	if err != nil {
		return false
	}
	if !ok {
		return false
	}

	return true
}

func RequireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !IsAuthenticated(r) {
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}
		next(w, r)
	}
}
