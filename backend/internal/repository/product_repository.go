package repository

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"dnsc_microservice/internal/models"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

// ProductRepository persists products and variants.
type ProductRepository interface {
	ListProducts(ctx context.Context, q models.ProductListQuery) ([]models.Product, error)
	GetProduct(ctx context.Context, id string) (*models.Product, error)
	CreateProduct(ctx context.Context, p *models.Product) (*models.Product, error)
	UpdateProduct(ctx context.Context, p *models.Product) (*models.Product, error)
	ListVariants(ctx context.Context, productID string) ([]models.AttributeCombination, error)
	UpdateVariants(ctx context.Context, variants []models.AttributeCombination) error
	LinkMedia(ctx context.Context, productID, mediaID string) error
}

type productRepository struct {
	db *pgxpool.Pool
}

func NewProductRepository(db *pgxpool.Pool) ProductRepository {
	return &productRepository{db: db}
}

const productSelectSQL = `
	SELECT p.id, COALESCE(p.barcode,''), COALESCE(p.basic_code,''), p.id_type, COALESCE(p.type_desc,''),
		p.description, COALESCE(p.description_it,''), COALESCE(p.description_en,''), COALESCE(p.description_es,''), COALESCE(p.description_bg,''),
		COALESCE(p.note,''), p.id_material_group, COALESCE(pg.description_it,''),
		COALESCE(p.um,''), COALESCE(u.description_it,''),
		p.sales_enabled, p.locked, p.stock_enabled, p.pricelist_enabled,
		p.sales_price, p.sales_price_include_vat, COALESCE(p.id_vat,''), p.purchase_price, p.manage_variant,
		p.gross_weight, p.net_weight, p.dim_height, p.dim_length, p.dim_width, p.box_qty,
		COALESCE(p.manufacturer,''), COALESCE(p.default_storage,''),
		p.create_date, p.modify_date
	FROM products p
	LEFT JOIN product_groups pg ON pg.id = p.id_material_group
	LEFT JOIN units_of_measure u ON u.um = p.um`

func (r *productRepository) ListProducts(ctx context.Context, q models.ProductListQuery) ([]models.Product, error) {
	sql := productSelectSQL + ` WHERE p.company_id = $1 AND p.deleted = false`
	args := []any{defaultCompanyID}
	n := 2

	if s := strings.TrimSpace(q.Description); s != "" {
		sql += fmt.Sprintf(` AND (p.description ILIKE $%d OR p.description_it ILIKE $%d)`, n, n)
		args = append(args, "%"+s+"%")
		n++
	}
	if q.SalesEnabled != nil {
		sql += fmt.Sprintf(` AND p.sales_enabled = $%d`, n)
		args = append(args, *q.SalesEnabled)
		n++
	}
	if q.Locked != nil {
		sql += fmt.Sprintf(` AND p.locked = $%d`, n)
		args = append(args, *q.Locked)
		n++
	}
	if q.StockEnabled != nil {
		sql += fmt.Sprintf(` AND p.stock_enabled = $%d`, n)
		args = append(args, *q.StockEnabled)
		n++
	}
	if q.PricelistEnabled != nil {
		sql += fmt.Sprintf(` AND p.pricelist_enabled = $%d`, n)
		args = append(args, *q.PricelistEnabled)
		n++
	}
	if s := strings.TrimSpace(q.CreateDateMillis); s != "" {
		if ms, err := strconv.ParseInt(s, 10, 64); err == nil {
			sql += fmt.Sprintf(` AND p.create_date >= to_timestamp($%d::double precision / 1000.0)`, n)
			args = append(args, ms)
			n++
		}
	}
	if s := strings.TrimSpace(q.ModifyDateMillis); s != "" {
		if ms, err := strconv.ParseInt(s, 10, 64); err == nil {
			sql += fmt.Sprintf(` AND p.modify_date >= to_timestamp($%d::double precision / 1000.0)`, n)
			args = append(args, ms)
			n++
		}
	}
	sql += fmt.Sprintf(` ORDER BY p.id LIMIT $%d OFFSET $%d`, n, n+1)
	args = append(args, limitOrDefault(q.Limit), q.Offset)

	rows, err := r.db.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("list products: %w", err)
	}
	defer rows.Close()
	return scanProducts(rows)
}

func (r *productRepository) GetProduct(ctx context.Context, id string) (*models.Product, error) {
	row := r.db.QueryRow(ctx, productSelectSQL+` WHERE p.id = $1 AND p.company_id = $2 AND p.deleted = false`,
		strings.TrimSpace(id), defaultCompanyID)
	p, err := scanProductRow(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("get product: %w", err)
	}
	return p, nil
}

