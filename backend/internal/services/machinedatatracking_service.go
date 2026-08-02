package services

import (
	"context"
	"fmt"

	"dnsc_microservice/internal/models"
	"dnsc_microservice/internal/repository"
)

// MachineDataTrackingService handles machine production tracking.
type MachineDataTrackingService interface {
	List(ctx context.Context, q models.MachineDataTrackingListQuery) ([]models.MachineDataTracking, error)
	Get(ctx context.Context, id int32) (*models.MachineDataTracking, error)
	Create(ctx context.Context, item models.MachineDataTracking) (*models.MachineDataTracking, error)
	CreateBatch(ctx context.Context, items []models.MachineDataTracking) ([]models.MachineDataTracking, error)
	Update(ctx context.Context, id int32, item models.MachineDataTracking) (*models.MachineDataTracking, error)
	Delete(ctx context.Context, id int32) error
}

type machineDataTrackingService struct {
	repo repository.MachineDataTrackingRepository
}

func NewMachineDataTrackingService(repo repository.MachineDataTrackingRepository) MachineDataTrackingService {
	return &machineDataTrackingService{repo: repo}
}

func (s *machineDataTrackingService) List(ctx context.Context, q models.MachineDataTrackingListQuery) ([]models.MachineDataTracking, error) {
	return s.repo.List(ctx, q)
}

func (s *machineDataTrackingService) Get(ctx context.Context, id int32) (*models.MachineDataTracking, error) {
	if id <= 0 {
		return nil, fmt.Errorf("invalid id")
	}
	return s.repo.Get(ctx, id)
}

func (s *machineDataTrackingService) Create(ctx context.Context, item models.MachineDataTracking) (*models.MachineDataTracking, error) {
	return s.repo.Create(ctx, &item)
}

func (s *machineDataTrackingService) CreateBatch(ctx context.Context, items []models.MachineDataTracking) ([]models.MachineDataTracking, error) {
	if len(items) == 0 {
		return nil, fmt.Errorf("list is required")
	}
	return s.repo.CreateBatch(ctx, items)
}

func (s *machineDataTrackingService) Update(ctx context.Context, id int32, item models.MachineDataTracking) (*models.MachineDataTracking, error) {
	if id <= 0 {
		return nil, fmt.Errorf("invalid id")
	}
	return s.repo.Update(ctx, id, &item)
}

func (s *machineDataTrackingService) Delete(ctx context.Context, id int32) error {
	if id <= 0 {
		return fmt.Errorf("invalid id")
	}
	return s.repo.Delete(ctx, id)
}
