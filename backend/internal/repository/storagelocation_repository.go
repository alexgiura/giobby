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

// StorageLocationRepository persists storage locations.
type StorageLocationRepository interface {
	ListStorageLocations(ctx context.Context, q models.StorageLocationListQuery) ([]models.StorageLocation, error)
	GetStorageLocation(ctx context.Context, idLocation string) (*models.StorageLocation, error)
	CreateStorageLocation(ctx context.Context, loc *models.StorageLocation) (*models.StorageLocation, error)
	UpdateStorageLocation(ctx context.Context, idLocation string, loc *models.StorageLocation) (*models.StorageLocation, error)
	DeleteStorageLocation(ctx context.Context, idLocation string) error
}

type storageLocationRepository struct {
	db *pgxpool.Pool
}

func NewStorageLocationRepository(db *pgxpool.Pool) StorageLocationRepository {
	return &storageLocationRepository{db: db}
}

const storageLocationSelect = `
	SELECT l.id_location, l.id_storage, COALESCE(l.description,'')
	FROM storage_locations l
	JOIN storages s ON s.id = l.id_storage
	WHERE s.company_id = $1 AND s.deleted = false`

func (r *storageLocationRepository) ListStorageLocations(ctx context.Context, q models.StorageLocationListQuery) ([]models.StorageLocation, error) {
	sql := storageLocationSelect
	args := []any{defaultCompanyID}
	n := 2
	if s := strings.TrimSpace(q.IDStorage); s != "" {
		sql += fmt.Sprintf(` AND l.id_storage = $%d`, n)
		args = append(args, s)
		n++
	}
	if q.ManagedStorage != nil {
		managed := 0
		if *q.ManagedStorage {
			managed = 1
		}
		sql += fmt.Sprintf(` AND s.managed = $%d`, n)
		args = append(args, managed)
		n++
	}
	sql += fmt.Sprintf(` ORDER BY l.id_location LIMIT $%d OFFSET $%d`, n, n+1)
	args = append(args, limitOrDefault(q.Limit), q.Offset)

	rows, err := r.db.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("list storage locations: %w", err)
	}
	defer rows.Close()
	return scanStorageLocations(rows)
}

func (r *storageLocationRepository) GetStorageLocation(ctx context.Context, idLocation string) (*models.StorageLocation, error) {
	row := r.db.QueryRow(ctx, storageLocationSelect+` AND l.id_location = $2`,
		defaultCompanyID, strings.TrimSpace(idLocation))
	return scanStorageLocationRow(row)
}

func (r *storageLocationRepository) CreateStorageLocation(ctx context.Context, loc *models.StorageLocation) (*models.StorageLocation, error) {
	loc.IDLocation = strings.TrimSpace(loc.IDLocation)
	loc.IDStorage = strings.TrimSpace(loc.IDStorage)
	if loc.IDLocation == "" || loc.IDStorage == "" {
		return nil, fmt.Errorf("idLocation and idStorage are required")
	}
	var exists bool
	if err := r.db.QueryRow(ctx, `
		SELECT EXISTS(SELECT 1 FROM storages WHERE id=$1 AND company_id=$2 AND deleted=false)
	`, loc.IDStorage, defaultCompanyID).Scan(&exists); err != nil {
		return nil, err
	}
	if !exists {
		return nil, fmt.Errorf("storage not found")
	}
	_, err := r.db.Exec(ctx, `
		INSERT INTO storage_locations (id_location, id_storage, description) VALUES ($1,$2,$3)
	`, loc.IDLocation, loc.IDStorage, emptyAsNull(loc.Description))
	if err != nil {
		return nil, fmt.Errorf("create storage location: %w", err)
	}
	return r.GetStorageLocation(ctx, loc.IDLocation)
}

func (r *storageLocationRepository) UpdateStorageLocation(ctx context.Context, idLocation string, loc *models.StorageLocation) (*models.StorageLocation, error) {
	tag, err := r.db.Exec(ctx, `
		UPDATE storage_locations l SET description=$2, id_storage=COALESCE(NULLIF($3,''), l.id_storage)
		FROM storages s
		WHERE l.id_location=$1 AND l.id_storage=s.id AND s.company_id=$4 AND s.deleted=false
	`, strings.TrimSpace(idLocation), emptyAsNull(loc.Description), strings.TrimSpace(loc.IDStorage), defaultCompanyID)
	if err != nil {
		return nil, fmt.Errorf("update storage location: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return nil, fmt.Errorf("storage location not found")
	}
	return r.GetStorageLocation(ctx, idLocation)
}

func (r *storageLocationRepository) DeleteStorageLocation(ctx context.Context, idLocation string) error {
	tag, err := r.db.Exec(ctx, `
		DELETE FROM storage_locations l
		USING storages s
		WHERE l.id_location=$1 AND l.id_storage=s.id AND s.company_id=$2
	`, strings.TrimSpace(idLocation), defaultCompanyID)
	if err != nil {
		return fmt.Errorf("delete storage location: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("storage location not found")
	}
	return nil
}

func scanStorageLocations(rows pgx.Rows) ([]models.StorageLocation, error) {
	var out []models.StorageLocation
	for rows.Next() {
		loc, err := scanStorageLocationFromScanner(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *loc)
	}
	return out, rows.Err()
}

func scanStorageLocationRow(row pgx.Row) (*models.StorageLocation, error) {
	loc, err := scanStorageLocationFromScanner(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return loc, nil
}

func scanStorageLocationFromScanner(s interface{ Scan(...any) error }) (*models.StorageLocation, error) {
	var loc models.StorageLocation
	if err := s.Scan(&loc.IDLocation, &loc.IDStorage, &loc.Description); err != nil {
		return nil, err
	}
	return &loc, nil
}
