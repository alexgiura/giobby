package repository

import (
	"github.com/jackc/pgx/v4/pgxpool"
)

// Repository holds all repository interfaces.
type Repository struct {
	Auth                AuthRepository
	Reference           ReferenceRepository
	Company             CompanyRepository
	Contact             ContactRepository
	Customer            CustomerRepository
	Vendor              VendorRepository
	Product             ProductRepository
	ProductGroup        ProductGroupRepository
	Attribute           AttributeRepository
	Pricelist           PricelistRepository
	Storage             StorageRepository
	StorageLocation     StorageLocationRepository
	Stock               StockRepository
	Lot                 LotRepository
	MachineDataTracking MachineDataTrackingRepository
	SaleDocument        SaleDocumentRepository
	PurchaseDocument    PurchaseDocumentRepository
	Accounting          AccountingRepository
	Crm                 CrmRepository
	Calendar            CalendarRepository
	PersonalActivity    PersonalActivityRepository
	User                UserRepository
	Settings            SettingsRepository
	Message             MessageRepository
	LoggedUser          LoggedUserRepository
	Notify              NotifyRepository
}

// NewRepository initializes all repositories.
func NewRepository(db *pgxpool.Pool) *Repository {
	contacts := NewContactRepository(db)
	return &Repository{
		Auth:                NewAuthRepository(db),
		Reference:           NewReferenceRepository(db),
		Company:             NewCompanyRepository(db),
		Contact:             contacts,
		Customer:            NewCustomerRepository(db, contacts),
		Vendor:              NewVendorRepository(db, contacts),
		Product:             NewProductRepository(db),
		ProductGroup:        NewProductGroupRepository(db),
		Attribute:           NewAttributeRepository(db),
		Pricelist:           NewPricelistRepository(db),
		Storage:             NewStorageRepository(db),
		StorageLocation:     NewStorageLocationRepository(db),
		Stock:               NewStockRepository(db),
		Lot:                 NewLotRepository(db),
		MachineDataTracking: NewMachineDataTrackingRepository(db),
		SaleDocument:        NewSaleDocumentRepository(db),
		PurchaseDocument:    NewPurchaseDocumentRepository(db),
		Accounting:          NewAccountingRepository(db),
		Crm:                 NewCrmRepository(db),
		Calendar:            NewCalendarRepository(db),
		PersonalActivity:    NewPersonalActivityRepository(db),
		User:                NewUserRepository(db),
		Settings:            NewSettingsRepository(db),
		Message:             NewMessageRepository(db),
		LoggedUser:          NewLoggedUserRepository(db),
		Notify:              NewNotifyRepository(db),
	}
}
