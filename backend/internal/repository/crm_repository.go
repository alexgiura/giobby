package repository

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"dnsc_microservice/internal/models"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

// CrmRepository persists CRM accounts, tasks and activities.
type CrmRepository interface {
	CreateAccount(ctx context.Context, account *models.CrmAccount) (*models.CrmAccount, error)
	GlobalReport(ctx context.Context) (map[string]any, error)
	AccountReport(ctx context.Context, idCrmAccount int32) (map[string]any, error)
	ListTasks(ctx context.Context, q models.CrmTaskListQuery) ([]models.CrmTask, error)
	GetTask(ctx context.Context, id int32) (*models.CrmTask, error)
	CreateTask(ctx context.Context, idParent int32, task *models.CrmTask) (*models.CrmTask, error)
	UpdateTask(ctx context.Context, idParent, id int32, task *models.CrmTask) (*models.CrmTask, error)
	PatchTask(ctx context.Context, idParent, id int32, fields map[string]any) (*models.CrmTask, error)
	DeleteTask(ctx context.Context, idParent, id int32) error
	ListActivities(ctx context.Context, idParent int32, q models.CrmActivityListQuery) ([]models.CrmActivity, error)
	GetActivity(ctx context.Context, idParent, id int32) (*models.CrmActivity, error)
	CreateActivity(ctx context.Context, idParent int32, act *models.CrmActivity) (*models.CrmActivity, error)
	UpdateActivity(ctx context.Context, idParent, id int32, act *models.CrmActivity) (*models.CrmActivity, error)
	PatchActivity(ctx context.Context, idParent, id int32, fields map[string]any) (*models.CrmActivity, error)
	DeleteActivity(ctx context.Context, idParent, id int32) error
	SaveAttachment(ctx context.Context, idParent int32, filename, contentType string, data []byte) (*models.CrmAttachmentResult, error)
}

type crmRepository struct {
	db *pgxpool.Pool
}

func NewCrmRepository(db *pgxpool.Pool) CrmRepository {
	return &crmRepository{db: db}
}

func (r *crmRepository) CreateAccount(ctx context.Context, account *models.CrmAccount) (*models.CrmAccount, error) {
	contactID, err := strconv.ParseInt(strings.TrimSpace(account.IDContact), 10, 32)
	if err != nil || contactID <= 0 {
		return nil, fmt.Errorf("invalid idContact")
	}
	name := strings.TrimSpace(account.Name)
	if name == "" {
		var contactName string
		err := r.db.QueryRow(ctx, `
			SELECT COALESCE(NULLIF(TRIM(name || ' ' || COALESCE(last_name,'')), ''), email)
			FROM contacts WHERE id=$1 AND company_id=$2 AND deleted=false
		`, contactID, defaultCompanyID).Scan(&contactName)
		if err != nil {
			return nil, fmt.Errorf("contact not found")
		}
		name = contactName
	}
	var id int32
	err = r.db.QueryRow(ctx, `
		INSERT INTO crm_accounts (contact_id, name) VALUES ($1,$2) RETURNING id
	`, contactID, name).Scan(&id)
	if err != nil {
		return nil, fmt.Errorf("create crm account: %w", err)
	}
	return &models.CrmAccount{ID: id, IDContact: account.IDContact, Name: name}, nil
}

func (r *crmRepository) GlobalReport(ctx context.Context) (map[string]any, error) {
	var accounts, tasks, activities, opportunities int
	_ = r.db.QueryRow(ctx, `SELECT COUNT(*) FROM crm_accounts WHERE company_id=$1 AND deleted=false`, defaultCompanyID).Scan(&accounts)
	_ = r.db.QueryRow(ctx, `SELECT COUNT(*) FROM crm_tasks WHERE company_id=$1 AND deleted=false`, defaultCompanyID).Scan(&tasks)
	_ = r.db.QueryRow(ctx, `SELECT COUNT(*) FROM crm_activities WHERE company_id=$1 AND deleted=false`, defaultCompanyID).Scan(&activities)
	_ = r.db.QueryRow(ctx, `
		SELECT COUNT(*) FROM crm_tasks WHERE company_id=$1 AND deleted=false AND type='TASK_CRM_OPPORTUNITY'
	`, defaultCompanyID).Scan(&opportunities)
	return map[string]any{
		"totalAccounts":     accounts,
		"totalTasks":        tasks,
		"totalActivities":   activities,
		"openOpportunities": opportunities,
	}, nil
}

