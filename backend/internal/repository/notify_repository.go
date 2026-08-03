package repository

import (
	"context"
	"encoding/json"
	"fmt"

	"dnsc_microservice/internal/models"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

type NotifyRepository interface {
	ListNotifies(ctx context.Context, channel string, q models.NotifyListQuery) ([]models.AppNotify, error)
	ListUnreadNotifies(ctx context.Context, channel string) ([]models.AppNotify, error)
	GetNotify(ctx context.Context, channel string, id int64) (*models.AppNotify, error)
	MarkNotifyRead(ctx context.Context, channel string, id int64) (*models.AppNotify, error)
	ListSocialPosts(ctx context.Context, limit, offset int, idLang string) ([]models.SocialPost, error)
	ListEmails(ctx context.Context, limit, offset int, hashcode string) ([]models.EmailItem, error)
	ListGlobalNotifications(ctx context.Context) ([]models.GlobalNotification, error)
	SaveIntegrationEvent(ctx context.Context, provider string, payload any) error
}

type notifyRepository struct{ db *pgxpool.Pool }

func NewNotifyRepository(db *pgxpool.Pool) NotifyRepository {
	return &notifyRepository{db: db}
}

func (r *notifyRepository) ListNotifies(ctx context.Context, channel string, q models.NotifyListQuery) ([]models.AppNotify, error) {
	limit, offset := limitOrDefault(q.Limit), q.Offset
	order := "ASC"
	if q.DescendingOrder {
		order = "DESC"
	}
	query := fmt.Sprintf(`
		SELECT id, channel, title, body, unread, created_at
		FROM app_notifies WHERE channel = $1`)
	args := []any{channel}
	n := 2
	if q.OnlyUnread {
		query += ` AND unread = true`
	}
	query += fmt.Sprintf(` ORDER BY created_at %s LIMIT $%d OFFSET $%d`, order, n, n+1)
	args = append(args, limit, offset)
	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanNotifies(rows)
}

func (r *notifyRepository) ListUnreadNotifies(ctx context.Context, channel string) ([]models.AppNotify, error) {
	return r.ListNotifies(ctx, channel, models.NotifyListQuery{
		OnlyUnread: true,
		ListQuery:  models.ListQuery{Limit: 100},
	})
}

func (r *notifyRepository) GetNotify(ctx context.Context, channel string, id int64) (*models.AppNotify, error) {
	row := r.db.QueryRow(ctx, `
		SELECT id, channel, title, body, unread, created_at
		FROM app_notifies WHERE channel = $1 AND id = $2`, channel, id)
	return scanNotify(row)
}

func (r *notifyRepository) MarkNotifyRead(ctx context.Context, channel string, id int64) (*models.AppNotify, error) {
	row := r.db.QueryRow(ctx, `
		UPDATE app_notifies SET unread = false WHERE channel = $1 AND id = $2
		RETURNING id, channel, title, body, unread, created_at`, channel, id)
	n, err := scanNotify(row)
	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("notify not found")
	}
	return n, err
}

func (r *notifyRepository) ListSocialPosts(ctx context.Context, limit, offset int, idLang string) ([]models.SocialPost, error) {
	limit = limitOrDefault(limit)
	query := `SELECT id, id_lang, title, body, created_at FROM social_posts WHERE 1=1`
	args := []any{}
	n := 1
	if idLang != "" {
		query += fmt.Sprintf(` AND id_lang = $%d`, n)
		args = append(args, idLang)
		n++
	}
	query += fmt.Sprintf(` ORDER BY created_at DESC LIMIT $%d OFFSET $%d`, n, n+1)
	args = append(args, limit, offset)
	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []models.SocialPost
	for rows.Next() {
		var p models.SocialPost
		if err := rows.Scan(&p.ID, &p.IDLang, &p.Title, &p.Body, &p.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

func (r *notifyRepository) ListEmails(ctx context.Context, limit, offset int, hashcode string) ([]models.EmailItem, error) {
	limit = limitOrDefault(limit)
	query := `SELECT id, hashcode, subject, body, sender, recipient, created_at FROM emails WHERE 1=1`
	args := []any{}
	n := 1
	if hashcode != "" {
		query += fmt.Sprintf(` AND hashcode = $%d`, n)
		args = append(args, hashcode)
		n++
	}
	query += fmt.Sprintf(` ORDER BY created_at DESC LIMIT $%d OFFSET $%d`, n, n+1)
	args = append(args, limit, offset)
	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []models.EmailItem
	for rows.Next() {
		var e models.EmailItem
		if err := rows.Scan(&e.ID, &e.Hashcode, &e.Subject, &e.Body, &e.Sender, &e.Recipient, &e.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

func (r *notifyRepository) ListGlobalNotifications(ctx context.Context) ([]models.GlobalNotification, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, title, body, source_company, created_at
		FROM global_notifications ORDER BY created_at DESC LIMIT 100`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []models.GlobalNotification
	for rows.Next() {
		var g models.GlobalNotification
		if err := rows.Scan(&g.ID, &g.Title, &g.Body, &g.SourceCompany, &g.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, g)
	}
	return out, rows.Err()
}

func (r *notifyRepository) SaveIntegrationEvent(ctx context.Context, provider string, payload any) error {
	b, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	_, err = r.db.Exec(ctx, `
		INSERT INTO integration_events (provider, payload) VALUES ($1, $2::jsonb)`, provider, string(b))
	return err
}

type notifyScanner interface {
	Scan(dest ...any) error
}

func scanNotify(row notifyScanner) (*models.AppNotify, error) {
	var n models.AppNotify
	err := row.Scan(&n.ID, &n.Channel, &n.Title, &n.Body, &n.Unread, &n.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &n, nil
}

func scanNotifies(rows interface {
	Next() bool
	Scan(dest ...any) error
	Err() error
}) ([]models.AppNotify, error) {
	var out []models.AppNotify
	for rows.Next() {
		n, err := scanNotify(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *n)
	}
	return out, rows.Err()
}
