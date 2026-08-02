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

// CalendarRepository persists calendars and events.
type CalendarRepository interface {
	ListCalendars(ctx context.Context, q models.CalendarListQuery) ([]models.Calendar, error)
	ListEvents(ctx context.Context, q models.CalendarEventListQuery) ([]models.CalendarEvent, error)
	GetEvent(ctx context.Context, id int32) (*models.CalendarEvent, error)
	CreateEvent(ctx context.Context, event *models.CalendarEvent) (*models.CalendarEvent, error)
	UpdateEvent(ctx context.Context, id int32, event *models.CalendarEvent) (*models.CalendarEvent, error)
	DeleteEvent(ctx context.Context, id int32) error
}

type calendarRepository struct {
	db *pgxpool.Pool
}

func NewCalendarRepository(db *pgxpool.Pool) CalendarRepository {
	return &calendarRepository{db: db}
}

func (r *calendarRepository) ListCalendars(ctx context.Context, q models.CalendarListQuery) ([]models.Calendar, error) {
	sql := `SELECT id, name, COALESCE(color,'') FROM calendars WHERE company_id=$1 AND deleted=false`
	args := []any{defaultCompanyID}
	n := 2
	if s := strings.TrimSpace(q.Name); s != "" {
		sql += fmt.Sprintf(` AND name ILIKE $%d`, n)
		args = append(args, "%"+s+"%")
		n++
	}
	sql += fmt.Sprintf(` ORDER BY id LIMIT $%d OFFSET $%d`, n, n+1)
	args = append(args, limitOrDefault(q.Limit), q.Offset)

	rows, err := r.db.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]models.Calendar, 0)
	for rows.Next() {
		var c models.Calendar
		if err := rows.Scan(&c.ID, &c.Name, &c.Color); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func (r *calendarRepository) ListEvents(ctx context.Context, q models.CalendarEventListQuery) ([]models.CalendarEvent, error) {
	sql := calendarEventSelect + ` WHERE e.company_id=$1 AND e.deleted=false`
	args := []any{defaultCompanyID}
	n := 2
	if s := strings.TrimSpace(q.StartDateMillis); s != "" {
		if t, err := parseMillisFilter(s); err == nil {
			sql += fmt.Sprintf(` AND e.end_date >= $%d`, n)
			args = append(args, t)
			n++
		}
	}
	if s := strings.TrimSpace(q.EndDateMillis); s != "" {
		if t, err := parseMillisFilter(s); err == nil {
			sql += fmt.Sprintf(` AND e.start_date <= $%d`, n)
			args = append(args, t)
			n++
		}
	}
	if s := strings.TrimSpace(q.FreeText); s != "" {
		sql += fmt.Sprintf(` AND (e.event ILIKE $%d OR COALESCE(e.place,'') ILIKE $%d)`, n, n)
		args = append(args, "%"+s+"%")
		n++
	}
	sql += fmt.Sprintf(` ORDER BY e.start_date LIMIT $%d OFFSET $%d`, n, n+1)
	args = append(args, limitOrDefault(q.Limit), q.Offset)

	rows, err := r.db.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanCalendarEvents(rows)
}

func (r *calendarRepository) GetEvent(ctx context.Context, id int32) (*models.CalendarEvent, error) {
	row := r.db.QueryRow(ctx, calendarEventSelect+` WHERE e.id=$1 AND e.company_id=$2 AND e.deleted=false`, id, defaultCompanyID)
	ev, err := scanCalendarEvent(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return ev, nil
}

func (r *calendarRepository) CreateEvent(ctx context.Context, event *models.CalendarEvent) (*models.CalendarEvent, error) {
	if event.IDCalendar <= 0 {
		return nil, fmt.Errorf("idCalendar is required")
	}
	if strings.TrimSpace(event.Event) == "" {
		return nil, fmt.Errorf("event is required")
	}
	var id int32
	err := r.db.QueryRow(ctx, `
		INSERT INTO calendar_events (id_calendar, event, place, start_date, end_date, color, privacy)
		VALUES ($1,$2,$3,$4,$5,$6,$7) RETURNING id
	`, event.IDCalendar, event.Event, emptyAsNull(event.Place),
		models.MillisToTime(event.StartDate), models.MillisToTime(event.EndDate),
		defaultStr(event.Color, "#3366FF"), defaultStr(event.Privacy, "PUBLIC")).Scan(&id)
	if err != nil {
		return nil, fmt.Errorf("create calendar event: %w", err)
	}
	return r.GetEvent(ctx, id)
}

func (r *calendarRepository) UpdateEvent(ctx context.Context, id int32, event *models.CalendarEvent) (*models.CalendarEvent, error) {
	tag, err := r.db.Exec(ctx, `
		UPDATE calendar_events SET
			id_calendar=$2, event=$3, place=$4, start_date=$5, end_date=$6, color=$7, privacy=$8, updated_at=now()
		WHERE id=$1 AND company_id=$9 AND deleted=false
	`, id, nullInt32(event.IDCalendar), event.Event, emptyAsNull(event.Place),
		models.MillisToTime(event.StartDate), models.MillisToTime(event.EndDate),
		defaultStr(event.Color, "#3366FF"), defaultStr(event.Privacy, "PUBLIC"), defaultCompanyID)
	if err != nil {
		return nil, err
	}
	if tag.RowsAffected() == 0 {
		return nil, fmt.Errorf("calendar event not found")
	}
	return r.GetEvent(ctx, id)
}

func (r *calendarRepository) DeleteEvent(ctx context.Context, id int32) error {
	tag, err := r.db.Exec(ctx, `
		UPDATE calendar_events SET deleted=true, updated_at=now() WHERE id=$1 AND company_id=$2 AND deleted=false
	`, id, defaultCompanyID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("calendar event not found")
	}
	return nil
}

const calendarEventSelect = `
	SELECT e.id, e.event, COALESCE(e.place,''), e.start_date, e.end_date, COALESCE(e.color,''),
		e.id_calendar, COALESCE(e.privacy,'')
	FROM calendar_events e`

func scanCalendarEvents(rows pgx.Rows) ([]models.CalendarEvent, error) {
	var out []models.CalendarEvent
	for rows.Next() {
		ev, err := scanCalendarEvent(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *ev)
	}
	if out == nil {
		out = make([]models.CalendarEvent, 0)
	}
	return out, rows.Err()
}

func scanCalendarEvent(s interface{ Scan(...any) error }) (*models.CalendarEvent, error) {
	var ev models.CalendarEvent
	var start, end time.Time
	if err := s.Scan(&ev.ID, &ev.Event, &ev.Place, &start, &end, &ev.Color, &ev.IDCalendar, &ev.Privacy); err != nil {
		return nil, err
	}
	ev.StartDate = models.MillisPtr(start)
	ev.EndDate = models.MillisPtr(end)
	return &ev, nil
}
