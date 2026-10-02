package repository

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"dnsc_microservice/internal/models"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

const defaultCompanyID int64 = 1

// ContactRepository persists contacts, offices and sub-contacts.
type ContactRepository interface {
	ListContacts(ctx context.Context, q models.ContactListQuery) ([]models.Contact, error)
	GetContact(ctx context.Context, id int32, retrieveImage bool) (*models.Contact, error)
	CreateContact(ctx context.Context, c *models.Contact) (*models.Contact, error)
	UpdateContact(ctx context.Context, id int32, c *models.Contact) (*models.Contact, error)
	PatchContact(ctx context.Context, id int32, fields map[string]any) (*models.Contact, error)
	DeleteContact(ctx context.Context, id int32) error
	UpdateContactPhoto(ctx context.Context, id int32, data []byte, mimeType string) error
	ListContactSources(ctx context.Context, q models.ContactSourceListQuery) ([]models.ContactSource, error)
	ListOffices(ctx context.Context, contactID int32, q models.OfficeListQuery) ([]models.ContactOffice, error)
	CreateOffice(ctx context.Context, contactID int32, office *models.ContactOffice) (*models.ContactOffice, error)
	UpdateOffice(ctx context.Context, contactID int32, officeID int32, office *models.ContactOffice) (*models.ContactOffice, error)
	DeleteOffice(ctx context.Context, contactID int32, officeID int32) error
	ListSubContacts(ctx context.Context, contactID int32, q models.SubContactListQuery) ([]models.SubContactAssoc, error)
	CreateSubContact(ctx context.Context, contactID int32, assoc *models.SubContactAssoc) (*models.SubContactAssoc, error)
	DeleteSubContact(ctx context.Context, contactID int32, subContactID int32) error
}

type contactRepository struct {
	db *pgxpool.Pool
}

func NewContactRepository(db *pgxpool.Pool) ContactRepository {
	return &contactRepository{db: db}
}

const contactSelectCols = `c.id, c.type, c.name, c.last_name, c.address, c.city, c.country, c.state, c.postal_code,
	c.pr, c.fiscal_code, c.vat_code, c.phone1, c.phone2, c.mobile, c.fax, c.email, c.email2,
	c.visibility_type, c.status, c.id_user_owner, c.source, c.note`

var contactPatchColumns = map[string]string{
	"type":           "type",
	"name":           "name",
	"lastName":       "last_name",
	"address":        "address",
	"city":           "city",
	"country":        "country",
	"state":          "state",
	"postalCode":     "postal_code",
	"pr":             "pr",
	"fiscalCode":     "fiscal_code",
	"vatCode":        "vat_code",
	"phone1":         "phone1",
	"phone2":         "phone2",
	"mobile":         "mobile",
	"fax":            "fax",
	"email":          "email",
	"email2":         "email2",
	"visibilityType": "visibility_type",
	"status":         "status",
	"idUserOwner":    "id_user_owner",
	"source":         "source",
	"note":           "note",
}

