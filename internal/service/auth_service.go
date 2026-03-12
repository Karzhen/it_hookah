package service

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/karzhen/restaurant-lk/internal/auth"
	"github.com/karzhen/restaurant-lk/internal/domain"
	"github.com/karzhen/restaurant-lk/internal/dto"
	"github.com/karzhen/restaurant-lk/internal/repository"
	"github.com/karzhen/restaurant-lk/internal/utils/apperror"
	"github.com/karzhen/restaurant-lk/internal/utils/validation"
)

type AuthService struct {
	users         UserRepository
	refreshTokens RefreshTokenRepository
	passwords     auth.PasswordManager
	jwt           *auth.JWTManager
	refresh       *auth.RefreshTokenManager
	logger        *slog.Logger
}

type LoginResult struct {
	AccessToken  string
	RefreshToken string
	ExpiresIn    int64
	User         *domain.User
}

type RefreshResult struct {
	AccessToken  string
	RefreshToken string
	ExpiresIn    int64
}

func NewAuthService(
	users UserRepository,
	refreshTokens RefreshTokenRepository,
	passwords auth.PasswordManager,
	jwt *auth.JWTManager,
	refresh *auth.RefreshTokenManager,
	logger *slog.Logger,
) *AuthService {
	return &AuthService{
		users:         users,
		refreshTokens: refreshTokens,
		passwords:     passwords,
		jwt:           jwt,
		refresh:       refresh,
		logger:        logger,
	}
}

func (s *AuthService) Register(ctx context.Context, req dto.RegisterRequest) (*domain.User, error) {
	req.Email = validation.NormalizeEmail(req.Email)
	req.FirstName = strings.TrimSpace(req.FirstName)
	req.LastName = strings.TrimSpace(req.LastName)

	if !validation.IsValidEmail(req.Email) {
		return nil, apperror.New(apperror.CodeValidationError, "invalid email", http.StatusBadRequest)
	}
	if len(req.Password) < 8 {
		return nil, apperror.New(apperror.CodeValidationError, "password must be at least 8 characters", http.StatusBadRequest)
	}
	if validation.IsBlank(req.FirstName) {
		return nil, apperror.New(apperror.CodeValidationError, "first_name is required", http.StatusBadRequest)
	}
	if validation.IsBlank(req.LastName) {
		return nil, apperror.New(apperror.CodeValidationError, "last_name is required", http.StatusBadRequest)
	}
	if req.Age != nil && *req.Age <= 0 {
		return nil, apperror.New(apperror.CodeValidationError, "age must be positive", http.StatusBadRequest)
	}

	existing, err := s.users.GetByEmail(ctx, req.Email)
	if err == nil && existing != nil {
		return nil, apperror.New(apperror.CodeEmailExists, "email already exists", http.StatusConflict)
	}
	if err != nil && !errors.Is(err, repository.ErrNotFound) {
		return nil, apperror.Wrap(apperror.CodeInternal, "failed to check existing user", http.StatusInternalServerError, err)
	}

	passwordHash, err := s.passwords.HashPassword(req.Password)
	if err != nil {
		return nil, apperror.Wrap(apperror.CodeInternal, "failed to hash password", http.StatusInternalServerError, err)
	}

	now := time.Now().UTC()
	user := &domain.User{
		ID:           uuid.New(),
		Email:        req.Email,
		PasswordHash: passwordHash,
		FirstName:    req.FirstName,
		LastName:     req.LastName,
		MiddleName:   req.MiddleName,
		Phone:        req.Phone,
		Age:          req.Age,
		Role:         "user",
		IsActive:     true,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if err := s.users.Create(ctx, user); err != nil {
		return nil, apperror.Wrap(apperror.CodeInternal, "failed to create user", http.StatusInternalServerError, err)
	}

	return user, nil
}

func (s *AuthService) Login(ctx context.Context, req dto.LoginRequest, userAgent, ip string) (*LoginResult, error) {
	req.Email = validation.NormalizeEmail(req.Email)

	if !validation.IsValidEmail(req.Email) || validation.IsBlank(req.Password) {
		return nil, apperror.New(apperror.CodeInvalidCreds, "invalid email or password", http.StatusUnauthorized)
	}

	user, err := s.users.GetByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, apperror.New(apperror.CodeInvalidCreds, "invalid email or password", http.StatusUnauthorized)
		}
		return nil, apperror.Wrap(apperror.CodeInternal, "failed to fetch user", http.StatusInternalServerError, err)
	}

	if !user.IsActive {
		return nil, apperror.New(apperror.CodeForbidden, "user is inactive", http.StatusForbidden)
	}

	if err := s.passwords.ComparePassword(user.PasswordHash, req.Password); err != nil {
		return nil, apperror.New(apperror.CodeInvalidCreds, "invalid email or password", http.StatusUnauthorized)
	}

	accessToken, _, err := s.jwt.Generate(user.ID, user.Role)
	if err != nil {
		return nil, apperror.Wrap(apperror.CodeInternal, "failed to generate access token", http.StatusInternalServerError, err)
	}

	refreshToken, tokenHash, expiresAt, err := s.refresh.Generate()
	if err != nil {
		return nil, apperror.Wrap(apperror.CodeInternal, "failed to generate refresh token", http.StatusInternalServerError, err)
	}

	now := time.Now().UTC()
	dbToken := &domain.RefreshToken{
		ID:        uuid.New(),
		UserID:    user.ID,
		TokenHash: tokenHash,
		ExpiresAt: expiresAt,
		CreatedAt: now,
	}
	if userAgent != "" {
		dbToken.UserAgent = &userAgent
	}
	if ip != "" {
		dbToken.IP = &ip
	}

	if err := s.refreshTokens.Create(ctx, dbToken); err != nil {
		return nil, apperror.Wrap(apperror.CodeInternal, "failed to persist refresh token", http.StatusInternalServerError, err)
	}

	return &LoginResult{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    s.jwt.TTLSeconds(),
		User:         user,
	}, nil
}

