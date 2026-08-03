package models

import "time"

// AppNotify is a channel notification (ecommerce | social | task).
type AppNotify struct {
	ID        int64     `json:"id"`
	Channel   string    `json:"channel,omitempty"`
	Title     string    `json:"title"`
	Body      string    `json:"body"`
	Unread    bool      `json:"unread"`
	CreatedAt time.Time `json:"createdAt"`
}

// NotifyListQuery filters notify lists.
type NotifyListQuery struct {
	ListQuery
	OnlyUnread      bool
	DescendingOrder bool
}

// SocialPost is a public social post.
type SocialPost struct {
	ID        int64     `json:"id"`
	IDLang    *string   `json:"idLang,omitempty"`
	Title     string    `json:"title"`
	Body      string    `json:"body"`
	CreatedAt time.Time `json:"createdAt"`
}

// EmailItem maps emails table.
type EmailItem struct {
	ID        int64     `json:"id"`
	Hashcode  *string   `json:"hashcode,omitempty"`
	Subject   string    `json:"subject"`
	Body      string    `json:"body"`
	Sender    *string   `json:"sender,omitempty"`
	Recipient *string   `json:"recipient,omitempty"`
	CreatedAt time.Time `json:"createdAt"`
}

// GlobalNotification maps global_notifications.
type GlobalNotification struct {
	ID            int64     `json:"id"`
	Title         string    `json:"title"`
	Body          string    `json:"body"`
	SourceCompany *string   `json:"sourceCompany,omitempty"`
	CreatedAt     time.Time `json:"createdAt"`
}

// BindCommerceRequest body for stock update plugin.
type BindCommerceRequest struct {
	IDMaterials []string `json:"idMaterials"`
}

// IntegrationAck is a generic stub response.
type IntegrationAck struct {
	OK      bool   `json:"ok"`
	Message string `json:"message"`
}
