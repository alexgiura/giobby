package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"dnsc_microservice/internal/models"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

// PricelistRepository persists pricelists and schemes.
type PricelistRepository interface {
	ListPricelists(ctx context.Context, q models.PricelistListQuery) ([]models.Pricelist, error)
	CreatePricelist(ctx context.Context, p *models.Pricelist) (*models.Pricelist, error)
	GetPricelistRows(ctx context.Context, id int32, q models.PricelistDetailQuery) ([]models.PricelistRow, error)
	UpdatePricelist(ctx context.Context, id int32, p *models.Pricelist) (*models.Pricelist, error)
	DeletePricelist(ctx context.Context, id int32) error
	ListPricelistSchemes(ctx context.Context, q models.ListQuery) ([]models.PricelistScheme, error)
	GetPricelistScheme(ctx context.Context, id int32) (*models.PricelistScheme, error)
}

type pricelistRepository struct {
	db *pgxpool.Pool
}

func NewPricelistRepository(db *pgxpool.Pool) PricelistRepository {
	return &pricelistRepository{db: db}
}

const pricelistSelectSQL = `
	SELECT id, id_pricelist_scheme, description, COALESCE(type,''), COALESCE(note,''),
		valid_since, valid_until, priority, COALESCE(source_id,'')
	FROM pricelists WHERE company_id = $1 AND deleted = false`

func (r *pricelistRepository) ListPricelists(ctx context.Context, q models.PricelistListQuery) ([]models.Pricelist, error) {
	sql := pricelistSelectSQL
	args := []any{defaultCompanyID}
	n := 2
	if s := strings.TrimSpace(q.Type); s != "" {
		sql += fmt.Sprintf(` AND type = $%d`, n)
		args = append(args, s)
		n++
	}
	sql += fmt.Sprintf(` ORDER BY id LIMIT $%d OFFSET $%d`, n, n+1)
	args = append(args, limitOrDefault(q.Limit), q.Offset)

	rows, err := r.db.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("list pricelists: %w", err)
	}
	defer rows.Close()
	return scanPricelists(rows)
}

func (r *pricelistRepository) CreatePricelist(ctx context.Context, p *models.Pricelist) (*models.Pricelist, error) {
	row := r.db.QueryRow(ctx, `
		INSERT INTO pricelists (id_pricelist_scheme, description, type, note, valid_since, valid_until, priority, source_id)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
		RETURNING id
	`, p.IDPricelistScheme, p.Description, emptyAsNull(p.Type), emptyAsNull(p.Note),
		millisToTimePtr(p.ValidSince), millisToTimePtr(p.ValidUntil), p.Priority, emptyAsNull(p.SourceID))
	var id int32
	if err := row.Scan(&id); err != nil {
		return nil, fmt.Errorf("create pricelist: %w", err)
	}
	return r.getPricelistHeader(ctx, id)
}

func (r *pricelistRepository) GetPricelistRows(ctx context.Context, id int32, q models.PricelistDetailQuery) ([]models.PricelistRow, error) {
	if _, err := r.getPricelistHeader(ctx, id); err != nil {
		return nil, err
	}
	sql := `
		SELECT COALESCE(material_id,''), COALESCE(prod_description,''), price
		FROM pricelist_rows WHERE pricelist_id = $1`
	args := []any{id}
	n := 2
	if s := strings.TrimSpace(q.MaterialID); s != "" {
		sql += fmt.Sprintf(` AND material_id = $%d`, n)
		args = append(args, s)
		n++
	} else if s := strings.TrimSpace(q.ProdDescription); s != "" {
		sql += fmt.Sprintf(` AND prod_description ILIKE $%d`, n)
		args = append(args, "%"+s+"%")
		n++
	}
	sql += fmt.Sprintf(` ORDER BY id LIMIT $%d OFFSET $%d`, n, n+1)
	args = append(args, limitOrDefault(q.Limit), q.Offset)

	rows, err := r.db.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("get pricelist rows: %w", err)
	}
	defer rows.Close()
	var out []models.PricelistRow
	for rows.Next() {
		var row models.PricelistRow
		if err := rows.Scan(&row.MaterialID, &row.ProdDescription, &row.Price); err != nil {
			return nil, err
		}
		out = append(out, row)
	}
	return out, rows.Err()
}

