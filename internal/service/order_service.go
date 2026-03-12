package service

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/google/uuid"

	"github.com/karzhen/restaurant-lk/internal/domain"
	"github.com/karzhen/restaurant-lk/internal/repository"
	"github.com/karzhen/restaurant-lk/internal/utils/apperror"
)

type OrderManager struct {
	store OrderStore
}

func NewOrderManager(store OrderStore) *OrderManager {
	return &OrderManager{store: store}
}

func (s *OrderManager) Checkout(ctx context.Context, userID uuid.UUID) (*domain.Order, error) {
	order, err := s.store.CheckoutFromCart(ctx, userID)
	if err != nil {
		if errors.Is(err, repository.ErrEmptyCart) {
			return nil, apperror.New(apperror.CodeEmptyCart, "cart is empty", http.StatusBadRequest)
		}
		if errors.Is(err, repository.ErrInsufficientStock) {
			return nil, apperror.New(apperror.CodeInsufficientStock, "insufficient stock", http.StatusBadRequest)
		}
		if errors.Is(err, repository.ErrForbidden) {
			return nil, apperror.New(apperror.CodeForbidden, "checkout is forbidden", http.StatusForbidden)
		}
		return nil, apperror.Wrap(apperror.CodeInternal, "failed to checkout cart", http.StatusInternalServerError, err)
	}

	return order, nil
}

func (s *OrderManager) GetOrderByID(ctx context.Context, userID uuid.UUID, orderID uuid.UUID) (*domain.Order, error) {
	order, err := s.store.GetOrderByID(ctx, orderID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, apperror.New(apperror.CodeOrderNotFound, "order not found", http.StatusNotFound)
		}
		return nil, apperror.Wrap(apperror.CodeInternal, "failed to fetch order", http.StatusInternalServerError, err)
	}

	if order.UserID != userID {
		return nil, apperror.New(apperror.CodeForbidden, "insufficient permissions", http.StatusForbidden)
	}

	items, err := s.store.ListOrderItemsByOrderID(ctx, order.ID)
	if err != nil {
		return nil, apperror.Wrap(apperror.CodeInternal, "failed to fetch order items", http.StatusInternalServerError, err)
	}
	order.Items = items

	return order, nil
}

func (s *OrderManager) ListUserOrders(ctx context.Context, userID uuid.UUID) ([]domain.Order, error) {
	orders, err := s.store.ListOrdersByUserID(ctx, userID)
	if err != nil {
		return nil, apperror.Wrap(apperror.CodeInternal, "failed to list user orders", http.StatusInternalServerError, err)
	}

	if err := s.attachItems(ctx, orders); err != nil {
		return nil, err
	}

	return orders, nil
}

func (s *OrderManager) ListAllOrders(ctx context.Context, actorRole string, limit, offset int) ([]domain.Order, error) {
	if !isAdminRole(actorRole) {
		return nil, apperror.New(apperror.CodeForbidden, "insufficient permissions", http.StatusForbidden)
	}

	limit = normalizeLimit(limit)
	offset = normalizeOffset(offset)

	orders, err := s.store.ListAllOrders(ctx, limit, offset)
	if err != nil {
		return nil, apperror.Wrap(apperror.CodeInternal, "failed to list orders", http.StatusInternalServerError, err)
	}

	if err := s.attachItems(ctx, orders); err != nil {
		return nil, err
	}

	return orders, nil
}

func (s *OrderManager) ChangeOrderStatus(
	ctx context.Context,
	actorRole string,
	orderID uuid.UUID,
	status domain.OrderStatus,
) (*domain.Order, error) {
	if !isAdminRole(actorRole) {
		return nil, apperror.New(apperror.CodeForbidden, "insufficient permissions", http.StatusForbidden)
	}

	status = domain.OrderStatus(strings.ToLower(strings.TrimSpace(string(status))))
	if !isValidOrderStatus(status) {
		return nil, apperror.New(apperror.CodeInvalidOrderStatus, "invalid order status", http.StatusBadRequest)
	}

	if err := s.store.UpdateOrderStatus(ctx, orderID, status); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, apperror.New(apperror.CodeOrderNotFound, "order not found", http.StatusNotFound)
		}
		if errors.Is(err, repository.ErrInvalidOrderStatus) {
			return nil, apperror.New(apperror.CodeInvalidOrderStatus, "invalid order status", http.StatusBadRequest)
		}
		return nil, apperror.Wrap(apperror.CodeInternal, "failed to update order status", http.StatusInternalServerError, err)
	}

	order, err := s.store.GetOrderByID(ctx, orderID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, apperror.New(apperror.CodeOrderNotFound, "order not found", http.StatusNotFound)
		}
		return nil, apperror.Wrap(apperror.CodeInternal, "failed to fetch updated order", http.StatusInternalServerError, err)
	}

	items, err := s.store.ListOrderItemsByOrderID(ctx, order.ID)
	if err != nil {
		return nil, apperror.Wrap(apperror.CodeInternal, "failed to fetch order items", http.StatusInternalServerError, err)
	}
	order.Items = items

	return order, nil
}

func (s *OrderManager) attachItems(ctx context.Context, orders []domain.Order) error {
	for i := range orders {
		items, err := s.store.ListOrderItemsByOrderID(ctx, orders[i].ID)
		if err != nil {
			return apperror.Wrap(apperror.CodeInternal, "failed to fetch order items", http.StatusInternalServerError, err)
		}
		orders[i].Items = items
	}
	return nil
}

func isAdminRole(role string) bool {
	return strings.EqualFold(strings.TrimSpace(role), "admin")
}

func isValidOrderStatus(status domain.OrderStatus) bool {
	switch status {
	case domain.OrderStatusPending,
		domain.OrderStatusConfirmed,
		domain.OrderStatusPreparing,
		domain.OrderStatusCompleted,
		domain.OrderStatusCancelled:
		return true
	default:
		return false
	}
}

func normalizeLimit(limit int) int {
	if limit <= 0 {
		return 20
	}
	if limit > 100 {
		return 100
	}
	return limit
}

func normalizeOffset(offset int) int {
	if offset < 0 {
		return 0
	}
	return offset
}
