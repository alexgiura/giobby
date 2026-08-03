package services

import (
	"context"
	"fmt"
	"strings"

	"dnsc_microservice/internal/models"
	"dnsc_microservice/internal/repository"

	"github.com/google/uuid"
)

type LoggedUserService interface {
	GetInfo(ctx context.Context, user *models.User) (*models.LoggedUserInfo, error)
	ChangeLanguage(ctx context.Context, userID uuid.UUID, language string) error
	SetImage(ctx context.Context, userID uuid.UUID, imageURL string) error
	AddDeviceToken(ctx context.Context, userID uuid.UUID, in models.DeviceToken) (*models.DeviceToken, error)
	DeleteDeviceToken(ctx context.Context, userID uuid.UUID, id int64) error
}

type loggedUserService struct{ repo repository.LoggedUserRepository }

func NewLoggedUserService(repo repository.LoggedUserRepository) LoggedUserService {
	return &loggedUserService{repo: repo}
}

func (s *loggedUserService) GetInfo(ctx context.Context, user *models.User) (*models.LoggedUserInfo, error) {
	lang, img, err := s.repo.GetOrCreateProfile(ctx, user.ID)
	if err != nil {
		return nil, err
	}
	return &models.LoggedUserInfo{
		ID:       user.ID,
		Username: user.Username,
		Email:    user.Email,
		Language: lang,
		ImageURL: img,
		IsActive: user.IsActive,
	}, nil
}

func (s *loggedUserService) ChangeLanguage(ctx context.Context, userID uuid.UUID, language string) error {
	language = strings.TrimSpace(language)
	if language == "" {
		return fmt.Errorf("language is required")
	}
	return s.repo.SetLanguage(ctx, userID, language)
}

func (s *loggedUserService) SetImage(ctx context.Context, userID uuid.UUID, imageURL string) error {
	imageURL = strings.TrimSpace(imageURL)
	if imageURL == "" {
		return fmt.Errorf("image url is required")
	}
	return s.repo.SetImageURL(ctx, userID, imageURL)
}

func (s *loggedUserService) AddDeviceToken(ctx context.Context, userID uuid.UUID, in models.DeviceToken) (*models.DeviceToken, error) {
	if strings.TrimSpace(in.DeviceToken) == "" {
		return nil, fmt.Errorf("deviceToken is required")
	}
	return s.repo.AddDeviceToken(ctx, userID, in.OS, in.DeviceToken)
}

func (s *loggedUserService) DeleteDeviceToken(ctx context.Context, userID uuid.UUID, id int64) error {
	return s.repo.DeleteDeviceToken(ctx, userID, id)
}