func (s *AuthService) Refresh(ctx context.Context, plainRefreshToken, userAgent, ip string) (*RefreshResult, error) {
	plainRefreshToken = strings.TrimSpace(plainRefreshToken)
	if plainRefreshToken == "" {
		return nil, apperror.New(apperror.CodeInvalidRefresh, "invalid refresh token", http.StatusUnauthorized)
	}

	hash := s.refresh.Hash(plainRefreshToken)
	existing, err := s.refreshTokens.GetByTokenHash(ctx, hash)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, apperror.New(apperror.CodeInvalidRefresh, "invalid refresh token", http.StatusUnauthorized)
		}
		return nil, apperror.Wrap(apperror.CodeInternal, "failed to fetch refresh token", http.StatusInternalServerError, err)
	}

	if existing.RevokedAt != nil || time.Now().After(existing.ExpiresAt) {
		return nil, apperror.New(apperror.CodeInvalidRefresh, "invalid refresh token", http.StatusUnauthorized)
	}

	user, err := s.users.GetByID(ctx, existing.UserID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, apperror.New(apperror.CodeInvalidRefresh, "invalid refresh token", http.StatusUnauthorized)
		}
		return nil, apperror.Wrap(apperror.CodeInternal, "failed to fetch user", http.StatusInternalServerError, err)
	}

	if !user.IsActive {
		return nil, apperror.New(apperror.CodeForbidden, "user is inactive", http.StatusForbidden)
	}

	if err := s.refreshTokens.RevokeByID(ctx, existing.ID); err != nil {
		if !errors.Is(err, repository.ErrNotFound) {
			return nil, apperror.Wrap(apperror.CodeInternal, "failed to revoke token", http.StatusInternalServerError, err)
		}
	}

	newRefreshToken, newHash, expiresAt, err := s.refresh.Generate()
	if err != nil {
		return nil, apperror.Wrap(apperror.CodeInternal, "failed to generate refresh token", http.StatusInternalServerError, err)
	}

	now := time.Now().UTC()
	dbToken := &domain.RefreshToken{
		ID:        uuid.New(),
		UserID:    user.ID,
		TokenHash: newHash,
		ExpiresAt: expiresAt,
		CreatedAt: now,
	}
	if userAgent != "" {
		dbToken.UserAgent = &userAgent
	}
	if ip != "" {
		dbToken.IP = &ip
	}

	if err := s.refreshTokens.Create(ctx, dbToken); err != nil {
		return nil, apperror.Wrap(apperror.CodeInternal, "failed to persist refresh token", http.StatusInternalServerError, err)
	}

	accessToken, _, err := s.jwt.Generate(user.ID, user.Role)
	if err != nil {
		return nil, apperror.Wrap(apperror.CodeInternal, "failed to generate access token", http.StatusInternalServerError, err)
	}

	return &RefreshResult{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
		ExpiresIn:    s.jwt.TTLSeconds(),
	}, nil
}

func (s *AuthService) Logout(ctx context.Context, plainRefreshToken string) error {
	plainRefreshToken = strings.TrimSpace(plainRefreshToken)
	if plainRefreshToken == "" {
		return apperror.New(apperror.CodeInvalidRefresh, "invalid refresh token", http.StatusUnauthorized)
	}

	hash := s.refresh.Hash(plainRefreshToken)
	err := s.refreshTokens.RevokeByTokenHash(ctx, hash)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return apperror.New(apperror.CodeInvalidRefresh, "invalid refresh token", http.StatusUnauthorized)
		}
		return apperror.Wrap(apperror.CodeInternal, "failed to revoke token", http.StatusInternalServerError, err)
	}

	return nil
}

func (s *AuthService) ChangePassword(ctx context.Context, userID uuid.UUID, oldPassword, newPassword string) error {
	if len(newPassword) < 8 {
		return apperror.New(apperror.CodeValidationError, "new password must be at least 8 characters", http.StatusBadRequest)
	}

	user, err := s.users.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return apperror.New(apperror.CodeUserNotFound, "user not found", http.StatusNotFound)
		}
		return apperror.Wrap(apperror.CodeInternal, "failed to fetch user", http.StatusInternalServerError, err)
	}

	if err := s.passwords.ComparePassword(user.PasswordHash, oldPassword); err != nil {
		return apperror.New(apperror.CodeInvalidCreds, "old password is incorrect", http.StatusUnauthorized)
	}

	newHash, err := s.passwords.HashPassword(newPassword)
	if err != nil {
		return apperror.Wrap(apperror.CodeInternal, "failed to hash password", http.StatusInternalServerError, err)
	}

	if err := s.users.UpdatePassword(ctx, userID, newHash); err != nil {
		return apperror.Wrap(apperror.CodeInternal, "failed to update password", http.StatusInternalServerError, err)
	}

	return nil
}
