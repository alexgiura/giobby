package services

import (
	"context"
	"dnsc_microservice/internal/auth"
	"dnsc_microservice/internal/models"
	"dnsc_microservice/internal/repository"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

const bcryptCost = 12

// AuthService handles local registration, login, token issue/refresh, logout.
type AuthService interface {
	Register(ctx context.Context, username, password string, email *string) (*models.TokenResponse, error)
	Login(ctx context.Context, username, password string) (*models.TokenResponse, error)
	Refresh(ctx context.Context, refreshToken string) (*models.TokenResponse, error)
	Logout(ctx context.Context, sessionID uuid.UUID) error
	LogoutByAccessToken(ctx context.Context, accessToken string) error
	ValidateAccessToken(ctx context.Context, accessToken string) (*models.User, error)
	ToPublic(u *models.User) models.UserPublic
}

type authService struct {
	repo       repository.AuthRepository
	jwtSecret  string
	accessTTL  time.Duration
	refreshTTL time.Duration
}

// NewAuthService creates local OAuth-style auth service.
func NewAuthService(repo repository.AuthRepository, jwtSecret string, accessTTL, refreshTTL time.Duration) AuthService {
	if accessTTL <= 0 {
		accessTTL = time.Hour
	}
	if refreshTTL <= 0 {
		refreshTTL = 7 * 24 * time.Hour
	}
	return &authService{
		repo:       repo,
		jwtSecret:  jwtSecret,
		accessTTL:  accessTTL,
		refreshTTL: refreshTTL,
	}
}

func (s *authService) Register(ctx context.Context, username, password string, email *string) (*models.TokenResponse, error) {
	username = strings.TrimSpace(username)
	password = strings.TrimSpace(password)
	if username == "" || password == "" {
		return nil, fmt.Errorf("username and password are required")
	}
	if len(password) < 8 {
		return nil, fmt.Errorf("password must be at least 8 characters")
	}

	existing, err := s.repo.GetUserByUsername(ctx, username)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, errors.New("username already exists")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	var emailVal *string
	if email != nil {
		v := strings.TrimSpace(*email)
		if v != "" {
			emailVal = &v
		}
	}

	user := &models.User{
		ID:           uuid.New(),
		Username:     username,
		Email:        emailVal,
		PasswordHash: string(hash),
		IsActive:     true,
	}
	if err := s.repo.CreateUser(ctx, user); err != nil {
		return nil, err
	}
	return s.issueTokens(ctx, user)
}

func (s *authService) Login(ctx context.Context, username, password string) (*models.TokenResponse, error) {
	username = strings.TrimSpace(username)
	password = strings.TrimSpace(password)
	if username == "" || password == "" {
		return nil, fmt.Errorf("username and password are required")
	}

	u, err := s.repo.GetUserByUsername(ctx, username)
	if err != nil {
		return nil, err
	}
	if u == nil || !u.IsActive {
		return nil, errors.New("invalid credentials")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)); err != nil {
		return nil, errors.New("invalid credentials")
	}
	return s.issueTokens(ctx, u)
}

func (s *authService) Refresh(ctx context.Context, refreshToken string) (*models.TokenResponse, error) {
	refreshToken = strings.TrimSpace(refreshToken)
	if refreshToken == "" {
		return nil, errors.New("refresh_token required")
	}

	hash := auth.HashRefreshToken(refreshToken)
	sess, err := s.repo.GetSessionByRefreshTokenHash(ctx, hash)
	if err != nil {
		return nil, err
	}
	if sess == nil {
		return nil, errors.New("invalid refresh token")
	}
	if sess.RevokedAt != nil {
		return nil, errors.New("refresh token revoked")
	}
	if sess.ExpiresAt.Before(time.Now().UTC()) {
		_ = s.repo.DeleteSession(ctx, sess.ID)
		return nil, errors.New("refresh token expired")
	}

	u, err := s.repo.GetUserByID(ctx, sess.UserID)
	if err != nil {
		return nil, err
	}
	if u == nil || !u.IsActive {
		return nil, errors.New("user invalid")
	}

	if err := s.repo.DeleteSession(ctx, sess.ID); err != nil {
		return nil, err
	}
	return s.issueTokens(ctx, u)
}

func (s *authService) Logout(ctx context.Context, sessionID uuid.UUID) error {
	return s.repo.RevokeSession(ctx, sessionID)
}

func (s *authService) LogoutByAccessToken(ctx context.Context, accessToken string) error {
	_, sessionID, err := auth.ParseAccessToken(accessToken, s.jwtSecret)
	if err != nil {
		return errors.New("invalid access token")
	}
	return s.Logout(ctx, sessionID)
}

func (s *authService) ValidateAccessToken(ctx context.Context, accessToken string) (*models.User, error) {
	userID, sessionID, err := auth.ParseAccessToken(accessToken, s.jwtSecret)
	if err != nil {
		return nil, errors.New("invalid access token")
	}

	sess, err := s.repo.GetSessionByID(ctx, sessionID)
	if err != nil {
		return nil, err
	}
	if sess == nil || sess.RevokedAt != nil || sess.UserID != userID {
		return nil, errors.New("session invalid")
	}
	if sess.ExpiresAt.Before(time.Now().UTC()) {
		return nil, errors.New("session expired")
	}

	u, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if u == nil || !u.IsActive {
		return nil, errors.New("user invalid")
	}
	return u, nil
}

func (s *authService) ToPublic(u *models.User) models.UserPublic {
	if u == nil {
		return models.UserPublic{}
	}
	return models.UserPublic{ID: u.ID, Username: u.Username, Email: u.Email}
}

func (s *authService) issueTokens(ctx context.Context, user *models.User) (*models.TokenResponse, error) {
	if user == nil {
		return nil, fmt.Errorf("user is nil")
	}

	refreshToken, err := auth.GenerateRefreshToken()
	if err != nil {
		return nil, err
	}

	sessionID := uuid.New()
	refreshExp := time.Now().UTC().Add(s.refreshTTL)
	hash := auth.HashRefreshToken(refreshToken)
	if err := s.repo.InsertSession(ctx, sessionID, user.ID, hash, refreshExp); err != nil {
		return nil, err
	}

	accessToken, expiresIn, err := auth.IssueAccessToken(user.ID, sessionID, s.jwtSecret, s.accessTTL)
	if err != nil {
		_ = s.repo.DeleteSession(ctx, sessionID)
		return nil, err
	}

	remaining := int64(time.Until(refreshExp).Seconds())
	if remaining < 0 {
		remaining = 0
	}

	return &models.TokenResponse{
		AccessToken:      accessToken,
		ExpiresIn:        expiresIn,
		RefreshExpiresIn: remaining,
		RefreshToken:     refreshToken,
		TokenType:        "Bearer",
	}, nil
}
