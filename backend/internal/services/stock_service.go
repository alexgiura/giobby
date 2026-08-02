package services

import (
	"context"
	"fmt"

	"dnsc_microservice/internal/models"
	"dnsc_microservice/internal/repository"
)

// StockService handles stock movements and availability.
type StockService interface {
	CreateStock(ctx context.Context, s models.Stock) (*models.Stock, error)
	UpdateStock(ctx context.Context, s models.Stock) (*models.Stock, error)
	ListAvailability(ctx context.Context, q models.StockAvailabilityQuery) ([]models.StockAvailability, error)
	ListAvailabilityReport(ctx context.Context, q models.StockAvailabilityReportQuery) ([]models.StockAvailability, error)
}

type stockService struct {
	repo repository.StockRepository
}

func NewStockService(repo repository.StockRepository) StockService {
	return &stockService{repo: repo}
}

func (s *stockService) CreateStock(ctx context.Context, st models.Stock) (*models.Stock, error) {
	return s.repo.CreateStock(ctx, &st)
}

func (s *stockService) UpdateStock(ctx context.Context, st models.Stock) (*models.Stock, error) {
	if st.ID <= 0 {
		return nil, fmt.Errorf("stock id is required")
	}
	return s.repo.UpdateStock(ctx, &st)
}

func (s *stockService) ListAvailability(ctx context.Context, q models.StockAvailabilityQuery) ([]models.StockAvailability, error) {
	return s.repo.ListAvailability(ctx, q)
}

func (s *stockService) ListAvailabilityReport(ctx context.Context, q models.StockAvailabilityReportQuery) ([]models.StockAvailability, error) {
	return s.repo.ListAvailabilityReport(ctx, q)
}
