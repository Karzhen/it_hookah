package service

import (
	"context"

	"github.com/google/uuid"

	"github.com/karzhen/restaurant-lk/internal/domain"
)

type UserRepository interface {
	Create(ctx context.Context, user *domain.User) error
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
	GetByID(ctx context.Context, userID uuid.UUID) (*domain.User, error)
	UpdateProfile(ctx context.Context, userID uuid.UUID, input domain.UpdateUserProfile) (*domain.User, error)
	UpdatePassword(ctx context.Context, userID uuid.UUID, passwordHash string) error
}

type RefreshTokenRepository interface {
	Create(ctx context.Context, token *domain.RefreshToken) error
	GetByTokenHash(ctx context.Context, tokenHash string) (*domain.RefreshToken, error)
	RevokeByID(ctx context.Context, tokenID uuid.UUID) error
	RevokeByTokenHash(ctx context.Context, tokenHash string) error
}
