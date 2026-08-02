package services

import (
	"context"
	"fmt"
	"strings"

	"dnsc_microservice/internal/models"
	"dnsc_microservice/internal/repository"
)

// ReferenceService handles reference/lookup data.
type ReferenceService interface {
	ListCountries(ctx context.Context, q models.CountryListQuery) ([]models.Country, error)
	GetCountry(ctx context.Context, id string) (*models.Country, error)

	ListCities(ctx context.Context, q models.CityListQuery) ([]models.City, error)
	CreateCity(ctx context.Context, city models.City) (*models.City, error)

	ListCurrencies(ctx context.Context, q models.ListQuery) ([]models.Currency, error)
	GetCurrency(ctx context.Context, code string) (*models.Currency, error)

	ListUms(ctx context.Context, q models.UmListQuery) ([]models.Um, error)
	GetUm(ctx context.Context, idUm string, showDeleted bool) (*models.Um, error)
	CreateUm(ctx context.Context, um models.Um) error
	UpdateUm(ctx context.Context, idUm string, um models.Um) error
	DeleteUm(ctx context.Context, idUm string) error

	ListOfficeTypes(ctx context.Context, q models.OfficeTypeListQuery) ([]models.OfficeType, error)
	ListContactRoles(ctx context.Context, q models.ContactRoleListQuery) ([]models.ContactRole, error)

	ListPaymentTerms(ctx context.Context, q models.PaymentTermListQuery) ([]models.PaymentTerm, error)
	GetPaymentTerm(ctx context.Context, id int32) (*models.PaymentTerm, error)
	CreatePaymentTerm(ctx context.Context, term models.PaymentTerm) (*models.PaymentTerm, error)
}

type referenceService struct {
	repo repository.ReferenceRepository
}

func NewReferenceService(repo repository.ReferenceRepository) ReferenceService {
	return &referenceService{repo: repo}
}

func (s *referenceService) ListCountries(ctx context.Context, q models.CountryListQuery) ([]models.Country, error) {
	return s.repo.ListCountries(ctx, q)
}

func (s *referenceService) GetCountry(ctx context.Context, id string) (*models.Country, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, fmt.Errorf("country id is required")
	}
	return s.repo.GetCountry(ctx, id)
}

func (s *referenceService) ListCities(ctx context.Context, q models.CityListQuery) ([]models.City, error) {
	return s.repo.ListCities(ctx, q)
}

func (s *referenceService) CreateCity(ctx context.Context, city models.City) (*models.City, error) {
	city.Name = strings.TrimSpace(city.Name)
	if city.Name == "" {
		return nil, fmt.Errorf("city name is required")
	}
	return s.repo.CreateCity(ctx, &city)
}

func (s *referenceService) ListCurrencies(ctx context.Context, q models.ListQuery) ([]models.Currency, error) {
	return s.repo.ListCurrencies(ctx, q)
}

func (s *referenceService) GetCurrency(ctx context.Context, code string) (*models.Currency, error) {
	code = strings.TrimSpace(code)
	if code == "" {
		return nil, fmt.Errorf("currency code is required")
	}
	return s.repo.GetCurrency(ctx, code)
}

func (s *referenceService) ListUms(ctx context.Context, q models.UmListQuery) ([]models.Um, error) {
	return s.repo.ListUms(ctx, q)
}

func (s *referenceService) GetUm(ctx context.Context, idUm string, showDeleted bool) (*models.Um, error) {
	idUm = strings.TrimSpace(idUm)
	if idUm == "" {
		return nil, fmt.Errorf("idUm is required")
	}
	return s.repo.GetUm(ctx, idUm, showDeleted)
}

func (s *referenceService) CreateUm(ctx context.Context, um models.Um) error {
	um.Um = strings.TrimSpace(um.Um)
	if um.Um == "" {
		return fmt.Errorf("um code is required")
	}
	return s.repo.CreateUm(ctx, &um)
}

func (s *referenceService) UpdateUm(ctx context.Context, idUm string, um models.Um) error {
	idUm = strings.TrimSpace(idUm)
	if idUm == "" {
		return fmt.Errorf("idUm is required")
	}
	return s.repo.UpdateUm(ctx, idUm, &um)
}

func (s *referenceService) DeleteUm(ctx context.Context, idUm string) error {
	idUm = strings.TrimSpace(idUm)
	if idUm == "" {
		return fmt.Errorf("idUm is required")
	}
	return s.repo.DeleteUm(ctx, idUm)
}

func (s *referenceService) ListOfficeTypes(ctx context.Context, q models.OfficeTypeListQuery) ([]models.OfficeType, error) {
	return s.repo.ListOfficeTypes(ctx, q)
}

func (s *referenceService) ListContactRoles(ctx context.Context, q models.ContactRoleListQuery) ([]models.ContactRole, error) {
	return s.repo.ListContactRoles(ctx, q)
}

func (s *referenceService) ListPaymentTerms(ctx context.Context, q models.PaymentTermListQuery) ([]models.PaymentTerm, error) {
	return s.repo.ListPaymentTerms(ctx, q)
}

func (s *referenceService) GetPaymentTerm(ctx context.Context, id int32) (*models.PaymentTerm, error) {
	if id <= 0 {
		return nil, fmt.Errorf("invalid payment term id")
	}
	return s.repo.GetPaymentTerm(ctx, id)
}

func (s *referenceService) CreatePaymentTerm(ctx context.Context, term models.PaymentTerm) (*models.PaymentTerm, error) {
	term.Description = strings.TrimSpace(term.Description)
	if term.Description == "" {
		return nil, fmt.Errorf("description is required")
	}
	return s.repo.CreatePaymentTerm(ctx, &term)
}
