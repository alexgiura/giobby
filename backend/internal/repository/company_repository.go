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

// CompanyRepository persists company and settings data.
type CompanyRepository interface {
	GetCompany(ctx context.Context, id int64) (*models.Company, error)

	ListAccountCenters(ctx context.Context, companyID int64, q models.AccountCenterListQuery) ([]models.AccountCenter, error)
	GetAccountCenter(ctx context.Context, companyID int64, id int32) (*models.AccountCenter, error)
	CreateAccountCenter(ctx context.Context, companyID int64, ac *models.AccountCenter) (*models.AccountCenter, error)
	UpdateAccountCenter(ctx context.Context, companyID int64, id int32, ac *models.AccountCenter) (*models.AccountCenter, error)
	DeleteAccountCenter(ctx context.Context, companyID int64, id int32) error

	ListAccountCodes(ctx context.Context, companyID int64, q models.AccountCodeListQuery) ([]models.AccountCode, error)

	ListBanks(ctx context.Context, companyID int64, q models.BankListQuery) ([]models.BankCashdesk, error)
	GetBank(ctx context.Context, companyID int64, id int32) (*models.BankCashdesk, error)
	CreateBank(ctx context.Context, companyID int64, b *models.BankCashdesk) (*models.BankCashdesk, error)
	UpdateBank(ctx context.Context, companyID int64, id int32, b *models.BankCashdesk) (*models.BankCashdesk, error)
	DeleteBank(ctx context.Context, companyID int64, id int32) error

	ListBupNumerators(ctx context.Context, companyID int64, q models.BupNumeratorQuery) ([]models.BupNumerator, error)
	GetCurrencyChange(ctx context.Context, companyID int64, sourceCurrency string, asOf *time.Time) (*models.CurrencyChange, error)

	ListVatRates(ctx context.Context, companyID int64, q models.VatListQuery) ([]models.VatRate, error)
	GetVatRate(ctx context.Context, companyID int64, idVat string, q models.VatGetQuery) (*models.VatRate, error)
}

type companyRepository struct {
	db *pgxpool.Pool
}

func NewCompanyRepository(db *pgxpool.Pool) CompanyRepository {
	return &companyRepository{db: db}
}

func (r *companyRepository) GetCompany(ctx context.Context, id int64) (*models.Company, error) {
	row := r.db.QueryRow(ctx, `
		SELECT id, name, vat_number, email, phone, address, city, zip, country, base_currency
		FROM companies WHERE id = $1
	`, id)
	var c models.Company
	var vat, email, phone, address, city, zip, country, baseCurr *string
	if err := row.Scan(&c.ID, &c.Name, &vat, &email, &phone, &address, &city, &zip, &country, &baseCurr); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("get company: %w", err)
	}
	c.VatNumber = vat
	c.Email = email
	c.Phone = phone
	c.Address = address
	c.City = city
	c.Zip = zip
	c.Country = country
	c.BaseCurrency = baseCurr
	return &c, nil
}

func (r *companyRepository) ListAccountCenters(ctx context.Context, companyID int64, q models.AccountCenterListQuery) ([]models.AccountCenter, error) {
	sql := `SELECT id, name, COALESCE(type,''), COALESCE(description,'') FROM account_centers WHERE company_id = $1`
	args := []any{companyID}
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
		return nil, fmt.Errorf("list account centers: %w", err)
	}
	defer rows.Close()
	var out []models.AccountCenter
	for rows.Next() {
		var ac models.AccountCenter
		if err := rows.Scan(&ac.ID, &ac.Name, &ac.Type, &ac.Description); err != nil {
			return nil, err
		}
		out = append(out, ac)
	}
	return out, rows.Err()
}

func (r *companyRepository) GetAccountCenter(ctx context.Context, companyID int64, id int32) (*models.AccountCenter, error) {
	row := r.db.QueryRow(ctx, `
		SELECT id, name, COALESCE(type,''), COALESCE(description,'')
		FROM account_centers WHERE company_id = $1 AND id = $2
	`, companyID, id)
	var ac models.AccountCenter
	if err := row.Scan(&ac.ID, &ac.Name, &ac.Type, &ac.Description); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("get account center: %w", err)
	}
	return &ac, nil
}

func (r *companyRepository) CreateAccountCenter(ctx context.Context, companyID int64, ac *models.AccountCenter) (*models.AccountCenter, error) {
	row := r.db.QueryRow(ctx, `
		INSERT INTO account_centers (company_id, name, type, description)
		VALUES ($1, $2, $3, $4)
		RETURNING id, name, COALESCE(type,''), COALESCE(description,'')
	`, companyID, ac.Name, nullIfEmpty(ac.Type), nullIfEmpty(ac.Description))
	var out models.AccountCenter
	if err := row.Scan(&out.ID, &out.Name, &out.Type, &out.Description); err != nil {
		return nil, fmt.Errorf("create account center: %w", err)
	}
	return &out, nil
}

