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

// ReferenceRepository persists lookup/reference data.
type ReferenceRepository interface {
	ListCountries(ctx context.Context, q models.CountryListQuery) ([]models.Country, error)
	GetCountry(ctx context.Context, id string) (*models.Country, error)

	ListCities(ctx context.Context, q models.CityListQuery) ([]models.City, error)
	CreateCity(ctx context.Context, city *models.City) (*models.City, error)

	ListCurrencies(ctx context.Context, q models.ListQuery) ([]models.Currency, error)
	GetCurrency(ctx context.Context, code string) (*models.Currency, error)

	ListUms(ctx context.Context, q models.UmListQuery) ([]models.Um, error)
	GetUm(ctx context.Context, idUm string, showDeleted bool) (*models.Um, error)
	CreateUm(ctx context.Context, um *models.Um) error
	UpdateUm(ctx context.Context, idUm string, um *models.Um) error
	DeleteUm(ctx context.Context, idUm string) error

	ListOfficeTypes(ctx context.Context, q models.OfficeTypeListQuery) ([]models.OfficeType, error)
	ListContactRoles(ctx context.Context, q models.ContactRoleListQuery) ([]models.ContactRole, error)

	ListPaymentTerms(ctx context.Context, q models.PaymentTermListQuery) ([]models.PaymentTerm, error)
	GetPaymentTerm(ctx context.Context, id int32) (*models.PaymentTerm, error)
	CreatePaymentTerm(ctx context.Context, term *models.PaymentTerm) (*models.PaymentTerm, error)
}

type referenceRepository struct {
	db *pgxpool.Pool
}

func NewReferenceRepository(db *pgxpool.Pool) ReferenceRepository {
	return &referenceRepository{db: db}
}

func (r *referenceRepository) ListCountries(ctx context.Context, q models.CountryListQuery) ([]models.Country, error) {
	sql := `SELECT id, description FROM countries WHERE 1=1`
	args := []any{}
	n := 1
	if s := strings.TrimSpace(q.Description); s != "" {
		sql += fmt.Sprintf(` AND description ILIKE $%d`, n)
		args = append(args, "%"+s+"%")
		n++
	}
	sql += fmt.Sprintf(` ORDER BY description LIMIT $%d OFFSET $%d`, n, n+1)
	args = append(args, limitOrDefault(q.Limit), q.Offset)

	rows, err := r.db.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("list countries: %w", err)
	}
	defer rows.Close()

	var out []models.Country
	for rows.Next() {
		var c models.Country
		if err := rows.Scan(&c.ID, &c.Description); err != nil {
			return nil, fmt.Errorf("scan country: %w", err)
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func (r *referenceRepository) GetCountry(ctx context.Context, id string) (*models.Country, error) {
	row := r.db.QueryRow(ctx, `SELECT id, description FROM countries WHERE id = $1`, strings.TrimSpace(id))
	var c models.Country
	if err := row.Scan(&c.ID, &c.Description); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("get country: %w", err)
	}
	return &c, nil
}

func (r *referenceRepository) ListCities(ctx context.Context, q models.CityListQuery) ([]models.City, error) {
	sql := `SELECT id, name, COALESCE(pr,''), COALESCE(zip,''), COALESCE(country,''), COALESCE(state,'') FROM cities WHERE 1=1`
	args := []any{}
	n := 1
	if s := strings.TrimSpace(q.Name); s != "" {
		sql += fmt.Sprintf(` AND name ILIKE $%d`, n)
		args = append(args, "%"+s+"%")
		n++
	}
	sql += fmt.Sprintf(` ORDER BY name LIMIT $%d OFFSET $%d`, n, n+1)
	args = append(args, limitOrDefault(q.Limit), q.Offset)

	rows, err := r.db.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("list cities: %w", err)
	}
	defer rows.Close()
	return scanCities(rows)
}

func (r *referenceRepository) CreateCity(ctx context.Context, city *models.City) (*models.City, error) {
	if city == nil {
		return nil, fmt.Errorf("city is nil")
	}
	row := r.db.QueryRow(ctx, `
		INSERT INTO cities (name, pr, zip, country, state)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, name, COALESCE(pr,''), COALESCE(zip,''), COALESCE(country,''), COALESCE(state,'')
	`, city.Name, nullIfEmpty(city.Pr), nullIfEmpty(city.Zip), nullIfEmpty(city.Country), nullIfEmpty(city.State))
	var out models.City
	if err := row.Scan(&out.ID, &out.Name, &out.Pr, &out.Zip, &out.Country, &out.State); err != nil {
		return nil, fmt.Errorf("create city: %w", err)
	}
	return &out, nil
}

