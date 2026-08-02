package services

import (
	"context"
	"fmt"
	"strings"
	"time"

	"dnsc_microservice/internal/models"
	"dnsc_microservice/internal/repository"
)

const defaultCompanyID int64 = 1

// CompanyService handles company profile and settings.
type CompanyService interface {
	GetCompany(ctx context.Context, id int64) (*models.Company, error)

	ListAccountCenters(ctx context.Context, q models.AccountCenterListQuery) ([]models.AccountCenter, error)
	GetAccountCenter(ctx context.Context, id int32) (*models.AccountCenter, error)
	CreateAccountCenter(ctx context.Context, ac models.AccountCenter) (*models.AccountCenter, error)
	UpdateAccountCenter(ctx context.Context, id int32, ac models.AccountCenter) (*models.AccountCenter, error)
	DeleteAccountCenter(ctx context.Context, id int32) error

	ListAccountCodes(ctx context.Context, q models.AccountCodeListQuery) ([]models.AccountCode, error)

	ListBanks(ctx context.Context, q models.BankListQuery) ([]models.BankCashdesk, error)
	GetBank(ctx context.Context, id int32) (*models.BankCashdesk, error)
	CreateBank(ctx context.Context, b models.BankCashdesk) (*models.BankCashdesk, error)
	UpdateBank(ctx context.Context, id int32, b models.BankCashdesk) (*models.BankCashdesk, error)
	DeleteBank(ctx context.Context, id int32) error

	ListBupNumerators(ctx context.Context, q models.BupNumeratorQuery) ([]models.BupNumerator, error)
	GetCurrencyChange(ctx context.Context, sourceCurrency string, asOf *time.Time) (*models.CurrencyChange, error)

	ListVatRates(ctx context.Context, q models.VatListQuery) ([]models.VatRate, error)
	GetVatRate(ctx context.Context, idVat string, q models.VatGetQuery) (*models.VatRate, error)
}

type companyService struct {
	repo repository.CompanyRepository
}

func NewCompanyService(repo repository.CompanyRepository) CompanyService {
	return &companyService{repo: repo}
}

func (s *companyService) GetCompany(ctx context.Context, id int64) (*models.Company, error) {
	if id <= 0 {
		return nil, fmt.Errorf("invalid company id")
	}
	return s.repo.GetCompany(ctx, id)
}

func (s *companyService) ListAccountCenters(ctx context.Context, q models.AccountCenterListQuery) ([]models.AccountCenter, error) {
	return s.repo.ListAccountCenters(ctx, defaultCompanyID, q)
}

func (s *companyService) GetAccountCenter(ctx context.Context, id int32) (*models.AccountCenter, error) {
	if id <= 0 {
		return nil, fmt.Errorf("invalid account center id")
	}
	return s.repo.GetAccountCenter(ctx, defaultCompanyID, id)
}

func (s *companyService) CreateAccountCenter(ctx context.Context, ac models.AccountCenter) (*models.AccountCenter, error) {
	ac.Name = strings.TrimSpace(ac.Name)
	if ac.Name == "" {
		return nil, fmt.Errorf("name is required")
	}
	return s.repo.CreateAccountCenter(ctx, defaultCompanyID, &ac)
}

func (s *companyService) UpdateAccountCenter(ctx context.Context, id int32, ac models.AccountCenter) (*models.AccountCenter, error) {
	ac.Name = strings.TrimSpace(ac.Name)
	if id <= 0 || ac.Name == "" {
		return nil, fmt.Errorf("invalid account center")
	}
	out, err := s.repo.UpdateAccountCenter(ctx, defaultCompanyID, id, &ac)
	if err != nil {
		return nil, err
	}
	if out == nil {
		return nil, fmt.Errorf("account center not found")
	}
	return out, nil
}

func (s *companyService) DeleteAccountCenter(ctx context.Context, id int32) error {
	if id <= 0 {
		return fmt.Errorf("invalid account center id")
	}
	return s.repo.DeleteAccountCenter(ctx, defaultCompanyID, id)
}

func (s *companyService) ListAccountCodes(ctx context.Context, q models.AccountCodeListQuery) ([]models.AccountCode, error) {
	return s.repo.ListAccountCodes(ctx, defaultCompanyID, q)
}

func (s *companyService) ListBanks(ctx context.Context, q models.BankListQuery) ([]models.BankCashdesk, error) {
	return s.repo.ListBanks(ctx, defaultCompanyID, q)
}

func (s *companyService) GetBank(ctx context.Context, id int32) (*models.BankCashdesk, error) {
	if id <= 0 {
		return nil, fmt.Errorf("invalid bank id")
	}
	return s.repo.GetBank(ctx, defaultCompanyID, id)
}

func (s *companyService) CreateBank(ctx context.Context, b models.BankCashdesk) (*models.BankCashdesk, error) {
	b.Description = strings.TrimSpace(b.Description)
	if b.Description == "" {
		return nil, fmt.Errorf("description is required")
	}
	return s.repo.CreateBank(ctx, defaultCompanyID, &b)
}

func (s *companyService) UpdateBank(ctx context.Context, id int32, b models.BankCashdesk) (*models.BankCashdesk, error) {
	b.Description = strings.TrimSpace(b.Description)
	if id <= 0 || b.Description == "" {
		return nil, fmt.Errorf("invalid bank")
	}
	return s.repo.UpdateBank(ctx, defaultCompanyID, id, &b)
}

func (s *companyService) DeleteBank(ctx context.Context, id int32) error {
	if id <= 0 {
		return fmt.Errorf("invalid bank id")
	}
	return s.repo.DeleteBank(ctx, defaultCompanyID, id)
}

func (s *companyService) ListBupNumerators(ctx context.Context, q models.BupNumeratorQuery) ([]models.BupNumerator, error) {
	return s.repo.ListBupNumerators(ctx, defaultCompanyID, q)
}

func (s *companyService) GetCurrencyChange(ctx context.Context, sourceCurrency string, asOf *time.Time) (*models.CurrencyChange, error) {
	sourceCurrency = strings.TrimSpace(sourceCurrency)
	if sourceCurrency == "" {
		return nil, fmt.Errorf("source currency is required")
	}
	return s.repo.GetCurrencyChange(ctx, defaultCompanyID, sourceCurrency, asOf)
}

func (s *companyService) ListVatRates(ctx context.Context, q models.VatListQuery) ([]models.VatRate, error) {
	return s.repo.ListVatRates(ctx, defaultCompanyID, q)
}

func (s *companyService) GetVatRate(ctx context.Context, idVat string, q models.VatGetQuery) (*models.VatRate, error) {
	idVat = strings.TrimSpace(idVat)
	if idVat == "" {
		return nil, fmt.Errorf("idVat is required")
	}
	return s.repo.GetVatRate(ctx, defaultCompanyID, idVat, q)
}
