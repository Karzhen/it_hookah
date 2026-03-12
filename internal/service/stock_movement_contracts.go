package service

import (
	"context"

	"github.com/google/uuid"

	"github.com/karzhen/restaurant-lk/internal/domain"
)

type StockMovementStore interface {
	CreateStockMovement(ctx context.Context, movement *domain.StockMovement) error
	ListStockMovements(ctx context.Context, filter domain.StockMovementFilter) ([]domain.StockMovement, error)
	ListStockMovementsByProductID(
		ctx context.Context,
		productID uuid.UUID,
		operation *domain.StockMovementOperation,
		limit,
		offset int,
	) ([]domain.StockMovement, error)
}

type StockMovementService interface {
	RecordMovement(ctx context.Context, movement *domain.StockMovement) error
	ListMovements(ctx context.Context, filter domain.StockMovementFilter) ([]domain.StockMovement, error)
	ListProductMovements(
		ctx context.Context,
		productID uuid.UUID,
		operation *domain.StockMovementOperation,
		limit,
		offset int,
	) ([]domain.StockMovement, error)
}
