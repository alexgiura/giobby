package services

import (
	"context"
	"fmt"
	"strings"

	"dnsc_microservice/internal/models"
	"dnsc_microservice/internal/repository"
)

// VendorService handles vendors.
type VendorService interface {
	ListVendors(ctx context.Context, q models.VendorListQuery) ([]models.Vendor, error)
	GetVendor(ctx context.Context, id string, onlyVendorData bool) (*models.Vendor, error)
	CreateVendor(ctx context.Context, v models.Vendor) (*models.Vendor, error)
	UpdateVendor(ctx context.Context, id string, v models.Vendor) (*models.Vendor, error)
	PatchVendor(ctx context.Context, id string, fields map[string]any) (*models.Vendor, error)
	DeleteVendor(ctx context.Context, id string) error
	GetVendorReport(ctx context.Context, id string) (*models.PartnerReport, error)
}

type vendorService struct {
	repo repository.VendorRepository
}

func NewVendorService(repo repository.VendorRepository) VendorService {
	return &vendorService{repo: repo}
}

func (s *vendorService) ListVendors(ctx context.Context, q models.VendorListQuery) ([]models.Vendor, error) {
	return s.repo.ListVendors(ctx, q)
}

func (s *vendorService) GetVendor(ctx context.Context, id string, onlyVendorData bool) (*models.Vendor, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, fmt.Errorf("invalid vendor id")
	}
	return s.repo.GetVendor(ctx, id, onlyVendorData)
}

func (s *vendorService) CreateVendor(ctx context.Context, v models.Vendor) (*models.Vendor, error) {
	return s.repo.CreateVendor(ctx, &v)
}

func (s *vendorService) UpdateVendor(ctx context.Context, id string, v models.Vendor) (*models.Vendor, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, fmt.Errorf("invalid vendor id")
	}
	return s.repo.UpdateVendor(ctx, id, &v)
}

func (s *vendorService) PatchVendor(ctx context.Context, id string, fields map[string]any) (*models.Vendor, error) {
	id = strings.TrimSpace(id)
	if id == "" || len(fields) == 0 {
		return nil, fmt.Errorf("invalid patch")
	}
	return s.repo.PatchVendor(ctx, id, fields)
}

func (s *vendorService) DeleteVendor(ctx context.Context, id string) error {
	id = strings.TrimSpace(id)
	if id == "" {
		return fmt.Errorf("invalid vendor id")
	}
	return s.repo.DeleteVendor(ctx, id)
}

func (s *vendorService) GetVendorReport(ctx context.Context, id string) (*models.PartnerReport, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, fmt.Errorf("invalid vendor id")
	}
	return s.repo.GetVendorReport(ctx, id)
}
