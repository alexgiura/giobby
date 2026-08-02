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

// UserRepository persists Giobby company users.
type UserRepository interface {
	ListUsers(ctx context.Context, q models.CompanyUserListQuery) ([]models.CompanyUser, error)
	GetUser(ctx context.Context, id int32, retrieveImage bool) (*models.CompanyUser, error)
	CreateUser(ctx context.Context, user *models.CompanyUser, passwordHash string) (*models.CompanyUser, error)
	UpdateUser(ctx context.Context, id int32, user *models.CompanyUser, passwordHash string) (*models.CompanyUser, error)
	DeleteUser(ctx context.Context, id int32) error
	ListAgentCommissions(ctx context.Context, q models.AgentCommissionListQuery) ([]models.AgentCommission, error)
	AddAuthProfiles(ctx context.Context, idUser int32, profiles []string) error
}

type userRepository struct {
	db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) UserRepository {
	return &userRepository{db: db}
}

const companyUserSelect = `
	SELECT id, is_active, username, COALESCE(email,''), COALESCE(firstname,''), COALESCE(lastname,''),
		COALESCE(mobilephone,''), is_agent1, is_agent2, is_employee`

func (r *userRepository) ListUsers(ctx context.Context, q models.CompanyUserListQuery) ([]models.CompanyUser, error) {
	sql := companyUserSelect
	if q.RetrieveImage {
		sql += `, image_data, COALESCE(image_mime,'')`
	}
	sql += ` FROM company_users WHERE company_id=$1 AND deleted=false`
	args := []any{defaultCompanyID}
	n := 2
	if s := strings.TrimSpace(q.IsAgent1); s != "" {
		if b, err := parseBoolString(s); err == nil {
			sql += fmt.Sprintf(` AND is_agent1 = $%d`, n)
			args = append(args, b)
			n++
		}
	}
	if s := strings.TrimSpace(q.IsAgent2); s != "" {
		if b, err := parseBoolString(s); err == nil {
			sql += fmt.Sprintf(` AND is_agent2 = $%d`, n)
			args = append(args, b)
			n++
		}
	}
	sql += fmt.Sprintf(` ORDER BY id LIMIT $%d OFFSET $%d`, n, n+1)
	args = append(args, limitOrDefault(q.Limit), q.Offset)

	rows, err := r.db.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]models.CompanyUser, 0)
	for rows.Next() {
		u, err := scanCompanyUser(rows, q.RetrieveImage)
		if err != nil {
			return nil, err
		}
		out = append(out, *u)
	}
	return out, rows.Err()
}

func (r *userRepository) GetUser(ctx context.Context, id int32, retrieveImage bool) (*models.CompanyUser, error) {
	sql := companyUserSelect
	if retrieveImage {
		sql += `, image_data, COALESCE(image_mime,'')`
	}
	sql += ` FROM company_users WHERE id=$1 AND company_id=$2 AND deleted=false`
	row := r.db.QueryRow(ctx, sql, id, defaultCompanyID)
	u, err := scanCompanyUser(row, retrieveImage)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return u, nil
}

func (r *userRepository) CreateUser(ctx context.Context, user *models.CompanyUser, passwordHash string) (*models.CompanyUser, error) {
	var id int32
	var pwd any
	if passwordHash != "" {
		pwd = passwordHash
	}
	err := r.db.QueryRow(ctx, `
		INSERT INTO company_users (
			username, password_hash, email, firstname, lastname, mobilephone,
			is_active, is_agent1, is_agent2, is_employee
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10) RETURNING id
	`, user.Username, pwd, emptyAsNull(user.Email), emptyAsNull(user.Firstname), emptyAsNull(user.Lastname),
		emptyAsNull(user.Mobilephone), user.Active, user.IsAgent1, user.IsAgent2, user.IsEmployee).Scan(&id)
	if err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}
	return r.GetUser(ctx, id, false)
}

func (r *userRepository) UpdateUser(ctx context.Context, id int32, user *models.CompanyUser, passwordHash string) (*models.CompanyUser, error) {
	if passwordHash != "" {
		tag, err := r.db.Exec(ctx, `
			UPDATE company_users SET
				username=$2, password_hash=$3, email=$4, firstname=$5, lastname=$6, mobilephone=$7,
				is_active=$8, is_agent1=$9, is_agent2=$10, is_employee=$11, updated_at=now()
			WHERE id=$1 AND company_id=$12 AND deleted=false
		`, id, user.Username, passwordHash, emptyAsNull(user.Email), emptyAsNull(user.Firstname), emptyAsNull(user.Lastname),
			emptyAsNull(user.Mobilephone), user.Active, user.IsAgent1, user.IsAgent2, user.IsEmployee, defaultCompanyID)
		if err != nil {
			return nil, err
		}
		if tag.RowsAffected() == 0 {
			return nil, fmt.Errorf("user not found")
		}
	} else {
		tag, err := r.db.Exec(ctx, `
			UPDATE company_users SET
				username=$2, email=$3, firstname=$4, lastname=$5, mobilephone=$6,
				is_active=$7, is_agent1=$8, is_agent2=$9, is_employee=$10, updated_at=now()
			WHERE id=$1 AND company_id=$11 AND deleted=false
		`, id, user.Username, emptyAsNull(user.Email), emptyAsNull(user.Firstname), emptyAsNull(user.Lastname),
			emptyAsNull(user.Mobilephone), user.Active, user.IsAgent1, user.IsAgent2, user.IsEmployee, defaultCompanyID)
		if err != nil {
			return nil, err
		}
		if tag.RowsAffected() == 0 {
			return nil, fmt.Errorf("user not found")
		}
	}
	return r.GetUser(ctx, id, false)
}

