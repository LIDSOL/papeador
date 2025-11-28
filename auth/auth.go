package auth

import (
	"net/http"

	"lidsol.org/papeador/security"
)

func IsAuthenticated(r *http.Request) bool {
	cookieJWT, err := r.Cookie("jwt")
	if err != nil {
		return false
	}
	cookieUsername, err := r.Cookie("username")
	if err != nil {
		return false
	}

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
			w.Header().Set("HX-Redirect", "/login")
			w.WriteHeader(http.StatusOK)
			return
		}
		next(w, r)
	}
}
