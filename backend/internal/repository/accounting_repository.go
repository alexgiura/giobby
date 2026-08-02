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

// AccountingRepository persists manual account movements.
type AccountingRepository interface {
	ListAccountMovements(ctx context.Context, q models.AccountMovementListQuery) ([]models.AccountMovementRegistration, error)
	GetAccountMovement(ctx context.Context, idDoc int32) (*models.AccountMovementRegistration, error)
	CreateAccountMovement(ctx context.Context, reg *models.AccountMovementRegistration) (*models.AccountMovementRegistration, error)
	UpdateAccountMovement(ctx context.Context, reg *models.AccountMovementRegistration) (*models.AccountMovementRegistration, error)
	DeleteAccountMovement(ctx context.Context, idDoc int32) error
}

type accountingRepository struct {
	db *pgxpool.Pool
}

func NewAccountingRepository(db *pgxpool.Pool) AccountingRepository {
	return &accountingRepository{db: db}
}

const accountMovementDocSQL = `
	SELECT id, COALESCE(id_bu,''), COALESCE(id_bup,''), COALESCE(id_numerator,0), COALESCE(id_accountmovement_type,0),
		doc_date, reg_date, id_account_center, COALESCE(description,''), COALESCE(entry_type,'')
	FROM account_movement_documents WHERE company_id = $1 AND deleted = false`

func (r *accountingRepository) ListAccountMovements(ctx context.Context, q models.AccountMovementListQuery) ([]models.AccountMovementRegistration, error) {
	sql := accountMovementDocSQL
	args := []any{defaultCompanyID}
	n := 2

	if q.IDDoc != nil {
		sql += fmt.Sprintf(` AND id = $%d`, n)
		args = append(args, *q.IDDoc)
		n++
	}
	if s := strings.TrimSpace(q.EntryType); s != "" {
		sql += fmt.Sprintf(` AND entry_type = $%d`, n)
		args = append(args, s)
		n++
	}
	if s := strings.TrimSpace(q.IDBu); s != "" {
		sql += fmt.Sprintf(` AND id_bu = $%d`, n)
		args = append(args, s)
		n++
	}
	if s := strings.TrimSpace(q.IDBup); s != "" {
		sql += fmt.Sprintf(` AND id_bup = $%d`, n)
		args = append(args, s)
		n++
	}
	if q.AccountCenter != nil {
		sql += fmt.Sprintf(` AND id_account_center = $%d`, n)
		args = append(args, *q.AccountCenter)
		n++
	}
	if s := strings.TrimSpace(q.FreeText); s != "" {
		sql += fmt.Sprintf(` AND (description ILIKE $%d OR EXISTS (
			SELECT 1 FROM account_movement_rows r WHERE r.document_id = account_movement_documents.id
			AND (r.description ILIKE $%d OR r.account_code ILIKE $%d)
		))`, n, n, n)
		args = append(args, "%"+s+"%")
		n++
	}
	if s := strings.TrimSpace(q.FromDateMillis); s != "" {
		if ms, err := parseMillisFilter(s); err == nil {
			sql += fmt.Sprintf(` AND doc_date >= $%d`, n)
			args = append(args, ms)
			n++
		}
	}
	if s := strings.TrimSpace(q.ToDateMillis); s != "" {
		if ms, err := parseMillisFilter(s); err == nil {
			sql += fmt.Sprintf(` AND doc_date <= $%d`, n)
			args = append(args, ms)
			n++
		}
	}
	if !q.ShowZero {
		sql += ` AND EXISTS (
			SELECT 1 FROM account_movement_rows r WHERE r.document_id = account_movement_documents.id AND r.amount <> 0
		)`
	}

	sql += fmt.Sprintf(` ORDER BY id DESC LIMIT $%d OFFSET $%d`, n, n+1)
	args = append(args, limitOrDefault(q.Limit), q.Offset)

	rows, err := r.db.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("list account movements: %w", err)
	}
	defer rows.Close()

	out := make([]models.AccountMovementRegistration, 0)
	for rows.Next() {
		reg, err := scanAccountMovementDoc(rows)
		if err != nil {
			return nil, err
		}
		reg.Rows, err = r.loadRows(ctx, reg.IDDoc)
		if err != nil {
			return nil, err
		}
		out = append(out, *reg)
	}
	return out, rows.Err()
}

func (r *accountingRepository) GetAccountMovement(ctx context.Context, idDoc int32) (*models.AccountMovementRegistration, error) {
	row := r.db.QueryRow(ctx, accountMovementDocSQL+` AND id = $2`, defaultCompanyID, idDoc)
	reg, err := scanAccountMovementDoc(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("get account movement: %w", err)
	}
	reg.Rows, err = r.loadRows(ctx, idDoc)
	if err != nil {
		return nil, err
	}
	return reg, nil
}

