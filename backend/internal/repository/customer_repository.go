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

// CustomerRepository persists customers (linked to contacts).
type CustomerRepository interface {
	ListCustomers(ctx context.Context, q models.CustomerListQuery) ([]models.Customer, error)
	GetCustomer(ctx context.Context, id string, onlyCustomerData bool) (*models.Customer, error)
	CreateCustomer(ctx context.Context, c *models.Customer) (*models.Customer, error)
	UpdateCustomer(ctx context.Context, id string, c *models.Customer) (*models.Customer, error)
	PatchCustomer(ctx context.Context, id string, fields map[string]any) (*models.Customer, error)
	DeleteCustomer(ctx context.Context, id string) error
	GetCustomerReport(ctx context.Context, id string) (*models.PartnerReport, error)
}

type customerRepository struct {
	db       *pgxpool.Pool
	contacts ContactRepository
}

func NewCustomerRepository(db *pgxpool.Pool, contacts ContactRepository) CustomerRepository {
	return &customerRepository{db: db, contacts: contacts}
}

var customerPatchColumns = map[string]string{
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
	"locked":           "locked",
	"recipientCode":    "recipient_code",
}

func (r *customerRepository) ListCustomers(ctx context.Context, q models.CustomerListQuery) ([]models.Customer, error) {
	companyID := defaultCompanyID
	if q.IDCompany != nil {
		companyID = *q.IDCompany
	}
	sql := customerSelectSQL() + `
		FROM customers cu
		JOIN contacts c ON c.id = cu.contact_id
		WHERE c.company_id = $1`
	args := []any{companyID}
	n := 2
	sql, args, n = appendCustomerFilters(sql, args, n, q.Name, q.FreeText, q.Deleted)
	sql += fmt.Sprintf(` ORDER BY cu.id LIMIT $%d OFFSET $%d`, n, n+1)
	args = append(args, limitOrDefault(q.Limit), q.Offset)

	rows, err := r.db.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("list customers: %w", err)
	}
	defer rows.Close()
	return scanCustomers(rows, true)
}

func (r *customerRepository) GetCustomer(ctx context.Context, id string, onlyCustomerData bool) (*models.Customer, error) {
	sql := customerBaseSelectSQL()
	if !onlyCustomerData {
		sql += `, ` + contactSelectCols
	}
	sql += `
		FROM customers cu
		JOIN contacts c ON c.id = cu.contact_id
		WHERE cu.id = $1 AND c.company_id = $2 AND cu.deleted = false AND c.deleted = false`
	row := r.db.QueryRow(ctx, sql, strings.TrimSpace(id), defaultCompanyID)
	cust, err := scanCustomerFromScanner(row, !onlyCustomerData)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("get customer: %w", err)
	}
	return cust, nil
}

func (r *customerRepository) CreateCustomer(ctx context.Context, c *models.Customer) (*models.Customer, error) {
	var contactID int32
	if c.Contact.ID > 0 {
		contactID = c.Contact.ID
		tag, err := r.db.Exec(ctx, `
			UPDATE contacts SET is_customer=true, updated_at=now()
			WHERE id=$1 AND company_id=$2 AND deleted=false
		`, contactID, defaultCompanyID)
		if err != nil {
			return nil, fmt.Errorf("enable customer contact: %w", err)
		}
		if tag.RowsAffected() == 0 {
			return nil, fmt.Errorf("contact not found")
		}
	} else {
		created, err := r.contacts.CreateContact(ctx, &c.Contact)
		if err != nil {
			return nil, err
		}
		contactID = created.ID
		if _, err := r.db.Exec(ctx, `UPDATE contacts SET is_customer=true WHERE id=$1`, contactID); err != nil {
			return nil, fmt.Errorf("mark customer contact: %w", err)
		}
	}

	customerID := fmt.Sprintf("C%05d", contactID)
	row := r.db.QueryRow(ctx, `
		INSERT INTO customers (
			id, contact_id, id_payment_term, id_user_agent1, id_user_agent2, currency, credit,
			dsc1, dsc2, dsc3, dsc4, price_list_id, price_list_type, price_list_enabled,
			bank, iban, swift, locked, recipient_code
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19)
		RETURNING id
	`, customerID, contactID, c.IDPaymentTerm, c.IDUserAgent1, c.IDUserAgent2, c.Currency, c.Credit,
		c.Dsc1, c.Dsc2, c.Dsc3, c.Dsc4, c.PriceListID, c.PriceListType, c.PriceListEnabled,
		c.Bank, c.Iban, c.Swift, c.Locked, c.RecipientCode)
	var outID string
	if err := row.Scan(&outID); err != nil {
		return nil, fmt.Errorf("create customer: %w", err)
	}
	return r.GetCustomer(ctx, outID, false)
}

