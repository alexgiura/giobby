package services

import (
	"context"
	"fmt"
	"strings"

	"dnsc_microservice/internal/models"
	"dnsc_microservice/internal/repository"

	"golang.org/x/crypto/bcrypt"
)

const userBcryptCost = 12

// UserService handles Giobby company users.
type UserService interface {
	ListUsers(ctx context.Context, q models.CompanyUserListQuery) ([]models.CompanyUser, error)
	GetUser(ctx context.Context, id int32, retrieveImage bool) (*models.CompanyUser, error)
	CreateUser(ctx context.Context, user models.CompanyUser) (*models.CompanyUser, error)
	UpdateUser(ctx context.Context, id int32, user models.CompanyUser) (*models.CompanyUser, error)
	DeleteUser(ctx context.Context, id int32) error
	ListAgentCommissions(ctx context.Context, q models.AgentCommissionListQuery) ([]models.AgentCommission, error)
	AddAuthProfiles(ctx context.Context, idUser int32, profiles []string) error
}

type userService struct {
	repo repository.UserRepository
}

func NewUserService(repo repository.UserRepository) UserService {
	return &userService{repo: repo}
}

func (s *userService) ListUsers(ctx context.Context, q models.CompanyUserListQuery) ([]models.CompanyUser, error) {
	return s.repo.ListUsers(ctx, q)
}

func (s *userService) GetUser(ctx context.Context, id int32, retrieveImage bool) (*models.CompanyUser, error) {
	if id <= 0 {
		return nil, fmt.Errorf("invalid user id")
	}
	return s.repo.GetUser(ctx, id, retrieveImage)
}

func (s *userService) CreateUser(ctx context.Context, user models.CompanyUser) (*models.CompanyUser, error) {
	if strings.TrimSpace(user.Username) == "" {
		return nil, fmt.Errorf("username is required")
	}
	hash, err := hashUserPassword(user.Password)
	if err != nil {
		return nil, err
	}
	user.Password = ""
	return s.repo.CreateUser(ctx, &user, hash)
}

func (s *userService) UpdateUser(ctx context.Context, id int32, user models.CompanyUser) (*models.CompanyUser, error) {
	if id <= 0 {
		return nil, fmt.Errorf("invalid user id")
	}
	hash, err := hashUserPassword(user.Password)
	if err != nil {
		return nil, err
	}
	user.Password = ""
	return s.repo.UpdateUser(ctx, id, &user, hash)
}

func (s *userService) DeleteUser(ctx context.Context, id int32) error {
	if id <= 0 {
		return fmt.Errorf("invalid user id")
	}
	return s.repo.DeleteUser(ctx, id)
}

func (s *userService) ListAgentCommissions(ctx context.Context, q models.AgentCommissionListQuery) ([]models.AgentCommission, error) {
	return s.repo.ListAgentCommissions(ctx, q)
}

func (s *userService) AddAuthProfiles(ctx context.Context, idUser int32, profiles []string) error {
	if idUser <= 0 {
		return fmt.Errorf("invalid user id")
	}
	if len(profiles) == 0 {
		return fmt.Errorf("at least one profile is required")
	}
	return s.repo.AddAuthProfiles(ctx, idUser, profiles)
}

func hashUserPassword(password string) (string, error) {
	if strings.TrimSpace(password) == "" {
		return "", nil
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), userBcryptCost)
	if err != nil {
		return "", fmt.Errorf("hash password: %w", err)
	}
	return string(hash), nil
}
