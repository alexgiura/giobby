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

// MachineDataTrackingRepository persists production machine tracking records.
type MachineDataTrackingRepository interface {
	List(ctx context.Context, q models.MachineDataTrackingListQuery) ([]models.MachineDataTracking, error)
	Get(ctx context.Context, id int32) (*models.MachineDataTracking, error)
	Create(ctx context.Context, item *models.MachineDataTracking) (*models.MachineDataTracking, error)
	CreateBatch(ctx context.Context, items []models.MachineDataTracking) ([]models.MachineDataTracking, error)
	Update(ctx context.Context, id int32, item *models.MachineDataTracking) (*models.MachineDataTracking, error)
	Delete(ctx context.Context, id int32) error
}

type machineDataTrackingRepository struct {
	db *pgxpool.Pool
}

func NewMachineDataTrackingRepository(db *pgxpool.Pool) MachineDataTrackingRepository {
	return &machineDataTrackingRepository{db: db}
}

const mdtSelect = `
	SELECT id, COALESCE(id_machine,''), COALESCE(id_operator,''), COALESCE(materials,''), COALESCE(id_lot,''),
		box_qty, box_number, production_waste, COALESCE(barcode,''), COALESCE(start_time,''), COALESCE(end_time,''),
		COALESCE(duration,''), COALESCE(machine_status,''), COALESCE(alarm_code,''), COALESCE(alarm_time,''),
		COALESCE(track_date,''), COALESCE(id_attachment,'')
	FROM machine_data_tracking WHERE company_id = $1`

func (r *machineDataTrackingRepository) List(ctx context.Context, q models.MachineDataTrackingListQuery) ([]models.MachineDataTracking, error) {
	sql := mdtSelect
	args := []any{defaultCompanyID}
	n := 2

	if q.Deleted != nil {
		sql += fmt.Sprintf(` AND deleted = $%d`, n)
		args = append(args, *q.Deleted)
		n++
	} else {
		sql += ` AND deleted = false`
	}
	if q.ID != nil {
		sql += fmt.Sprintf(` AND id = $%d`, n)
		args = append(args, *q.ID)
		n++
	}
	if s := strings.TrimSpace(q.IDMachine); s != "" {
		sql += fmt.Sprintf(` AND id_machine = $%d`, n)
		args = append(args, s)
		n++
	}
	if s := strings.TrimSpace(q.IDLot); s != "" {
		sql += fmt.Sprintf(` AND id_lot = $%d`, n)
		args = append(args, s)
		n++
	}
	sql += fmt.Sprintf(` ORDER BY id DESC LIMIT $%d OFFSET $%d`, n, n+1)
	args = append(args, limitOrDefault(q.Limit), q.Offset)

	rows, err := r.db.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("list machine data tracking: %w", err)
	}
	defer rows.Close()
	return scanMachineDataTrackingRows(rows)
}

func (r *machineDataTrackingRepository) Get(ctx context.Context, id int32) (*models.MachineDataTracking, error) {
	row := r.db.QueryRow(ctx, mdtSelect+` AND id = $2 AND deleted = false`, defaultCompanyID, id)
	item, err := scanMachineDataTrackingRow(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("get machine data tracking: %w", err)
	}
	return item, nil
}

func (r *machineDataTrackingRepository) Create(ctx context.Context, item *models.MachineDataTracking) (*models.MachineDataTracking, error) {
	row := r.db.QueryRow(ctx, `
		INSERT INTO machine_data_tracking (
			id_machine, id_operator, materials, id_lot, box_qty, box_number, production_waste,
			barcode, start_time, end_time, duration, machine_status, alarm_code, alarm_time, track_date, id_attachment
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16)
		RETURNING id
	`, emptyAsNull(item.IDMachine), emptyAsNull(item.IDOperator), emptyAsNull(item.Materials), emptyAsNull(item.IDLot),
		item.BoxQty, item.BoxNumber, item.ProductionWaste, emptyAsNull(item.Barcode), emptyAsNull(item.StartTime),
		emptyAsNull(item.EndTime), emptyAsNull(item.Duration), emptyAsNull(item.MachineStatus), emptyAsNull(item.AlarmCode),
		emptyAsNull(item.AlarmTime), emptyAsNull(item.Date), emptyAsNull(item.IDAttachment))
	var id int32
	if err := row.Scan(&id); err != nil {
		return nil, fmt.Errorf("create machine data tracking: %w", err)
	}
	return r.Get(ctx, id)
}

