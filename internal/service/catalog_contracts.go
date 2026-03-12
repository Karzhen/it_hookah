package service

import (
	"context"

	"github.com/google/uuid"

	"github.com/karzhen/restaurant-lk/internal/domain"
)

type CatalogRepository interface {
	ListCategories(ctx context.Context, activeOnly bool) ([]domain.ProductCategory, error)
	GetCategoryByID(ctx context.Context, id uuid.UUID) (*domain.ProductCategory, error)
	CreateCategory(ctx context.Context, category *domain.ProductCategory) error
	UpdateCategory(ctx context.Context, category *domain.ProductCategory) error
	DeactivateCategory(ctx context.Context, id uuid.UUID) error

	ListFlavors(ctx context.Context, activeOnly bool) ([]domain.TobaccoFlavor, error)
	GetFlavorByID(ctx context.Context, id uuid.UUID) (*domain.TobaccoFlavor, error)
	GetFlavorsByIDs(ctx context.Context, ids []uuid.UUID) ([]domain.TobaccoFlavor, error)
	CreateFlavor(ctx context.Context, flavor *domain.TobaccoFlavor) error
	UpdateFlavor(ctx context.Context, flavor *domain.TobaccoFlavor) error
	DeactivateFlavor(ctx context.Context, id uuid.UUID) error

	ListStrengths(ctx context.Context, activeOnly bool) ([]domain.TobaccoStrength, error)
	GetStrengthByID(ctx context.Context, id uuid.UUID) (*domain.TobaccoStrength, error)
	CreateStrength(ctx context.Context, strength *domain.TobaccoStrength) error
	UpdateStrength(ctx context.Context, strength *domain.TobaccoStrength) error
	DeactivateStrength(ctx context.Context, id uuid.UUID) error

	ListProducts(ctx context.Context, filter domain.ProductFilter) ([]domain.Product, error)
	GetProductByID(ctx context.Context, id uuid.UUID, adminView bool) (*domain.Product, error)
	CreateProduct(ctx context.Context, input domain.ProductUpsert) error
	UpdateProduct(ctx context.Context, input domain.ProductUpsert) error
	DeactivateProduct(ctx context.Context, id uuid.UUID) error
	ApplyProductStockOperation(
		ctx context.Context,
		id uuid.UUID,
		operation domain.StockMovementOperation,
		quantity int,
		reason *string,
		createdByUserID *uuid.UUID,
	) error
}
