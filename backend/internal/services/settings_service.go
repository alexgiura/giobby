package services

import (
	"context"
	"fmt"
	"strings"

	"dnsc_microservice/internal/models"
	"dnsc_microservice/internal/repository"
)

// SettingsService handles application settings.
type SettingsService interface {
	ListSettings(ctx context.Context, q models.AppSettingListQuery) ([]models.AppSetting, error)
	UpsertSetting(ctx context.Context, s models.AppSetting) (*models.AppSetting, error)
}

type settingsService struct {
	repo repository.SettingsRepository
}

func NewSettingsService(repo repository.SettingsRepository) SettingsService {
	return &settingsService{repo: repo}
}

func (s *settingsService) ListSettings(ctx context.Context, q models.AppSettingListQuery) ([]models.AppSetting, error) {
	return s.repo.ListSettings(ctx, q)
}

func (s *settingsService) UpsertSetting(ctx context.Context, setting models.AppSetting) (*models.AppSetting, error) {
	if strings.TrimSpace(setting.Key1) == "" {
		return nil, fmt.Errorf("key1 is required")
	}
	return s.repo.UpsertSetting(ctx, &setting)
}
