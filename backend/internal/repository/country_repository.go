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

// CountryRepository persists countries.
type CountryRepository interface {
	ListCountries(ctx context.Context, q models.CountryListQuery) ([]models.Country, error)
	GetCountry(ctx context.Context, id string) (*models.Country, error)
	CreateCountry(ctx context.Context, country *models.Country) (*models.Country, error)
	UpdateCountry(ctx context.Context, id string, country *models.Country) (*models.Country, error)
	DeleteCountry(ctx context.Context, id string) error
	RecoverCountry(ctx context.Context, id string) (*models.Country, error)
}

type countryRepository struct {
	db *pgxpool.Pool
}

func NewCountryRepository(db *pgxpool.Pool) CountryRepository {
	return &countryRepository{db: db}
}

func (r *countryRepository) ListCountries(ctx context.Context, q models.CountryListQuery) ([]models.Country, error) {
	sql := `SELECT id, description FROM countries WHERE 1=1`
	args := []any{}
	n := 1
	if !q.ShowDeleted {
		sql += ` AND deleted_at IS NULL`
	}
	if s := strings.TrimSpace(q.Description); s != "" {
		sql += fmt.Sprintf(` AND description ILIKE $%d`, n)
		args = append(args, "%"+s+"%")
		n++
	}
	sql += fmt.Sprintf(` ORDER BY description LIMIT $%d OFFSET $%d`, n, n+1)
	args = append(args, limitOrDefault(q.Limit), q.Offset)

	rows, err := r.db.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("list countries: %w", err)
	}
	defer rows.Close()

	var out []models.Country
	for rows.Next() {
		var c models.Country
		if err := rows.Scan(&c.ID, &c.Description); err != nil {
			return nil, fmt.Errorf("scan country: %w", err)
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func (r *countryRepository) GetCountry(ctx context.Context, id string) (*models.Country, error) {
	row := r.db.QueryRow(ctx, `SELECT id, description FROM countries WHERE id = $1 AND deleted_at IS NULL`, strings.TrimSpace(id))
	var c models.Country
	if err := row.Scan(&c.ID, &c.Description); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("get country: %w", err)
	}
	return &c, nil
}

func (r *countryRepository) CreateCountry(ctx context.Context, country *models.Country) (*models.Country, error) {
	if country == nil {
		return nil, fmt.Errorf("country is nil")
	}
	row := r.db.QueryRow(ctx, `
		INSERT INTO countries (id, description)
		VALUES ($1, $2)
		RETURNING id, description
	`, country.ID, country.Description)
	var out models.Country
	if err := row.Scan(&out.ID, &out.Description); err != nil {
		return nil, fmt.Errorf("create country: %w", err)
	}
	return &out, nil
}

func (r *countryRepository) UpdateCountry(ctx context.Context, id string, country *models.Country) (*models.Country, error) {
	if country == nil {
		return nil, fmt.Errorf("country is nil")
	}
	row := r.db.QueryRow(ctx, `
		UPDATE countries SET description = $2
		WHERE id = $1 AND deleted_at IS NULL
		RETURNING id, description
	`, strings.TrimSpace(id), country.Description)
	var out models.Country
	if err := row.Scan(&out.ID, &out.Description); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("country not found")
		}
		return nil, fmt.Errorf("update country: %w", err)
	}
	return &out, nil
}

func (r *countryRepository) DeleteCountry(ctx context.Context, id string) error {
	tag, err := r.db.Exec(ctx, `
		UPDATE countries SET deleted_at = now()
		WHERE id = $1 AND deleted_at IS NULL
	`, strings.TrimSpace(id))
	if err != nil {
		return fmt.Errorf("delete country: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("country not found")
	}
	return nil
}

func (r *countryRepository) RecoverCountry(ctx context.Context, id string) (*models.Country, error) {
	row := r.db.QueryRow(ctx, `
		UPDATE countries SET deleted_at = NULL
		WHERE id = $1 AND deleted_at IS NOT NULL
		RETURNING id, description
	`, strings.TrimSpace(id))
	var out models.Country
	if err := row.Scan(&out.ID, &out.Description); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("country not found")
		}
		return nil, fmt.Errorf("recover country: %w", err)
	}
	return &out, nil
}
