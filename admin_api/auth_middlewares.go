package adminapi

import (
	"net/http"
	"strings"
)

func AuthenticationMiddleware(next http.HandlerFunc) http.HandlerFunc {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        authHeader := r.Header.Get("Authorization")
        
        // Check for "Bearer <token>"
        parts := strings.Split(authHeader, " ")
        if len(parts) != 2 || parts[0] != "Bearer" {
            http.Error(w, "Unauthorized: Malformed header", http.StatusUnauthorized)
            return
        }

        // Validating token
        if !isTokenValid(parts[1]) {
            http.Error(w, "Unauthorized: Invalid or expired token", http.StatusUnauthorized)
            return
        }

        next(w, r)
    })
}