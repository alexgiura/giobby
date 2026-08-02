package services

import (
	"context"
	"fmt"

	"dnsc_microservice/internal/models"
	"dnsc_microservice/internal/repository"
)

// PersonalActivityService handles personal activities.
type PersonalActivityService interface {
	List(ctx context.Context, q models.PersonalActivityListQuery) ([]models.PersonalActivity, error)
	Get(ctx context.Context, id int32) (*models.PersonalActivity, error)
	Create(ctx context.Context, a models.PersonalActivity) (*models.PersonalActivity, error)
	Update(ctx context.Context, id int32, a models.PersonalActivity) (*models.PersonalActivity, error)
	Patch(ctx context.Context, id int32, fields map[string]any) (*models.PersonalActivity, error)
	Delete(ctx context.Context, id int32) error
}

type personalActivityService struct {
	repo repository.PersonalActivityRepository
}

func NewPersonalActivityService(repo repository.PersonalActivityRepository) PersonalActivityService {
	return &personalActivityService{repo: repo}
}

func (s *personalActivityService) List(ctx context.Context, q models.PersonalActivityListQuery) ([]models.PersonalActivity, error) {
	return s.repo.List(ctx, q)
}

func (s *personalActivityService) Get(ctx context.Context, id int32) (*models.PersonalActivity, error) {
	if id <= 0 {
		return nil, fmt.Errorf("invalid activity id")
	}
	return s.repo.Get(ctx, id)
}

func (s *personalActivityService) Create(ctx context.Context, a models.PersonalActivity) (*models.PersonalActivity, error) {
	return s.repo.Create(ctx, &a)
}

func (s *personalActivityService) Update(ctx context.Context, id int32, a models.PersonalActivity) (*models.PersonalActivity, error) {
	if id <= 0 {
		return nil, fmt.Errorf("invalid activity id")
	}
	return s.repo.Update(ctx, id, &a)
}

func (s *personalActivityService) Patch(ctx context.Context, id int32, fields map[string]any) (*models.PersonalActivity, error) {
	if id <= 0 {
		return nil, fmt.Errorf("invalid activity id")
	}
	return s.repo.Patch(ctx, id, fields)
}

func (s *personalActivityService) Delete(ctx context.Context, id int32) error {
	if id <= 0 {
		return fmt.Errorf("invalid activity id")
	}
	return s.repo.Delete(ctx, id)
}
