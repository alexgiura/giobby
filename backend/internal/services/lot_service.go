package services

import (
	"context"
	"fmt"
	"strings"

	"dnsc_microservice/internal/models"
	"dnsc_microservice/internal/repository"
)

// LotService handles product lots.
type LotService interface {
	ListLots(ctx context.Context, q models.LotListQuery) ([]models.Lot, error)
	GetLot(ctx context.Context, idLot string) (*models.Lot, error)
	CreateLot(ctx context.Context, lot models.Lot) (*models.Lot, error)
	UpdateLot(ctx context.Context, idLot string, lot models.Lot) (*models.Lot, error)
	DeleteLot(ctx context.Context, idLot string) error
}

type lotService struct {
	repo repository.LotRepository
}

func NewLotService(repo repository.LotRepository) LotService {
	return &lotService{repo: repo}
}

func (s *lotService) ListLots(ctx context.Context, q models.LotListQuery) ([]models.Lot, error) {
	return s.repo.ListLots(ctx, q)
}

func (s *lotService) GetLot(ctx context.Context, idLot string) (*models.Lot, error) {
	idLot = strings.TrimSpace(idLot)
	if idLot == "" {
		return nil, fmt.Errorf("invalid lot id")
	}
	return s.repo.GetLot(ctx, idLot)
}

func (s *lotService) CreateLot(ctx context.Context, lot models.Lot) (*models.Lot, error) {
	return s.repo.CreateLot(ctx, &lot)
}

func (s *lotService) UpdateLot(ctx context.Context, idLot string, lot models.Lot) (*models.Lot, error) {
	idLot = strings.TrimSpace(idLot)
	if idLot == "" {
		return nil, fmt.Errorf("invalid lot id")
	}
	return s.repo.UpdateLot(ctx, idLot, &lot)
}

func (s *lotService) DeleteLot(ctx context.Context, idLot string) error {
	idLot = strings.TrimSpace(idLot)
	if idLot == "" {
		return fmt.Errorf("invalid lot id")
	}
	return s.repo.DeleteLot(ctx, idLot)
}
