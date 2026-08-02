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

// VendorRepository persists vendors (linked to contacts).
type VendorRepository interface {
	ListVendors(ctx context.Context, q models.VendorListQuery) ([]models.Vendor, error)
	GetVendor(ctx context.Context, id string, onlyVendorData bool) (*models.Vendor, error)
	CreateVendor(ctx context.Context, v *models.Vendor) (*models.Vendor, error)
	UpdateVendor(ctx context.Context, id string, v *models.Vendor) (*models.Vendor, error)
	PatchVendor(ctx context.Context, id string, fields map[string]any) (*models.Vendor, error)
	DeleteVendor(ctx context.Context, id string) error
	GetVendorReport(ctx context.Context, id string) (*models.PartnerReport, error)
}

type vendorRepository struct {
	db       *pgxpool.Pool
	contacts ContactRepository
}

func NewVendorRepository(db *pgxpool.Pool, contacts ContactRepository) VendorRepository {
	return &vendorRepository{db: db, contacts: contacts}
}

var vendorPatchColumns = map[string]string{
	"idPaymentTerm":    "id_payment_term",
	"idUserAgent1":     "id_user_agent1",
	"idUserAgent2":     "id_user_agent2",
	"currency":         "currency",
	"credit":           "credit",
	"dsc1":             "dsc1",
	"dsc2":             "dsc2",
	"dsc3":             "dsc3",
	"dsc4":             "dsc4",
	"priceListID":      "price_list_id",
	"priceListType":    "price_list_type",
	"priceListEnabled": "price_list_enabled",
	"bank":             "bank",
	"iban":             "iban",
	"swift":            "swift",
}

func (r *vendorRepository) ListVendors(ctx context.Context, q models.VendorListQuery) ([]models.Vendor, error) {
	companyID := defaultCompanyID
	if q.IDCompany != nil {
		companyID = *q.IDCompany
	}
	sql := vendorSelectSQL() + `
		FROM vendors v
		JOIN contacts c ON c.id = v.contact_id
		WHERE c.company_id = $1`
	args := []any{companyID}
	n := 2
	sql, args, n = appendVendorFilters(sql, args, n, q.Name, q.FreeText, q.Deleted)
	sql += fmt.Sprintf(` ORDER BY v.id LIMIT $%d OFFSET $%d`, n, n+1)
	args = append(args, limitOrDefault(q.Limit), q.Offset)

	rows, err := r.db.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("list vendors: %w", err)
	}
	defer rows.Close()
	return scanVendors(rows, true)
}

func (r *vendorRepository) GetVendor(ctx context.Context, id string, onlyVendorData bool) (*models.Vendor, error) {
	sql := vendorBaseSelectSQL()
	if !onlyVendorData {
		sql += `, ` + contactSelectCols
	}
	sql += `
		FROM vendors v
		JOIN contacts c ON c.id = v.contact_id
		WHERE v.id = $1 AND c.company_id = $2 AND v.deleted = false AND c.deleted = false`
	row := r.db.QueryRow(ctx, sql, strings.TrimSpace(id), defaultCompanyID)
	vendor, err := scanVendorFromScanner(row, !onlyVendorData)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("get vendor: %w", err)
	}
	return vendor, nil
}

func (r *vendorRepository) CreateVendor(ctx context.Context, v *models.Vendor) (*models.Vendor, error) {
	var contactID int32
	if v.Contact.ID > 0 {
		contactID = v.Contact.ID
		tag, err := r.db.Exec(ctx, `
			UPDATE contacts SET is_vendor=true, updated_at=now()
			WHERE id=$1 AND company_id=$2 AND deleted=false
		`, contactID, defaultCompanyID)
		if err != nil {
			return nil, fmt.Errorf("enable vendor contact: %w", err)
		}
		if tag.RowsAffected() == 0 {
			return nil, fmt.Errorf("contact not found")
		}
	} else {
		created, err := r.contacts.CreateContact(ctx, &v.Contact)
		if err != nil {
			return nil, err
		}
		contactID = created.ID
		if _, err := r.db.Exec(ctx, `UPDATE contacts SET is_vendor=true WHERE id=$1`, contactID); err != nil {
			return nil, fmt.Errorf("mark vendor contact: %w", err)
		}
	}

	vendorID := fmt.Sprintf("V%05d", contactID)
	row := r.db.QueryRow(ctx, `
		INSERT INTO vendors (
			id, contact_id, id_payment_term, id_user_agent1, id_user_agent2, currency, credit,
			dsc1, dsc2, dsc3, dsc4, price_list_id, price_list_type, price_list_enabled,
			bank, iban, swift
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17)
		RETURNING id
	`, vendorID, contactID, v.IDPaymentTerm, v.IDUserAgent1, v.IDUserAgent2, v.Currency, v.Credit,
		v.Dsc1, v.Dsc2, v.Dsc3, v.Dsc4, v.PriceListID, v.PriceListType, v.PriceListEnabled,
		v.Bank, v.Iban, v.Swift)
	var outID string
	if err := row.Scan(&outID); err != nil {
		return nil, fmt.Errorf("create vendor: %w", err)
	}
	return r.GetVendor(ctx, outID, false)
}

