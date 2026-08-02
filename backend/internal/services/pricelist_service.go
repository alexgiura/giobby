package services

import (
	"context"
	"fmt"

	"dnsc_microservice/internal/models"
	"dnsc_microservice/internal/repository"
)

// PricelistService handles pricelists and schemes.
type PricelistService interface {
	ListPricelists(ctx context.Context, q models.PricelistListQuery) ([]models.Pricelist, error)
	CreatePricelist(ctx context.Context, p models.Pricelist) (*models.Pricelist, error)
	GetPricelistRows(ctx context.Context, id int32, q models.PricelistDetailQuery) ([]models.PricelistRow, error)
	UpdatePricelist(ctx context.Context, id int32, p models.Pricelist) (*models.Pricelist, error)
	DeletePricelist(ctx context.Context, id int32) error
	ListPricelistSchemes(ctx context.Context, q models.ListQuery) ([]models.PricelistScheme, error)
	GetPricelistScheme(ctx context.Context, id int32) (*models.PricelistScheme, error)
}

type pricelistService struct {
	repo repository.PricelistRepository
}

func NewPricelistService(repo repository.PricelistRepository) PricelistService {
	return &pricelistService{repo: repo}
}

func (s *pricelistService) ListPricelists(ctx context.Context, q models.PricelistListQuery) ([]models.Pricelist, error) {
	return s.repo.ListPricelists(ctx, q)
}

func (s *pricelistService) CreatePricelist(ctx context.Context, p models.Pricelist) (*models.Pricelist, error) {
	if p.Description == "" {
		return nil, fmt.Errorf("description is required")
	}
	return s.repo.CreatePricelist(ctx, &p)
}

func (s *pricelistService) GetPricelistRows(ctx context.Context, id int32, q models.PricelistDetailQuery) ([]models.PricelistRow, error) {
	if id <= 0 {
		return nil, fmt.Errorf("invalid pricelist id")
	}
	return s.repo.GetPricelistRows(ctx, id, q)
}

func (s *pricelistService) UpdatePricelist(ctx context.Context, id int32, p models.Pricelist) (*models.Pricelist, error) {
	if id <= 0 {
		return nil, fmt.Errorf("invalid pricelist id")
	}
	return s.repo.UpdatePricelist(ctx, id, &p)
}

func (s *pricelistService) DeletePricelist(ctx context.Context, id int32) error {
	if id <= 0 {
		return fmt.Errorf("invalid pricelist id")
	}
	return s.repo.DeletePricelist(ctx, id)
}

func (s *pricelistService) ListPricelistSchemes(ctx context.Context, q models.ListQuery) ([]models.PricelistScheme, error) {
	return s.repo.ListPricelistSchemes(ctx, q)
}

func (s *pricelistService) GetPricelistScheme(ctx context.Context, id int32) (*models.PricelistScheme, error) {
	if id <= 0 {
		return nil, fmt.Errorf("invalid scheme id")
	}
	return s.repo.GetPricelistScheme(ctx, id)
}
