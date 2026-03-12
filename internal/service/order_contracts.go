package service

import (
	"context"

	"github.com/google/uuid"

	"github.com/karzhen/restaurant-lk/internal/domain"
)

type OrderStore interface {
	CreateOrder(ctx context.Context, order *domain.Order) error
	CreateOrderItems(ctx context.Context, items []domain.OrderItem) error
	CheckoutFromCart(ctx context.Context, userID uuid.UUID) (*domain.Order, error)
	GetOrderByID(ctx context.Context, orderID uuid.UUID) (*domain.Order, error)
	ListOrdersByUserID(ctx context.Context, userID uuid.UUID) ([]domain.Order, error)
	ListAllOrders(ctx context.Context, limit, offset int) ([]domain.Order, error)
	ListOrderItemsByOrderID(ctx context.Context, orderID uuid.UUID) ([]domain.OrderItem, error)
	UpdateOrderStatus(ctx context.Context, orderID uuid.UUID, status domain.OrderStatus) error
	UpdateOrderTotalAmount(ctx context.Context, orderID uuid.UUID, totalAmount string) error
}

type OrderService interface {
	Checkout(ctx context.Context, userID uuid.UUID) (*domain.Order, error)
	GetOrderByID(ctx context.Context, userID uuid.UUID, orderID uuid.UUID) (*domain.Order, error)
	ListUserOrders(ctx context.Context, userID uuid.UUID) ([]domain.Order, error)
	ListAllOrders(ctx context.Context, actorRole string, limit, offset int) ([]domain.Order, error)
	ChangeOrderStatus(ctx context.Context, actorRole string, orderID uuid.UUID, status domain.OrderStatus) (*domain.Order, error)
}