func (r *vendorRepository) UpdateVendor(ctx context.Context, id string, v *models.Vendor) (*models.Vendor, error) {
	existing, err := r.GetVendor(ctx, id, false)
	if err != nil || existing == nil {
		return existing, err
	}
	if _, err := r.contacts.UpdateContact(ctx, existing.Contact.ID, &v.Contact); err != nil {
		return nil, err
	}
	row := r.db.QueryRow(ctx, `
		UPDATE vendors vnd SET
			id_payment_term=$2, id_user_agent1=$3, id_user_agent2=$4, currency=$5, credit=$6,
			dsc1=$7, dsc2=$8, dsc3=$9, dsc4=$10, price_list_id=$11, price_list_type=$12,
			price_list_enabled=$13, bank=$14, iban=$15, swift=$16
		FROM contacts c
		WHERE vnd.id=$1 AND vnd.contact_id=c.id AND c.company_id=$17 AND vnd.deleted=false AND c.deleted=false
		RETURNING vnd.id
	`, strings.TrimSpace(id), v.IDPaymentTerm, v.IDUserAgent1, v.IDUserAgent2, v.Currency, v.Credit,
		v.Dsc1, v.Dsc2, v.Dsc3, v.Dsc4, v.PriceListID, v.PriceListType, v.PriceListEnabled,
		v.Bank, v.Iban, v.Swift, defaultCompanyID)
	var outID string
	if err := row.Scan(&outID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("update vendor: %w", err)
	}
	return r.GetVendor(ctx, outID, false)
}

func (r *vendorRepository) PatchVendor(ctx context.Context, id string, fields map[string]any) (*models.Vendor, error) {
	existing, err := r.GetVendor(ctx, id, false)
	if err != nil || existing == nil {
		return existing, err
	}
	if contactFields, ok := fields["contact"].(map[string]any); ok && len(contactFields) > 0 {
		if _, err := r.contacts.PatchContact(ctx, existing.Contact.ID, contactFields); err != nil {
			return nil, err
		}
	}
	topLevel := make(map[string]any, len(fields))
	for k, v := range fields {
		if k != "contact" {
			topLevel[k] = v
		}
	}
	if len(topLevel) > 0 {
		setSQL, args, err := buildPatchSet(vendorPatchColumns, topLevel, 2)
		if err != nil {
			return nil, err
		}
		if setSQL != "" {
			args = append([]any{strings.TrimSpace(id)}, args...)
			sql := fmt.Sprintf(`
				UPDATE vendors vnd SET %s
				FROM contacts c
				WHERE vnd.id=$1 AND vnd.contact_id=c.id AND c.company_id=$%d AND vnd.deleted=false AND c.deleted=false
			`, setSQL, len(args)+1)
			args = append(args, defaultCompanyID)
			tag, err := r.db.Exec(ctx, sql, args...)
			if err != nil {
				return nil, fmt.Errorf("patch vendor: %w", err)
			}
			if tag.RowsAffected() == 0 {
				return nil, nil
			}
		}
	}
	return r.GetVendor(ctx, id, false)
}

