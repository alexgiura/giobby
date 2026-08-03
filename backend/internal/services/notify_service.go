package services

import (
	"context"

	"dnsc_microservice/internal/models"
	"dnsc_microservice/internal/repository"
)

type NotifyService interface {
	ListNotifies(ctx context.Context, channel string, q models.NotifyListQuery) ([]models.AppNotify, error)
	ListUnreadNotifies(ctx context.Context, channel string) ([]models.AppNotify, error)
	GetNotify(ctx context.Context, channel string, id int64) (*models.AppNotify, error)
	MarkNotifyRead(ctx context.Context, channel string, id int64) (*models.AppNotify, error)
	ListSocialPosts(ctx context.Context, limit, offset int, idLang string) ([]models.SocialPost, error)
	ListEmails(ctx context.Context, limit, offset int, hashcode string) ([]models.EmailItem, error)
	ListGlobalNotifications(ctx context.Context) ([]models.GlobalNotification, error)
	BindCommerceUpdateStock(ctx context.Context, in models.BindCommerceRequest) (*models.IntegrationAck, error)
	TilbySales(ctx context.Context, payload any) (*models.IntegrationAck, error)
}

type notifyService struct{ repo repository.NotifyRepository }

func NewNotifyService(repo repository.NotifyRepository) NotifyService {
	return &notifyService{repo: repo}
}

func (s *notifyService) ListNotifies(ctx context.Context, channel string, q models.NotifyListQuery) ([]models.AppNotify, error) {
	return s.repo.ListNotifies(ctx, channel, q)
}
func (s *notifyService) ListUnreadNotifies(ctx context.Context, channel string) ([]models.AppNotify, error) {
	return s.repo.ListUnreadNotifies(ctx, channel)
}
func (s *notifyService) GetNotify(ctx context.Context, channel string, id int64) (*models.AppNotify, error) {
	return s.repo.GetNotify(ctx, channel, id)
}
func (s *notifyService) MarkNotifyRead(ctx context.Context, channel string, id int64) (*models.AppNotify, error) {
	return s.repo.MarkNotifyRead(ctx, channel, id)
}
func (s *notifyService) ListSocialPosts(ctx context.Context, limit, offset int, idLang string) ([]models.SocialPost, error) {
	return s.repo.ListSocialPosts(ctx, limit, offset, idLang)
}
func (s *notifyService) ListEmails(ctx context.Context, limit, offset int, hashcode string) ([]models.EmailItem, error) {
	return s.repo.ListEmails(ctx, limit, offset, hashcode)
}
func (s *notifyService) ListGlobalNotifications(ctx context.Context) ([]models.GlobalNotification, error) {
	return s.repo.ListGlobalNotifications(ctx)
}
func (s *notifyService) BindCommerceUpdateStock(ctx context.Context, in models.BindCommerceRequest) (*models.IntegrationAck, error) {
	if err := s.repo.SaveIntegrationEvent(ctx, "bindcommerce", in); err != nil {
		return nil, err
	}
	return &models.IntegrationAck{OK: true, Message: "stock update accepted"}, nil
}
func (s *notifyService) TilbySales(ctx context.Context, payload any) (*models.IntegrationAck, error) {
	if err := s.repo.SaveIntegrationEvent(ctx, "tilby", payload); err != nil {
		return nil, err
	}
	return &models.IntegrationAck{OK: true, Message: "tilby sale accepted"}, nil
}
