package service

import (
	"context"

	"github.com/google/uuid"

	"github.com/karzhen/restaurant-lk/internal/domain"
)

type TagStore interface {
	CreateTag(ctx context.Context, tag *domain.Tag) error
	UpdateTag(ctx context.Context, tag *domain.Tag) error
	DeactivateTag(ctx context.Context, tagID uuid.UUID) error
	GetTagByID(ctx context.Context, tagID uuid.UUID, adminView bool) (*domain.Tag, error)
	GetTagByCode(ctx context.Context, code string, adminView bool) (*domain.Tag, error)
	GetTagByName(ctx context.Context, name string, adminView bool) (*domain.Tag, error)
	GetTagsByIDs(ctx context.Context, tagIDs []uuid.UUID, activeOnly bool) ([]domain.Tag, error)
	ListTags(ctx context.Context, activeOnly bool, limit, offset int) ([]domain.Tag, error)

	SetProductTags(ctx context.Context, productID uuid.UUID, tagIDs []uuid.UUID) error
	ListProductTags(ctx context.Context, productID uuid.UUID, activeOnly bool) ([]domain.Tag, error)

	SetMixTags(ctx context.Context, mixID uuid.UUID, tagIDs []uuid.UUID) error
	ListMixTags(ctx context.Context, mixID uuid.UUID, activeOnly bool) ([]domain.Tag, error)
}

type TagService interface {
	ListPublicTags(ctx context.Context, limit, offset int) ([]domain.Tag, error)
	ListAdminTags(ctx context.Context, limit, offset int) ([]domain.Tag, error)
	GetPublicTagByID(ctx context.Context, tagID uuid.UUID) (*domain.Tag, error)
	GetAdminTagByID(ctx context.Context, tagID uuid.UUID) (*domain.Tag, error)
	CreateTag(ctx context.Context, tag *domain.Tag) (*domain.Tag, error)
	UpdateTag(ctx context.Context, tagID uuid.UUID, tag *domain.Tag) (*domain.Tag, error)
	DeactivateTag(ctx context.Context, tagID uuid.UUID) error

	SetProductTags(ctx context.Context, productID uuid.UUID, tagIDs []uuid.UUID) error
	ListProductTags(ctx context.Context, productID uuid.UUID, adminView bool) ([]domain.Tag, error)

	SetMixTags(ctx context.Context, mixID uuid.UUID, tagIDs []uuid.UUID) error
	ListMixTags(ctx context.Context, mixID uuid.UUID, adminView bool) ([]domain.Tag, error)
}
