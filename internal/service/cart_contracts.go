package service

import (
	"context"

	"github.com/google/uuid"

	"github.com/karzhen/restaurant-lk/internal/domain"
)

type CartStore interface {
	GetCartByUserID(ctx context.Context, userID uuid.UUID) (*domain.Cart, error)
	GetOrCreateCartByUserID(ctx context.Context, userID uuid.UUID) (*domain.Cart, error)
	ListCartItems(ctx context.Context, cartID uuid.UUID) ([]domain.CartItem, error)
	AddOrIncrementItem(ctx context.Context, userID uuid.UUID, productID uuid.UUID, quantity int, maxStock int) error
	SetItemQuantity(ctx context.Context, userID uuid.UUID, productID uuid.UUID, quantity int, maxStock int) error
	RemoveItem(ctx context.Context, userID uuid.UUID, productID uuid.UUID) error
	ClearCart(ctx context.Context, userID uuid.UUID) error
}

type ProductReader interface {
	GetProductByID(ctx context.Context, id uuid.UUID, adminView bool) (*domain.Product, error)
}

type CartService interface {
	GetCart(ctx context.Context, userID uuid.UUID) (*domain.Cart, []domain.CartItem, error)
	AddToCart(ctx context.Context, userID uuid.UUID, productID uuid.UUID, quantity int) (*domain.Cart, []domain.CartItem, error)
	UpdateCartItemQuantity(ctx context.Context, userID uuid.UUID, productID uuid.UUID, quantity int) (*domain.Cart, []domain.CartItem, error)
	RemoveFromCart(ctx context.Context, userID uuid.UUID, productID uuid.UUID) (*domain.Cart, []domain.CartItem, error)
	ClearCart(ctx context.Context, userID uuid.UUID) error
}