func (r *pricelistRepository) UpdatePricelist(ctx context.Context, id int32, p *models.Pricelist) (*models.Pricelist, error) {
	tag, err := r.db.Exec(ctx, `
		UPDATE pricelists SET
			id_pricelist_scheme=$2, description=$3, type=$4, note=$5,
			valid_since=$6, valid_until=$7, priority=$8, source_id=$9
		WHERE id=$1 AND company_id=$10 AND deleted=false
	`, id, p.IDPricelistScheme, p.Description, emptyAsNull(p.Type), emptyAsNull(p.Note),
		millisToTimePtr(p.ValidSince), millisToTimePtr(p.ValidUntil), p.Priority, emptyAsNull(p.SourceID), defaultCompanyID)
	if err != nil {
		return nil, fmt.Errorf("update pricelist: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return nil, fmt.Errorf("pricelist not found")
	}
	return r.getPricelistHeader(ctx, id)
}

func (r *pricelistRepository) DeletePricelist(ctx context.Context, id int32) error {
	tag, err := r.db.Exec(ctx, `
		UPDATE pricelists SET deleted=true WHERE id=$1 AND company_id=$2 AND deleted=false
	`, id, defaultCompanyID)
	if err != nil {
		return fmt.Errorf("delete pricelist: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("pricelist not found")
	}
	return nil
}

func (r *pricelistRepository) ListPricelistSchemes(ctx context.Context, q models.ListQuery) ([]models.PricelistScheme, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, description FROM pricelist_schemes
		WHERE company_id = $1
		ORDER BY id LIMIT $2 OFFSET $3
	`, defaultCompanyID, limitOrDefault(q.Limit), q.Offset)
	if err != nil {
		return nil, fmt.Errorf("list pricelist schemes: %w", err)
	}
	defer rows.Close()
	var out []models.PricelistScheme
	for rows.Next() {
		var s models.PricelistScheme
		if err := rows.Scan(&s.ID, &s.Description); err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

func (r *pricelistRepository) GetPricelistScheme(ctx context.Context, id int32) (*models.PricelistScheme, error) {
	row := r.db.QueryRow(ctx, `
		SELECT id, description FROM pricelist_schemes WHERE id = $1 AND company_id = $2
	`, id, defaultCompanyID)
	var s models.PricelistScheme
	if err := row.Scan(&s.ID, &s.Description); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("get pricelist scheme: %w", err)
	}
	return &s, nil
}

func (r *pricelistRepository) getPricelistHeader(ctx context.Context, id int32) (*models.Pricelist, error) {
	row := r.db.QueryRow(ctx, pricelistSelectSQL+` AND id = $2`, defaultCompanyID, id)
	p, err := scanPricelistRow(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("pricelist not found")
		}
		return nil, err
	}
	return p, nil
}

func scanPricelists(rows pgx.Rows) ([]models.Pricelist, error) {
	var out []models.Pricelist
	for rows.Next() {
		p, err := scanPricelistFromScanner(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *p)
	}
	return out, rows.Err()
}

func scanPricelistRow(row pgx.Row) (*models.Pricelist, error) {
	return scanPricelistFromScanner(row)
}

func scanPricelistFromScanner(s interface{ Scan(...any) error }) (*models.Pricelist, error) {
	var p models.Pricelist
	var validSince, validUntil *time.Time
	if err := s.Scan(&p.ID, &p.IDPricelistScheme, &p.Description, &p.Type, &p.Note,
		&validSince, &validUntil, &p.Priority, &p.SourceID); err != nil {
		return nil, err
	}
	if validSince != nil {
		p.ValidSince = models.MillisPtr(*validSince)
	}
	if validUntil != nil {
		p.ValidUntil = models.MillisPtr(*validUntil)
	}
	return &p, nil
}

func millisToTimePtr(ms *int64) *time.Time {
	if ms == nil {
		return nil
	}
	t := time.UnixMilli(*ms)
	return &t
}
