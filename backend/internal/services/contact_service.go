package services

import (
	"context"
	"fmt"
	"strings"

	"dnsc_microservice/internal/models"
	"dnsc_microservice/internal/repository"
)

// ContactService handles contacts, offices and sub-contacts.
type ContactService interface {
	ListContacts(ctx context.Context, q models.ContactListQuery) ([]models.Contact, error)
	GetContact(ctx context.Context, id int32, retrieveImage bool) (*models.Contact, error)
	CreateContact(ctx context.Context, c models.Contact) (*models.Contact, error)
	UpdateContact(ctx context.Context, id int32, c models.Contact) (*models.Contact, error)
	PatchContact(ctx context.Context, id int32, fields map[string]any) (*models.Contact, error)
	DeleteContact(ctx context.Context, id int32) error
	UpdateContactPhoto(ctx context.Context, id int32, data []byte, mimeType string) error
	ListContactSources(ctx context.Context, q models.ContactSourceListQuery) ([]models.ContactSource, error)
	ListOffices(ctx context.Context, contactID int32, q models.OfficeListQuery) ([]models.ContactOffice, error)
	CreateOffice(ctx context.Context, contactID int32, office models.ContactOffice) (*models.ContactOffice, error)
	UpdateOffice(ctx context.Context, contactID, officeID int32, office models.ContactOffice) (*models.ContactOffice, error)
	DeleteOffice(ctx context.Context, contactID, officeID int32) error
	ListSubContacts(ctx context.Context, contactID int32, q models.SubContactListQuery) ([]models.SubContactAssoc, error)
	CreateSubContact(ctx context.Context, contactID int32, assoc models.SubContactAssoc) (*models.SubContactAssoc, error)
	DeleteSubContact(ctx context.Context, contactID, subContactID int32) error
}

type contactService struct {
	repo repository.ContactRepository
}

func NewContactService(repo repository.ContactRepository) ContactService {
	return &contactService{repo: repo}
}

func (s *contactService) ListContacts(ctx context.Context, q models.ContactListQuery) ([]models.Contact, error) {
	return s.repo.ListContacts(ctx, q)
}

func (s *contactService) GetContact(ctx context.Context, id int32, retrieveImage bool) (*models.Contact, error) {
	if id <= 0 {
		return nil, fmt.Errorf("invalid contact id")
	}
	return s.repo.GetContact(ctx, id, retrieveImage)
}

func (s *contactService) CreateContact(ctx context.Context, c models.Contact) (*models.Contact, error) {
	c.Name = strings.TrimSpace(c.Name)
	if c.Name == "" {
		return nil, fmt.Errorf("name is required")
	}
	if c.Type == "" {
		c.Type = "COMPANY"
	}
	return s.repo.CreateContact(ctx, &c)
}

func (s *contactService) UpdateContact(ctx context.Context, id int32, c models.Contact) (*models.Contact, error) {
	c.Name = strings.TrimSpace(c.Name)
	if id <= 0 || c.Name == "" {
		return nil, fmt.Errorf("invalid contact")
	}
	return s.repo.UpdateContact(ctx, id, &c)
}

func (s *contactService) PatchContact(ctx context.Context, id int32, fields map[string]any) (*models.Contact, error) {
	if id <= 0 || len(fields) == 0 {
		return nil, fmt.Errorf("invalid patch")
	}
	return s.repo.PatchContact(ctx, id, fields)
}

func (s *contactService) DeleteContact(ctx context.Context, id int32) error {
	if id <= 0 {
		return fmt.Errorf("invalid contact id")
	}
	return s.repo.DeleteContact(ctx, id)
}

func (s *contactService) UpdateContactPhoto(ctx context.Context, id int32, data []byte, mimeType string) error {
	if id <= 0 || len(data) == 0 {
		return fmt.Errorf("invalid photo upload")
	}
	return s.repo.UpdateContactPhoto(ctx, id, data, mimeType)
}

func (s *contactService) ListContactSources(ctx context.Context, q models.ContactSourceListQuery) ([]models.ContactSource, error) {
	return s.repo.ListContactSources(ctx, q)
}

func (s *contactService) ListOffices(ctx context.Context, contactID int32, q models.OfficeListQuery) ([]models.ContactOffice, error) {
	if contactID <= 0 {
		return nil, fmt.Errorf("invalid contact id")
	}
	return s.repo.ListOffices(ctx, contactID, q)
}

func (s *contactService) CreateOffice(ctx context.Context, contactID int32, office models.ContactOffice) (*models.ContactOffice, error) {
	if contactID <= 0 {
		return nil, fmt.Errorf("invalid contact id")
	}
	return s.repo.CreateOffice(ctx, contactID, &office)
}

func (s *contactService) UpdateOffice(ctx context.Context, contactID, officeID int32, office models.ContactOffice) (*models.ContactOffice, error) {
	if contactID <= 0 || officeID <= 0 {
		return nil, fmt.Errorf("invalid office")
	}
	return s.repo.UpdateOffice(ctx, contactID, officeID, &office)
}

func (s *contactService) DeleteOffice(ctx context.Context, contactID, officeID int32) error {
	if contactID <= 0 || officeID <= 0 {
		return fmt.Errorf("invalid office")
	}
	return s.repo.DeleteOffice(ctx, contactID, officeID)
}

func (s *contactService) ListSubContacts(ctx context.Context, contactID int32, q models.SubContactListQuery) ([]models.SubContactAssoc, error) {
	if contactID <= 0 {
		return nil, fmt.Errorf("invalid contact id")
	}
	return s.repo.ListSubContacts(ctx, contactID, q)
}

func (s *contactService) CreateSubContact(ctx context.Context, contactID int32, assoc models.SubContactAssoc) (*models.SubContactAssoc, error) {
	if contactID <= 0 || assoc.IDSubContact <= 0 {
		return nil, fmt.Errorf("invalid subcontact association")
	}
	return s.repo.CreateSubContact(ctx, contactID, &assoc)
}

func (s *contactService) DeleteSubContact(ctx context.Context, contactID, subContactID int32) error {
	if contactID <= 0 || subContactID <= 0 {
		return fmt.Errorf("invalid subcontact")
	}
	return s.repo.DeleteSubContact(ctx, contactID, subContactID)
}
