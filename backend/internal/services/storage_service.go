package services

import (
	"context"
	"fmt"
	"strings"

	"dnsc_microservice/internal/models"
	"dnsc_microservice/internal/repository"
)

// StorageService handles warehouses.
type StorageService interface {
	ListStorages(ctx context.Context, q models.StorageListQuery) ([]models.Storage, error)
	GetStorage(ctx context.Context, id string) (*models.Storage, error)
	CreateStorage(ctx context.Context, s models.Storage) (*models.Storage, error)
	UpdateStorage(ctx context.Context, id string, s models.Storage) (*models.Storage, error)
	DeleteStorage(ctx context.Context, id string) error
}

type storageService struct {
	repo repository.StorageRepository
}

func NewStorageService(repo repository.StorageRepository) StorageService {
	return &storageService{repo: repo}
}

func (s *storageService) ListStorages(ctx context.Context, q models.StorageListQuery) ([]models.Storage, error) {
	return s.repo.ListStorages(ctx, q)
}

func (s *storageService) GetStorage(ctx context.Context, id string) (*models.Storage, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, fmt.Errorf("invalid storage id")
	}
	return s.repo.GetStorage(ctx, id)
}

func (s *storageService) CreateStorage(ctx context.Context, st models.Storage) (*models.Storage, error) {
	return s.repo.CreateStorage(ctx, &st)
}

func (s *storageService) UpdateStorage(ctx context.Context, id string, st models.Storage) (*models.Storage, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, fmt.Errorf("invalid storage id")
	}
	return s.repo.UpdateStorage(ctx, id, &st)
}

func (s *storageService) DeleteStorage(ctx context.Context, id string) error {
	id = strings.TrimSpace(id)
	if id == "" {
		return fmt.Errorf("invalid storage id")
	}
	return s.repo.DeleteStorage(ctx, id)
}
