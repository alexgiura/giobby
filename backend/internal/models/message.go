package models

import "time"

// Message maps messages table (Giobby MessageApi).
type Message struct {
	ID               int64     `json:"id"`
	IDGroup          *int64    `json:"idGroup,omitempty"`
	Unread           bool      `json:"unread"`
	SentDate         time.Time `json:"sentDate"`
	Body             string    `json:"body"`
	UserFrom         *string   `json:"userFrom,omitempty"`
	UserFromFullName *string   `json:"userFromFullName,omitempty"`
	UserToFullName   *string   `json:"userToFullName,omitempty"`
	IDTransaction    *int32    `json:"idTransaction,omitempty"`
	IDChannel        *string   `json:"idChannel,omitempty"`
	CompanyTo        *int32    `json:"companyTo,omitempty"`
	UsersTo          []string  `json:"usersTo,omitempty"`
	SentMsg          bool      `json:"sentMsg"`
}

// MessageListQuery filters GET /messages.
type MessageListQuery struct {
	ListQuery
	OnlyUnread bool
	IDGroup    *int64
}

// MessageContact is a messageable contact.
type MessageContact struct {
	ID         int64   `json:"id"`
	Name       string  `json:"name"`
	Email      *string `json:"email,omitempty"`
	InternalID *string `json:"internalId,omitempty"`
}

// MessageGroup is a chat thread id.
type MessageGroup struct {
	ID        int64     `json:"id"`
	Title     *string   `json:"title,omitempty"`
	CreatedAt time.Time `json:"createdAt,omitempty"`
}