func (r *productRepository) CreateProduct(ctx context.Context, p *models.Product) (*models.Product, error) {
	var nextID int64
	if err := r.db.QueryRow(ctx, `SELECT nextval('product_code_seq')`).Scan(&nextID); err != nil {
		return nil, fmt.Errorf("next product id: %w", err)
	}
	p.ID = fmt.Sprintf("M%05d", nextID)
	if p.Description == "" {
		p.Description = p.DescriptionIT
	}
	if p.Description == "" {
		return nil, fmt.Errorf("description is required")
	}

	row := r.db.QueryRow(ctx, `
		INSERT INTO products (
			id, barcode, basic_code, id_type, type_desc, description, description_it, description_en, description_es, description_bg,
			note, id_material_group, um, sales_enabled, locked, stock_enabled, pricelist_enabled,
			sales_price, sales_price_include_vat, id_vat, purchase_price, manage_variant,
			gross_weight, net_weight, dim_height, dim_length, dim_width, box_qty, manufacturer, default_storage
		) VALUES (
			$1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21,$22,$23,$24,$25,$26,$27,$28,$29,$30
		) RETURNING create_date, modify_date
	`, p.ID, emptyAsNull(p.Barcode), emptyAsNull(p.BasicCode), defaultInt32(p.IDType, 1), emptyAsNull(p.TypeDesc),
		p.Description, emptyAsNull(p.DescriptionIT), emptyAsNull(p.DescriptionEN), emptyAsNull(p.DescriptionES), emptyAsNull(p.DescriptionBG),
		emptyAsNull(p.Note), p.IDMaterialGroup, emptyAsNull(p.Um), p.SalesEnabled, p.Locked, p.StockEnabled, p.PricelistEnabled,
		p.SalesPrice, p.SalesPriceIncludeVat, emptyAsNull(p.IDVat), p.PurchasePrice, p.ManageVariant,
		p.GrossWeight, p.NetWeight, p.DimHeight, p.DimLength, p.DimWidth, p.BoxQty, emptyAsNull(p.Manufacturer), emptyAsNull(p.DefaultStorage))

	var created, modified time.Time
	if err := row.Scan(&created, &modified); err != nil {
		return nil, fmt.Errorf("create product: %w", err)
	}
	p.CreateDate = models.MillisPtr(created)
	p.ModifyDate = models.MillisPtr(modified)
	return r.GetProduct(ctx, p.ID)
}