func (r *accountingRepository) CreateAccountMovement(ctx context.Context, reg *models.AccountMovementRegistration) (*models.AccountMovementRegistration, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	var id int32
	err = tx.QueryRow(ctx, `
		INSERT INTO account_movement_documents (
			id_bu, id_bup, id_numerator, id_accountmovement_type, doc_date, reg_date,
			id_account_center, description, entry_type
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9) RETURNING id
	`, emptyAsNull(reg.Bu), emptyAsNull(reg.Bup), nullInt32(reg.IDNumerator), nullInt32(reg.IDAccountmovementType),
		models.MillisToTime(reg.DocDate), models.MillisToTime(reg.RegDate),
		reg.IDAccountCenter, emptyAsNull(reg.Description), inferEntryType(reg)).Scan(&id)
	if err != nil {
		return nil, fmt.Errorf("insert account movement: %w", err)
	}
	if err := r.insertRows(ctx, tx, id, reg.Rows); err != nil {
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return r.GetAccountMovement(ctx, id)
}

func (r *accountingRepository) UpdateAccountMovement(ctx context.Context, reg *models.AccountMovementRegistration) (*models.AccountMovementRegistration, error) {
	if reg.IDDoc <= 0 {
		return nil, fmt.Errorf("idDoc is required")
	}
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	tag, err := tx.Exec(ctx, `
		UPDATE account_movement_documents SET
			id_bu=$2, id_bup=$3, id_numerator=$4, id_accountmovement_type=$5,
			doc_date=$6, reg_date=$7, id_account_center=$8, description=$9, entry_type=$10, updated_at=now()
		WHERE id=$1 AND company_id=$11 AND deleted=false
	`, reg.IDDoc, emptyAsNull(reg.Bu), emptyAsNull(reg.Bup), nullInt32(reg.IDNumerator), nullInt32(reg.IDAccountmovementType),
		models.MillisToTime(reg.DocDate), models.MillisToTime(reg.RegDate),
		reg.IDAccountCenter, emptyAsNull(reg.Description), inferEntryType(reg), defaultCompanyID)
	if err != nil {
		return nil, err
	}
	if tag.RowsAffected() == 0 {
		return nil, fmt.Errorf("account movement not found")
	}
	if _, err := tx.Exec(ctx, `DELETE FROM account_movement_rows WHERE document_id=$1`, reg.IDDoc); err != nil {
		return nil, err
	}
	if err := r.insertRows(ctx, tx, reg.IDDoc, reg.Rows); err != nil {
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return r.GetAccountMovement(ctx, reg.IDDoc)
}

func (r *accountingRepository) DeleteAccountMovement(ctx context.Context, idDoc int32) error {
	tag, err := r.db.Exec(ctx, `
		UPDATE account_movement_documents SET deleted=true, updated_at=now()
		WHERE id=$1 AND company_id=$2 AND deleted=false
	`, idDoc, defaultCompanyID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("account movement not found")
	}
	return nil
}

func (r *accountingRepository) loadRows(ctx context.Context, docID int32) ([]models.AccountMovement, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, COALESCE(description,''), COALESCE(from_account_code,''), COALESCE(to_account_code,''),
			COALESCE(account_code,''), COALESCE(account_sign,''), amount
		FROM account_movement_rows WHERE document_id=$1 ORDER BY line_no
	`, docID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]models.AccountMovement, 0)
	for rows.Next() {
		var m models.AccountMovement
		if err := rows.Scan(&m.ID, &m.Description, &m.FromAccountCode, &m.ToAccountCode,
			&m.AccountCode, &m.AccountSign, &m.Amount); err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

func (r *accountingRepository) insertRows(ctx context.Context, tx pgx.Tx, docID int32, rows []models.AccountMovement) error {
	for i, row := range rows {
		lineNo := int32(i + 1)
		if _, err := tx.Exec(ctx, `
			INSERT INTO account_movement_rows (
				document_id, line_no, description, from_account_code, to_account_code,
				account_code, account_sign, amount
			) VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
		`, docID, lineNo, emptyAsNull(row.Description), emptyAsNull(row.FromAccountCode), emptyAsNull(row.ToAccountCode),
			emptyAsNull(row.AccountCode), emptyAsNull(row.AccountSign), row.Amount); err != nil {
			return err
		}
	}
	return nil
}

func scanAccountMovementDoc(s interface{ Scan(...any) error }) (*models.AccountMovementRegistration, error) {
	var reg models.AccountMovementRegistration
	var docDate, regDate time.Time
	var idAccountCenter *int32
	var entryType string
	if err := s.Scan(&reg.IDDoc, &reg.Bu, &reg.Bup, &reg.IDNumerator, &reg.IDAccountmovementType,
		&docDate, &regDate, &idAccountCenter, &reg.Description, &entryType); err != nil {
		return nil, err
	}
	_ = entryType
	reg.DocDate = models.MillisPtr(docDate)
	reg.RegDate = models.MillisPtr(regDate)
	reg.IDAccountCenter = idAccountCenter
	return &reg, nil
}

func inferEntryType(reg *models.AccountMovementRegistration) string {
	// Default economic entry; P for patrimonial if account codes suggest it.
	for _, row := range reg.Rows {
		if strings.HasPrefix(row.AccountCode, "1") || strings.HasPrefix(row.AccountCode, "2") {
			return "P"
		}
	}
	return "E"
}

func parseMillisFilter(s string) (time.Time, error) {
	ms, err := parseInt64String(s)
	if err != nil {
		return time.Time{}, err
	}
	return time.UnixMilli(ms), nil
}

func parseInt64String(s string) (int64, error) {
	var n int64
	_, err := fmt.Sscan(s, &n)
	return n, err
}
