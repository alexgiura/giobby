package services

import (
	"context"
	"fmt"

	"dnsc_microservice/internal/models"
	"dnsc_microservice/internal/repository"
)

// ProductGroupService handles product categories.
type ProductGroupService interface {
	ListProductGroups(ctx context.Context, q models.ProductGroupListQuery) ([]models.ProductGroup, error)
	GetProductGroup(ctx context.Context, id int32) (*models.ProductGroup, error)
	CreateProductGroup(ctx context.Context, g models.ProductGroup) (*models.ProductGroup, error)
	UpdateProductGroup(ctx context.Context, id int32, g models.ProductGroup) (*models.ProductGroup, error)
	DeleteProductGroup(ctx context.Context, id int32) error
}

type productGroupService struct {
	repo repository.ProductGroupRepository
}

func NewProductGroupService(repo repository.ProductGroupRepository) ProductGroupService {
	return &productGroupService{repo: repo}
}

func (s *productGroupService) ListProductGroups(ctx context.Context, q models.ProductGroupListQuery) ([]models.ProductGroup, error) {
	return s.repo.ListProductGroups(ctx, q)
}

func (s *productGroupService) GetProductGroup(ctx context.Context, id int32) (*models.ProductGroup, error) {
	if id <= 0 {
		return nil, fmt.Errorf("invalid product group id")
	}
	return s.repo.GetProductGroup(ctx, id)
}

func (s *productGroupService) CreateProductGroup(ctx context.Context, g models.ProductGroup) (*models.ProductGroup, error) {
	return s.repo.CreateProductGroup(ctx, &g)
}

func (s *productGroupService) UpdateProductGroup(ctx context.Context, id int32, g models.ProductGroup) (*models.ProductGroup, error) {
	if id <= 0 {
		return nil, fmt.Errorf("invalid product group id")
	}
	return s.repo.UpdateProductGroup(ctx, id, &g)
}

func (s *productGroupService) DeleteProductGroup(ctx context.Context, id int32) error {
	if id <= 0 {
		return fmt.Errorf("invalid product group id")
	}
	return s.repo.DeleteProductGroup(ctx, id)
}