func (r *productRepository) UpdateProduct(ctx context.Context, p *models.Product) (*models.Product, error) {
	if strings.TrimSpace(p.ID) == "" {
		return nil, fmt.Errorf("product id is required")
	}
	tag, err := r.db.Exec(ctx, `
		UPDATE products SET
			barcode=$2, basic_code=$3, id_type=$4, type_desc=$5, description=$6,
			description_it=$7, description_en=$8, description_es=$9, description_bg=$10, note=$11,
			id_material_group=$12, um=$13, sales_enabled=$14, locked=$15, stock_enabled=$16, pricelist_enabled=$17,
			sales_price=$18, sales_price_include_vat=$19, id_vat=$20, purchase_price=$21, manage_variant=$22,
			gross_weight=$23, net_weight=$24, dim_height=$25, dim_length=$26, dim_width=$27, box_qty=$28,
			manufacturer=$29, default_storage=$30, modify_date=now()
		WHERE id=$1 AND company_id=$31 AND deleted=false
	`, p.ID, emptyAsNull(p.Barcode), emptyAsNull(p.BasicCode), defaultInt32(p.IDType, 1), emptyAsNull(p.TypeDesc), p.Description,
		emptyAsNull(p.DescriptionIT), emptyAsNull(p.DescriptionEN), emptyAsNull(p.DescriptionES), emptyAsNull(p.DescriptionBG), emptyAsNull(p.Note),
		p.IDMaterialGroup, emptyAsNull(p.Um), p.SalesEnabled, p.Locked, p.StockEnabled, p.PricelistEnabled,
		p.SalesPrice, p.SalesPriceIncludeVat, emptyAsNull(p.IDVat), p.PurchasePrice, p.ManageVariant,
		p.GrossWeight, p.NetWeight, p.DimHeight, p.DimLength, p.DimWidth, p.BoxQty, emptyAsNull(p.Manufacturer), emptyAsNull(p.DefaultStorage), defaultCompanyID)
	if err != nil {
		return nil, fmt.Errorf("update product: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return nil, fmt.Errorf("product not found")
	}
	return r.GetProduct(ctx, p.ID)
}

func (r *productRepository) ListVariants(ctx context.Context, productID string) ([]models.AttributeCombination, error) {
	rows, err := r.db.Query(ctx, `
		SELECT product_id, id, COALESCE(attribute_ids_concat,''), COALESCE(barcode,''), delta_price, COALESCE(description,'')
		FROM product_variants
		WHERE product_id = $1
		ORDER BY id
	`, strings.TrimSpace(productID))
	if err != nil {
		return nil, fmt.Errorf("list variants: %w", err)
	}
	defer rows.Close()
	var out []models.AttributeCombination
	for rows.Next() {
		var v models.AttributeCombination
		if err := rows.Scan(&v.IDMaterial, &v.IDAttributeCombination, &v.AttributeIdsConcat, &v.Barcode, &v.DeltaPrice, &v.Description); err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, rows.Err()
}

func (r *productRepository) UpdateVariants(ctx context.Context, variants []models.AttributeCombination) error {
	if len(variants) == 0 {
		return fmt.Errorf("variants are required")
	}
	productID := strings.TrimSpace(variants[0].IDMaterial)
	if productID == "" {
		return fmt.Errorf("idMaterial is required")
	}
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, `DELETE FROM product_variants WHERE product_id = $1`, productID); err != nil {
		return fmt.Errorf("clear variants: %w", err)
	}
	for _, v := range variants {
		if _, err := tx.Exec(ctx, `
			INSERT INTO product_variants (product_id, attribute_ids_concat, barcode, delta_price, description)
			VALUES ($1,$2,$3,$4,$5)
		`, productID, emptyAsNull(v.AttributeIdsConcat), emptyAsNull(v.Barcode), v.DeltaPrice, emptyAsNull(v.Description)); err != nil {
			return fmt.Errorf("insert variant: %w", err)
		}
	}
	if _, err := tx.Exec(ctx, `UPDATE products SET manage_variant=true, modify_date=now() WHERE id=$1`, productID); err != nil {
		return fmt.Errorf("mark variants: %w", err)
	}
	return tx.Commit(ctx)
}

func (r *productRepository) LinkMedia(ctx context.Context, productID, mediaID string) error {
	tag, err := r.db.Exec(ctx, `
		UPDATE products SET repo_media_id=$2, modify_date=now()
		WHERE id=$1 AND company_id=$3 AND deleted=false
	`, strings.TrimSpace(productID), strings.TrimSpace(mediaID), defaultCompanyID)
	if err != nil {
		return fmt.Errorf("link media: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("product not found")
	}
	return nil
}

func scanProducts(rows pgx.Rows) ([]models.Product, error) {
	var out []models.Product
	for rows.Next() {
		p, err := scanProductFromScanner(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *p)
	}
	return out, rows.Err()
}

func scanProductRow(row pgx.Row) (*models.Product, error) {
	p, err := scanProductFromScanner(row)
	if err != nil {
		return nil, err
	}
	return p, nil
}

func scanProductFromScanner(s interface{ Scan(...any) error }) (*models.Product, error) {
	var p models.Product
	var groupID *int32
	var created, modified time.Time
	if err := s.Scan(
		&p.ID, &p.Barcode, &p.BasicCode, &p.IDType, &p.TypeDesc,
		&p.Description, &p.DescriptionIT, &p.DescriptionEN, &p.DescriptionES, &p.DescriptionBG,
		&p.Note, &groupID, &p.MaterialGroupDesc,
		&p.Um, &p.UmDescription,
		&p.SalesEnabled, &p.Locked, &p.StockEnabled, &p.PricelistEnabled,
		&p.SalesPrice, &p.SalesPriceIncludeVat, &p.IDVat, &p.PurchasePrice, &p.ManageVariant,
		&p.GrossWeight, &p.NetWeight, &p.DimHeight, &p.DimLength, &p.DimWidth, &p.BoxQty,
		&p.Manufacturer, &p.DefaultStorage,
		&created, &modified,
	); err != nil {
		return nil, err
	}
	p.IDMaterialGroup = groupID
	p.CreateDate = models.MillisPtr(created)
	p.ModifyDate = models.MillisPtr(modified)
	return &p, nil
}

func defaultInt32(v, def int32) int32 {
	if v == 0 {
		return def
	}
	return v
}
