package repository

import (
	"context"
	"fmt"
	"strings"

	"dnsc_microservice/internal/models"

	"github.com/jackc/pgx/v4/pgxpool"
)

// StockRepository persists stock movements and availability queries.
type StockRepository interface {
	CreateStock(ctx context.Context, s *models.Stock) (*models.Stock, error)
	UpdateStock(ctx context.Context, s *models.Stock) (*models.Stock, error)
	ListAvailability(ctx context.Context, q models.StockAvailabilityQuery) ([]models.StockAvailability, error)
	ListAvailabilityReport(ctx context.Context, q models.StockAvailabilityReportQuery) ([]models.StockAvailability, error)
}

type stockRepository struct {
	db *pgxpool.Pool
}

func NewStockRepository(db *pgxpool.Pool) StockRepository {
	return &stockRepository{db: db}
}

func (r *stockRepository) CreateStock(ctx context.Context, s *models.Stock) (*models.Stock, error) {
	if s.IDDocumentType == 0 {
		s.IDDocumentType = 109
	}
	row := r.db.QueryRow(ctx, `
		INSERT INTO stocks (
			id_storage, id_location, id_lot, id_material, um, quantity, value,
			number_of_containers, stock_state, id_document_type, id_attribute_combination,
			attribute_combination_desc, storage_desc
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)
		RETURNING id
	`, emptyAsNull(s.IDStorage), emptyAsNull(s.IDLocation), emptyAsNull(s.IDLot), emptyAsNull(s.IDMaterial),
		emptyAsNull(s.Um), s.Quantity, s.Value, s.NumberOfContainers, s.StockState, s.IDDocumentType,
		s.IDAttributeCombination, emptyAsNull(s.AttributeCombinationDesc), emptyAsNull(s.StorageDesc))
	var id int32
	if err := row.Scan(&id); err != nil {
		return nil, fmt.Errorf("create stock: %w", err)
	}
	s.ID = id
	return s, nil
}

func (r *stockRepository) UpdateStock(ctx context.Context, s *models.Stock) (*models.Stock, error) {
	if s.ID <= 0 {
		return nil, fmt.Errorf("stock id is required")
	}
	tag, err := r.db.Exec(ctx, `
		UPDATE stocks SET
			id_storage=$2, id_location=$3, id_lot=$4, id_material=$5, um=$6, quantity=$7, value=$8,
			number_of_containers=$9, stock_state=$10, id_document_type=$11, id_attribute_combination=$12,
			attribute_combination_desc=$13, storage_desc=$14
		WHERE id=$1
	`, s.ID, emptyAsNull(s.IDStorage), emptyAsNull(s.IDLocation), emptyAsNull(s.IDLot), emptyAsNull(s.IDMaterial),
		emptyAsNull(s.Um), s.Quantity, s.Value, s.NumberOfContainers, s.StockState, s.IDDocumentType,
		s.IDAttributeCombination, emptyAsNull(s.AttributeCombinationDesc), emptyAsNull(s.StorageDesc))
	if err != nil {
		return nil, fmt.Errorf("update stock: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return nil, fmt.Errorf("stock not found")
	}
	return s, nil
}

func (r *stockRepository) ListAvailability(ctx context.Context, q models.StockAvailabilityQuery) ([]models.StockAvailability, error) {
	return r.queryAvailability(ctx, q.IDMaterial, q.IDStorage, q.IDLocation, q.ZeroQtyRows, q.Limit, q.Offset, false)
}

func (r *stockRepository) ListAvailabilityReport(ctx context.Context, q models.StockAvailabilityReportQuery) ([]models.StockAvailability, error) {
	includeLot := q.LotsDetail != nil && *q.LotsDetail
	return r.queryAvailability(ctx, q.IDMaterial, q.IDStorage, q.IDLocation, q.ZeroQtyRows, q.Limit, q.Offset, includeLot)
}

func (r *stockRepository) queryAvailability(ctx context.Context, idMaterial, idStorage, idLocation string, zeroQty *bool, limit, offset int, includeLot bool) ([]models.StockAvailability, error) {
	groupLot := ""
	selectLot := `'' AS id_lot`
	if includeLot {
		groupLot = `, st.id_lot`
		selectLot = `COALESCE(st.id_lot,'')`
	}
	sql := fmt.Sprintf(`
		SELECT st.id_material, COALESCE(p.description,''), COALESCE(st.id_storage,''), COALESCE(st.storage_desc,''),
			COALESCE(st.id_location,''), COALESCE(sl.description,''), %s,
			COALESCE(SUM(st.quantity),0), 0::float8, COALESCE(SUM(st.quantity),0), COALESCE(MAX(st.um),'')
		FROM stocks st
		LEFT JOIN products p ON p.id = st.id_material
		LEFT JOIN storage_locations sl ON sl.id_location = st.id_location
		WHERE 1=1`, selectLot)
	args := []any{}
	n := 1
	if s := strings.TrimSpace(idMaterial); s != "" {
		sql += fmt.Sprintf(` AND st.id_material = $%d`, n)
		args = append(args, s)
		n++
	}
	if s := strings.TrimSpace(idStorage); s != "" {
		sql += fmt.Sprintf(` AND st.id_storage = $%d`, n)
		args = append(args, s)
		n++
	}
	if s := strings.TrimSpace(idLocation); s != "" {
		sql += fmt.Sprintf(` AND st.id_location = $%d`, n)
		args = append(args, s)
		n++
	}
	sql += fmt.Sprintf(` GROUP BY st.id_material, p.description, st.id_storage, st.storage_desc, st.id_location, sl.description%s`, groupLot)
	if zeroQty == nil || !*zeroQty {
		sql += ` HAVING COALESCE(SUM(st.quantity),0) <> 0`
	}
	sql += fmt.Sprintf(` ORDER BY st.id_material LIMIT $%d OFFSET $%d`, n, n+1)
	args = append(args, limitOrDefault(limit), offset)

	rows, err := r.db.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("stock availability: %w", err)
	}
	defer rows.Close()
	var out []models.StockAvailability
	for rows.Next() {
		var a models.StockAvailability
		if err := rows.Scan(&a.IDMaterial, &a.MaterialDesc, &a.IDStorage, &a.StorageDesc,
			&a.IDLocation, &a.LocationDesc, &a.IDLot, &a.Quantity, &a.CommittedQty, &a.AvailableQty, &a.Um); err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, rows.Err()
}
