package repository

import (
	"context"
	"fmt"

	"dnsc_microservice/internal/models"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

type LoggedUserRepository interface {
	GetOrCreateProfile(ctx context.Context, userID uuid.UUID) (language string, imageURL *string, err error)
	SetLanguage(ctx context.Context, userID uuid.UUID, language string) error
	SetImageURL(ctx context.Context, userID uuid.UUID, imageURL string) error
	AddDeviceToken(ctx context.Context, userID uuid.UUID, os *string, token string) (*models.DeviceToken, error)
	DeleteDeviceToken(ctx context.Context, userID uuid.UUID, id int64) error
}

type loggedUserRepository struct{ db *pgxpool.Pool }

func NewLoggedUserRepository(db *pgxpool.Pool) LoggedUserRepository {
	return &loggedUserRepository{db: db}
}

func (r *loggedUserRepository) GetOrCreateProfile(ctx context.Context, userID uuid.UUID) (string, *string, error) {
	var lang string
	var img *string
	err := r.db.QueryRow(ctx, `
		SELECT language, image_url FROM logged_user_profiles WHERE user_id = $1`, userID).Scan(&lang, &img)
	if err == nil {
		return lang, img, nil
	}
	if err != pgx.ErrNoRows {
		return "", nil, err
	}
	_, err = r.db.Exec(ctx, `
		INSERT INTO logged_user_profiles (user_id, language) VALUES ($1, 'en')
		ON CONFLICT (user_id) DO NOTHING`, userID)
	if err != nil {
		return "", nil, err
	}
	return "en", nil, nil
}

func (r *loggedUserRepository) SetLanguage(ctx context.Context, userID uuid.UUID, language string) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO logged_user_profiles (user_id, language, updated_at)
		VALUES ($1, $2, now())
		ON CONFLICT (user_id) DO UPDATE SET language = EXCLUDED.language, updated_at = now()`,
		userID, language)
	return err
}

func (r *loggedUserRepository) SetImageURL(ctx context.Context, userID uuid.UUID, imageURL string) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO logged_user_profiles (user_id, language, image_url, updated_at)
		VALUES ($1, 'en', $2, now())
		ON CONFLICT (user_id) DO UPDATE SET image_url = EXCLUDED.image_url, updated_at = now()`,
		userID, imageURL)
	return err
}

func (r *loggedUserRepository) AddDeviceToken(ctx context.Context, userID uuid.UUID, os *string, token string) (*models.DeviceToken, error) {
	var d models.DeviceToken
	err := r.db.QueryRow(ctx, `
		INSERT INTO logged_user_device_tokens (user_id, os, device_token)
		VALUES ($1, $2, $3)
		ON CONFLICT (user_id, device_token) DO UPDATE SET os = EXCLUDED.os
		RETURNING id, os, device_token, created_at`, userID, os, token).
		Scan(&d.ID, &d.OS, &d.DeviceToken, &d.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &d, nil
}

func (r *loggedUserRepository) DeleteDeviceToken(ctx context.Context, userID uuid.UUID, id int64) error {
	ct, err := r.db.Exec(ctx, `
		DELETE FROM logged_user_device_tokens WHERE id = $1 AND user_id = $2`, id, userID)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return fmt.Errorf("device token not found")
	}
	return nil
}