func (r *contactRepository) ListContacts(ctx context.Context, q models.ContactListQuery) ([]models.Contact, error) {
	sql := `SELECT ` + contactSelectCols
	args := []any{}
	n := 1
	if q.RetrieveImage {
		sql += `, c.photo_data, c.photo_mime`
	}
	sql += ` FROM contacts c WHERE c.company_id = $1`
	args = append(args, defaultCompanyID)
	n = 2

	if q.Deleted != nil {
		sql += fmt.Sprintf(` AND c.deleted = $%d`, n)
		args = append(args, *q.Deleted)
		n++
	} else {
		sql += ` AND c.deleted = false`
	}
	if s := strings.TrimSpace(q.Type); s != "" {
		sql += fmt.Sprintf(` AND c.type = $%d`, n)
		args = append(args, s)
		n++
	}
	if s := strings.TrimSpace(q.VisibilityType); s != "" {
		sql += fmt.Sprintf(` AND c.visibility_type = $%d`, n)
		args = append(args, s)
		n++
	}
	if s := strings.TrimSpace(q.FullName); s != "" {
		sql += fmt.Sprintf(` AND (c.name ILIKE $%d OR c.last_name ILIKE $%d OR CONCAT(c.name, ' ', COALESCE(c.last_name,'')) ILIKE $%d)`, n, n, n)
		args = append(args, "%"+s+"%")
		n++
	}
	if s := strings.TrimSpace(q.FreeText); s != "" {
		sql += fmt.Sprintf(` AND (c.name ILIKE $%d OR c.last_name ILIKE $%d OR c.city ILIKE $%d OR c.pr ILIKE $%d OR c.state ILIKE $%d)`, n, n, n, n, n)
		args = append(args, "%"+s+"%")
		n++
	}
	if s := strings.TrimSpace(q.Email); s != "" {
		sql += fmt.Sprintf(` AND c.email ILIKE $%d`, n)
		args = append(args, "%"+s+"%")
		n++
	}
	if s := strings.TrimSpace(q.FiscalCode); s != "" {
		sql += fmt.Sprintf(` AND c.fiscal_code ILIKE $%d`, n)
		args = append(args, "%"+s+"%")
		n++
	}
	if s := strings.TrimSpace(q.VatCode); s != "" {
		sql += fmt.Sprintf(` AND c.vat_code ILIKE $%d`, n)
		args = append(args, "%"+s+"%")
		n++
	}
	if q.OnlyCustomers {
		sql += ` AND c.is_customer = true`
	}
	if q.OnlyVendors {
		sql += ` AND c.is_vendor = true`
	}
	if q.OnlyLeads {
		sql += ` AND c.is_lead = true`
	}
	sql += fmt.Sprintf(` ORDER BY c.id LIMIT $%d OFFSET $%d`, n, n+1)
	args = append(args, limitOrDefault(q.Limit), q.Offset)

	rows, err := r.db.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("list contacts: %w", err)
	}
	defer rows.Close()
	return scanContacts(rows, q.RetrieveImage)
}

func (r *contactRepository) GetContact(ctx context.Context, id int32, retrieveImage bool) (*models.Contact, error) {
	sql := `SELECT ` + contactSelectCols
	if retrieveImage {
		sql += `, c.photo_data, c.photo_mime`
	}
	sql += ` FROM contacts c WHERE c.id = $1 AND c.company_id = $2 AND c.deleted = false`
	row := r.db.QueryRow(ctx, sql, id, defaultCompanyID)
	return scanContactRow(row, retrieveImage)
}

func (r *contactRepository) CreateContact(ctx context.Context, c *models.Contact) (*models.Contact, error) {
	contactType := strings.TrimSpace(c.Type)
	if contactType == "" {
		contactType = "COMPANY"
	}
	visibility := "ALL"
	if c.VisibilityType != nil && strings.TrimSpace(*c.VisibilityType) != "" {
		visibility = strings.TrimSpace(*c.VisibilityType)
	}
	row := r.db.QueryRow(ctx, `
		INSERT INTO contacts (
			company_id, type, name, last_name, address, city, country, state, postal_code, pr,
			fiscal_code, vat_code, phone1, phone2, mobile, fax, email, email2,
			visibility_type, status, id_user_owner, source, note
		) VALUES (
			$1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21,$22,$23
		) RETURNING id
	`, defaultCompanyID, contactType, c.Name, c.LastName, c.Address, c.City, c.Country, c.State, c.PostalCode, c.Pr,
		c.FiscalCode, c.VatCode, c.Phone1, c.Phone2, c.Mobile, c.Fax, c.Email, c.Email2,
		visibility, c.Status, c.IDUserOwner, c.Source, c.Note)
	var id int32
	if err := row.Scan(&id); err != nil {
		return nil, fmt.Errorf("create contact: %w", err)
	}
	return r.GetContact(ctx, id, false)
}