func (r *userRepository) DeleteUser(ctx context.Context, id int32) error {
	tag, err := r.db.Exec(ctx, `
		UPDATE company_users SET deleted=true, updated_at=now() WHERE id=$1 AND company_id=$2 AND deleted=false
	`, id, defaultCompanyID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("user not found")
	}
	return nil
}

func (r *userRepository) ListAgentCommissions(ctx context.Context, q models.AgentCommissionListQuery) ([]models.AgentCommission, error) {
	sql := `
		SELECT id, user_id, COALESCE(id_bu,''), COALESCE(id_customer,''), id_material_group,
			COALESCE(id_material,''), commission_rate, COALESCE(description,'')
		FROM user_agent_commissions WHERE company_id=$1`
	args := []any{defaultCompanyID}
	n := 2
	if q.IDUser != nil {
		sql += fmt.Sprintf(` AND user_id = $%d`, n)
		args = append(args, *q.IDUser)
		n++
	}
	if s := strings.TrimSpace(q.IDBu); s != "" {
		sql += fmt.Sprintf(` AND id_bu = $%d`, n)
		args = append(args, s)
		n++
	}
	if s := strings.TrimSpace(q.IDCustomer); s != "" {
		sql += fmt.Sprintf(` AND id_customer = $%d`, n)
		args = append(args, s)
		n++
	}
	if q.IDMaterialGroup != nil {
		sql += fmt.Sprintf(` AND id_material_group = $%d`, n)
		args = append(args, *q.IDMaterialGroup)
		n++
	}
	if s := strings.TrimSpace(q.IDMaterial); s != "" {
		sql += fmt.Sprintf(` AND id_material = $%d`, n)
		args = append(args, s)
		n++
	}
	sql += fmt.Sprintf(` ORDER BY id LIMIT $%d OFFSET $%d`, n, n+1)
	args = append(args, limitOrDefault(q.Limit), q.Offset)

	rows, err := r.db.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]models.AgentCommission, 0)
	for rows.Next() {
		var c models.AgentCommission
		if err := rows.Scan(&c.ID, &c.IDUser, &c.IDBu, &c.IDCustomer, &c.IDMaterialGroup,
			&c.IDMaterial, &c.CommissionRate, &c.Description); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func (r *userRepository) AddAuthProfiles(ctx context.Context, idUser int32, profiles []string) error {
	var exists int
	if err := r.db.QueryRow(ctx, `SELECT 1 FROM company_users WHERE id=$1 AND company_id=$2 AND deleted=false`, idUser, defaultCompanyID).Scan(&exists); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("user not found")
		}
		return err
	}
	for _, p := range profiles {
		name := strings.TrimSpace(p)
		if name == "" {
			continue
		}
		_, _ = r.db.Exec(ctx, `
			INSERT INTO user_auth_profiles (user_id, profile_name) VALUES ($1,$2)
			ON CONFLICT (user_id, profile_name) DO NOTHING
		`, idUser, name)
	}
	return nil
}

func scanCompanyUser(s interface{ Scan(...any) error }, withImage bool) (*models.CompanyUser, error) {
	var u models.CompanyUser
	if !withImage {
		if err := s.Scan(&u.ID, &u.Active, &u.Username, &u.Email, &u.Firstname, &u.Lastname,
			&u.Mobilephone, &u.IsAgent1, &u.IsAgent2, &u.IsEmployee); err != nil {
			return nil, err
		}
		return &u, nil
	}
	var image []byte
	var mime string
	if err := s.Scan(&u.ID, &u.Active, &u.Username, &u.Email, &u.Firstname, &u.Lastname,
		&u.Mobilephone, &u.IsAgent1, &u.IsAgent2, &u.IsEmployee, &image, &mime); err != nil {
		return nil, err
	}
	if len(image) > 0 {
		u.Image = base64.StdEncoding.EncodeToString(image)
	}
	return &u, nil
}

func parseBoolString(s string) (bool, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "true", "1", "yes":
		return true, nil
	case "false", "0", "no":
		return false, nil
	default:
		return false, fmt.Errorf("invalid bool")
	}
}
