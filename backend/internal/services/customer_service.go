package services

import (
	"context"
	"fmt"
	"strings"

	"dnsc_microservice/internal/models"
	"dnsc_microservice/internal/repository"
)

// CustomerService handles customers.
type CustomerService interface {
	ListCustomers(ctx context.Context, q models.CustomerListQuery) ([]models.Customer, error)
	GetCustomer(ctx context.Context, id string, onlyCustomerData bool) (*models.Customer, error)
	CreateCustomer(ctx context.Context, c models.Customer) (*models.Customer, error)
	UpdateCustomer(ctx context.Context, id string, c models.Customer) (*models.Customer, error)
	PatchCustomer(ctx context.Context, id string, fields map[string]any) (*models.Customer, error)
	DeleteCustomer(ctx context.Context, id string) error
	GetCustomerReport(ctx context.Context, id string) (*models.PartnerReport, error)
}

type customerService struct {
	repo repository.CustomerRepository
}

func NewCustomerService(repo repository.CustomerRepository) CustomerService {
	return &customerService{repo: repo}
}

func (s *customerService) ListCustomers(ctx context.Context, q models.CustomerListQuery) ([]models.Customer, error) {
	return s.repo.ListCustomers(ctx, q)
}

func (s *customerService) GetCustomer(ctx context.Context, id string, onlyCustomerData bool) (*models.Customer, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, fmt.Errorf("invalid customer id")
	}
	return s.repo.GetCustomer(ctx, id, onlyCustomerData)
}

func (s *customerService) CreateCustomer(ctx context.Context, c models.Customer) (*models.Customer, error) {
	return s.repo.CreateCustomer(ctx, &c)
}

func (s *customerService) UpdateCustomer(ctx context.Context, id string, c models.Customer) (*models.Customer, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, fmt.Errorf("invalid customer id")
	}
	return s.repo.UpdateCustomer(ctx, id, &c)
}

func (s *customerService) PatchCustomer(ctx context.Context, id string, fields map[string]any) (*models.Customer, error) {
	id = strings.TrimSpace(id)
	if id == "" || len(fields) == 0 {
		return nil, fmt.Errorf("invalid patch")
	}
	return s.repo.PatchCustomer(ctx, id, fields)
}

func (s *customerService) DeleteCustomer(ctx context.Context, id string) error {
	id = strings.TrimSpace(id)
	if id == "" {
		return fmt.Errorf("invalid customer id")
	}
	return s.repo.DeleteCustomer(ctx, id)
}

func (s *customerService) GetCustomerReport(ctx context.Context, id string) (*models.PartnerReport, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, fmt.Errorf("invalid customer id")
	}
	return s.repo.GetCustomerReport(ctx, id)
}