func (r *customerRepository) UpdateCustomer(ctx context.Context, id string, c *models.Customer) (*models.Customer, error) {
	existing, err := r.GetCustomer(ctx, id, false)
	if err != nil || existing == nil {
		return existing, err
	}
	if _, err := r.contacts.UpdateContact(ctx, existing.Contact.ID, &c.Contact); err != nil {
		return nil, err
	}
	row := r.db.QueryRow(ctx, `
		UPDATE customers cu SET
			id_payment_term=$2, id_user_agent1=$3, id_user_agent2=$4, currency=$5, credit=$6,
			dsc1=$7, dsc2=$8, dsc3=$9, dsc4=$10, price_list_id=$11, price_list_type=$12,
			price_list_enabled=$13, bank=$14, iban=$15, swift=$16, locked=$17, recipient_code=$18
		FROM contacts c
		WHERE cu.id=$1 AND cu.contact_id=c.id AND c.company_id=$19 AND cu.deleted=false AND c.deleted=false
		RETURNING cu.id
	`, strings.TrimSpace(id), c.IDPaymentTerm, c.IDUserAgent1, c.IDUserAgent2, c.Currency, c.Credit,
		c.Dsc1, c.Dsc2, c.Dsc3, c.Dsc4, c.PriceListID, c.PriceListType, c.PriceListEnabled,
		c.Bank, c.Iban, c.Swift, c.Locked, c.RecipientCode, defaultCompanyID)
	var outID string
	if err := row.Scan(&outID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("update customer: %w", err)
	}
	return r.GetCustomer(ctx, outID, false)
}

func (r *customerRepository) PatchCustomer(ctx context.Context, id string, fields map[string]any) (*models.Customer, error) {
	existing, err := r.GetCustomer(ctx, id, false)
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
		setSQL, args, err := buildPatchSet(customerPatchColumns, topLevel, 2)
		if err != nil {
			return nil, err
		}
		if setSQL != "" {
			args = append([]any{strings.TrimSpace(id)}, args...)
			sql := fmt.Sprintf(`
				UPDATE customers cu SET %s
				FROM contacts c
				WHERE cu.id=$1 AND cu.contact_id=c.id AND c.company_id=$%d AND cu.deleted=false AND c.deleted=false
			`, setSQL, len(args)+1)
			args = append(args, defaultCompanyID)
			tag, err := r.db.Exec(ctx, sql, args...)
			if err != nil {
				return nil, fmt.Errorf("patch customer: %w", err)
			}
			if tag.RowsAffected() == 0 {
				return nil, nil
			}
		}
	}
	return r.GetCustomer(ctx, id, false)
}

func (r *customerRepository) DeleteCustomer(ctx context.Context, id string) error {
	tag, err := r.db.Exec(ctx, `
		UPDATE customers cu SET deleted=true
		FROM contacts c
		WHERE cu.id=$1 AND cu.contact_id=c.id AND c.company_id=$2 AND cu.deleted=false
	`, strings.TrimSpace(id), defaultCompanyID)
	if err != nil {
		return fmt.Errorf("delete customer: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("customer not found")
	}
	return nil
}

func (r *customerRepository) GetCustomerReport(ctx context.Context, id string) (*models.PartnerReport, error) {
	return &models.PartnerReport{
		ID:            strings.TrimSpace(id),
		TotalInvoiced: 0,
		TotalPaid:     0,
		TotalDue:      0,
	}, nil
}


func customerSelectSQL() string {
	return customerBaseSelectSQL() + `, ` + contactSelectCols
}

func customerBaseSelectSQL() string {
	return `SELECT cu.id, cu.contact_id, cu.id_payment_term, cu.id_user_agent1, cu.id_user_agent2,
		cu.currency, cu.credit, cu.dsc1, cu.dsc2, cu.dsc3, cu.dsc4, cu.price_list_id, cu.price_list_type,
		cu.price_list_enabled, cu.bank, cu.iban, cu.swift, cu.locked, cu.recipient_code`
}

func scanCustomers(rows pgx.Rows, withContact bool) ([]models.Customer, error) {
	var out []models.Customer
	for rows.Next() {
		c, err := scanCustomerFromScanner(rows, withContact)
		if err != nil {
			return nil, err
		}
		out = append(out, *c)
	}
	return out, rows.Err()
}

func scanCustomerFromScanner(s interface{ Scan(...any) error }, withContact bool) (*models.Customer, error) {
	var cust models.Customer
	var contactID int32
	dest := []any{
		&cust.ID, &contactID, &cust.IDPaymentTerm, &cust.IDUserAgent1, &cust.IDUserAgent2,
		&cust.Currency, &cust.Credit, &cust.Dsc1, &cust.Dsc2, &cust.Dsc3, &cust.Dsc4,
		&cust.PriceListID, &cust.PriceListType, &cust.PriceListEnabled, &cust.Bank, &cust.Iban,
		&cust.Swift, &cust.Locked, &cust.RecipientCode,
	}
	if withContact {
		dest = append(dest,
			&cust.Contact.ID, &cust.Contact.Type, &cust.Contact.Name, &cust.Contact.LastName,
			&cust.Contact.Address, &cust.Contact.City, &cust.Contact.Country, &cust.Contact.State,
			&cust.Contact.PostalCode, &cust.Contact.Pr, &cust.Contact.FiscalCode, &cust.Contact.VatCode,
			&cust.Contact.Phone1, &cust.Contact.Phone2, &cust.Contact.Mobile, &cust.Contact.Fax,
			&cust.Contact.Email, &cust.Contact.Email2, &cust.Contact.VisibilityType, &cust.Contact.Status,
			&cust.Contact.IDUserOwner, &cust.Contact.Source, &cust.Contact.Note,
		)
	}
	if err := s.Scan(dest...); err != nil {
		return nil, err
	}
	if !withContact {
		cust.Contact.ID = contactID
	}
	return &cust, nil
}

func appendCustomerFilters(sql string, args []any, n int, name, freeText string, deleted *bool) (string, []any, int) {
	if deleted != nil {
		sql += fmt.Sprintf(` AND cu.deleted = $%d`, n)
		args = append(args, *deleted)
		n++
	} else {
		sql += ` AND cu.deleted = false`
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

