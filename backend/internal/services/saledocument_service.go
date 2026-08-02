package services

import (
	"context"
	"fmt"

	"dnsc_microservice/internal/models"
	"dnsc_microservice/internal/repository"
)

// SaleDocumentService handles sales documents (SaleDocument swagger tag).
type SaleDocumentService interface {
	ListDocuments(ctx context.Context, kind string, q models.SaleDocumentListQuery) ([]models.Document, error)
	GetDocument(ctx context.Context, id int32, kind string, onlyHeader bool) (*models.Document, error)
	CreateDocument(ctx context.Context, kind string, doc models.Document) (*models.Document, error)
	UpdateDocument(ctx context.Context, id int32, kind string, doc models.Document) (*models.Document, error)
	DeleteDocument(ctx context.Context, id int32, kind string) error
	Transform(ctx context.Context, sourceID int32, targetKind string, docDate *int64, docNumber string) (*models.Document, error)
	TransformList(ctx context.Context, targetKind string, payload models.DocumentListTransformation) ([]models.Document, error)
	GetPDF(ctx context.Context, id int32) ([]byte, error)
	GetAttachment(ctx context.Context, idDoc int32) (*models.SaleDocumentAttachment, error)
	SaveAttachment(ctx context.Context, idDoc int32, filename, contentType string, data []byte) error
	DeleteAttachment(ctx context.Context, idDoc int32) error
	UpdateOrderPayments(ctx context.Context, id int32, rows []models.DocumentPaymentRowBase) error
	AddInvoicePayment(ctx context.Context, id int32, payload models.DocumentPaymentPayload) error
	ListCustomerPayments(ctx context.Context, q models.ListQuery) ([]models.CustomerPayment, error)
	GetCustomerPayment(ctx context.Context, id int32) (*models.CustomerPayment, error)
	CreateCustomerPayment(ctx context.Context, p models.CustomerPayment) (*models.CustomerPayment, error)
	UpdateCustomerPayment(ctx context.Context, id int32, p models.CustomerPayment) (*models.CustomerPayment, error)
	DeleteCustomerPayment(ctx context.Context, id int32) error
	GetCustomerPaymentByDoc(ctx context.Context, idDoc int32) (*models.CustomerPayment, error)
	OpenBalance(ctx context.Context, idCustomer string, limit, offset int) ([]models.OpenBalanceRow, error)
	Reports(ctx context.Context, reportType, startDate, endDate string) (map[string]any, error)
	EInvoice(ctx context.Context, idDoc int32) (map[string]any, error)
}

type saleDocumentService struct {
	repo repository.SaleDocumentRepository
}

func NewSaleDocumentService(repo repository.SaleDocumentRepository) SaleDocumentService {
	return &saleDocumentService{repo: repo}
}

func (s *saleDocumentService) ListDocuments(ctx context.Context, kind string, q models.SaleDocumentListQuery) ([]models.Document, error) {
	return s.repo.ListDocuments(ctx, kind, q)
}

func (s *saleDocumentService) GetDocument(ctx context.Context, id int32, kind string, onlyHeader bool) (*models.Document, error) {
	if id <= 0 {
		return nil, fmt.Errorf("invalid document id")
	}
	return s.repo.GetDocument(ctx, id, kind, onlyHeader)
}

func (s *saleDocumentService) CreateDocument(ctx context.Context, kind string, doc models.Document) (*models.Document, error) {
	if doc.IDContact <= 0 && doc.IDCustomer == "" {
		return nil, fmt.Errorf("idContact or idCustomer is required")
	}
	return s.repo.CreateDocument(ctx, kind, &doc)
}

func (s *saleDocumentService) UpdateDocument(ctx context.Context, id int32, kind string, doc models.Document) (*models.Document, error) {
	if id <= 0 {
		return nil, fmt.Errorf("invalid document id")
	}
	return s.repo.UpdateDocument(ctx, id, kind, &doc)
}