func (r *vendorRepository) DeleteVendor(ctx context.Context, id string) error {
	tag, err := r.db.Exec(ctx, `
		UPDATE vendors vnd SET deleted=true
		FROM contacts c
		WHERE vnd.id=$1 AND vnd.contact_id=c.id AND c.company_id=$2 AND vnd.deleted=false
	`, strings.TrimSpace(id), defaultCompanyID)
	if err != nil {
		return fmt.Errorf("delete vendor: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("vendor not found")
	}
	return nil
}

func (r *vendorRepository) GetVendorReport(ctx context.Context, id string) (*models.PartnerReport, error) {
	return &models.PartnerReport{
		ID:            strings.TrimSpace(id),
		TotalInvoiced: 0,
		TotalPaid:     0,
		TotalDue:      0,
	}, nil
}


func vendorSelectSQL() string {
	return vendorBaseSelectSQL() + `, ` + contactSelectCols
}

func vendorBaseSelectSQL() string {
	return `SELECT v.id, v.contact_id, v.id_payment_term, v.id_user_agent1, v.id_user_agent2,
		v.currency, v.credit, v.dsc1, v.dsc2, v.dsc3, v.dsc4, v.price_list_id, v.price_list_type,
		v.price_list_enabled, v.bank, v.iban, v.swift`
}

func scanVendors(rows pgx.Rows, withContact bool) ([]models.Vendor, error) {
	var out []models.Vendor
	for rows.Next() {
		v, err := scanVendorFromScanner(rows, withContact)
		if err != nil {
			return nil, err
		}
		out = append(out, *v)
	}
	return out, rows.Err()
}

func scanVendorFromScanner(s interface{ Scan(...any) error }, withContact bool) (*models.Vendor, error) {
	var vendor models.Vendor
	var contactID int32
	dest := []any{
		&vendor.ID, &contactID, &vendor.IDPaymentTerm, &vendor.IDUserAgent1, &vendor.IDUserAgent2,
		&vendor.Currency, &vendor.Credit, &vendor.Dsc1, &vendor.Dsc2, &vendor.Dsc3, &vendor.Dsc4,
		&vendor.PriceListID, &vendor.PriceListType, &vendor.PriceListEnabled, &vendor.Bank, &vendor.Iban,
		&vendor.Swift,
	}
	if withContact {
		dest = append(dest,
			&vendor.Contact.ID, &vendor.Contact.Type, &vendor.Contact.Name, &vendor.Contact.LastName,
			&vendor.Contact.Address, &vendor.Contact.City, &vendor.Contact.Country, &vendor.Contact.State,
			&vendor.Contact.PostalCode, &vendor.Contact.Pr, &vendor.Contact.FiscalCode, &vendor.Contact.VatCode,
			&vendor.Contact.Phone1, &vendor.Contact.Phone2, &vendor.Contact.Mobile, &vendor.Contact.Fax,
			&vendor.Contact.Email, &vendor.Contact.Email2, &vendor.Contact.VisibilityType, &vendor.Contact.Status,
			&vendor.Contact.IDUserOwner, &vendor.Contact.Source, &vendor.Contact.Note,
		)
	}
	if err := s.Scan(dest...); err != nil {
		return nil, err
	}
	if !withContact {
		vendor.Contact.ID = contactID
	}
	return &vendor, nil
}

func appendVendorFilters(sql string, args []any, n int, name, freeText string, deleted *bool) (string, []any, int) {
	if deleted != nil {
		sql += fmt.Sprintf(` AND v.deleted = $%d`, n)
		args = append(args, *deleted)
		n++
	} else {
		sql += ` AND v.deleted = false`
	}
	if s := strings.TrimSpace(name); s != "" {
		sql += fmt.Sprintf(` AND (c.name ILIKE $%d OR c.last_name ILIKE $%d OR CONCAT(c.name, ' ', COALESCE(c.last_name,'')) ILIKE $%d)`, n, n, n)
		args = append(args, "%"+s+"%")
		n++
	}
	if s := strings.TrimSpace(freeText); s != "" {
		sql += fmt.Sprintf(` AND (c.name ILIKE $%d OR c.last_name ILIKE $%d OR c.city ILIKE $%d OR c.pr ILIKE $%d OR c.state ILIKE $%d)`, n, n, n, n, n)
		args = append(args, "%"+s+"%")
		n++
	}
	return sql, args, n
}