func (r *contactRepository) UpdateContact(ctx context.Context, id int32, c *models.Contact) (*models.Contact, error) {
	visibility := "ALL"
	if c.VisibilityType != nil && strings.TrimSpace(*c.VisibilityType) != "" {
		visibility = strings.TrimSpace(*c.VisibilityType)
	}
	row := r.db.QueryRow(ctx, `
		UPDATE contacts SET
			type=$3, name=$4, last_name=$5, address=$6, city=$7, country=$8, state=$9, postal_code=$10, pr=$11,
			fiscal_code=$12, vat_code=$13, phone1=$14, phone2=$15, mobile=$16, fax=$17, email=$18, email2=$19,
			visibility_type=$20, status=$21, id_user_owner=$22, source=$23, note=$24, updated_at=now()
		WHERE id=$1 AND company_id=$2 AND deleted=false
		RETURNING id
	`, id, defaultCompanyID, strings.TrimSpace(c.Type), c.Name, c.LastName, c.Address, c.City, c.Country, c.State,
		c.PostalCode, c.Pr, c.FiscalCode, c.VatCode, c.Phone1, c.Phone2, c.Mobile, c.Fax, c.Email, c.Email2,
		visibility, c.Status, c.IDUserOwner, c.Source, c.Note)
	var outID int32
	if err := row.Scan(&outID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("update contact: %w", err)
	}
	return r.GetContact(ctx, outID, false)
}

func (r *contactRepository) PatchContact(ctx context.Context, id int32, fields map[string]any) (*models.Contact, error) {
	if len(fields) == 0 {
		return r.GetContact(ctx, id, false)
	}
	setSQL, args, err := buildPatchSet(contactPatchColumns, fields, 3)
	if err != nil {
		return nil, err
	}
	if setSQL == "" {
		return r.GetContact(ctx, id, false)
	}
	args = append([]any{id, defaultCompanyID}, args...)
	sql := fmt.Sprintf(`UPDATE contacts SET %s, updated_at=now() WHERE id=$1 AND company_id=$2 AND deleted=false RETURNING id`, setSQL)
	var outID int32
	if err := r.db.QueryRow(ctx, sql, args...).Scan(&outID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("patch contact: %w", err)
	}
	return r.GetContact(ctx, outID, false)
}

func (r *contactRepository) DeleteContact(ctx context.Context, id int32) error {
	tag, err := r.db.Exec(ctx, `
		UPDATE contacts SET deleted=true, updated_at=now()
		WHERE id=$1 AND company_id=$2 AND deleted=false
	`, id, defaultCompanyID)
	if err != nil {
		return fmt.Errorf("delete contact: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("contact not found")
	}
	return nil
}

func (r *contactRepository) UpdateContactPhoto(ctx context.Context, id int32, data []byte, mimeType string) error {
	tag, err := r.db.Exec(ctx, `
		UPDATE contacts SET photo_data=$3, photo_mime=$4, updated_at=now()
		WHERE id=$1 AND company_id=$2 AND deleted=false
	`, id, defaultCompanyID, data, nullIfEmpty(mimeType))
	if err != nil {
		return fmt.Errorf("update contact photo: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("contact not found")
	}
	return nil
}

func (r *contactRepository) ListContactSources(ctx context.Context, q models.ContactSourceListQuery) ([]models.ContactSource, error) {
	sql := `SELECT id, name FROM contact_sources WHERE 1=1`
	args := []any{}
	n := 1
	if s := strings.TrimSpace(q.FreeText); s != "" {
		sql += fmt.Sprintf(` AND name ILIKE $%d`, n)
		args = append(args, "%"+s+"%")
		n++
	}
	sql += fmt.Sprintf(` ORDER BY name LIMIT $%d OFFSET $%d`, n, n+1)
	args = append(args, limitOrDefault(q.Limit), q.Offset)

	rows, err := r.db.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("list contact sources: %w", err)
	}
	defer rows.Close()
	var out []models.ContactSource
	for rows.Next() {
		var cs models.ContactSource
		if err := rows.Scan(&cs.ID, &cs.Name); err != nil {
			return nil, err
		}
		out = append(out, cs)
	}
	return out, rows.Err()
}

func (r *contactRepository) ListOffices(ctx context.Context, contactID int32, q models.OfficeListQuery) ([]models.ContactOffice, error) {
	sql := `
		SELECT o.id, o.id_office_type, o.name, o.address, o.city, o.pr, o.postal_code, o.country, o.state,
			o.email, o.phone1, o.phone2, o.fax, o.default_dest
		FROM contact_offices o
		JOIN contacts c ON c.id = o.contact_id
		WHERE o.contact_id = $1 AND c.company_id = $2`
	args := []any{contactID, defaultCompanyID}
	n := 3
	if q.DefaultDest != nil {
		sql += fmt.Sprintf(` AND o.default_dest = $%d`, n)
		args = append(args, *q.DefaultDest)
		n++
	}
	sql += fmt.Sprintf(` ORDER BY o.id LIMIT $%d OFFSET $%d`, n, n+1)
	args = append(args, limitOrDefault(q.Limit), q.Offset)

	rows, err := r.db.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("list offices: %w", err)
	}
	defer rows.Close()
	return scanOffices(rows)
}

