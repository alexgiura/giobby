package services

import (
	"context"
	"fmt"

	"dnsc_microservice/internal/models"
	"dnsc_microservice/internal/repository"
)

// AccountingService handles manual account movements (Accounting swagger tag).
type AccountingService interface {
	ListAccountMovements(ctx context.Context, q models.AccountMovementListQuery) ([]models.AccountMovementRegistration, error)
	CreateAccountMovement(ctx context.Context, reg models.AccountMovementRegistration) (*models.AccountMovementRegistration, error)
	UpdateAccountMovement(ctx context.Context, reg models.AccountMovementRegistration) (*models.AccountMovementRegistration, error)
	DeleteAccountMovement(ctx context.Context, idDoc int32) error
}

type accountingService struct {
	repo repository.AccountingRepository
}

func NewAccountingService(repo repository.AccountingRepository) AccountingService {
	return &accountingService{repo: repo}
}

func (s *accountingService) ListAccountMovements(ctx context.Context, q models.AccountMovementListQuery) ([]models.AccountMovementRegistration, error) {
	return s.repo.ListAccountMovements(ctx, q)
}

func (s *accountingService) CreateAccountMovement(ctx context.Context, reg models.AccountMovementRegistration) (*models.AccountMovementRegistration, error) {
	if len(reg.Rows) == 0 {
		return nil, fmt.Errorf("at least one movement row is required")
	}
	return s.repo.CreateAccountMovement(ctx, &reg)
}

func (s *accountingService) UpdateAccountMovement(ctx context.Context, reg models.AccountMovementRegistration) (*models.AccountMovementRegistration, error) {
	if reg.IDDoc <= 0 {
		return nil, fmt.Errorf("idDoc is required")
	}
	if len(reg.Rows) == 0 {
		return nil, fmt.Errorf("at least one movement row is required")
	}
	return s.repo.UpdateAccountMovement(ctx, &reg)
}

func (s *accountingService) DeleteAccountMovement(ctx context.Context, idDoc int32) error {
	if idDoc <= 0 {
		return fmt.Errorf("invalid document id")
	}
	return s.repo.DeleteAccountMovement(ctx, idDoc)
}
