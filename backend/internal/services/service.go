package services

import (
	"dnsc_microservice/internal/config"
	"dnsc_microservice/internal/repository"
)

// AppServices holds all service interfaces.
type AppServices struct {
	Auth                AuthService
	Country             CountryService
	Reference           ReferenceService
	Company             CompanyService
	Contact             ContactService
	Customer            CustomerService
	Vendor              VendorService
	Product             ProductService
	ProductGroup        ProductGroupService
	Attribute           AttributeService
	Pricelist           PricelistService
	Storage             StorageService
	StorageLocation     StorageLocationService
	Stock               StockService
	Lot                 LotService
	MachineDataTracking MachineDataTrackingService
	SaleDocument        SaleDocumentService
	PurchaseDocument    PurchaseDocumentService
	Accounting          AccountingService
	Crm                 CrmService
	Calendar            CalendarService
	PersonalActivity    PersonalActivityService
	User                UserService
	Settings            SettingsService
	Message             MessageService
	LoggedUser          LoggedUserService
	Notify              NotifyService
}

// NewAppServices initializes all services.
func NewAppServices(repos *repository.Repository, cfg *config.Config) *AppServices {
	return &AppServices{
		Auth: NewAuthService(
			repos.Auth,
			cfg.JWTSecret,
			cfg.AccessTokenTTL(),
			cfg.RefreshTokenTTL(),
		),
		Country:             NewCountryService(repos.Country),
		Reference:           NewReferenceService(repos.Reference),
		Company:             NewCompanyService(repos.Company),
		Contact:             NewContactService(repos.Contact),
		Customer:            NewCustomerService(repos.Customer),
		Vendor:              NewVendorService(repos.Vendor),
		Product:             NewProductService(repos.Product),
		ProductGroup:        NewProductGroupService(repos.ProductGroup),
		Attribute:           NewAttributeService(repos.Attribute),
		Pricelist:           NewPricelistService(repos.Pricelist),
		Storage:             NewStorageService(repos.Storage),
		StorageLocation:     NewStorageLocationService(repos.StorageLocation),
		Stock:               NewStockService(repos.Stock),
		Lot:                 NewLotService(repos.Lot),
		MachineDataTracking: NewMachineDataTrackingService(repos.MachineDataTracking),
		SaleDocument:        NewSaleDocumentService(repos.SaleDocument),
		PurchaseDocument:    NewPurchaseDocumentService(repos.PurchaseDocument),
		Accounting:          NewAccountingService(repos.Accounting),
		Crm:                 NewCrmService(repos.Crm),
		Calendar:            NewCalendarService(repos.Calendar),
		PersonalActivity:    NewPersonalActivityService(repos.PersonalActivity),
		User:                NewUserService(repos.User),
		Settings:            NewSettingsService(repos.Settings),
		Message:             NewMessageService(repos.Message),
		LoggedUser:          NewLoggedUserService(repos.LoggedUser),
		Notify:              NewNotifyService(repos.Notify),
	}
}