func (r *contactRepository) CreateOffice(ctx context.Context, contactID int32, office *models.ContactOffice) (*models.ContactOffice, error) {
	row := r.db.QueryRow(ctx, `
		INSERT INTO contact_offices (
			contact_id, id_office_type, name, address, city, pr, postal_code, country, state,
			email, phone1, phone2, fax, default_dest
		)
		SELECT $1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14
		FROM contacts c WHERE c.id = $1 AND c.company_id = $15 AND c.deleted = false
		RETURNING id, id_office_type, name, address, city, pr, postal_code, country, state,
			email, phone1, phone2, fax, default_dest
	`, contactID, office.IDOfficeType, office.Name, office.Address, office.City, office.Pr, office.PostalCode,
		office.Country, office.State, office.Email, office.Phone1, office.Phone2, office.Fax, office.DefaultDest, defaultCompanyID)
	return scanOfficeRow(row)
}

func (r *contactRepository) UpdateOffice(ctx context.Context, contactID int32, officeID int32, office *models.ContactOffice) (*models.ContactOffice, error) {
	row := r.db.QueryRow(ctx, `
		UPDATE contact_offices o SET
			id_office_type=$4, name=$5, address=$6, city=$7, pr=$8, postal_code=$9, country=$10, state=$11,
			email=$12, phone1=$13, phone2=$14, fax=$15, default_dest=$16
		FROM contacts c
		WHERE o.id=$3 AND o.contact_id=$1 AND o.contact_id=c.id AND c.company_id=$2 AND c.deleted=false
		RETURNING o.id, o.id_office_type, o.name, o.address, o.city, o.pr, o.postal_code, o.country, o.state,
			o.email, o.phone1, o.phone2, o.fax, o.default_dest
	`, contactID, defaultCompanyID, officeID, office.IDOfficeType, office.Name, office.Address, office.City,
		office.Pr, office.PostalCode, office.Country, office.State, office.Email, office.Phone1, office.Phone2,
		office.Fax, office.DefaultDest)
	return scanOfficeRow(row)
}

func (r *contactRepository) DeleteOffice(ctx context.Context, contactID int32, officeID int32) error {
	tag, err := r.db.Exec(ctx, `
		DELETE FROM contact_offices o
		USING contacts c
		WHERE o.id=$3 AND o.contact_id=$1 AND o.contact_id=c.id AND c.company_id=$2
	`, contactID, defaultCompanyID, officeID)
	if err != nil {
		return fmt.Errorf("delete office: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("office not found")
	}
	return nil
}

func (r *contactRepository) ListSubContacts(ctx context.Context, contactID int32, q models.SubContactListQuery) ([]models.SubContactAssoc, error) {
	sql := `
		SELECT a.id, a.child_contact_id, a.id_contact_role
		FROM contact_subcontact_assocs a
		JOIN contacts parent ON parent.id = a.parent_contact_id
		JOIN contacts child ON child.id = a.child_contact_id
		WHERE a.parent_contact_id = $1 AND parent.company_id = $2`
	args := []any{contactID, defaultCompanyID}
	n := 3
	if s := strings.TrimSpace(q.Type); s != "" {
		sql += fmt.Sprintf(` AND child.type = $%d`, n)
		args = append(args, s)
		n++
	}
	if q.Deleted != nil {
		sql += fmt.Sprintf(` AND child.deleted = $%d`, n)
		args = append(args, *q.Deleted)
		n++
	} else {
		sql += ` AND child.deleted = false`
	}
	sql += fmt.Sprintf(` ORDER BY a.id LIMIT $%d OFFSET $%d`, n, n+1)
	args = append(args, limitOrDefault(q.Limit), q.Offset)

	rows, err := r.db.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("list subcontacts: %w", err)
	}
	defer rows.Close()
	var out []models.SubContactAssoc
	for rows.Next() {
		var a models.SubContactAssoc
		if err := rows.Scan(&a.ID, &a.IDSubContact, &a.IDContactRole); err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

func (r *contactRepository) CreateSubContact(ctx context.Context, contactID int32, assoc *models.SubContactAssoc) (*models.SubContactAssoc, error) {
	row := r.db.QueryRow(ctx, `
		INSERT INTO contact_subcontact_assocs (parent_contact_id, child_contact_id, id_contact_role)
		SELECT $1, $2, $3
		FROM contacts parent
		JOIN contacts child ON child.id = $2
		WHERE parent.id = $1 AND parent.company_id = $4 AND parent.deleted = false AND child.deleted = false
		RETURNING id, child_contact_id, id_contact_role
	`, contactID, assoc.IDSubContact, assoc.IDContactRole, defaultCompanyID)
	var out models.SubContactAssoc
	if err := row.Scan(&out.ID, &out.IDSubContact, &out.IDContactRole); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("contact not found")
		}
		return nil, fmt.Errorf("create subcontact: %w", err)
	}
	return &out, nil
}

