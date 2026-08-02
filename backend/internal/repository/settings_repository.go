package repository

import (
	"context"
	"fmt"
	"strings"

	"dnsc_microservice/internal/models"

	"github.com/jackc/pgx/v4/pgxpool"
)

// SettingsRepository persists application settings.
type SettingsRepository interface {
	ListSettings(ctx context.Context, q models.AppSettingListQuery) ([]models.AppSetting, error)
	UpsertSetting(ctx context.Context, s *models.AppSetting) (*models.AppSetting, error)
}

type settingsRepository struct {
	db *pgxpool.Pool
}

func NewSettingsRepository(db *pgxpool.Pool) SettingsRepository {
	return &settingsRepository{db: db}
}

func (r *settingsRepository) ListSettings(ctx context.Context, q models.AppSettingListQuery) ([]models.AppSetting, error) {
	sql := `
		SELECT COALESCE(key1,''), COALESCE(key2,''), COALESCE(key3,''), COALESCE(key4,''), COALESCE(key5,''), COALESCE(value,'')
		FROM app_settings WHERE company_id=$1`
	args := []any{defaultCompanyID}
	n := 2
	if s := strings.TrimSpace(q.Key1); s != "" {
		sql += fmt.Sprintf(` AND key1 = $%d`, n)
		args = append(args, s)
		n++
	}
	if s := strings.TrimSpace(q.Key2); s != "" {
		sql += fmt.Sprintf(` AND key2 = $%d`, n)
		args = append(args, s)
		n++
	}
	if s := strings.TrimSpace(q.Key3); s != "" {
		sql += fmt.Sprintf(` AND key3 = $%d`, n)
		args = append(args, s)
		n++
	}
	if s := strings.TrimSpace(q.Key4); s != "" {
		sql += fmt.Sprintf(` AND key4 = $%d`, n)
		args = append(args, s)
		n++
	}
	if s := strings.TrimSpace(q.Key5); s != "" {
		sql += fmt.Sprintf(` AND key5 = $%d`, n)
		args = append(args, s)
		n++
	}
	sql += ` ORDER BY key1, key2, key3, key4, key5`

	rows, err := r.db.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]models.AppSetting, 0)
	for rows.Next() {
		var s models.AppSetting
		if err := rows.Scan(&s.Key1, &s.Key2, &s.Key3, &s.Key4, &s.Key5, &s.Value); err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

func (r *settingsRepository) UpsertSetting(ctx context.Context, s *models.AppSetting) (*models.AppSetting, error) {
	_, err := r.db.Exec(ctx, `
		INSERT INTO app_settings (key1, key2, key3, key4, key5, value)
		VALUES ($1,$2,$3,$4,$5,$6)
		ON CONFLICT (company_id, key1, key2, key3, key4, key5)
		DO UPDATE SET value = EXCLUDED.value, updated_at = now()
	`, defaultStr(s.Key1, ""), defaultStr(s.Key2, ""), defaultStr(s.Key3, ""),
		defaultStr(s.Key4, ""), defaultStr(s.Key5, ""), emptyAsNull(s.Value))
	if err != nil {
		return nil, fmt.Errorf("upsert setting: %w", err)
	}
	items, err := r.ListSettings(ctx, models.AppSettingListQuery{
		Key1: s.Key1, Key2: s.Key2, Key3: s.Key3, Key4: s.Key4, Key5: s.Key5,
	})
	if err != nil || len(items) == 0 {
		return s, err
	}
	return &items[0], nil
}