func (r *companyRepository) UpdateAccountCenter(ctx context.Context, companyID int64, id int32, ac *models.AccountCenter) (*models.AccountCenter, error) {
	row := r.db.QueryRow(ctx, `
		UPDATE account_centers SET name = $3, type = $4, description = $5
		WHERE company_id = $1 AND id = $2
		RETURNING id, name, COALESCE(type,''), COALESCE(description,'')
	`, companyID, id, ac.Name, nullIfEmpty(ac.Type), nullIfEmpty(ac.Description))
	var out models.AccountCenter
	if err := row.Scan(&out.ID, &out.Name, &out.Type, &out.Description); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("update account center: %w", err)
	}
	return &out, nil
}

func (r *companyRepository) DeleteAccountCenter(ctx context.Context, companyID int64, id int32) error {
	tag, err := r.db.Exec(ctx, `DELETE FROM account_centers WHERE company_id = $1 AND id = $2`, companyID, id)
	if err != nil {
		return fmt.Errorf("delete account center: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("account center not found")
	}
	return nil
}

func (r *companyRepository) ListAccountCodes(ctx context.Context, companyID int64, q models.AccountCodeListQuery) ([]models.AccountCode, error) {
	sql := `SELECT id, code, description, entry, entry_type, disabled FROM account_codes WHERE company_id = $1`
	args := []any{companyID}
	n := 2
	if !q.ShowDisabled {
		sql += ` AND disabled = false`
	}
	if s := strings.TrimSpace(q.Entry); s != "" {
		sql += fmt.Sprintf(` AND entry = $%d`, n)
		args = append(args, s)
		n++
	}
	if s := strings.TrimSpace(q.EntryType); s != "" {
		sql += fmt.Sprintf(` AND entry_type = $%d`, n)
		args = append(args, s)
		n++
	}
	sql += fmt.Sprintf(` ORDER BY code LIMIT $%d OFFSET $%d`, n, n+1)
	args = append(args, limitOrDefault(q.Limit), q.Offset)

	rows, err := r.db.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("list account codes: %w", err)
	}
	defer rows.Close()
	var out []models.AccountCode
	for rows.Next() {
		var ac models.AccountCode
		if err := rows.Scan(&ac.ID, &ac.Code, &ac.Description, &ac.Entry, &ac.EntryType, &ac.Disabled); err != nil {
			return nil, err
		}
		out = append(out, ac)
	}
	return out, rows.Err()
}

func (r *companyRepository) ListBanks(ctx context.Context, companyID int64, q models.BankListQuery) ([]models.BankCashdesk, error) {
	sql := `SELECT id, description, is_cashdesk, account_code, deleted, iban, sia, swift, cuc
		FROM banks WHERE company_id = $1`
	args := []any{companyID}
	n := 2
	if !q.ShowDeleted {
		sql += ` AND deleted = false`
	}
	if q.IsCashdesk != nil {
		sql += fmt.Sprintf(` AND is_cashdesk = $%d`, n)
		args = append(args, *q.IsCashdesk)
		n++
	}
	if s := strings.TrimSpace(q.AccountCode); s != "" {
		sql += fmt.Sprintf(` AND account_code ILIKE $%d`, n)
		args = append(args, "%"+s+"%")
		n++
	}
	if s := strings.TrimSpace(q.Iban); s != "" {
		sql += fmt.Sprintf(` AND iban ILIKE $%d`, n)
		args = append(args, "%"+s+"%")
		n++
	}
	if s := strings.TrimSpace(q.Sia); s != "" {
		sql += fmt.Sprintf(` AND sia ILIKE $%d`, n)
		args = append(args, "%"+s+"%")
		n++
	}
	if s := strings.TrimSpace(q.Swift); s != "" {
		sql += fmt.Sprintf(` AND swift ILIKE $%d`, n)
		args = append(args, "%"+s+"%")
		n++
	}
	if s := strings.TrimSpace(q.Cuc); s != "" {
		sql += fmt.Sprintf(` AND cuc ILIKE $%d`, n)
		args = append(args, "%"+s+"%")
		n++
	}
	sql += fmt.Sprintf(` ORDER BY id LIMIT $%d OFFSET $%d`, n, n+1)
	args = append(args, limitOrDefault(q.Limit), q.Offset)

	rows, err := r.db.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("list banks: %w", err)
	}
	defer rows.Close()
	return scanBanks(rows)
}

