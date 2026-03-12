package service

import (
	"context"
	"errors"
	"net/http"

	"github.com/google/uuid"

	"github.com/karzhen/restaurant-lk/internal/domain"
	"github.com/karzhen/restaurant-lk/internal/repository"
	"github.com/karzhen/restaurant-lk/internal/utils/apperror"
)

type CartManager struct {
	store    CartStore
	products ProductReader
}

func NewCartManager(store CartStore, products ProductReader) *CartManager {
	return &CartManager{
		store:    store,
		products: products,
	}
}

func (s *CartManager) GetCart(ctx context.Context, userID uuid.UUID) (*domain.Cart, []domain.CartItem, error) {
	cart, err := s.store.GetOrCreateCartByUserID(ctx, userID)
	if err != nil {
		return nil, nil, apperror.Wrap(apperror.CodeInternal, "failed to fetch cart", http.StatusInternalServerError, err)
	}

	items, err := s.store.ListCartItems(ctx, cart.ID)
	if err != nil {
		return nil, nil, apperror.Wrap(apperror.CodeInternal, "failed to fetch cart items", http.StatusInternalServerError, err)
	}

	return cart, items, nil
}

func (s *CartManager) AddToCart(ctx context.Context, userID uuid.UUID, productID uuid.UUID, quantity int) (*domain.Cart, []domain.CartItem, error) {
	if quantity <= 0 {
		return nil, nil, apperror.New(apperror.CodeValidationError, "quantity must be positive", http.StatusBadRequest)
	}

	product, err := s.products.GetProductByID(ctx, productID, true)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, nil, apperror.New(apperror.CodeProductNotFound, "product not found", http.StatusNotFound)
		}
		return nil, nil, apperror.Wrap(apperror.CodeInternal, "failed to fetch product", http.StatusInternalServerError, err)
	}

	if !product.IsActive {
		return nil, nil, apperror.New(apperror.CodeValidationError, "inactive product cannot be added to cart", http.StatusBadRequest)
	}
	if product.StockQuantity <= 0 {
		return nil, nil, apperror.New(apperror.CodeInsufficientStock, "insufficient stock", http.StatusBadRequest)
	}

	if err := s.store.AddOrIncrementItem(ctx, userID, productID, quantity, product.StockQuantity); err != nil {
		if errors.Is(err, repository.ErrInsufficientStock) {
			return nil, nil, apperror.New(apperror.CodeInsufficientStock, "insufficient stock", http.StatusBadRequest)
		}
		return nil, nil, apperror.Wrap(apperror.CodeInternal, "failed to add product to cart", http.StatusInternalServerError, err)
	}

	cart, items, err := s.getOrCreateCartWithItems(ctx, userID)
	if err != nil {
		return nil, nil, err
	}

	return cart, items, nil
}

func (s *CartManager) UpdateCartItemQuantity(ctx context.Context, userID uuid.UUID, productID uuid.UUID, quantity int) (*domain.Cart, []domain.CartItem, error) {
	if quantity <= 0 {
		return nil, nil, apperror.New(apperror.CodeValidationError, "quantity must be positive", http.StatusBadRequest)
	}

	product, err := s.products.GetProductByID(ctx, productID, true)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, nil, apperror.New(apperror.CodeProductNotFound, "product not found", http.StatusNotFound)
		}
		return nil, nil, apperror.Wrap(apperror.CodeInternal, "failed to fetch product", http.StatusInternalServerError, err)
	}

	if !product.IsActive {
		return nil, nil, apperror.New(apperror.CodeValidationError, "inactive product cannot be updated in cart", http.StatusBadRequest)
	}
	if product.StockQuantity <= 0 || quantity > product.StockQuantity {
		return nil, nil, apperror.New(apperror.CodeInsufficientStock, "insufficient stock", http.StatusBadRequest)
	}

	if err := s.store.SetItemQuantity(ctx, userID, productID, quantity, product.StockQuantity); err != nil {
		if errors.Is(err, repository.ErrInsufficientStock) {
			return nil, nil, apperror.New(apperror.CodeInsufficientStock, "insufficient stock", http.StatusBadRequest)
		}
		if errors.Is(err, repository.ErrNotFound) {
			return nil, nil, apperror.New(apperror.CodeValidationError, "cart item not found", http.StatusNotFound)
		}
		return nil, nil, apperror.Wrap(apperror.CodeInternal, "failed to update cart item", http.StatusInternalServerError, err)
	}

	cart, items, err := s.getOrCreateCartWithItems(ctx, userID)
	if err != nil {
		return nil, nil, err
	}

	return cart, items, nil
}

func (s *CartManager) RemoveFromCart(ctx context.Context, userID uuid.UUID, productID uuid.UUID) (*domain.Cart, []domain.CartItem, error) {
	if err := s.store.RemoveItem(ctx, userID, productID); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, nil, apperror.New(apperror.CodeValidationError, "cart item not found", http.StatusNotFound)
		}
		return nil, nil, apperror.Wrap(apperror.CodeInternal, "failed to remove cart item", http.StatusInternalServerError, err)
	}

	cart, items, err := s.getOrCreateCartWithItems(ctx, userID)
	if err != nil {
		return nil, nil, err
	}

	return cart, items, nil
}

func (s *CartManager) ClearCart(ctx context.Context, userID uuid.UUID) error {
	if err := s.store.ClearCart(ctx, userID); err != nil {
		return apperror.Wrap(apperror.CodeInternal, "failed to clear cart", http.StatusInternalServerError, err)
	}

	return nil
}

func (s *CartManager) getOrCreateCartWithItems(ctx context.Context, userID uuid.UUID) (*domain.Cart, []domain.CartItem, error) {
	cart, err := s.store.GetOrCreateCartByUserID(ctx, userID)
	if err != nil {
		return nil, nil, apperror.Wrap(apperror.CodeInternal, "failed to fetch cart", http.StatusInternalServerError, err)
	}

	items, err := s.store.ListCartItems(ctx, cart.ID)
	if err != nil {
		return nil, nil, apperror.Wrap(apperror.CodeInternal, "failed to fetch cart items", http.StatusInternalServerError, err)
	}

	return cart, items, nil
}