func (r *referenceRepository) ListCurrencies(ctx context.Context, q models.ListQuery) ([]models.Currency, error) {
	rows, err := r.db.Query(ctx, `
		SELECT code, description FROM currencies
		ORDER BY code LIMIT $1 OFFSET $2
	`, limitOrDefault(q.Limit), q.Offset)
	if err != nil {
		return nil, fmt.Errorf("list currencies: %w", err)
	}
	defer rows.Close()

	var out []models.Currency
	for rows.Next() {
		var c models.Currency
		if err := rows.Scan(&c.Code, &c.Description); err != nil {
			return nil, fmt.Errorf("scan currency: %w", err)
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func (r *referenceRepository) GetCurrency(ctx context.Context, code string) (*models.Currency, error) {
	row := r.db.QueryRow(ctx, `SELECT code, description FROM currencies WHERE code = $1`, strings.TrimSpace(code))
	var c models.Currency
	if err := row.Scan(&c.Code, &c.Description); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("get currency: %w", err)
	}
	return &c, nil
}

func (r *referenceRepository) ListUms(ctx context.Context, q models.UmListQuery) ([]models.Um, error) {
	sql := `SELECT um,
		COALESCE(description_it,''), COALESCE(description_en,''), COALESCE(description_es,''), COALESCE(description_bg,''),
		COALESCE(short_description_it,''), COALESCE(short_description_en,''), COALESCE(short_description_es,''), COALESCE(short_description_bg,'')
		FROM units_of_measure WHERE 1=1`
	args := []any{}
	n := 1
	if !q.ShowDeleted {
		sql += ` AND deleted_at IS NULL`
	}
	if s := strings.TrimSpace(q.Description); s != "" {
		sql += fmt.Sprintf(` AND (description_it ILIKE $%d OR description_en ILIKE $%d)`, n, n)
		args = append(args, "%"+s+"%")
		n++
	}
	if s := strings.TrimSpace(q.ShortDescription); s != "" {
		sql += fmt.Sprintf(` AND (short_description_it ILIKE $%d OR short_description_en ILIKE $%d)`, n, n)
		args = append(args, "%"+s+"%")
		n++
	}
	sql += fmt.Sprintf(` ORDER BY um LIMIT $%d OFFSET $%d`, n, n+1)
	args = append(args, limitOrDefault(q.Limit), q.Offset)

	rows, err := r.db.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("list ums: %w", err)
	}
	defer rows.Close()
	return scanUms(rows)
}

func (r *referenceRepository) GetUm(ctx context.Context, idUm string, showDeleted bool) (*models.Um, error) {
	sql := `SELECT um,
		COALESCE(description_it,''), COALESCE(description_en,''), COALESCE(description_es,''), COALESCE(description_bg,''),
		COALESCE(short_description_it,''), COALESCE(short_description_en,''), COALESCE(short_description_es,''), COALESCE(short_description_bg,'')
		FROM units_of_measure WHERE um = $1`
	if !showDeleted {
		sql += ` AND deleted_at IS NULL`
	}
	row := r.db.QueryRow(ctx, sql, strings.TrimSpace(idUm))
	return scanUmRow(row)
}

func (r *referenceRepository) CreateUm(ctx context.Context, um *models.Um) error {
	if um == nil || strings.TrimSpace(um.Um) == "" {
		return fmt.Errorf("um code is required")
	}
	_, err := r.db.Exec(ctx, `
		INSERT INTO units_of_measure (
			um, description_it, description_en, description_es, description_bg,
			short_description_it, short_description_en, short_description_es, short_description_bg
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
	`, um.Um, um.DescriptionIT, um.DescriptionEN, um.DescriptionES, um.DescriptionBG,
		um.ShortDescriptionIT, um.ShortDescriptionEN, um.ShortDescriptionES, um.ShortDescriptionBG)
	if err != nil {
		return fmt.Errorf("create um: %w", err)
	}
	return nil
}

func (r *referenceRepository) UpdateUm(ctx context.Context, idUm string, um *models.Um) error {
	tag, err := r.db.Exec(ctx, `
		UPDATE units_of_measure SET
			description_it = $2, description_en = $3, description_es = $4, description_bg = $5,
			short_description_it = $6, short_description_en = $7, short_description_es = $8, short_description_bg = $9
		WHERE um = $1 AND deleted_at IS NULL
	`, strings.TrimSpace(idUm), um.DescriptionIT, um.DescriptionEN, um.DescriptionES, um.DescriptionBG,
		um.ShortDescriptionIT, um.ShortDescriptionEN, um.ShortDescriptionES, um.ShortDescriptionBG)
	if err != nil {
		return fmt.Errorf("update um: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("um not found")
	}
	return nil
}

func (r *referenceRepository) DeleteUm(ctx context.Context, idUm string) error {
	tag, err := r.db.Exec(ctx, `
		UPDATE units_of_measure SET deleted_at = now() WHERE um = $1 AND deleted_at IS NULL
	`, strings.TrimSpace(idUm))
	if err != nil {
		return fmt.Errorf("delete um: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("um not found")
	}
	return nil
}

func (r *referenceRepository) ListOfficeTypes(ctx context.Context, q models.OfficeTypeListQuery) ([]models.OfficeType, error) {
	sql := `SELECT id, description FROM office_types WHERE 1=1`
	args := []any{}
	n := 1
	if s := strings.TrimSpace(q.Description); s != "" {
		sql += fmt.Sprintf(` AND description ILIKE $%d`, n)
		args = append(args, "%"+s+"%")
		n++
	}
	sql += fmt.Sprintf(` ORDER BY description LIMIT $%d OFFSET $%d`, n, n+1)
	args = append(args, limitOrDefault(q.Limit), q.Offset)

	rows, err := r.db.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("list office types: %w", err)
	}
	defer rows.Close()
	return scanOfficeTypes(rows)
}

func (r *referenceRepository) ListContactRoles(ctx context.Context, q models.ContactRoleListQuery) ([]models.ContactRole, error) {
	sql := `SELECT id, description FROM contact_roles WHERE 1=1`
	args := []any{}
	n := 1
	if s := strings.TrimSpace(q.Description); s != "" {
		sql += fmt.Sprintf(` AND description ILIKE $%d`, n)
		args = append(args, "%"+s+"%")
		n++
	}
	sql += fmt.Sprintf(` ORDER BY description LIMIT $%d OFFSET $%d`, n, n+1)
	args = append(args, limitOrDefault(q.Limit), q.Offset)

	rows, err := r.db.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("list contact roles: %w", err)
	}
	defer rows.Close()
	return scanContactRoles(rows)
}

func (r *referenceRepository) ListPaymentTerms(ctx context.Context, q models.PaymentTermListQuery) ([]models.PaymentTerm, error) {
	sql := `SELECT id, id_payment_type, description, end_of_month, extra_days FROM payment_terms WHERE 1=1`
	args := []any{}
	n := 1
	if s := strings.TrimSpace(q.Description); s != "" {
		sql += fmt.Sprintf(` AND description ILIKE $%d`, n)
		args = append(args, "%"+s+"%")
		n++
	}
	sql += fmt.Sprintf(` ORDER BY id LIMIT $%d OFFSET $%d`, n, n+1)
	args = append(args, limitOrDefault(q.Limit), q.Offset)

	rows, err := r.db.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("list payment terms: %w", err)
	}
	defer rows.Close()

	var terms []models.PaymentTerm
	for rows.Next() {
		var t models.PaymentTerm
		var idPaymentType *int32
		if err := rows.Scan(&t.ID, &idPaymentType, &t.Description, &t.EndOfMonth, &t.ExtraDays); err != nil {
			return nil, fmt.Errorf("scan payment term: %w", err)
		}
		t.IDPaymentType = idPaymentType
		terms = append(terms, t)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	for i := range terms {
		pos, err := r.loadPaymentTermPos(ctx, terms[i].ID)
		if err != nil {
			return nil, err
		}
		terms[i].Pos = pos
	}
	return terms, nil
}

func (r *referenceRepository) GetPaymentTerm(ctx context.Context, id int32) (*models.PaymentTerm, error) {
	row := r.db.QueryRow(ctx, `
		SELECT id, id_payment_type, description, end_of_month, extra_days FROM payment_terms WHERE id = $1
	`, id)
	t, err := scanPaymentTermRow(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("get payment term: %w", err)
	}
	if t == nil {
		return nil, nil
	}
	pos, err := r.loadPaymentTermPos(ctx, t.ID)
	if err != nil {
		return nil, err
	}
	t.Pos = pos
	return t, nil
}

func (r *referenceRepository) CreatePaymentTerm(ctx context.Context, term *models.PaymentTerm) (*models.PaymentTerm, error) {
	if term == nil {
		return nil, fmt.Errorf("payment term is nil")
	}
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	var id int32
	var idPaymentType *int32
	err = tx.QueryRow(ctx, `
		INSERT INTO payment_terms (id_payment_type, description, end_of_month, extra_days)
		VALUES ($1, $2, $3, $4)
		RETURNING id, id_payment_type
	`, term.IDPaymentType, term.Description, term.EndOfMonth, term.ExtraDays).Scan(&id, &idPaymentType)
	if err != nil {
		return nil, fmt.Errorf("insert payment term: %w", err)
	}

	pos, err := r.insertPaymentTermPos(ctx, tx, id, term.Pos)
	if err != nil {
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}

	return &models.PaymentTerm{
		ID:            id,
		IDPaymentType: idPaymentType,
		Description:   term.Description,
		EndOfMonth:    term.EndOfMonth,
		ExtraDays:     term.ExtraDays,
		Pos:           pos,
	}, nil
}

func (r *referenceRepository) loadPaymentTermPos(ctx context.Context, termID int32) ([]models.PaymentTermPos, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, days, percentage FROM payment_term_positions
		WHERE payment_term_id = $1 ORDER BY position_order, id
	`, termID)
	if err != nil {
		return nil, fmt.Errorf("load payment term pos: %w", err)
	}
	defer rows.Close()

	var out []models.PaymentTermPos
	for rows.Next() {
		var p models.PaymentTermPos
		if err := rows.Scan(&p.ID, &p.Days, &p.Percentage); err != nil {
			return nil, fmt.Errorf("scan pos: %w", err)
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

func (r *referenceRepository) insertPaymentTermPos(ctx context.Context, tx pgx.Tx, termID int32, pos []models.PaymentTermPos) ([]models.PaymentTermPos, error) {
	var out []models.PaymentTermPos
	for i, p := range pos {
		var id int32
		err := tx.QueryRow(ctx, `
			INSERT INTO payment_term_positions (payment_term_id, days, percentage, position_order)
			VALUES ($1, $2, $3, $4) RETURNING id
		`, termID, p.Days, p.Percentage, i+1).Scan(&id)
		if err != nil {
			return nil, fmt.Errorf("insert pos: %w", err)
		}
		out = append(out, models.PaymentTermPos{ID: id, Days: p.Days, Percentage: p.Percentage})
	}
	return out, nil
}

func scanCities(rows pgx.Rows) ([]models.City, error) {
	var out []models.City
	for rows.Next() {
		var c models.City
		if err := rows.Scan(&c.ID, &c.Name, &c.Pr, &c.Zip, &c.Country, &c.State); err != nil {
			return nil, fmt.Errorf("scan city: %w", err)
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func scanUms(rows pgx.Rows) ([]models.Um, error) {
	var out []models.Um
	for rows.Next() {
		u, err := scanUmRow(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *u)
	}
	return out, rows.Err()
}

func scanUmRow(row pgx.Row) (*models.Um, error) {
	var u models.Um
	if err := row.Scan(&u.Um,
		&u.DescriptionIT, &u.DescriptionEN, &u.DescriptionES, &u.DescriptionBG,
		&u.ShortDescriptionIT, &u.ShortDescriptionEN, &u.ShortDescriptionES, &u.ShortDescriptionBG); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("scan um: %w", err)
	}
	return &u, nil
}

func scanOfficeTypes(rows pgx.Rows) ([]models.OfficeType, error) {
	var out []models.OfficeType
	for rows.Next() {
		var o models.OfficeType
		if err := rows.Scan(&o.ID, &o.Description); err != nil {
			return nil, fmt.Errorf("scan office type: %w", err)
		}
		out = append(out, o)
	}
	return out, rows.Err()
}

func scanContactRoles(rows pgx.Rows) ([]models.ContactRole, error) {
	var out []models.ContactRole
	for rows.Next() {
		var c models.ContactRole
		if err := rows.Scan(&c.ID, &c.Description); err != nil {
			return nil, fmt.Errorf("scan contact role: %w", err)
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func scanPaymentTermRow(row pgx.Row) (*models.PaymentTerm, error) {
	var t models.PaymentTerm
	var idPaymentType *int32
	if err := row.Scan(&t.ID, &idPaymentType, &t.Description, &t.EndOfMonth, &t.ExtraDays); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	t.IDPaymentType = idPaymentType
	return &t, nil
}

func limitOrDefault(limit int) int {
	if limit <= 0 {
		return 100
	}
	if limit > 1000 {
		return 1000
	}
	return limit
}

func nullIfEmpty(s string) *string {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	return &s
}
