package models

import (
	"time"

	"github.com/google/uuid"
)

// User maps the users table.
type User struct {
	ID           uuid.UUID `json:"id"`
	Username     string    `json:"username"`
	Email        *string   `json:"email,omitempty"`
	PasswordHash string    `json:"-"`
	IsActive     bool      `json:"is_active"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// UserPublic is returned from profile endpoints (no password).
type UserPublic struct {
	ID       uuid.UUID `json:"id"`
	Username string    `json:"username"`
	Email    *string   `json:"email,omitempty"`
}

// UserSession maps user_sessions (refresh token storage).
type UserSession struct {
	ID               uuid.UUID
	UserID           uuid.UUID
	RefreshTokenHash string
	ExpiresAt        time.Time
	RevokedAt        *time.Time
	CreatedAt        time.Time
}

// RegisterRequest body for POST /api/auth/register.
type RegisterRequest struct {
	Username string  `json:"username"`
	Password string  `json:"password"`
	Email    *string `json:"email,omitempty"`
}

// LoginRequest body for POST /api/auth/login.
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// RefreshRequest body for POST /api/auth/refresh (OAuth-style).
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
	GrantType    string `json:"grant_type,omitempty"`
}

// TokenResponse matches Giobby OAuth token JSON (access + refresh).
type TokenResponse struct {
	AccessToken      string `json:"access_token"`
	ExpiresIn        int64  `json:"expires_in"`
	RefreshExpiresIn int64  `json:"refresh_expires_in"`
	RefreshToken     string `json:"refresh_token"`
	TokenType        string `json:"token_type"`
}

// MeResponse body for GET /api/auth/me.
type MeResponse struct {
	User UserPublic `json:"user"`
}
