package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"dnsc_microservice/internal/models"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

// AuthRepository persists local users and refresh-token sessions.
type AuthRepository interface {
	CreateUser(ctx context.Context, user *models.User) error
	GetUserByUsername(ctx context.Context, username string) (*models.User, error)
	GetUserByID(ctx context.Context, id uuid.UUID) (*models.User, error)
	InsertSession(ctx context.Context, sessionID, userID uuid.UUID, refreshTokenHash string, expiresAt time.Time) error
	GetSessionByID(ctx context.Context, sessionID uuid.UUID) (*models.UserSession, error)
	GetSessionByRefreshTokenHash(ctx context.Context, refreshTokenHash string) (*models.UserSession, error)
	RevokeSession(ctx context.Context, sessionID uuid.UUID) error
	DeleteSession(ctx context.Context, sessionID uuid.UUID) error
}

type authRepository struct {
	db *pgxpool.Pool
}

func NewAuthRepository(db *pgxpool.Pool) AuthRepository {
	return &authRepository{db: db}
}

func (r *authRepository) CreateUser(ctx context.Context, user *models.User) error {
	if user == nil {
		return fmt.Errorf("user is nil")
	}
	_, err := r.db.Exec(ctx, `
		INSERT INTO users (id, username, email, password_hash, is_active)
		VALUES ($1, $2, $3, $4, $5)
	`, user.ID, user.Username, user.Email, user.PasswordHash, user.IsActive)
	if err != nil {
		return fmt.Errorf("create user: %w", err)
	}
	return nil
}

func (r *authRepository) GetUserByUsername(ctx context.Context, username string) (*models.User, error) {
	username = strings.TrimSpace(username)
	row := r.db.QueryRow(ctx, `
		SELECT id, username, email, password_hash, is_active, created_at, updated_at
		FROM users WHERE username = $1
	`, username)
	return scanUser(row)
}

func (r *authRepository) GetUserByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	row := r.db.QueryRow(ctx, `
		SELECT id, username, email, password_hash, is_active, created_at, updated_at
		FROM users WHERE id = $1
	`, id)
	return scanUser(row)
}

func scanUser(row pgx.Row) (*models.User, error) {
	var u models.User
	if err := row.Scan(&u.ID, &u.Username, &u.Email, &u.PasswordHash, &u.IsActive, &u.CreatedAt, &u.UpdatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("scan user: %w", err)
	}
	return &u, nil
}

func (r *authRepository) InsertSession(ctx context.Context, sessionID, userID uuid.UUID, refreshTokenHash string, expiresAt time.Time) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO user_sessions (id, user_id, refresh_token_hash, expires_at)
		VALUES ($1, $2, $3, $4)
	`, sessionID, userID, refreshTokenHash, expiresAt)
	if err != nil {
		return fmt.Errorf("insert session: %w", err)
	}
	return nil
}

func (r *authRepository) GetSessionByID(ctx context.Context, sessionID uuid.UUID) (*models.UserSession, error) {
	row := r.db.QueryRow(ctx, `
		SELECT id, user_id, refresh_token_hash, expires_at, revoked_at, created_at
		FROM user_sessions WHERE id = $1
	`, sessionID)
	return scanSession(row)
}

func (r *authRepository) GetSessionByRefreshTokenHash(ctx context.Context, refreshTokenHash string) (*models.UserSession, error) {
	row := r.db.QueryRow(ctx, `
		SELECT id, user_id, refresh_token_hash, expires_at, revoked_at, created_at
		FROM user_sessions WHERE refresh_token_hash = $1
	`, refreshTokenHash)
	return scanSession(row)
}

func scanSession(row pgx.Row) (*models.UserSession, error) {
	var s models.UserSession
	if err := row.Scan(&s.ID, &s.UserID, &s.RefreshTokenHash, &s.ExpiresAt, &s.RevokedAt, &s.CreatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("scan session: %w", err)
	}
	return &s, nil
}

func (r *authRepository) RevokeSession(ctx context.Context, sessionID uuid.UUID) error {
	tag, err := r.db.Exec(ctx, `
		UPDATE user_sessions SET revoked_at = now() WHERE id = $1 AND revoked_at IS NULL
	`, sessionID)
	if err != nil {
		return fmt.Errorf("revoke session: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("session not found")
	}
	return nil
}

func (r *authRepository) DeleteSession(ctx context.Context, sessionID uuid.UUID) error {
	_, err := r.db.Exec(ctx, `DELETE FROM user_sessions WHERE id = $1`, sessionID)
	if err != nil {
		return fmt.Errorf("delete session: %w", err)
	}
	return nil
}
