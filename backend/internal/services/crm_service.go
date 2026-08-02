package services

import (
	"context"
	"fmt"
	"strings"

	"dnsc_microservice/internal/models"
	"dnsc_microservice/internal/repository"
)

// CrmService handles CRM accounts, tasks and activities.
type CrmService interface {
	CreateAccount(ctx context.Context, account models.CrmAccount) (*models.CrmAccount, error)
	GlobalReport(ctx context.Context) (map[string]any, error)
	AccountReport(ctx context.Context, idCrmAccount int32) (map[string]any, error)
	ListTasks(ctx context.Context, q models.CrmTaskListQuery) ([]models.CrmTask, error)
	GetTask(ctx context.Context, id int32) (*models.CrmTask, error)
	CreateTask(ctx context.Context, idParent int32, task models.CrmTask) (*models.CrmTask, error)
	UpdateTask(ctx context.Context, idParent, id int32, task models.CrmTask) (*models.CrmTask, error)
	PatchTask(ctx context.Context, idParent, id int32, fields map[string]any) (*models.CrmTask, error)
	DeleteTask(ctx context.Context, idParent, id int32) error
	ListActivities(ctx context.Context, idParent int32, q models.CrmActivityListQuery) ([]models.CrmActivity, error)
	GetActivity(ctx context.Context, idParent, id int32) (*models.CrmActivity, error)
	CreateActivity(ctx context.Context, idParent int32, act models.CrmActivity) (*models.CrmActivity, error)
	UpdateActivity(ctx context.Context, idParent, id int32, act models.CrmActivity) (*models.CrmActivity, error)
	PatchActivity(ctx context.Context, idParent, id int32, fields map[string]any) (*models.CrmActivity, error)
	DeleteActivity(ctx context.Context, idParent, id int32) error
	SaveAttachment(ctx context.Context, idParent int32, filename, contentType string, data []byte) (*models.CrmAttachmentResult, error)
}

type crmService struct {
	repo repository.CrmRepository
}

func NewCrmService(repo repository.CrmRepository) CrmService {
	return &crmService{repo: repo}
}

func (s *crmService) CreateAccount(ctx context.Context, account models.CrmAccount) (*models.CrmAccount, error) {
	if strings.TrimSpace(account.IDContact) == "" {
		return nil, fmt.Errorf("idContact is required")
	}
	return s.repo.CreateAccount(ctx, &account)
}

func (s *crmService) GlobalReport(ctx context.Context) (map[string]any, error) {
	return s.repo.GlobalReport(ctx)
}

func (s *crmService) AccountReport(ctx context.Context, idCrmAccount int32) (map[string]any, error) {
	if idCrmAccount <= 0 {
		return nil, fmt.Errorf("invalid crm account id")
	}
	return s.repo.AccountReport(ctx, idCrmAccount)
}

func (s *crmService) ListTasks(ctx context.Context, q models.CrmTaskListQuery) ([]models.CrmTask, error) {
	return s.repo.ListTasks(ctx, q)
}

func (s *crmService) GetTask(ctx context.Context, id int32) (*models.CrmTask, error) {
	if id <= 0 {
		return nil, fmt.Errorf("invalid task id")
	}
	return s.repo.GetTask(ctx, id)
}

func (s *crmService) CreateTask(ctx context.Context, idParent int32, task models.CrmTask) (*models.CrmTask, error) {
	if idParent <= 0 {
		return nil, fmt.Errorf("invalid parent id")
	}
	return s.repo.CreateTask(ctx, idParent, &task)
}

func (s *crmService) UpdateTask(ctx context.Context, idParent, id int32, task models.CrmTask) (*models.CrmTask, error) {
	if idParent <= 0 || id <= 0 {
		return nil, fmt.Errorf("invalid ids")
	}
	return s.repo.UpdateTask(ctx, idParent, id, &task)
}

func (s *crmService) PatchTask(ctx context.Context, idParent, id int32, fields map[string]any) (*models.CrmTask, error) {
	if idParent <= 0 || id <= 0 {
		return nil, fmt.Errorf("invalid ids")
	}
	return s.repo.PatchTask(ctx, idParent, id, fields)
}

func (s *crmService) DeleteTask(ctx context.Context, idParent, id int32) error {
	if idParent <= 0 || id <= 0 {
		return fmt.Errorf("invalid ids")
	}
	return s.repo.DeleteTask(ctx, idParent, id)
}

func (s *crmService) ListActivities(ctx context.Context, idParent int32, q models.CrmActivityListQuery) ([]models.CrmActivity, error) {
	if idParent <= 0 {
		return nil, fmt.Errorf("invalid parent id")
	}
	return s.repo.ListActivities(ctx, idParent, q)
}

func (s *crmService) GetActivity(ctx context.Context, idParent, id int32) (*models.CrmActivity, error) {
	if idParent <= 0 || id <= 0 {
		return nil, fmt.Errorf("invalid ids")
	}
	return s.repo.GetActivity(ctx, idParent, id)
}

func (s *crmService) CreateActivity(ctx context.Context, idParent int32, act models.CrmActivity) (*models.CrmActivity, error) {
	if idParent <= 0 {
		return nil, fmt.Errorf("invalid parent id")
	}
	return s.repo.CreateActivity(ctx, idParent, &act)
}

func (s *crmService) UpdateActivity(ctx context.Context, idParent, id int32, act models.CrmActivity) (*models.CrmActivity, error) {
	if idParent <= 0 || id <= 0 {
		return nil, fmt.Errorf("invalid ids")
	}
	return s.repo.UpdateActivity(ctx, idParent, id, &act)
}

func (s *crmService) PatchActivity(ctx context.Context, idParent, id int32, fields map[string]any) (*models.CrmActivity, error) {
	if idParent <= 0 || id <= 0 {
		return nil, fmt.Errorf("invalid ids")
	}
	return s.repo.PatchActivity(ctx, idParent, id, fields)
}

func (s *crmService) DeleteActivity(ctx context.Context, idParent, id int32) error {
	if idParent <= 0 || id <= 0 {
		return fmt.Errorf("invalid ids")
	}
	return s.repo.DeleteActivity(ctx, idParent, id)
}

func (s *crmService) SaveAttachment(ctx context.Context, idParent int32, filename, contentType string, data []byte) (*models.CrmAttachmentResult, error) {
	if idParent <= 0 {
		return nil, fmt.Errorf("invalid parent id")
	}
	if len(data) == 0 {
		return nil, fmt.Errorf("attachment data is required")
	}
	return s.repo.SaveAttachment(ctx, idParent, filename, contentType, data)
}
