package middleware

import (
	"net/http"
	"strings"

	"github.com/rs/cors"
)

// CorsMiddleware sets CORS with credentials; allowedOrigins must be explicit (not *) when AllowCredentials is true.
func CorsMiddleware(allowedOrigins []string) func(http.Handler) http.Handler {
	if len(allowedOrigins) == 0 {
		allowedOrigins = []string{"http://localhost:5173"}
	}
	return func(next http.Handler) http.Handler {
		c := cors.New(cors.Options{
			AllowedOrigins:   allowedOrigins,
			AllowedMethods:   []string{http.MethodGet, http.MethodPost, http.MethodPatch, http.MethodPut, http.MethodDelete, http.MethodOptions},
			AllowedHeaders:   []string{"*"},
			AllowCredentials: true,
		})
		return c.Handler(next)
	}
}

// APIKeyMiddleware validates API key from X-API-Key header
// Add this middleware to routes that need API key authentication
func APIKeyMiddleware(validAPIKey string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			apiKey := r.Header.Get("X-API-Key")
			if apiKey == "" {
				apiKey = strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
			}

			if apiKey == "" || apiKey != validAPIKey {
				http.Error(w, "Invalid or missing API key", http.StatusUnauthorized)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