func (r *contactRepository) DeleteSubContact(ctx context.Context, contactID int32, subContactID int32) error {
	tag, err := r.db.Exec(ctx, `
		DELETE FROM contact_subcontact_assocs a
		USING contacts parent
		WHERE a.parent_contact_id = $1 AND a.child_contact_id = $3
			AND parent.id = a.parent_contact_id AND parent.company_id = $2
	`, contactID, defaultCompanyID, subContactID)
	if err != nil {
		return fmt.Errorf("delete subcontact: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("subcontact not found")
	}
	return nil
}

func scanContacts(rows pgx.Rows, retrieveImage bool) ([]models.Contact, error) {
	var out []models.Contact
	for rows.Next() {
		c, err := scanContactFromScanner(rows, retrieveImage)
		if err != nil {
			return nil, err
		}
		out = append(out, *c)
	}
	return out, rows.Err()
}

func scanContactRow(row pgx.Row, retrieveImage bool) (*models.Contact, error) {
	c, err := scanContactFromScanner(row, retrieveImage)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return c, nil
}

func scanContactFromScanner(s interface{ Scan(...any) error }, retrieveImage bool) (*models.Contact, error) {
	var c models.Contact
	dest := []any{
		&c.ID, &c.Type, &c.Name, &c.LastName, &c.Address, &c.City, &c.Country, &c.State, &c.PostalCode,
		&c.Pr, &c.FiscalCode, &c.VatCode, &c.Phone1, &c.Phone2, &c.Mobile, &c.Fax, &c.Email, &c.Email2,
		&c.VisibilityType, &c.Status, &c.IDUserOwner, &c.Source, &c.Note,
	}
	var photoData []byte
	var photoMime *string
	if retrieveImage {
		dest = append(dest, &photoData, &photoMime)
	}
	if err := s.Scan(dest...); err != nil {
		return nil, err
	}
	if retrieveImage && len(photoData) > 0 {
		encoded := base64.StdEncoding.EncodeToString(photoData)
		c.PhotoBase64 = &encoded
	}
	return &c, nil
}

func scanOffices(rows pgx.Rows) ([]models.ContactOffice, error) {
	var out []models.ContactOffice
	for rows.Next() {
		o, err := scanOfficeFromScanner(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *o)
	}
	return out, rows.Err()
}

func scanOfficeRow(row pgx.Row) (*models.ContactOffice, error) {
	o, err := scanOfficeFromScanner(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return o, nil
}

func scanOfficeFromScanner(s interface{ Scan(...any) error }) (*models.ContactOffice, error) {
	var o models.ContactOffice
	if err := s.Scan(&o.ID, &o.IDOfficeType, &o.Name, &o.Address, &o.City, &o.Pr, &o.PostalCode,
		&o.Country, &o.State, &o.Email, &o.Phone1, &o.Phone2, &o.Fax, &o.DefaultDest); err != nil {
		return nil, err
	}
	return &o, nil
}