func (r *companyRepository) GetBank(ctx context.Context, companyID int64, id int32) (*models.BankCashdesk, error) {
	row := r.db.QueryRow(ctx, `
		SELECT id, description, is_cashdesk, account_code, deleted, iban, sia, swift, cuc
		FROM banks WHERE company_id = $1 AND id = $2
	`, companyID, id)
	return scanBankRow(row)
}

func (r *companyRepository) CreateBank(ctx context.Context, companyID int64, b *models.BankCashdesk) (*models.BankCashdesk, error) {
	row := r.db.QueryRow(ctx, `
		INSERT INTO banks (company_id, description, is_cashdesk, account_code, iban, sia, swift, cuc)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
		RETURNING id, description, is_cashdesk, account_code, deleted, iban, sia, swift, cuc
	`, companyID, b.Description, b.IsCashdesk, b.AccountCode, b.Iban, b.Sia, b.Swift, b.Cuc)
	return scanBankRow(row)
}

func (r *companyRepository) UpdateBank(ctx context.Context, companyID int64, id int32, b *models.BankCashdesk) (*models.BankCashdesk, error) {
	row := r.db.QueryRow(ctx, `
		UPDATE banks SET description=$3, is_cashdesk=$4, account_code=$5, iban=$6, sia=$7, swift=$8, cuc=$9
		WHERE company_id=$1 AND id=$2 AND deleted=false
		RETURNING id, description, is_cashdesk, account_code, deleted, iban, sia, swift, cuc
	`, companyID, id, b.Description, b.IsCashdesk, b.AccountCode, b.Iban, b.Sia, b.Swift, b.Cuc)
	out, err := scanBankRow(row)
	if err != nil {
		return nil, err
	}
	if out == nil {
		return nil, fmt.Errorf("bank not found")
	}
	return out, nil
}

func (r *companyRepository) DeleteBank(ctx context.Context, companyID int64, id int32) error {
	tag, err := r.db.Exec(ctx, `UPDATE banks SET deleted=true WHERE company_id=$1 AND id=$2 AND deleted=false`, companyID, id)
	if err != nil {
		return fmt.Errorf("delete bank: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("bank not found")
	}
	return nil
}

func (r *companyRepository) ListBupNumerators(ctx context.Context, companyID int64, q models.BupNumeratorQuery) ([]models.BupNumerator, error) {
	sql := `
		SELECT bu.id, bu.name, bup.id, bup.name, dn.id, dn.id_document_type, dn.numerator
		FROM document_numerators dn
		JOIN business_unit_points bup ON bup.id = dn.bup_id
		JOIN business_units bu ON bu.id = bup.bu_id
		WHERE bu.company_id = $1`
	args := []any{companyID}
	n := 2
	if q.IDDocumentType != nil {
		sql += fmt.Sprintf(` AND dn.id_document_type = $%d`, n)
		args = append(args, *q.IDDocumentType)
		n++
	}
	if s := strings.TrimSpace(q.IDPlugin); s != "" {
		sql += fmt.Sprintf(` AND dn.id_plugin = $%d`, n)
		args = append(args, s)
		n++
	}
	sql += ` ORDER BY bu.id, bup.id, dn.id_document_type`

	rows, err := r.db.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("list bup numerators: %w", err)
	}
	defer rows.Close()
	var out []models.BupNumerator
	for rows.Next() {
		var item models.BupNumerator
		if err := rows.Scan(&item.IDBu, &item.BuName, &item.IDBup, &item.BupName, &item.IDNumerator, &item.IDDocumentType, &item.Numerator); err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func (r *companyRepository) GetCurrencyChange(ctx context.Context, companyID int64, sourceCurrency string, asOf *time.Time) (*models.CurrencyChange, error) {
	var baseCurrency string
	err := r.db.QueryRow(ctx, `SELECT COALESCE(base_currency,'EUR') FROM companies WHERE id=$1`, companyID).Scan(&baseCurrency)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	sourceCurrency = strings.TrimSpace(strings.ToUpper(sourceCurrency))
	if sourceCurrency == baseCurrency {
		return &models.CurrencyChange{SourceCurrency: sourceCurrency, TargetCurrency: baseCurrency, Rate: 1}, nil
	}

	sql := `
		SELECT source_currency, target_currency, rate FROM currency_rates
		WHERE company_id = $1 AND source_currency = $2 AND target_currency = $3`
	args := []any{companyID, sourceCurrency, baseCurrency}
	if asOf != nil {
		sql += ` AND effective_date <= $4 ORDER BY effective_date DESC LIMIT 1`
		args = append(args, asOf.Format("2006-01-02"))
	} else {
		sql += ` ORDER BY effective_date DESC LIMIT 1`
	}

	row := r.db.QueryRow(ctx, sql, args...)
	var cc models.CurrencyChange
	if err := row.Scan(&cc.SourceCurrency, &cc.TargetCurrency, &cc.Rate); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("get currency change: %w", err)
	}
	return &cc, nil
}

func (r *companyRepository) ListVatRates(ctx context.Context, companyID int64, q models.VatListQuery) ([]models.VatRate, error) {
	sql := `SELECT id, id_bu, country, state, contact_type, tax_super_type, rate, description
		FROM vat_rates WHERE company_id = $1`
	args := []any{companyID}
	n := 2
	sql, args, n = appendVatFilters(sql, args, n, q.IDBu, q.Country, q.State, q.ContactType, q.TaxSuperType, q.Rate)
	sql += fmt.Sprintf(` ORDER BY id LIMIT $%d OFFSET $%d`, n, n+1)
	args = append(args, limitOrDefault(q.Limit), q.Offset)

	rows, err := r.db.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("list vat rates: %w", err)
	}
	defer rows.Close()
	return scanVatRates(rows)
}