func (r *crmRepository) AccountReport(ctx context.Context, idCrmAccount int32) (map[string]any, error) {
	row := r.db.QueryRow(ctx, `
		SELECT a.id, a.name, c.id, COALESCE(c.name,''), COALESCE(c.email,'')
		FROM crm_accounts a
		JOIN contacts c ON c.id = a.contact_id
		WHERE a.id=$1 AND a.company_id=$2 AND a.deleted=false
	`, idCrmAccount, defaultCompanyID)
	var accountID, contactID int32
	var accountName, contactName, email string
	if err := row.Scan(&accountID, &accountName, &contactID, &contactName, &email); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	var taskCount, activityCount int
	_ = r.db.QueryRow(ctx, `SELECT COUNT(*) FROM crm_tasks WHERE id_root=$1 AND deleted=false`, idCrmAccount).Scan(&taskCount)
	_ = r.db.QueryRow(ctx, `SELECT COUNT(*) FROM crm_activities WHERE id_root=$1 AND deleted=false`, idCrmAccount).Scan(&activityCount)
	return map[string]any{
		"idCrmAccount":    accountID,
		"name":            accountName,
		"idContact":       contactID,
		"contactName":     contactName,
		"email":           email,
		"tasksCount":      taskCount,
		"activitiesCount": activityCount,
	}, nil
}

func (r *crmRepository) ListTasks(ctx context.Context, q models.CrmTaskListQuery) ([]models.CrmTask, error) {
	sql := crmTaskSelect + ` WHERE t.company_id = $1 AND t.deleted = false`
	args := []any{defaultCompanyID}
	n := 2
	if q.IDCrmAccount != nil {
		sql += fmt.Sprintf(` AND t.id_root = $%d`, n)
		args = append(args, *q.IDCrmAccount)
		n++
	}
	if s := strings.TrimSpace(q.Types); s != "" {
		types := strings.Split(s, ",")
		sql += fmt.Sprintf(` AND t.type = ANY($%d)`, n)
		args = append(args, types)
		n++
	}
	if s := strings.TrimSpace(q.Status); s != "" {
		sql += fmt.Sprintf(` AND t.status = $%d`, n)
		args = append(args, s)
		n++
	}
	sql += fmt.Sprintf(` ORDER BY t.id DESC LIMIT $%d OFFSET $%d`, n, n+1)
	args = append(args, limitOrDefault(q.Limit), q.Offset)

	rows, err := r.db.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]models.CrmTask, 0)
	for rows.Next() {
		task, err := scanCrmTask(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *task)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if q.IncludeActivities {
		for i := range out {
			acts, err := r.ListActivities(ctx, out[i].ID, models.CrmActivityListQuery{ListQuery: models.ListQuery{Limit: 50}})
			if err != nil {
				return nil, err
			}
			out[i].Attendees = acts
		}
	}
	return out, nil
}

