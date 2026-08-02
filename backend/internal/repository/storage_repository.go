package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"dnsc_microservice/internal/models"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

// StorageRepository persists warehouses/storages.
type StorageRepository interface {
	ListStorages(ctx context.Context, q models.StorageListQuery) ([]models.Storage, error)
	GetStorage(ctx context.Context, id string) (*models.Storage, error)
	CreateStorage(ctx context.Context, s *models.Storage) (*models.Storage, error)
	UpdateStorage(ctx context.Context, id string, s *models.Storage) (*models.Storage, error)
	DeleteStorage(ctx context.Context, id string) error
}

type storageRepository struct {
	db *pgxpool.Pool
}

func NewStorageRepository(db *pgxpool.Pool) StorageRepository {
	return &storageRepository{db: db}
}

func (r *storageRepository) ListStorages(ctx context.Context, q models.StorageListQuery) ([]models.Storage, error) {
	sql := `SELECT id, description, managed FROM storages WHERE company_id = $1 AND deleted = false`
	args := []any{defaultCompanyID}
	n := 2
	if s := strings.TrimSpace(q.ID); s != "" {
		sql += fmt.Sprintf(` AND id = $%d`, n)
		args = append(args, s)
		n++
	}
	if q.Managed != nil {
		managed := 0
		if *q.Managed {
			managed = 1
		}
		sql += fmt.Sprintf(` AND managed = $%d`, n)
		args = append(args, managed)
		n++
	}
	sql += fmt.Sprintf(` ORDER BY id LIMIT $%d OFFSET $%d`, n, n+1)
	args = append(args, limitOrDefault(q.Limit), q.Offset)

	rows, err := r.db.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("list storages: %w", err)
	}
	defer rows.Close()
	var out []models.Storage
	for rows.Next() {
		var s models.Storage
		if err := rows.Scan(&s.ID, &s.Description, &s.Managed); err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

func (r *storageRepository) GetStorage(ctx context.Context, id string) (*models.Storage, error) {
	row := r.db.QueryRow(ctx, `
		SELECT id, description, managed FROM storages
		WHERE id = $1 AND company_id = $2 AND deleted = false
	`, strings.TrimSpace(id), defaultCompanyID)
	var s models.Storage
	if err := row.Scan(&s.ID, &s.Description, &s.Managed); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("get storage: %w", err)
	}
	return &s, nil
}

func (r *storageRepository) CreateStorage(ctx context.Context, s *models.Storage) (*models.Storage, error) {
	s.ID = strings.TrimSpace(s.ID)
	if s.ID == "" || strings.TrimSpace(s.Description) == "" {
		return nil, fmt.Errorf("id and description are required")
	}
	_, err := r.db.Exec(ctx, `
		INSERT INTO storages (id, description, managed) VALUES ($1,$2,$3)
	`, s.ID, s.Description, s.Managed)
	if err != nil {
		return nil, fmt.Errorf("create storage: %w", err)
	}
	return r.GetStorage(ctx, s.ID)
}

func (r *storageRepository) UpdateStorage(ctx context.Context, id string, s *models.Storage) (*models.Storage, error) {
	tag, err := r.db.Exec(ctx, `
		UPDATE storages SET description=$2, managed=$3
		WHERE id=$1 AND company_id=$4 AND deleted=false
	`, strings.TrimSpace(id), s.Description, s.Managed, defaultCompanyID)
	if err != nil {
		return nil, fmt.Errorf("update storage: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return nil, fmt.Errorf("storage not found")
	}
	return r.GetStorage(ctx, id)
}

func (r *storageRepository) DeleteStorage(ctx context.Context, id string) error {
	tag, err := r.db.Exec(ctx, `
		UPDATE storages SET deleted=true WHERE id=$1 AND company_id=$2 AND deleted=false
	`, strings.TrimSpace(id), defaultCompanyID)
	if err != nil {
		return fmt.Errorf("delete storage: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("storage not found")
	}
	return nil
}
