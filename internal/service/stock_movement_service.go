package service

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/karzhen/restaurant-lk/internal/domain"
	"github.com/karzhen/restaurant-lk/internal/repository"
	"github.com/karzhen/restaurant-lk/internal/utils/apperror"
)

type StockMovementManager struct {
	store    StockMovementStore
	products ProductReader
}

func NewStockMovementManager(store StockMovementStore, products ProductReader) *StockMovementManager {
	return &StockMovementManager{
		store:    store,
		products: products,
	}
}

func (s *StockMovementManager) RecordMovement(ctx context.Context, movement *domain.StockMovement) error {
	if movement == nil {
		return apperror.New(apperror.CodeValidationError, "stock movement is required", http.StatusBadRequest)
	}
	if movement.ID == uuid.Nil {
		movement.ID = uuid.New()
	}
	if movement.CreatedAt.IsZero() {
		movement.CreatedAt = time.Now().UTC()
	}
	if movement.ProductID == uuid.Nil {
		return apperror.New(apperror.CodeValidationError, "product_id is required", http.StatusBadRequest)
	}
	if movement.Quantity <= 0 {
		return apperror.New(apperror.CodeValidationError, "quantity must be positive", http.StatusBadRequest)
	}

	if err := s.store.CreateStockMovement(ctx, movement); err != nil {
		return apperror.Wrap(apperror.CodeInternal, "failed to record stock movement", http.StatusInternalServerError, err)
	}
	return nil
}

func (s *StockMovementManager) ListMovements(ctx context.Context, filter domain.StockMovementFilter) ([]domain.StockMovement, error) {
	filter.Limit = normalizeStockMovementLimit(filter.Limit)
	filter.Offset = normalizeStockMovementOffset(filter.Offset)

	if filter.Operation != nil {
		op := domain.StockMovementOperation(strings.ToLower(strings.TrimSpace(string(*filter.Operation))))
		if !isValidStockMovementOperation(op) {
			return nil, apperror.New(apperror.CodeValidationError, "invalid operation", http.StatusBadRequest)
		}
		filter.Operation = &op
	}

	if filter.ProductID != nil {
		if err := s.ensureProductExists(ctx, *filter.ProductID); err != nil {
			return nil, err
		}
	}

	items, err := s.store.ListStockMovements(ctx, filter)
	if err != nil {
		return nil, apperror.Wrap(apperror.CodeInternal, "failed to list stock movements", http.StatusInternalServerError, err)
	}
	return items, nil
}

func (s *StockMovementManager) ListProductMovements(
	ctx context.Context,
	productID uuid.UUID,
	operation *domain.StockMovementOperation,
	limit,
	offset int,
) ([]domain.StockMovement, error) {
	if err := s.ensureProductExists(ctx, productID); err != nil {
		return nil, err
	}

	limit = normalizeStockMovementLimit(limit)
	offset = normalizeStockMovementOffset(offset)

	if operation != nil {
		op := domain.StockMovementOperation(strings.ToLower(strings.TrimSpace(string(*operation))))
		if !isValidStockMovementOperation(op) {
			return nil, apperror.New(apperror.CodeValidationError, "invalid operation", http.StatusBadRequest)
		}
		operation = &op
	}

	items, err := s.store.ListStockMovementsByProductID(ctx, productID, operation, limit, offset)
	if err != nil {
		return nil, apperror.Wrap(apperror.CodeInternal, "failed to list product stock movements", http.StatusInternalServerError, err)
	}
	return items, nil
}

func (s *StockMovementManager) ensureProductExists(ctx context.Context, productID uuid.UUID) error {
	if productID == uuid.Nil {
		return apperror.New(apperror.CodeValidationError, "invalid product_id", http.StatusBadRequest)
	}
	if s.products == nil {
		return apperror.New(apperror.CodeInternal, "product reader is not configured", http.StatusInternalServerError)
	}

	_, err := s.products.GetProductByID(ctx, productID, true)
	if err == nil {
		return nil
	}
	if errors.Is(err, repository.ErrNotFound) {
		return apperror.New(apperror.CodeProductNotFound, "product not found", http.StatusNotFound)
	}
	return apperror.Wrap(apperror.CodeInternal, "failed to fetch product", http.StatusInternalServerError, err)
}

func normalizeStockMovementLimit(limit int) int {
	if limit <= 0 {
		return 20
	}
	if limit > 100 {
		return 100
	}
	return limit
}

func normalizeStockMovementOffset(offset int) int {
	if offset < 0 {
		return 0
	}
	return offset
}

func isValidStockMovementOperation(operation domain.StockMovementOperation) bool {
	switch operation {
	case domain.StockMovementOperationSet,
		domain.StockMovementOperationIncrement,
		domain.StockMovementOperationDecrement,
		domain.StockMovementOperationCheckoutDecrement:
		return true
	default:
		return false
	}
}
