package services

import (
	"context"
	"fmt"

	"dnsc_microservice/internal/models"
	"dnsc_microservice/internal/repository"
)

// AttributeService handles custom attributes and product characteristics.
type AttributeService interface {
	ListAttributes(ctx context.Context, q models.AttributeListQuery) ([]models.Attribute, error)
	CreateAttributeLinks(ctx context.Context, payload models.ProductAttributesPayload) ([]models.Attribute, error)
	ListProductCharacteristics(ctx context.Context) ([]models.ProductAttribute, error)
	CreateProductCharacteristic(ctx context.Context, a models.ProductAttribute) (*models.ProductAttribute, error)
	UpdateProductCharacteristic(ctx context.Context, a models.ProductAttribute) (*models.ProductAttribute, error)
}

type attributeService struct {
	repo repository.AttributeRepository
}

func NewAttributeService(repo repository.AttributeRepository) AttributeService {
	return &attributeService{repo: repo}
}

func (s *attributeService) ListAttributes(ctx context.Context, q models.AttributeListQuery) ([]models.Attribute, error) {
	return s.repo.ListAttributes(ctx, q)
}

func (s *attributeService) CreateAttributeLinks(ctx context.Context, payload models.ProductAttributesPayload) ([]models.Attribute, error) {
	if len(payload.Attributes) == 0 {
		return nil, fmt.Errorf("attributes are required")
	}
	return s.repo.CreateAttributeLinks(ctx, payload)
}

func (s *attributeService) ListProductCharacteristics(ctx context.Context) ([]models.ProductAttribute, error) {
	return s.repo.ListProductCharacteristics(ctx)
}

func (s *attributeService) CreateProductCharacteristic(ctx context.Context, a models.ProductAttribute) (*models.ProductAttribute, error) {
	return s.repo.CreateProductCharacteristic(ctx, &a)
}

func (s *attributeService) UpdateProductCharacteristic(ctx context.Context, a models.ProductAttribute) (*models.ProductAttribute, error) {
	if a.IDAttribute <= 0 {
		return nil, fmt.Errorf("idAttribute is required")
	}
	return s.repo.UpdateProductCharacteristic(ctx, &a)
}
