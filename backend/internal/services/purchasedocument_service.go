package services

import (
	"context"
	"fmt"

	"dnsc_microservice/internal/models"
	"dnsc_microservice/internal/repository"
)

// PurchaseDocumentService handles purchase documents (PuchaseDocument swagger tag).
type PurchaseDocumentService interface {
	ListDocuments(ctx context.Context, kind string, q models.PurchaseDocumentListQuery) ([]models.Document, error)
	GetDocument(ctx context.Context, id int32, kind string, onlyHeader bool) (*models.Document, error)
	CreateDocument(ctx context.Context, kind string, doc models.Document) (*models.Document, error)
	UpdateDocument(ctx context.Context, id int32, kind string, doc models.Document) (*models.Document, error)
	DeleteDocument(ctx context.Context, id int32, kind string) error
	TransformList(ctx context.Context, targetKind string, payload models.DocumentListTransformation) ([]models.Document, error)
	GetPDF(ctx context.Context, id int32) ([]byte, error)
	GetAttachment(ctx context.Context, idDoc int32) (*models.PurchaseDocumentAttachment, error)
	SaveAttachment(ctx context.Context, idDoc int32, filename, contentType string, data []byte) error
	DeleteAttachment(ctx context.Context, idDoc int32) error
	ListVendorPayments(ctx context.Context, q models.ListQuery) ([]models.VendorPayment, error)
	GetVendorPayment(ctx context.Context, id int32) (*models.VendorPayment, error)
	CreateVendorPayment(ctx context.Context, p models.VendorPayment) (*models.VendorPayment, error)
	UpdateVendorPayment(ctx context.Context, id int32, p models.VendorPayment) (*models.VendorPayment, error)
	DeleteVendorPayment(ctx context.Context, id int32) error
	GetVendorPaymentByDoc(ctx context.Context, idDoc int32) (*models.VendorPayment, error)
	OpenBalance(ctx context.Context, idVendor string, limit, offset int) ([]models.VendorOpenBalanceRow, error)
}

type purchaseDocumentService struct {
	repo repository.PurchaseDocumentRepository
}

func NewPurchaseDocumentService(repo repository.PurchaseDocumentRepository) PurchaseDocumentService {
	return &purchaseDocumentService{repo: repo}
}

func (s *purchaseDocumentService) ListDocuments(ctx context.Context, kind string, q models.PurchaseDocumentListQuery) ([]models.Document, error) {
	return s.repo.ListDocuments(ctx, kind, q)
}

func (s *purchaseDocumentService) GetDocument(ctx context.Context, id int32, kind string, onlyHeader bool) (*models.Document, error) {
	if id <= 0 {
		return nil, fmt.Errorf("invalid document id")
	}
	return s.repo.GetDocument(ctx, id, kind, onlyHeader)
}

func (s *purchaseDocumentService) CreateDocument(ctx context.Context, kind string, doc models.Document) (*models.Document, error) {
	if doc.IDContact <= 0 && doc.IDVendor == "" {
		return nil, fmt.Errorf("idContact or idVendor is required")
	}
	return s.repo.CreateDocument(ctx, kind, &doc)
}

func (s *purchaseDocumentService) UpdateDocument(ctx context.Context, id int32, kind string, doc models.Document) (*models.Document, error) {
	if id <= 0 {
		return nil, fmt.Errorf("invalid document id")
	}
	return s.repo.UpdateDocument(ctx, id, kind, &doc)
}

func (s *purchaseDocumentService) DeleteDocument(ctx context.Context, id int32, kind string) error {
	if id <= 0 {
		return fmt.Errorf("invalid document id")
	}
	return s.repo.DeleteDocument(ctx, id, kind)
}

func (s *purchaseDocumentService) TransformList(ctx context.Context, targetKind string, payload models.DocumentListTransformation) ([]models.Document, error) {
	return s.repo.TransformDocumentList(ctx, targetKind, payload)
}

func (s *purchaseDocumentService) GetPDF(ctx context.Context, id int32) ([]byte, error) {
	if id <= 0 {
		return nil, fmt.Errorf("invalid document id")
	}
	return s.repo.GetPDF(ctx, id)
}

func (s *purchaseDocumentService) GetAttachment(ctx context.Context, idDoc int32) (*models.PurchaseDocumentAttachment, error) {
	return s.repo.GetAttachment(ctx, idDoc)
}

func (s *purchaseDocumentService) SaveAttachment(ctx context.Context, idDoc int32, filename, contentType string, data []byte) error {
	if len(data) == 0 {
		return fmt.Errorf("attachment data is required")
	}
	return s.repo.SaveAttachment(ctx, idDoc, filename, contentType, data)
}

func (s *purchaseDocumentService) DeleteAttachment(ctx context.Context, idDoc int32) error {
	return s.repo.DeleteAttachment(ctx, idDoc)
}

func (s *purchaseDocumentService) ListVendorPayments(ctx context.Context, q models.ListQuery) ([]models.VendorPayment, error) {
	return s.repo.ListVendorPayments(ctx, q)
}

func (s *purchaseDocumentService) GetVendorPayment(ctx context.Context, id int32) (*models.VendorPayment, error) {
	return s.repo.GetVendorPayment(ctx, id)
}

func (s *purchaseDocumentService) CreateVendorPayment(ctx context.Context, p models.VendorPayment) (*models.VendorPayment, error) {
	return s.repo.CreateVendorPayment(ctx, &p)
}

func (s *purchaseDocumentService) UpdateVendorPayment(ctx context.Context, id int32, p models.VendorPayment) (*models.VendorPayment, error) {
	return s.repo.UpdateVendorPayment(ctx, id, &p)
}

func (s *purchaseDocumentService) DeleteVendorPayment(ctx context.Context, id int32) error {
	return s.repo.DeleteVendorPayment(ctx, id)
}

func (s *purchaseDocumentService) GetVendorPaymentByDoc(ctx context.Context, idDoc int32) (*models.VendorPayment, error) {
	return s.repo.GetVendorPaymentByDoc(ctx, idDoc)
}

func (s *purchaseDocumentService) OpenBalance(ctx context.Context, idVendor string, limit, offset int) ([]models.VendorOpenBalanceRow, error) {
	return s.repo.ListOpenBalance(ctx, idVendor, limit, offset)
}
