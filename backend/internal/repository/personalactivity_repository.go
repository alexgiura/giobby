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

// PersonalActivityRepository persists personal activities.
type PersonalActivityRepository interface {
	List(ctx context.Context, q models.PersonalActivityListQuery) ([]models.PersonalActivity, error)
	Get(ctx context.Context, id int32) (*models.PersonalActivity, error)
	Create(ctx context.Context, a *models.PersonalActivity) (*models.PersonalActivity, error)
	Update(ctx context.Context, id int32, a *models.PersonalActivity) (*models.PersonalActivity, error)
	Patch(ctx context.Context, id int32, fields map[string]any) (*models.PersonalActivity, error)
	Delete(ctx context.Context, id int32) error
}

type personalActivityRepository struct {
	db *pgxpool.Pool
}

func NewPersonalActivityRepository(db *pgxpool.Pool) PersonalActivityRepository {
	return &personalActivityRepository{db: db}
}

const personalActivitySelect = `
	SELECT id, COALESCE(description,''), COALESCE(root_name,''), activity_date, COALESCE(status,'')
	FROM personal_activities WHERE company_id=$1 AND deleted=false`

func (r *personalActivityRepository) List(ctx context.Context, q models.PersonalActivityListQuery) ([]models.PersonalActivity, error) {
	sql := personalActivitySelect
	args := []any{defaultCompanyID}
	n := 2
	if s := strings.TrimSpace(q.Status); s != "" {
		sql += fmt.Sprintf(` AND status = $%d`, n)
		args = append(args, s)
		n++
	}
	if s := strings.TrimSpace(q.Name); s != "" {
		sql += fmt.Sprintf(` AND (description ILIKE $%d OR root_name ILIKE $%d)`, n, n)
		args = append(args, "%"+s+"%")
		n++
	}
	sql += fmt.Sprintf(` ORDER BY id DESC LIMIT $%d OFFSET $%d`, n, n+1)
	args = append(args, limitOrDefault(q.Limit), q.Offset)

	rows, err := r.db.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]models.PersonalActivity, 0)
	for rows.Next() {
		a, err := scanPersonalActivity(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *a)
	}
	return out, rows.Err()
}

func (r *personalActivityRepository) Get(ctx context.Context, id int32) (*models.PersonalActivity, error) {
	row := r.db.QueryRow(ctx, personalActivitySelect+` AND id=$2`, defaultCompanyID, id)
	a, err := scanPersonalActivity(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return a, nil
}

func (r *personalActivityRepository) Create(ctx context.Context, a *models.PersonalActivity) (*models.PersonalActivity, error) {
	var id int32
	err := r.db.QueryRow(ctx, `
		INSERT INTO personal_activities (description, root_name, activity_date, status)
		VALUES ($1,$2,$3,$4) RETURNING id
	`, emptyAsNull(a.Description), emptyAsNull(a.RootName), millisToTimePtr(a.Date), defaultStr(a.Status, "OPEN")).Scan(&id)
	if err != nil {
		return nil, fmt.Errorf("create personal activity: %w", err)
	}
	return r.Get(ctx, id)
}

func (r *personalActivityRepository) Update(ctx context.Context, id int32, a *models.PersonalActivity) (*models.PersonalActivity, error) {
	tag, err := r.db.Exec(ctx, `
		UPDATE personal_activities SET description=$2, root_name=$3, activity_date=$4, status=$5, updated_at=now()
		WHERE id=$1 AND company_id=$6 AND deleted=false
	`, id, emptyAsNull(a.Description), emptyAsNull(a.RootName), millisToTimePtr(a.Date),
		defaultStr(a.Status, "OPEN"), defaultCompanyID)
	if err != nil {
		return nil, err
	}
	if tag.RowsAffected() == 0 {
		return nil, fmt.Errorf("personal activity not found")
	}
	return r.Get(ctx, id)
}

func (r *personalActivityRepository) Patch(ctx context.Context, id int32, fields map[string]any) (*models.PersonalActivity, error) {
	setSQL, args, err := buildPatchSet(personalActivityPatchColumns, fields, 3)
	if err != nil {
		return nil, err
	}
	if setSQL == "" {
		return r.Get(ctx, id)
	}
	args = append([]any{id, defaultCompanyID}, args...)
	sql := fmt.Sprintf(`UPDATE personal_activities SET %s, updated_at=now() WHERE id=$1 AND company_id=$2 AND deleted=false`, setSQL)
	tag, err := r.db.Exec(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	if tag.RowsAffected() == 0 {
		return nil, fmt.Errorf("personal activity not found")
	}
	return r.Get(ctx, id)
}

func (r *personalActivityRepository) Delete(ctx context.Context, id int32) error {
	tag, err := r.db.Exec(ctx, `
		UPDATE personal_activities SET deleted=true, updated_at=now() WHERE id=$1 AND company_id=$2 AND deleted=false
	`, id, defaultCompanyID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("personal activity not found")
	}
	return nil
}

var personalActivityPatchColumns = map[string]string{
	"description": "description",
	"rootName":    "root_name",
	"status":      "status",
}

func scanPersonalActivity(s interface{ Scan(...any) error }) (*models.PersonalActivity, error) {
	var a models.PersonalActivity
	var activityDate *time.Time
	if err := s.Scan(&a.ID, &a.Description, &a.RootName, &activityDate, &a.Status); err != nil {
		return nil, err
	}
	if activityDate != nil {
		a.Date = models.MillisPtr(*activityDate)
	}
	return &a, nil
}