func (r *machineDataTrackingRepository) CreateBatch(ctx context.Context, items []models.MachineDataTracking) ([]models.MachineDataTracking, error) {
	if len(items) == 0 {
		return nil, fmt.Errorf("list is required")
	}
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	var out []models.MachineDataTracking
	for _, item := range items {
		row := tx.QueryRow(ctx, `
			INSERT INTO machine_data_tracking (
				id_machine, id_operator, materials, id_lot, box_qty, box_number, production_waste,
				barcode, start_time, end_time, duration, machine_status, alarm_code, alarm_time, track_date, id_attachment
			) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16)
			RETURNING id
		`, emptyAsNull(item.IDMachine), emptyAsNull(item.IDOperator), emptyAsNull(item.Materials), emptyAsNull(item.IDLot),
			item.BoxQty, item.BoxNumber, item.ProductionWaste, emptyAsNull(item.Barcode), emptyAsNull(item.StartTime),
			emptyAsNull(item.EndTime), emptyAsNull(item.Duration), emptyAsNull(item.MachineStatus), emptyAsNull(item.AlarmCode),
			emptyAsNull(item.AlarmTime), emptyAsNull(item.Date), emptyAsNull(item.IDAttachment))
		var id int32
		if err := row.Scan(&id); err != nil {
			return nil, fmt.Errorf("create batch item: %w", err)
		}
		item.ID = id
		out = append(out, item)
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return out, nil
}

func (r *machineDataTrackingRepository) Update(ctx context.Context, id int32, item *models.MachineDataTracking) (*models.MachineDataTracking, error) {
	tag, err := r.db.Exec(ctx, `
		UPDATE machine_data_tracking SET
			id_machine=$2, id_operator=$3, materials=$4, id_lot=$5, box_qty=$6, box_number=$7, production_waste=$8,
			barcode=$9, start_time=$10, end_time=$11, duration=$12, machine_status=$13, alarm_code=$14,
			alarm_time=$15, track_date=$16, id_attachment=$17
		WHERE id=$1 AND company_id=$18 AND deleted=false
	`, id, emptyAsNull(item.IDMachine), emptyAsNull(item.IDOperator), emptyAsNull(item.Materials), emptyAsNull(item.IDLot),
		item.BoxQty, item.BoxNumber, item.ProductionWaste, emptyAsNull(item.Barcode), emptyAsNull(item.StartTime),
		emptyAsNull(item.EndTime), emptyAsNull(item.Duration), emptyAsNull(item.MachineStatus), emptyAsNull(item.AlarmCode),
		emptyAsNull(item.AlarmTime), emptyAsNull(item.Date), emptyAsNull(item.IDAttachment), defaultCompanyID)
	if err != nil {
		return nil, fmt.Errorf("update machine data tracking: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return nil, fmt.Errorf("machine data tracking not found")
	}
	return r.Get(ctx, id)
}

func (r *machineDataTrackingRepository) Delete(ctx context.Context, id int32) error {
	tag, err := r.db.Exec(ctx, `
		UPDATE machine_data_tracking SET deleted=true WHERE id=$1 AND company_id=$2 AND deleted=false
	`, id, defaultCompanyID)
	if err != nil {
		return fmt.Errorf("delete machine data tracking: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("machine data tracking not found")
	}
	return nil
}

func scanMachineDataTrackingRows(rows pgx.Rows) ([]models.MachineDataTracking, error) {
	var out []models.MachineDataTracking
	for rows.Next() {
		item, err := scanMachineDataTrackingFromScanner(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *item)
	}
	return out, rows.Err()
}

func scanMachineDataTrackingRow(row pgx.Row) (*models.MachineDataTracking, error) {
	return scanMachineDataTrackingFromScanner(row)
}

func scanMachineDataTrackingFromScanner(s interface{ Scan(...any) error }) (*models.MachineDataTracking, error) {
	var item models.MachineDataTracking
	if err := s.Scan(&item.ID, &item.IDMachine, &item.IDOperator, &item.Materials, &item.IDLot,
		&item.BoxQty, &item.BoxNumber, &item.ProductionWaste, &item.Barcode, &item.StartTime, &item.EndTime,
		&item.Duration, &item.MachineStatus, &item.AlarmCode, &item.AlarmTime, &item.Date, &item.IDAttachment); err != nil {
		return nil, err
	}
	return &item, nil
}
