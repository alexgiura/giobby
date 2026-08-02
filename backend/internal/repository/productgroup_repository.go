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

// ProductGroupRepository persists product categories.
type ProductGroupRepository interface {
	ListProductGroups(ctx context.Context, q models.ProductGroupListQuery) ([]models.ProductGroup, error)
	GetProductGroup(ctx context.Context, id int32) (*models.ProductGroup, error)
	CreateProductGroup(ctx context.Context, g *models.ProductGroup) (*models.ProductGroup, error)
	UpdateProductGroup(ctx context.Context, id int32, g *models.ProductGroup) (*models.ProductGroup, error)
	DeleteProductGroup(ctx context.Context, id int32) error
}

type productGroupRepository struct {
	db *pgxpool.Pool
}

func NewProductGroupRepository(db *pgxpool.Pool) ProductGroupRepository {
	return &productGroupRepository{db: db}
}

func (r *productGroupRepository) ListProductGroups(ctx context.Context, q models.ProductGroupListQuery) ([]models.ProductGroup, error) {
	sql := `SELECT id, id_parent, COALESCE(description_it,''), COALESCE(description_en,''), COALESCE(description_es,''), COALESCE(description_bg,'')
		FROM product_groups WHERE company_id = $1`
	args := []any{defaultCompanyID}
	n := 2
	if s := strings.TrimSpace(q.Description); s != "" {
		sql += fmt.Sprintf(` AND (description_it ILIKE $%d OR description_en ILIKE $%d)`, n, n)
		args = append(args, "%"+s+"%")
		n++
	}
	sql += fmt.Sprintf(` ORDER BY id LIMIT $%d OFFSET $%d`, n, n+1)
	args = append(args, limitOrDefault(q.Limit), q.Offset)

	rows, err := r.db.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("list product groups: %w", err)
	}
	defer rows.Close()
	return scanProductGroups(rows)
}

func (r *productGroupRepository) GetProductGroup(ctx context.Context, id int32) (*models.ProductGroup, error) {
	row := r.db.QueryRow(ctx, `
		SELECT id, id_parent, COALESCE(description_it,''), COALESCE(description_en,''), COALESCE(description_es,''), COALESCE(description_bg,'')
		FROM product_groups WHERE id = $1 AND company_id = $2
	`, id, defaultCompanyID)
	g, err := scanProductGroupRow(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("get product group: %w", err)
	}
	return g, nil
}

func (r *productGroupRepository) CreateProductGroup(ctx context.Context, g *models.ProductGroup) (*models.ProductGroup, error) {
	row := r.db.QueryRow(ctx, `
		INSERT INTO product_groups (id_parent, description_it, description_en, description_es, description_bg)
		VALUES ($1,$2,$3,$4,$5)
		RETURNING id
	`, g.IDParent, emptyAsNull(g.DescriptionIT), emptyAsNull(g.DescriptionEN), emptyAsNull(g.DescriptionES), emptyAsNull(g.DescriptionBG))
	var id int32
	if err := row.Scan(&id); err != nil {
		return nil, fmt.Errorf("create product group: %w", err)
	}
	return r.GetProductGroup(ctx, id)
}

func (r *productGroupRepository) UpdateProductGroup(ctx context.Context, id int32, g *models.ProductGroup) (*models.ProductGroup, error) {
	tag, err := r.db.Exec(ctx, `
		UPDATE product_groups SET id_parent=$2, description_it=$3, description_en=$4, description_es=$5, description_bg=$6
		WHERE id=$1 AND company_id=$7
	`, id, g.IDParent, emptyAsNull(g.DescriptionIT), emptyAsNull(g.DescriptionEN), emptyAsNull(g.DescriptionES), emptyAsNull(g.DescriptionBG), defaultCompanyID)
	if err != nil {
		return nil, fmt.Errorf("update product group: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return nil, fmt.Errorf("product group not found")
	}
	return r.GetProductGroup(ctx, id)
}

func (r *productGroupRepository) DeleteProductGroup(ctx context.Context, id int32) error {
	tag, err := r.db.Exec(ctx, `DELETE FROM product_groups WHERE id=$1 AND company_id=$2`, id, defaultCompanyID)
	if err != nil {
		return fmt.Errorf("delete product group: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("product group not found")
	}
	return nil
}

func scanProductGroups(rows pgx.Rows) ([]models.ProductGroup, error) {
	var out []models.ProductGroup
	for rows.Next() {
		g, err := scanProductGroupFromScanner(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *g)
	}
	return out, rows.Err()
}

func scanProductGroupRow(row pgx.Row) (*models.ProductGroup, error) {
	return scanProductGroupFromScanner(row)
}

func scanProductGroupFromScanner(s interface{ Scan(...any) error }) (*models.ProductGroup, error) {
	var g models.ProductGroup
	if err := s.Scan(&g.ID, &g.IDParent, &g.DescriptionIT, &g.DescriptionEN, &g.DescriptionES, &g.DescriptionBG); err != nil {
		return nil, err
	}
	return &g, nil
}