func (r *crmRepository) GetTask(ctx context.Context, id int32) (*models.CrmTask, error) {
	row := r.db.QueryRow(ctx, crmTaskSelect+` WHERE t.id=$1 AND t.company_id=$2 AND t.deleted=false`, id, defaultCompanyID)
	task, err := scanCrmTask(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return task, nil
}

func (r *crmRepository) CreateTask(ctx context.Context, idParent int32, task *models.CrmTask) (*models.CrmTask, error) {
	root, err := r.resolveRoot(ctx, idParent)
	if err != nil {
		return nil, err
	}
	var id int32
	err = r.db.QueryRow(ctx, `
		INSERT INTO crm_tasks (
			id_root, id_parent, type, name, description, place, status, start_date, end_date,
			id_doc_type, id_doc, doc_description, priority, privacy, id_user_assignee, user_assignee_description,
			id_contact_representative, contact_representative_desc, quantity, amount, probability, forecast_revenue
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21,$22)
		RETURNING id
	`, root, idParent, defaultStr(task.Type, "TASK_CRM"), emptyAsNull(task.Name), emptyAsNull(task.Description),
		emptyAsNull(task.Place), defaultStr(task.Status, "OPEN"), millisToTimePtr(task.StartDate), millisToTimePtr(task.EndDate),
		task.IDDocType, emptyAsNull(task.IDDoc), emptyAsNull(task.DocDescription), task.Priority, emptyAsNull(task.Privacy),
		task.IDUserAssignee, emptyAsNull(task.UserAssigneeDescription), task.IDContactRepresentative,
		emptyAsNull(task.ContactRepresentativeDescription), task.Quantity, task.Amount, task.Probability, task.ForecastRevenue).Scan(&id)
	if err != nil {
		return nil, fmt.Errorf("create crm task: %w", err)
	}
	return r.GetTask(ctx, id)
}

func (r *crmRepository) UpdateTask(ctx context.Context, idParent, id int32, task *models.CrmTask) (*models.CrmTask, error) {
	tag, err := r.db.Exec(ctx, `
		UPDATE crm_tasks SET
			type=$4, name=$5, description=$6, place=$7, status=$8, start_date=$9, end_date=$10,
			id_doc_type=$11, id_doc=$12, doc_description=$13, priority=$14, privacy=$15,
			id_user_assignee=$16, user_assignee_description=$17, id_contact_representative=$18,
			contact_representative_desc=$19, quantity=$20, amount=$21, probability=$22, forecast_revenue=$23, updated_at=now()
		WHERE id=$1 AND id_parent=$2 AND company_id=$3 AND deleted=false
	`, id, idParent, defaultCompanyID, defaultStr(task.Type, "TASK_CRM"), emptyAsNull(task.Name), emptyAsNull(task.Description),
		emptyAsNull(task.Place), defaultStr(task.Status, "OPEN"), millisToTimePtr(task.StartDate), millisToTimePtr(task.EndDate),
		task.IDDocType, emptyAsNull(task.IDDoc), emptyAsNull(task.DocDescription), task.Priority, emptyAsNull(task.Privacy),
		task.IDUserAssignee, emptyAsNull(task.UserAssigneeDescription), task.IDContactRepresentative,
		emptyAsNull(task.ContactRepresentativeDescription), task.Quantity, task.Amount, task.Probability, task.ForecastRevenue)
	if err != nil {
		return nil, err
	}
	if tag.RowsAffected() == 0 {
		return nil, fmt.Errorf("crm task not found")
	}
	return r.GetTask(ctx, id)
}

func (r *crmRepository) PatchTask(ctx context.Context, idParent, id int32, fields map[string]any) (*models.CrmTask, error) {
	setSQL, args, err := buildPatchSet(crmTaskPatchColumns, fields, 4)
	if err != nil {
		return nil, err
	}
	if setSQL == "" {
		return r.GetTask(ctx, id)
	}
	args = append([]any{id, idParent, defaultCompanyID}, args...)
	sql := fmt.Sprintf(`UPDATE crm_tasks SET %s, updated_at=now() WHERE id=$1 AND id_parent=$2 AND company_id=$3 AND deleted=false`, setSQL)
	tag, err := r.db.Exec(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	if tag.RowsAffected() == 0 {
		return nil, fmt.Errorf("crm task not found")
	}
	return r.GetTask(ctx, id)
}

func (r *crmRepository) DeleteTask(ctx context.Context, idParent, id int32) error {
	tag, err := r.db.Exec(ctx, `
		UPDATE crm_tasks SET deleted=true, updated_at=now()
		WHERE id=$1 AND id_parent=$2 AND company_id=$3 AND deleted=false
	`, id, idParent, defaultCompanyID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("crm task not found")
	}
	return nil
}

func (r *crmRepository) ListActivities(ctx context.Context, idParent int32, q models.CrmActivityListQuery) ([]models.CrmActivity, error) {
	sql := crmActivitySelect + ` WHERE a.id_parent = $1 AND a.company_id = $2 AND a.deleted = false`
	args := []any{idParent, defaultCompanyID}
	n := 3
	if s := strings.TrimSpace(q.Type); s != "" {
		sql += fmt.Sprintf(` AND a.type = $%d`, n)
		args = append(args, s)
		n++
	}
	if s := strings.TrimSpace(q.IDDoc); s != "" {
		sql += fmt.Sprintf(` AND a.id_doc = $%d`, n)
		args = append(args, s)
		n++
	}
	if q.IDDocType != nil {
		sql += fmt.Sprintf(` AND a.id_document_type = $%d`, n)
		args = append(args, *q.IDDocType)
		n++
	}
	sql += fmt.Sprintf(` ORDER BY a.id DESC LIMIT $%d OFFSET $%d`, n, n+1)
	args = append(args, limitOrDefault(q.Limit), q.Offset)

	rows, err := r.db.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]models.CrmActivity, 0)
	for rows.Next() {
		act, err := scanCrmActivity(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *act)
	}
	return out, rows.Err()
}

