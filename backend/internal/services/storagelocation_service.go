package services

import (
	"context"
	"fmt"
	"strings"

	"dnsc_microservice/internal/models"
	"dnsc_microservice/internal/repository"
)

// StorageLocationService handles storage locations.
type StorageLocationService interface {
	ListStorageLocations(ctx context.Context, q models.StorageLocationListQuery) ([]models.StorageLocation, error)
	GetStorageLocation(ctx context.Context, idLocation string) (*models.StorageLocation, error)
	CreateStorageLocation(ctx context.Context, loc models.StorageLocation) (*models.StorageLocation, error)
	UpdateStorageLocation(ctx context.Context, idLocation string, loc models.StorageLocation) (*models.StorageLocation, error)
	DeleteStorageLocation(ctx context.Context, idLocation string) error
}

type storageLocationService struct {
	repo repository.StorageLocationRepository
}

func NewStorageLocationService(repo repository.StorageLocationRepository) StorageLocationService {
	return &storageLocationService{repo: repo}
}

func (s *storageLocationService) ListStorageLocations(ctx context.Context, q models.StorageLocationListQuery) ([]models.StorageLocation, error) {
	return s.repo.ListStorageLocations(ctx, q)
}

func (s *storageLocationService) GetStorageLocation(ctx context.Context, idLocation string) (*models.StorageLocation, error) {
	idLocation = strings.TrimSpace(idLocation)
	if idLocation == "" {
		return nil, fmt.Errorf("invalid location id")
	}
	return s.repo.GetStorageLocation(ctx, idLocation)
}

func (s *storageLocationService) CreateStorageLocation(ctx context.Context, loc models.StorageLocation) (*models.StorageLocation, error) {
	return s.repo.CreateStorageLocation(ctx, &loc)
}

func (s *storageLocationService) UpdateStorageLocation(ctx context.Context, idLocation string, loc models.StorageLocation) (*models.StorageLocation, error) {
	idLocation = strings.TrimSpace(idLocation)
	if idLocation == "" {
		return nil, fmt.Errorf("invalid location id")
	}
	return s.repo.UpdateStorageLocation(ctx, idLocation, &loc)
}

func (s *storageLocationService) DeleteStorageLocation(ctx context.Context, idLocation string) error {
	idLocation = strings.TrimSpace(idLocation)
	if idLocation == "" {
		return fmt.Errorf("invalid location id")
	}
	return s.repo.DeleteStorageLocation(ctx, idLocation)
}
