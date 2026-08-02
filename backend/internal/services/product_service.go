package services

import (
	"context"
	"fmt"
	"strings"

	"dnsc_microservice/internal/models"
	"dnsc_microservice/internal/repository"
)

// ProductService handles products and variants.
type ProductService interface {
	ListProducts(ctx context.Context, q models.ProductListQuery) ([]models.Product, error)
	GetProduct(ctx context.Context, id string) (*models.Product, error)
	CreateProduct(ctx context.Context, p models.Product) (*models.Product, error)
	UpdateProduct(ctx context.Context, p models.Product) (*models.Product, error)
	ListVariants(ctx context.Context, productID string) ([]models.AttributeCombination, error)
	UpdateVariants(ctx context.Context, payload models.ProductVariantsPayload) error
	LinkMedia(ctx context.Context, productID string, media models.RepomediaRef) error
}

type productService struct {
	repo repository.ProductRepository
}

func NewProductService(repo repository.ProductRepository) ProductService {
	return &productService{repo: repo}
}

func (s *productService) ListProducts(ctx context.Context, q models.ProductListQuery) ([]models.Product, error) {
	return s.repo.ListProducts(ctx, q)
}

func (s *productService) GetProduct(ctx context.Context, id string) (*models.Product, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, fmt.Errorf("invalid product id")
	}
	return s.repo.GetProduct(ctx, id)
}

func (s *productService) CreateProduct(ctx context.Context, p models.Product) (*models.Product, error) {
	p.Description = strings.TrimSpace(p.Description)
	if p.Description == "" {
		p.Description = strings.TrimSpace(p.DescriptionIT)
	}
	if p.Description == "" {
		return nil, fmt.Errorf("description is required")
	}
	return s.repo.CreateProduct(ctx, &p)
}

func (s *productService) UpdateProduct(ctx context.Context, p models.Product) (*models.Product, error) {
	if strings.TrimSpace(p.ID) == "" {
		return nil, fmt.Errorf("product id is required")
	}
	return s.repo.UpdateProduct(ctx, &p)
}

func (s *productService) ListVariants(ctx context.Context, productID string) ([]models.AttributeCombination, error) {
	productID = strings.TrimSpace(productID)
	if productID == "" {
		return nil, fmt.Errorf("invalid product id")
	}
	return s.repo.ListVariants(ctx, productID)
}

func (s *productService) UpdateVariants(ctx context.Context, payload models.ProductVariantsPayload) error {
	if len(payload.Variants) == 0 {
		return fmt.Errorf("variants are required")
	}
	return s.repo.UpdateVariants(ctx, payload.Variants)
}

func (s *productService) LinkMedia(ctx context.Context, productID string, media models.RepomediaRef) error {
	productID = strings.TrimSpace(productID)
	if productID == "" || strings.TrimSpace(media.ID) == "" {
		return fmt.Errorf("invalid media link")
	}
	return s.repo.LinkMedia(ctx, productID, media.ID)
}
