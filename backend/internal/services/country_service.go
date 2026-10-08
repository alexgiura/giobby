package services

import (
	"context"
	"fmt"
	"strings"

	"dnsc_microservice/internal/models"
	"dnsc_microservice/internal/repository"
)

// CountryService handles country reference data.
type CountryService interface {
	ListCountries(ctx context.Context, q models.CountryListQuery) ([]models.Country, error)
	GetCountry(ctx context.Context, id string) (*models.Country, error)
	CreateCountry(ctx context.Context, country models.Country) (*models.Country, error)
	UpdateCountry(ctx context.Context, id string, country models.Country) (*models.Country, error)
	DeleteCountry(ctx context.Context, id string) error
	RecoverCountry(ctx context.Context, id string) (*models.Country, error)
}

type countryService struct {
	repo repository.CountryRepository
}

func NewCountryService(repo repository.CountryRepository) CountryService {
	return &countryService{repo: repo}
}

func (s *countryService) ListCountries(ctx context.Context, q models.CountryListQuery) ([]models.Country, error) {
	return s.repo.ListCountries(ctx, q)
}

func (s *countryService) GetCountry(ctx context.Context, id string) (*models.Country, error) {
	id = normalizeCountryID(id)
	if id == "" {
		return nil, fmt.Errorf("country id is required")
	}
	return s.repo.GetCountry(ctx, id)
}

func (s *countryService) CreateCountry(ctx context.Context, country models.Country) (*models.Country, error) {
	country.ID = normalizeCountryID(country.ID)
	country.Description = strings.TrimSpace(country.Description)
	if country.ID == "" {
		return nil, fmt.Errorf("country id is required")
	}
	if country.Description == "" {
		return nil, fmt.Errorf("country description is required")
	}
	return s.repo.CreateCountry(ctx, &country)
}

func (s *countryService) UpdateCountry(ctx context.Context, id string, country models.Country) (*models.Country, error) {
	id = normalizeCountryID(id)
	if id == "" {
		return nil, fmt.Errorf("country id is required")
	}
	if bodyID := normalizeCountryID(country.ID); bodyID != "" && bodyID != id {
		return nil, fmt.Errorf("country id in body does not match path")
	}
	country.Description = strings.TrimSpace(country.Description)
	if country.Description == "" {
		return nil, fmt.Errorf("country description is required")
	}
	return s.repo.UpdateCountry(ctx, id, &country)
}

func (s *countryService) DeleteCountry(ctx context.Context, id string) error {
	id = normalizeCountryID(id)
	if id == "" {
		return fmt.Errorf("country id is required")
	}
	return s.repo.DeleteCountry(ctx, id)
}

func (s *countryService) RecoverCountry(ctx context.Context, id string) (*models.Country, error) {
	id = normalizeCountryID(id)
	if id == "" {
		return nil, fmt.Errorf("country id is required")
	}
	return s.repo.RecoverCountry(ctx, id)
}

func normalizeCountryID(id string) string {
	return strings.ToUpper(strings.TrimSpace(id))
}