func (r *crmRepository) GetActivity(ctx context.Context, idParent, id int32) (*models.CrmActivity, error) {
	row := r.db.QueryRow(ctx, crmActivitySelect+` WHERE a.id=$1 AND a.id_parent=$2 AND a.company_id=$3 AND a.deleted=false`,
		id, idParent, defaultCompanyID)
	act, err := scanCrmActivity(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return act, nil
}

func (r *crmRepository) CreateActivity(ctx context.Context, idParent int32, act *models.CrmActivity) (*models.CrmActivity, error) {
	root, err := r.resolveRoot(ctx, idParent)
	if err != nil {
		return nil, err
	}
	var id int32
	err = r.db.QueryRow(ctx, `
		INSERT INTO crm_activities (
			id_root, id_parent, id_user, id_attachment, attachment_file_name, name, description, type,
			id_document_type, id_document_type_ext, id_doc, doc_description
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12) RETURNING id
	`, root, idParent, act.IDUser, emptyAsNull(act.IDAttachment), emptyAsNull(act.AttachmentFileName),
		emptyAsNull(act.Name), emptyAsNull(act.Description), emptyAsNull(act.Type),
		act.IDDocumentType, emptyAsNull(act.IDDocumentTypeExt), emptyAsNull(act.IDDoc), emptyAsNull(act.DocDescription)).Scan(&id)
	if err != nil {
		return nil, fmt.Errorf("create crm activity: %w", err)
	}
	return r.GetActivity(ctx, idParent, id)
}

func (r *crmRepository) UpdateActivity(ctx context.Context, idParent, id int32, act *models.CrmActivity) (*models.CrmActivity, error) {
	tag, err := r.db.Exec(ctx, `
		UPDATE crm_activities SET
			id_user=$4, id_attachment=$5, attachment_file_name=$6, name=$7, description=$8, type=$9,
			id_document_type=$10, id_document_type_ext=$11, id_doc=$12, doc_description=$13, updated_at=now()
		WHERE id=$1 AND id_parent=$2 AND company_id=$3 AND deleted=false
	`, id, idParent, defaultCompanyID, act.IDUser, emptyAsNull(act.IDAttachment), emptyAsNull(act.AttachmentFileName),
		emptyAsNull(act.Name), emptyAsNull(act.Description), emptyAsNull(act.Type),
		act.IDDocumentType, emptyAsNull(act.IDDocumentTypeExt), emptyAsNull(act.IDDoc), emptyAsNull(act.DocDescription))
	if err != nil {
		return nil, err
	}
	if tag.RowsAffected() == 0 {
		return nil, fmt.Errorf("crm activity not found")
	}
	return r.GetActivity(ctx, idParent, id)
}

func (r *crmRepository) PatchActivity(ctx context.Context, idParent, id int32, fields map[string]any) (*models.CrmActivity, error) {
	setSQL, args, err := buildPatchSet(crmActivityPatchColumns, fields, 4)
	if err != nil {
		return nil, err
	}
	if setSQL == "" {
		return r.GetActivity(ctx, idParent, id)
	}
	args = append([]any{id, idParent, defaultCompanyID}, args...)
	sql := fmt.Sprintf(`UPDATE crm_activities SET %s, updated_at=now() WHERE id=$1 AND id_parent=$2 AND company_id=$3 AND deleted=false`, setSQL)
	tag, err := r.db.Exec(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	if tag.RowsAffected() == 0 {
		return nil, fmt.Errorf("crm activity not found")
	}
	return r.GetActivity(ctx, idParent, id)
}

func (r *crmRepository) DeleteActivity(ctx context.Context, idParent, id int32) error {
	tag, err := r.db.Exec(ctx, `
		UPDATE crm_activities SET deleted=true, updated_at=now()
		WHERE id=$1 AND id_parent=$2 AND company_id=$3 AND deleted=false
	`, id, idParent, defaultCompanyID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("crm activity not found")
	}
	return nil
}

func (r *crmRepository) SaveAttachment(ctx context.Context, idParent int32, filename, contentType string, data []byte) (*models.CrmAttachmentResult, error) {
	var id int32
	if err := r.db.QueryRow(ctx, `
		INSERT INTO crm_attachments (id_parent, filename, content_type, data) VALUES ($1,$2,$3,$4) RETURNING id
	`, idParent, filename, contentType, data).Scan(&id); err != nil {
		return nil, err
	}
	return &models.CrmAttachmentResult{
		IDAttachment:       strconv.Itoa(int(id)),
		AttachmentFileName: filename,
	}, nil
}

func (r *crmRepository) resolveRoot(ctx context.Context, idParent int32) (int32, error) {
	var accountID int32
	err := r.db.QueryRow(ctx, `SELECT id FROM crm_accounts WHERE id=$1 AND company_id=$2 AND deleted=false`, idParent, defaultCompanyID).Scan(&accountID)
	if err == nil {
		return accountID, nil
	}
	var root int32
	err = r.db.QueryRow(ctx, `SELECT id_root FROM crm_tasks WHERE id=$1 AND company_id=$2 AND deleted=false`, idParent, defaultCompanyID).Scan(&root)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, fmt.Errorf("parent not found")
		}
		return 0, err
	}
	return root, nil
}

