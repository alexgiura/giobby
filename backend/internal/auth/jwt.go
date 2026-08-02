package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type accessClaims struct {
	jwt.RegisteredClaims
	Sid string `json:"sid"`
}

// IssueAccessToken signs a short-lived JWT access token.
func IssueAccessToken(userID, sessionID uuid.UUID, secret string, ttl time.Duration) (string, int64, error) {
	if ttl <= 0 {
		ttl = time.Hour
	}
	now := time.Now().UTC()
	exp := now.Add(ttl)
	claims := accessClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID.String(),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(exp),
		},
		Sid: sessionID.String(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", 0, fmt.Errorf("sign access token: %w", err)
	}
	return signed, int64(ttl.Seconds()), nil
}

// ParseAccessToken validates the JWT and returns user and session IDs.
func ParseAccessToken(tokenStr, secret string) (userID, sessionID uuid.UUID, err error) {
	claims := &accessClaims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (any, error) {
		if t.Method != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return []byte(secret), nil
	})
	if err != nil {
		return uuid.Nil, uuid.Nil, fmt.Errorf("parse access token: %w", err)
	}
	if !token.Valid {
		return uuid.Nil, uuid.Nil, fmt.Errorf("invalid access token")
	}
	userID, err = uuid.Parse(claims.Subject)
	if err != nil {
		return uuid.Nil, uuid.Nil, fmt.Errorf("invalid subject: %w", err)
	}
	sessionID, err = uuid.Parse(claims.Sid)
	if err != nil {
		return uuid.Nil, uuid.Nil, fmt.Errorf("invalid sid: %w", err)
	}
	return userID, sessionID, nil
}
