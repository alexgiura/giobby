package models

import (
	"time"

	"github.com/google/uuid"
)

// LoggedUserInfo is returned by GET /loggeduser.
type LoggedUserInfo struct {
	ID       uuid.UUID `json:"id"`
	Username string    `json:"username"`
	Email    *string   `json:"email,omitempty"`
	Language string    `json:"language"`
	ImageURL *string   `json:"imageUrl,omitempty"`
	IsActive bool      `json:"isActive"`
}

// ChangeLanguageRequest body for POST /loggeduser/changelanguage.
type ChangeLanguageRequest struct {
	Language string `json:"language"`
	Type     string `json:"type,omitempty"`
}

// DeviceToken maps logged_user_device_tokens.
type DeviceToken struct {
	ID          int64     `json:"id"`
	OS          *string   `json:"os,omitempty"`
	DeviceToken string    `json:"deviceToken"`
	CreatedAt   time.Time `json:"createdAt,omitempty"`
}