const crmTaskSelect = `
	SELECT t.id, t.id_root, t.id_parent, COALESCE(t.type,''), COALESCE(t.name,''), COALESCE(t.description,''),
		COALESCE(t.place,''), COALESCE(t.status,''), t.start_date, t.end_date, t.create_date,
		t.id_doc_type, COALESCE(t.id_doc,''), COALESCE(t.doc_description,''), COALESCE(t.priority,0), COALESCE(t.privacy,''),
		t.id_user_assignee, COALESCE(t.user_assignee_description,''), t.id_contact_representative,
		COALESCE(t.contact_representative_desc,''), t.quantity, t.amount, t.probability, t.forecast_revenue
	FROM crm_tasks t`

const crmActivitySelect = `
	SELECT a.id, a.id_parent, a.id_root, a.id_user, COALESCE(a.id_attachment,''), COALESCE(a.attachment_file_name,''),
		COALESCE(a.name,''), COALESCE(a.description,''), COALESCE(a.type,''),
		a.id_document_type, COALESCE(a.id_document_type_ext,''), COALESCE(a.id_doc,''), COALESCE(a.doc_description,'')
	FROM crm_activities a`

var crmTaskPatchColumns = map[string]string{
	"type": "type", "name": "name", "description": "description", "place": "place", "status": "status",
	"idDoc": "id_doc", "docDescription": "doc_description", "priority": "priority", "privacy": "privacy",
	"amount": "amount", "probability": "probability", "forecastRevenue": "forecast_revenue",
}

var crmActivityPatchColumns = map[string]string{
	"name": "name", "description": "description", "type": "type",
	"idDoc": "id_doc", "docDescription": "doc_description",
}

func scanCrmTask(s interface{ Scan(...any) error }) (*models.CrmTask, error) {
	var t models.CrmTask
	var start, end, created *time.Time
	if err := s.Scan(&t.ID, &t.IDRoot, &t.IDParent, &t.Type, &t.Name, &t.Description, &t.Place, &t.Status,
		&start, &end, &created, &t.IDDocType, &t.IDDoc, &t.DocDescription, &t.Priority, &t.Privacy,
		&t.IDUserAssignee, &t.UserAssigneeDescription, &t.IDContactRepresentative, &t.ContactRepresentativeDescription,
		&t.Quantity, &t.Amount, &t.Probability, &t.ForecastRevenue); err != nil {
		return nil, err
	}
	if start != nil {
		t.StartDate = models.MillisPtr(*start)
	}
	if end != nil {
		t.EndDate = models.MillisPtr(*end)
	}
	if created != nil {
		t.CreateDate = models.MillisPtr(*created)
	}
	return &t, nil
}

func scanCrmActivity(s interface{ Scan(...any) error }) (*models.CrmActivity, error) {
	var a models.CrmActivity
	if err := s.Scan(&a.ID, &a.IDParent, &a.IDRoot, &a.IDUser, &a.IDAttachment, &a.AttachmentFileName,
		&a.Name, &a.Description, &a.Type, &a.IDDocumentType, &a.IDDocumentTypeExt, &a.IDDoc, &a.DocDescription); err != nil {
		return nil, err
	}
	return &a, nil
}
