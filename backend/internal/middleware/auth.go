package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"dnsc_microservice/internal/models"
)

type ctxKey int

const userCtxKey ctxKey = 1

// UserFromContext returns the authenticated user set by AuthMiddleware.
func UserFromContext(ctx context.Context) (*models.User, bool) {
	u, ok := ctx.Value(userCtxKey).(*models.User)
	return u, ok
}

// AccessTokenValidator validates Bearer access tokens.
type AccessTokenValidator interface {
	ValidateAccessToken(ctx context.Context, accessToken string) (*models.User, error)
}

func isPublicPath(path, method string) bool {
	if method == http.MethodOptions {
		return true
	}
	if path == "/healthz" || path == "/health" {
		return method == http.MethodGet
	}
	public := map[string][]string{
		"/api/auth/register": {http.MethodPost},
		"/api/auth/login":    {http.MethodPost},
		"/api/auth/refresh":  {http.MethodPost},
	}
	if methods, ok := public[path]; ok {
		for _, m := range methods {
			if m == method {
				return true
			}
		}
	}
	return false
}

// AuthMiddleware enforces Authorization: Bearer <access_token> for protected routes.
func AuthMiddleware(validator AccessTokenValidator) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if isPublicPath(r.URL.Path, r.Method) {
				next.ServeHTTP(w, r)
				return
			}

			token, ok := bearerToken(r)
			if !ok {
				httpUnauthorized(w)
				return
			}
			user, err := validator.ValidateAccessToken(r.Context(), token)
			if err != nil || user == nil {
				httpUnauthorized(w)
				return
			}
			ctx := context.WithValue(r.Context(), userCtxKey, user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func bearerToken(r *http.Request) (string, bool) {
	h := strings.TrimSpace(r.Header.Get("Authorization"))
	if !strings.HasPrefix(h, "Bearer ") {
		return "", false
	}
	t := strings.TrimSpace(strings.TrimPrefix(h, "Bearer "))
	return t, t != ""
}

func httpUnauthorized(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	_ = json.NewEncoder(w).Encode(map[string]string{"code": "UNAUTHORIZED", "message": "Unauthorized"})
}