func (r *companyRepository) GetVatRate(ctx context.Context, companyID int64, idVat string, q models.VatGetQuery) (*models.VatRate, error) {
	sql := `SELECT id, id_bu, country, state, contact_type, tax_super_type, rate, description
		FROM vat_rates WHERE company_id = $1 AND id = $2`
	args := []any{companyID, strings.TrimSpace(idVat)}
	n := 3
	sql, args, _ = appendVatFilters(sql, args, n, q.IDBu, q.Country, q.State, q.ContactType, q.TaxSuperType, nil)
	sql += ` LIMIT 1`

	row := r.db.QueryRow(ctx, sql, args...)
	return scanVatRow(row)
}

func appendVatFilters(sql string, args []any, n int, idBu, country, state, contactType string, taxSuperType *int32, rate *float64) (string, []any, int) {
	if s := strings.TrimSpace(idBu); s != "" {
		sql += fmt.Sprintf(` AND id_bu = $%d`, n)
		args = append(args, s)
		n++
	}
	if s := strings.TrimSpace(country); s != "" {
		sql += fmt.Sprintf(` AND country = $%d`, n)
		args = append(args, s)
		n++
	}
	if s := strings.TrimSpace(state); s != "" {
		sql += fmt.Sprintf(` AND state = $%d`, n)
		args = append(args, s)
		n++
	}
	if s := strings.TrimSpace(contactType); s != "" {
		sql += fmt.Sprintf(` AND contact_type = $%d`, n)
		args = append(args, s)
		n++
	}
	if taxSuperType != nil {
		sql += fmt.Sprintf(` AND tax_super_type = $%d`, n)
		args = append(args, *taxSuperType)
		n++
	}
	if rate != nil {
		sql += fmt.Sprintf(` AND rate = $%d`, n)
		args = append(args, *rate)
		n++
	}
	return sql, args, n
}

func scanBanks(rows pgx.Rows) ([]models.BankCashdesk, error) {
	var out []models.BankCashdesk
	for rows.Next() {
		b, err := scanBankFromScanner(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *b)
	}
	return out, rows.Err()
}

func scanBankRow(row pgx.Row) (*models.BankCashdesk, error) {
	b, err := scanBankFromScanner(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return b, nil
}

func scanBankFromScanner(s interface{ Scan(...any) error }) (*models.BankCashdesk, error) {
	var b models.BankCashdesk
	if err := s.Scan(&b.ID, &b.Description, &b.IsCashdesk, &b.AccountCode, &b.Deleted, &b.Iban, &b.Sia, &b.Swift, &b.Cuc); err != nil {
		return nil, err
	}
	return &b, nil
}

func scanVatRates(rows pgx.Rows) ([]models.VatRate, error) {
	var out []models.VatRate
	for rows.Next() {
		v, err := scanVatFromScanner(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *v)
	}
	return out, rows.Err()
}

func scanVatRow(row pgx.Row) (*models.VatRate, error) {
	v, err := scanVatFromScanner(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return v, nil
}

func scanVatFromScanner(s interface{ Scan(...any) error }) (*models.VatRate, error) {
	var v models.VatRate
	if err := s.Scan(&v.IDVat, &v.IDBu, &v.Country, &v.State, &v.ContactType, &v.TaxSuperType, &v.Rate, &v.Description); err != nil {
		return nil, err
	}
	return &v, nil
}