func (s *saleDocumentService) DeleteDocument(ctx context.Context, id int32, kind string) error {
	if id <= 0 {
		return fmt.Errorf("invalid document id")
	}
	return s.repo.DeleteDocument(ctx, id, kind)
}

func (s *saleDocumentService) Transform(ctx context.Context, sourceID int32, targetKind string, docDate *int64, docNumber string) (*models.Document, error) {
	if sourceID <= 0 {
		return nil, fmt.Errorf("invalid source document id")
	}
	return s.repo.TransformDocument(ctx, sourceID, targetKind, docDate, docNumber)
}

func (s *saleDocumentService) TransformList(ctx context.Context, targetKind string, payload models.DocumentListTransformation) ([]models.Document, error) {
	return s.repo.TransformDocumentList(ctx, targetKind, payload)
}

func (s *saleDocumentService) GetPDF(ctx context.Context, id int32) ([]byte, error) {
	if id <= 0 {
		return nil, fmt.Errorf("invalid document id")
	}
	return s.repo.GetPDF(ctx, id)
}

func (s *saleDocumentService) GetAttachment(ctx context.Context, idDoc int32) (*models.SaleDocumentAttachment, error) {
	return s.repo.GetAttachment(ctx, idDoc)
}

func (s *saleDocumentService) SaveAttachment(ctx context.Context, idDoc int32, filename, contentType string, data []byte) error {
	if len(data) == 0 {
		return fmt.Errorf("attachment data is required")
	}
	return s.repo.SaveAttachment(ctx, idDoc, filename, contentType, data)
}

func (s *saleDocumentService) DeleteAttachment(ctx context.Context, idDoc int32) error {
	return s.repo.DeleteAttachment(ctx, idDoc)
}

func (s *saleDocumentService) UpdateOrderPayments(ctx context.Context, id int32, rows []models.DocumentPaymentRowBase) error {
	return s.repo.UpdateOrderPayments(ctx, id, rows)
}

func (s *saleDocumentService) AddInvoicePayment(ctx context.Context, id int32, payload models.DocumentPaymentPayload) error {
	return s.repo.AddInvoicePayment(ctx, id, payload)
}

func (s *saleDocumentService) ListCustomerPayments(ctx context.Context, q models.ListQuery) ([]models.CustomerPayment, error) {
	return s.repo.ListCustomerPayments(ctx, q)
}

func (s *saleDocumentService) GetCustomerPayment(ctx context.Context, id int32) (*models.CustomerPayment, error) {
	return s.repo.GetCustomerPayment(ctx, id)
}

func (s *saleDocumentService) CreateCustomerPayment(ctx context.Context, p models.CustomerPayment) (*models.CustomerPayment, error) {
	return s.repo.CreateCustomerPayment(ctx, &p)
}

func (s *saleDocumentService) UpdateCustomerPayment(ctx context.Context, id int32, p models.CustomerPayment) (*models.CustomerPayment, error) {
	return s.repo.UpdateCustomerPayment(ctx, id, &p)
}

func (s *saleDocumentService) DeleteCustomerPayment(ctx context.Context, id int32) error {
	return s.repo.DeleteCustomerPayment(ctx, id)
}

func (s *saleDocumentService) GetCustomerPaymentByDoc(ctx context.Context, idDoc int32) (*models.CustomerPayment, error) {
	return s.repo.GetCustomerPaymentByDoc(ctx, idDoc)
}

func (s *saleDocumentService) OpenBalance(ctx context.Context, idCustomer string, limit, offset int) ([]models.OpenBalanceRow, error) {
	return s.repo.ListOpenBalance(ctx, idCustomer, limit, offset)
}

func (s *saleDocumentService) Reports(ctx context.Context, reportType, startDate, endDate string) (map[string]any, error) {
	return s.repo.SaleReports(ctx, reportType, startDate, endDate)
}

func (s *saleDocumentService) EInvoice(ctx context.Context, idDoc int32) (map[string]any, error) {
	return s.repo.SubmitEInvoice(ctx, idDoc)
}
