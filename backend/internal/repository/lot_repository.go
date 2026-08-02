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

// LotRepository persists product lots.
type LotRepository interface {
	ListLots(ctx context.Context, q models.LotListQuery) ([]models.Lot, error)
	GetLot(ctx context.Context, idLot string) (*models.Lot, error)
	CreateLot(ctx context.Context, lot *models.Lot) (*models.Lot, error)
	UpdateLot(ctx context.Context, idLot string, lot *models.Lot) (*models.Lot, error)
	DeleteLot(ctx context.Context, idLot string) error
}

type lotRepository struct {
	db *pgxpool.Pool
}

func NewLotRepository(db *pgxpool.Pool) LotRepository {
	return &lotRepository{db: db}
}

const lotSelectSQL = `
	SELECT id_lot, COALESCE(id_material,''), incoming_date, expire_date, COALESCE(reference,'')
	FROM lots WHERE 1=1`

func (r *lotRepository) ListLots(ctx context.Context, q models.LotListQuery) ([]models.Lot, error) {
	sql := lotSelectSQL
	args := []any{}
	n := 1
	if s := strings.TrimSpace(q.IDMaterial); s != "" {
		sql += fmt.Sprintf(` AND id_material = $%d`, n)
		args = append(args, s)
		n++
	}
	if s := strings.TrimSpace(q.IDLot); s != "" {
		sql += fmt.Sprintf(` AND id_lot = $%d`, n)
		args = append(args, s)
		n++
	}
	sql += fmt.Sprintf(` ORDER BY id_lot LIMIT $%d OFFSET $%d`, n, n+1)
	args = append(args, limitOrDefault(q.Limit), q.Offset)

	rows, err := r.db.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("list lots: %w", err)
	}
	defer rows.Close()
	return scanLots(rows)
}

func (r *lotRepository) GetLot(ctx context.Context, idLot string) (*models.Lot, error) {
	row := r.db.QueryRow(ctx, lotSelectSQL+` AND id_lot = $1`, strings.TrimSpace(idLot))
	lot, err := scanLotRow(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("get lot: %w", err)
	}
	return lot, nil
}

func (r *lotRepository) CreateLot(ctx context.Context, lot *models.Lot) (*models.Lot, error) {
	lot.IDLot = strings.TrimSpace(lot.IDLot)
	if lot.IDLot == "" {
		return nil, fmt.Errorf("idLot is required")
	}
	_, err := r.db.Exec(ctx, `
		INSERT INTO lots (id_lot, id_material, incoming_date, expire_date, reference)
		VALUES ($1,$2,$3,$4,$5)
	`, lot.IDLot, emptyAsNull(lot.IDMaterial), millisToTimePtr(lot.IncomingDate), millisToTimePtr(lot.ExpireDate), emptyAsNull(lot.Reference))
	if err != nil {
		return nil, fmt.Errorf("create lot: %w", err)
	}
	return r.GetLot(ctx, lot.IDLot)
}

func (r *lotRepository) UpdateLot(ctx context.Context, idLot string, lot *models.Lot) (*models.Lot, error) {
	tag, err := r.db.Exec(ctx, `
		UPDATE lots SET id_material=$2, incoming_date=$3, expire_date=$4, reference=$5
		WHERE id_lot=$1
	`, strings.TrimSpace(idLot), emptyAsNull(lot.IDMaterial), millisToTimePtr(lot.IncomingDate), millisToTimePtr(lot.ExpireDate), emptyAsNull(lot.Reference))
	if err != nil {
		return nil, fmt.Errorf("update lot: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return nil, fmt.Errorf("lot not found")
	}
	return r.GetLot(ctx, idLot)
}

func (r *lotRepository) DeleteLot(ctx context.Context, idLot string) error {
	tag, err := r.db.Exec(ctx, `DELETE FROM lots WHERE id_lot=$1`, strings.TrimSpace(idLot))
	if err != nil {
		return fmt.Errorf("delete lot: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("lot not found")
	}
	return nil
}

func scanLots(rows pgx.Rows) ([]models.Lot, error) {
	var out []models.Lot
	for rows.Next() {
		lot, err := scanLotFromScanner(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *lot)
	}
	return out, rows.Err()
}

func scanLotRow(row pgx.Row) (*models.Lot, error) {
	return scanLotFromScanner(row)
}

func scanLotFromScanner(s interface{ Scan(...any) error }) (*models.Lot, error) {
	var lot models.Lot
	var incoming, expire *time.Time
	if err := s.Scan(&lot.IDLot, &lot.IDMaterial, &incoming, &expire, &lot.Reference); err != nil {
		return nil, err
	}
	if incoming != nil {
		lot.IncomingDate = models.MillisPtr(*incoming)
	}
	if expire != nil {
		lot.ExpireDate = models.MillisPtr(*expire)
	}
	return &lot, nil
}
