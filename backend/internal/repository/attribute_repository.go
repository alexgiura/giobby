package repository

import (
	"context"
	"fmt"
	"strings"

	"dnsc_microservice/internal/models"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

// AttributeRepository persists product attributes and characteristics.
type AttributeRepository interface {
	ListAttributes(ctx context.Context, q models.AttributeListQuery) ([]models.Attribute, error)
	CreateAttributeLinks(ctx context.Context, payload models.ProductAttributesPayload) ([]models.Attribute, error)
	ListProductCharacteristics(ctx context.Context) ([]models.ProductAttribute, error)
	CreateProductCharacteristic(ctx context.Context, a *models.ProductAttribute) (*models.ProductAttribute, error)
	UpdateProductCharacteristic(ctx context.Context, a *models.ProductAttribute) (*models.ProductAttribute, error)
}

type attributeRepository struct {
	db *pgxpool.Pool
}

func NewAttributeRepository(db *pgxpool.Pool) AttributeRepository {
	return &attributeRepository{db: db}
}

func (r *attributeRepository) ListAttributes(ctx context.Context, q models.AttributeListQuery) ([]models.Attribute, error) {
	sql := `
		SELECT doc_type, id_doc, id_attribute, sub_id, pos,
			COALESCE(attribute_value,''), COALESCE(attribute_name,''), doc_type_attribute_value, COALESCE(id_lang,'')
		FROM document_attributes WHERE 1=1`
	args := []any{}
	n := 1

	if s := strings.TrimSpace(q.IDDoc); s != "" {
		sql += fmt.Sprintf(` AND id_doc = $%d`, n)
		args = append(args, s)
		n++
	}
	if q.IDDocumentType != nil {
		sql += fmt.Sprintf(` AND doc_type = $%d`, n)
		args = append(args, *q.IDDocumentType)
		n++
	}
	if s := strings.TrimSpace(q.IDLanguage); s != "" {
		sql += fmt.Sprintf(` AND id_lang = $%d`, n)
		args = append(args, s)
		n++
	}
	sql += fmt.Sprintf(` ORDER BY id LIMIT $%d OFFSET $%d`, n, n+1)
	args = append(args, limitOrDefault(q.Limit), q.Offset)

	rows, err := r.db.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("list attributes: %w", err)
	}
	defer rows.Close()
	return scanAttributes(rows)
}

func (r *attributeRepository) CreateAttributeLinks(ctx context.Context, payload models.ProductAttributesPayload) ([]models.Attribute, error) {
	if len(payload.Attributes) == 0 {
		return nil, fmt.Errorf("attributes are required")
	}
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	var out []models.Attribute
	for _, a := range payload.Attributes {
		if strings.TrimSpace(a.IDDoc) == "" {
			return nil, fmt.Errorf("idDoc is required")
		}
		row := tx.QueryRow(ctx, `
			INSERT INTO document_attributes (
				doc_type, id_doc, id_attribute, sub_id, pos, attribute_value, attribute_name, doc_type_attribute_value, id_lang
			) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
			RETURNING doc_type, id_doc, id_attribute, sub_id, pos, attribute_value, attribute_name, doc_type_attribute_value, id_lang
		`, a.DocType, a.IDDoc, a.IDAttribute, a.SubID, a.Pos,
			emptyAsNull(a.AttributeValue), emptyAsNull(a.AttributeName), a.DocTypeAttributeValue, emptyAsNull(a.IDLang))
		created, err := scanAttributeRow(row)
		if err != nil {
			return nil, fmt.Errorf("create attribute link: %w", err)
		}
		out = append(out, *created)
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return out, nil
}

func (r *attributeRepository) ListProductCharacteristics(ctx context.Context) ([]models.ProductAttribute, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, COALESCE(attribute_name_it,''), COALESCE(attribute_name_en,''), COALESCE(attribute_name_es,''), COALESCE(attribute_name_bg,'')
		FROM product_characteristics
		WHERE company_id = $1
		ORDER BY id
	`, defaultCompanyID)
	if err != nil {
		return nil, fmt.Errorf("list product characteristics: %w", err)
	}
	defer rows.Close()
	var out []models.ProductAttribute
	for rows.Next() {
		var a models.ProductAttribute
		if err := rows.Scan(&a.IDAttribute, &a.AttributeNameIT, &a.AttributeNameEN, &a.AttributeNameES, &a.AttributeNameBG); err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

func (r *attributeRepository) CreateProductCharacteristic(ctx context.Context, a *models.ProductAttribute) (*models.ProductAttribute, error) {
	row := r.db.QueryRow(ctx, `
		INSERT INTO product_characteristics (attribute_name_it, attribute_name_en, attribute_name_es, attribute_name_bg)
		VALUES ($1,$2,$3,$4)
		RETURNING id
	`, emptyAsNull(a.AttributeNameIT), emptyAsNull(a.AttributeNameEN), emptyAsNull(a.AttributeNameES), emptyAsNull(a.AttributeNameBG))
	var id int32
	if err := row.Scan(&id); err != nil {
		return nil, fmt.Errorf("create product characteristic: %w", err)
	}
	a.IDAttribute = id
	return a, nil
}

func (r *attributeRepository) UpdateProductCharacteristic(ctx context.Context, a *models.ProductAttribute) (*models.ProductAttribute, error) {
	if a.IDAttribute <= 0 {
		return nil, fmt.Errorf("idAttribute is required")
	}
	tag, err := r.db.Exec(ctx, `
		UPDATE product_characteristics
		SET attribute_name_it=$2, attribute_name_en=$3, attribute_name_es=$4, attribute_name_bg=$5
		WHERE id=$1 AND company_id=$6
	`, a.IDAttribute, emptyAsNull(a.AttributeNameIT), emptyAsNull(a.AttributeNameEN), emptyAsNull(a.AttributeNameES), emptyAsNull(a.AttributeNameBG), defaultCompanyID)
	if err != nil {
		return nil, fmt.Errorf("update product characteristic: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return nil, fmt.Errorf("product characteristic not found")
	}
	return a, nil
}

func scanAttributes(rows pgx.Rows) ([]models.Attribute, error) {
	var out []models.Attribute
	for rows.Next() {
		a, err := scanAttributeFromScanner(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *a)
	}
	return out, rows.Err()
}

func scanAttributeRow(row pgx.Row) (*models.Attribute, error) {
	return scanAttributeFromScanner(row)
}

func scanAttributeFromScanner(s interface{ Scan(...any) error }) (*models.Attribute, error) {
	var a models.Attribute
	if err := s.Scan(&a.DocType, &a.IDDoc, &a.IDAttribute, &a.SubID, &a.Pos,
		&a.AttributeValue, &a.AttributeName, &a.DocTypeAttributeValue, &a.IDLang); err != nil {
		return nil, err
	}
	return &a, nil
}
