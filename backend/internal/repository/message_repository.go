package repository

import (
	"context"
	"fmt"

	"dnsc_microservice/internal/models"

	"github.com/jackc/pgx/v4/pgxpool"
)

type MessageRepository interface {
	ListMessages(ctx context.Context, q models.MessageListQuery) ([]models.Message, error)
	ListUnread(ctx context.Context) ([]models.Message, error)
	GetMessage(ctx context.Context, id int64) (*models.Message, error)
	CreateMessage(ctx context.Context, m *models.Message) (*models.Message, error)
	MarkMessageRead(ctx context.Context, id int64) (*models.Message, error)
	ListContacts(ctx context.Context, limit, offset int, name string) ([]models.MessageContact, error)
	ListGroups(ctx context.Context, limit, offset int) ([]models.MessageGroup, error)
	DeleteGroup(ctx context.Context, id int64) error
	MarkGroupRead(ctx context.Context, idGroup int64) error
}

type messageRepository struct{ db *pgxpool.Pool }

func NewMessageRepository(db *pgxpool.Pool) MessageRepository {
	return &messageRepository{db: db}
}

func (r *messageRepository) ListMessages(ctx context.Context, q models.MessageListQuery) ([]models.Message, error) {
	limit, offset := limitOrDefault(q.Limit), q.Offset
	query := `
		SELECT id, id_group, unread, sent_date, body, user_from, user_from_full_name, user_to_full_name,
		       id_transaction, id_channel, company_to, sent_msg
		FROM messages WHERE 1=1`
	args := []any{}
	n := 1
	if q.OnlyUnread {
		query += ` AND unread = true`
	}
	if q.IDGroup != nil {
		query += fmt.Sprintf(` AND id_group = $%d`, n)
		args = append(args, *q.IDGroup)
		n++
	}
	query += fmt.Sprintf(` ORDER BY sent_date DESC LIMIT $%d OFFSET $%d`, n, n+1)
	args = append(args, limit, offset)
	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanMessages(rows)
}

func (r *messageRepository) ListUnread(ctx context.Context) ([]models.Message, error) {
	return r.ListMessages(ctx, models.MessageListQuery{OnlyUnread: true, ListQuery: models.ListQuery{Limit: 100}})
}

func (r *messageRepository) GetMessage(ctx context.Context, id int64) (*models.Message, error) {
	row := r.db.QueryRow(ctx, `
		SELECT id, id_group, unread, sent_date, body, user_from, user_from_full_name, user_to_full_name,
		       id_transaction, id_channel, company_to, sent_msg
		FROM messages WHERE id = $1`, id)
	return scanMessage(row)
}

func (r *messageRepository) CreateMessage(ctx context.Context, m *models.Message) (*models.Message, error) {
	if m.IDGroup == nil {
		var gid int64
		err := r.db.QueryRow(ctx, `INSERT INTO message_groups (title) VALUES ($1) RETURNING id`, "Chat").Scan(&gid)
		if err != nil {
			return nil, err
		}
		m.IDGroup = &gid
	}
	row := r.db.QueryRow(ctx, `
		INSERT INTO messages (id_group, body, unread, user_from, user_from_full_name, user_to_full_name,
		                      id_transaction, id_channel, company_to, sent_msg)
		VALUES ($1,$2,true,$3,$4,$5,$6,$7,$8,$9)
		RETURNING id, id_group, unread, sent_date, body, user_from, user_from_full_name, user_to_full_name,
		          id_transaction, id_channel, company_to, sent_msg`,
		m.IDGroup, m.Body, m.UserFrom, m.UserFromFullName, m.UserToFullName,
		m.IDTransaction, m.IDChannel, m.CompanyTo, m.SentMsg)
	return scanMessage(row)
}

func (r *messageRepository) MarkMessageRead(ctx context.Context, id int64) (*models.Message, error) {
	row := r.db.QueryRow(ctx, `
		UPDATE messages SET unread = false WHERE id = $1
		RETURNING id, id_group, unread, sent_date, body, user_from, user_from_full_name, user_to_full_name,
		          id_transaction, id_channel, company_to, sent_msg`, id)
	return scanMessage(row)
}

func (r *messageRepository) ListContacts(ctx context.Context, limit, offset int, name string) ([]models.MessageContact, error) {
	limit = limitOrDefault(limit)
	query := `SELECT id, name, email, internal_id FROM message_contacts WHERE 1=1`
	args := []any{}
	n := 1
	if name != "" {
		query += fmt.Sprintf(` AND name ILIKE $%d`, n)
		args = append(args, "%"+name+"%")
		n++
	}
	query += fmt.Sprintf(` ORDER BY name LIMIT $%d OFFSET $%d`, n, n+1)
	args = append(args, limit, offset)
	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []models.MessageContact
	for rows.Next() {
		var c models.MessageContact
		if err := rows.Scan(&c.ID, &c.Name, &c.Email, &c.InternalID); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func (r *messageRepository) ListGroups(ctx context.Context, limit, offset int) ([]models.MessageGroup, error) {
	limit = limitOrDefault(limit)
	rows, err := r.db.Query(ctx, `
		SELECT id, title, created_at FROM message_groups
		ORDER BY updated_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []models.MessageGroup
	for rows.Next() {
		var g models.MessageGroup
		if err := rows.Scan(&g.ID, &g.Title, &g.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, g)
	}
	return out, rows.Err()
}

func (r *messageRepository) DeleteGroup(ctx context.Context, id int64) error {
	ct, err := r.db.Exec(ctx, `DELETE FROM message_groups WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return fmt.Errorf("message group not found")
	}
	return nil
}

func (r *messageRepository) MarkGroupRead(ctx context.Context, idGroup int64) error {
	_, err := r.db.Exec(ctx, `UPDATE messages SET unread = false WHERE id_group = $1`, idGroup)
	return err
}

type messageScanner interface {
	Scan(dest ...any) error
}

func scanMessage(row messageScanner) (*models.Message, error) {
	var m models.Message
	err := row.Scan(&m.ID, &m.IDGroup, &m.Unread, &m.SentDate, &m.Body, &m.UserFrom, &m.UserFromFullName,
		&m.UserToFullName, &m.IDTransaction, &m.IDChannel, &m.CompanyTo, &m.SentMsg)
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func scanMessages(rows interface {
	Next() bool
	Scan(dest ...any) error
	Err() error
}) ([]models.Message, error) {
	var out []models.Message
	for rows.Next() {
		m, err := scanMessage(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *m)
	}
	return out, rows.Err()
}
