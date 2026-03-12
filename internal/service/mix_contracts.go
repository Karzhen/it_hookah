package service

import (
	"context"

	"github.com/google/uuid"

	"github.com/karzhen/restaurant-lk/internal/domain"
)

type MixStore interface {
	CreateMix(ctx context.Context, mix *domain.Mix) error
	CreateMixItems(ctx context.Context, items []domain.MixItem) error
	GetMixByID(ctx context.Context, mixID uuid.UUID, adminView bool) (*domain.Mix, error)
	ListMixes(ctx context.Context, activeOnly bool, limit, offset int) ([]domain.Mix, error)
	ListMixItemsByMixID(ctx context.Context, mixID uuid.UUID) ([]domain.MixItem, error)
	UpdateMix(ctx context.Context, mix *domain.Mix) error
	ReplaceMixItems(ctx context.Context, mixID uuid.UUID, items []domain.MixItem) error
	DeactivateMix(ctx context.Context, mixID uuid.UUID) error
}

type MixService interface {
	GetPublicMixByID(ctx context.Context, mixID uuid.UUID) (*domain.Mix, error)
	ListPublicMixes(ctx context.Context, limit, offset int) ([]domain.Mix, error)
	GetAdminMixByID(ctx context.Context, mixID uuid.UUID) (*domain.Mix, error)
	ListAdminMixes(ctx context.Context, limit, offset int) ([]domain.Mix, error)
	CreateMix(ctx context.Context, mix *domain.Mix, items []domain.MixItem) (*domain.Mix, error)
	UpdateMix(ctx context.Context, mixID uuid.UUID, mix *domain.Mix, items []domain.MixItem) (*domain.Mix, error)
	DeactivateMix(ctx context.Context, mixID uuid.UUID) error
}
