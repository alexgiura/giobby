package services

import (
	"context"
	"fmt"
	"strings"

	"dnsc_microservice/internal/models"
	"dnsc_microservice/internal/repository"
)

type MessageService interface {
	ListMessages(ctx context.Context, q models.MessageListQuery) ([]models.Message, error)
	ListUnread(ctx context.Context) ([]models.Message, error)
	GetMessage(ctx context.Context, id int64) (*models.Message, error)
	CreateMessage(ctx context.Context, m models.Message) (*models.Message, error)
	MarkMessageRead(ctx context.Context, id int64) (*models.Message, error)
	ListContacts(ctx context.Context, limit, offset int, name string) ([]models.MessageContact, error)
	ListGroups(ctx context.Context, limit, offset int) ([]models.MessageGroup, error)
	DeleteGroup(ctx context.Context, id int64) error
	MarkGroupRead(ctx context.Context, idGroup int64) error
}

type messageService struct{ repo repository.MessageRepository }

func NewMessageService(repo repository.MessageRepository) MessageService {
	return &messageService{repo: repo}
}

func (s *messageService) ListMessages(ctx context.Context, q models.MessageListQuery) ([]models.Message, error) {
	return s.repo.ListMessages(ctx, q)
}
func (s *messageService) ListUnread(ctx context.Context) ([]models.Message, error) {
	return s.repo.ListUnread(ctx)
}
func (s *messageService) GetMessage(ctx context.Context, id int64) (*models.Message, error) {
	return s.repo.GetMessage(ctx, id)
}
func (s *messageService) CreateMessage(ctx context.Context, m models.Message) (*models.Message, error) {
	if strings.TrimSpace(m.Body) == "" {
		return nil, fmt.Errorf("body is required")
	}
	m.SentMsg = true
	return s.repo.CreateMessage(ctx, &m)
}
func (s *messageService) MarkMessageRead(ctx context.Context, id int64) (*models.Message, error) {
	return s.repo.MarkMessageRead(ctx, id)
}
func (s *messageService) ListContacts(ctx context.Context, limit, offset int, name string) ([]models.MessageContact, error) {
	return s.repo.ListContacts(ctx, limit, offset, name)
}
func (s *messageService) ListGroups(ctx context.Context, limit, offset int) ([]models.MessageGroup, error) {
	return s.repo.ListGroups(ctx, limit, offset)
}
func (s *messageService) DeleteGroup(ctx context.Context, id int64) error {
	return s.repo.DeleteGroup(ctx, id)
}
func (s *messageService) MarkGroupRead(ctx context.Context, idGroup int64) error {
	return s.repo.MarkGroupRead(ctx, idGroup)
}
