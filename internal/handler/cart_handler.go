package handler

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/karzhen/restaurant-lk/internal/domain"
	"github.com/karzhen/restaurant-lk/internal/dto"
	"github.com/karzhen/restaurant-lk/internal/repository"
	"github.com/karzhen/restaurant-lk/internal/service"
	"github.com/karzhen/restaurant-lk/internal/utils/apperror"
	"github.com/karzhen/restaurant-lk/internal/utils/ctxkeys"
	"github.com/karzhen/restaurant-lk/internal/utils/httpx"
)

type cartProductReader interface {
	GetProductByID(ctx context.Context, id uuid.UUID, adminView bool) (*domain.Product, error)
}

type CartHandler struct {
	cartService *service.CartManager
	products    cartProductReader
	logger      *slog.Logger
}

func NewCartHandler(cartService *service.CartManager, products cartProductReader, logger *slog.Logger) *CartHandler {
	return &CartHandler{
		cartService: cartService,
		products:    products,
		logger:      logger,
	}
}

func (h *CartHandler) GetCart(w http.ResponseWriter, r *http.Request) {
	userID, ok := ctxkeys.UserID(r.Context())
	if !ok {
		httpx.WriteError(h.logger, w, apperror.New(apperror.CodeUnauthorized, "unauthorized", http.StatusUnauthorized))
		return
	}

	cart, items, err := h.cartService.GetCart(r.Context(), userID)
	if err != nil {
		httpx.WriteError(h.logger, w, err)
		return
	}

	if cart == nil {
		empty := dto.CartResponse{
			ID:            "",
			UserID:        userID.String(),
			Items:         []dto.CartItemResponse{},
			TotalQuantity: 0,
			TotalAmount:   "0.00",
		}
		httpx.WriteJSON(w, http.StatusOK, empty)
		return
	}

	resp, err := h.buildCartResponse(r.Context(), cart, items)
	if err != nil {
		httpx.WriteError(h.logger, w, err)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, resp)
}

func (h *CartHandler) AddItem(w http.ResponseWriter, r *http.Request) {
	userID, ok := ctxkeys.UserID(r.Context())
	if !ok {
		httpx.WriteError(h.logger, w, apperror.New(apperror.CodeUnauthorized, "unauthorized", http.StatusUnauthorized))
		return
	}

	var req dto.AddCartItemRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.WriteError(h.logger, w, err)
		return
	}

	productID, err := uuid.Parse(strings.TrimSpace(req.ProductID))
	if err != nil {
		httpx.WriteError(h.logger, w, apperror.New(apperror.CodeValidationError, "invalid product_id", http.StatusBadRequest))
		return
	}

	cart, items, err := h.cartService.AddToCart(r.Context(), userID, productID, req.Quantity)
	if err != nil {
		httpx.WriteError(h.logger, w, normalizeCartError(err))
		return
	}

	resp, err := h.buildCartResponse(r.Context(), cart, items)
	if err != nil {
		httpx.WriteError(h.logger, w, err)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, resp)
}

func (h *CartHandler) UpdateItemQuantity(w http.ResponseWriter, r *http.Request) {
	userID, ok := ctxkeys.UserID(r.Context())
	if !ok {
		httpx.WriteError(h.logger, w, apperror.New(apperror.CodeUnauthorized, "unauthorized", http.StatusUnauthorized))
		return
	}

	itemID, err := uuid.Parse(strings.TrimSpace(chi.URLParam(r, "id")))
	if err != nil {
		httpx.WriteError(h.logger, w, apperror.New(apperror.CodeValidationError, "invalid cart item id", http.StatusBadRequest))
		return
	}

	var req dto.UpdateCartItemRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.WriteError(h.logger, w, err)
		return
	}

	_, items, err := h.cartService.GetCart(r.Context(), userID)
	if err != nil {
		httpx.WriteError(h.logger, w, err)
		return
	}

	productID, found := findProductIDByCartItemID(items, itemID)
	if !found {
		httpx.WriteError(h.logger, w, apperror.New(apperror.CodeCartItemNotFound, "cart item not found", http.StatusNotFound))
		return
	}

	cart, updatedItems, err := h.cartService.UpdateCartItemQuantity(r.Context(), userID, productID, req.Quantity)
	if err != nil {
		httpx.WriteError(h.logger, w, normalizeCartError(err))
		return
	}

	resp, err := h.buildCartResponse(r.Context(), cart, updatedItems)
	if err != nil {
		httpx.WriteError(h.logger, w, err)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, resp)
}

func (h *CartHandler) RemoveItem(w http.ResponseWriter, r *http.Request) {
	userID, ok := ctxkeys.UserID(r.Context())
	if !ok {
		httpx.WriteError(h.logger, w, apperror.New(apperror.CodeUnauthorized, "unauthorized", http.StatusUnauthorized))
		return
	}

	itemID, err := uuid.Parse(strings.TrimSpace(chi.URLParam(r, "id")))
	if err != nil {
		httpx.WriteError(h.logger, w, apperror.New(apperror.CodeValidationError, "invalid cart item id", http.StatusBadRequest))
		return
	}

	_, items, err := h.cartService.GetCart(r.Context(), userID)
	if err != nil {
		httpx.WriteError(h.logger, w, err)
		return
	}

	productID, found := findProductIDByCartItemID(items, itemID)
	if !found {
		httpx.WriteError(h.logger, w, apperror.New(apperror.CodeCartItemNotFound, "cart item not found", http.StatusNotFound))
		return
	}

	cart, updatedItems, err := h.cartService.RemoveFromCart(r.Context(), userID, productID)
	if err != nil {
		httpx.WriteError(h.logger, w, normalizeCartError(err))
		return
	}

	resp, err := h.buildCartResponse(r.Context(), cart, updatedItems)
	if err != nil {
		httpx.WriteError(h.logger, w, err)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, resp)
}

func (h *CartHandler) ClearCart(w http.ResponseWriter, r *http.Request) {
	userID, ok := ctxkeys.UserID(r.Context())
	if !ok {
		httpx.WriteError(h.logger, w, apperror.New(apperror.CodeUnauthorized, "unauthorized", http.StatusUnauthorized))
		return
	}

	if err := h.cartService.ClearCart(r.Context(), userID); err != nil {
		httpx.WriteError(h.logger, w, err)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, dto.MessageResponse{Message: "cart cleared"})
}

func (h *CartHandler) buildCartResponse(ctx context.Context, cart *domain.Cart, items []domain.CartItem) (dto.CartResponse, error) {
	result := dto.CartResponse{
		ID:            cart.ID.String(),
		UserID:        cart.UserID.String(),
		Items:         make([]dto.CartItemResponse, 0, len(items)),
		TotalQuantity: 0,
		TotalAmount:   "0.00",
	}

	totalAmountCents := int64(0)

	for _, item := range items {
		product, err := h.products.GetProductByID(ctx, item.ProductID, true)
		if err != nil {
			return dto.CartResponse{}, normalizeCartError(err)
		}

		priceCents, err := parseCents(product.Price)
		if err != nil {
			return dto.CartResponse{}, apperror.Wrap(apperror.CodeInternal, "failed to parse product price", http.StatusInternalServerError, err)
		}

		subtotalCents := priceCents * int64(item.Quantity)
		totalAmountCents += subtotalCents
		result.TotalQuantity += item.Quantity

		flavors := make([]dto.CartProductFlavorResponse, 0, len(product.Flavors))
		for _, flavor := range product.Flavors {
			flavors = append(flavors, dto.CartProductFlavorResponse{
				ID:   flavor.ID.String(),
				Name: flavor.Name,
			})
		}
		tags := make([]dto.CartProductTagResponse, 0, len(product.Tags))
		for _, tag := range product.Tags {
			tags = append(tags, dto.CartProductTagResponse{
				ID:   tag.ID.String(),
				Code: tag.Code,
				Name: tag.Name,
			})
		}

		var strength *dto.CartProductStrengthResponse
		if product.Strength != nil {
			strength = &dto.CartProductStrengthResponse{
				ID:    product.Strength.ID.String(),
				Name:  product.Strength.Name,
				Level: product.Strength.Level,
			}
		}

		result.Items = append(result.Items, dto.CartItemResponse{
			ID:       item.ID.String(),
			Quantity: item.Quantity,
			Subtotal: formatCents(subtotalCents),
			Product: dto.CartProductResponse{
				ID:          product.ID.String(),
				Name:        product.Name,
				Description: product.Description,
				Price:       product.Price,
				Category: dto.CartProductCategoryResponse{
					ID:   product.Category.ID.String(),
					Code: product.Category.Code,
					Name: product.Category.Name,
				},
				StockQuantity: product.StockQuantity,
				Unit:          product.Unit,
				IsActive:      product.IsActive,
				Strength:      strength,
				Flavors:       flavors,
				Tags:          tags,
			},
		})
	}

	result.TotalAmount = formatCents(totalAmountCents)
	return result, nil
}

func findProductIDByCartItemID(items []domain.CartItem, itemID uuid.UUID) (uuid.UUID, bool) {
	for _, item := range items {
		if item.ID == itemID {
			return item.ProductID, true
		}
	}
	return uuid.Nil, false
}

func normalizeCartError(err error) error {
	if errors.Is(err, repository.ErrNotFound) {
		return apperror.New(apperror.CodeProductNotFound, "product not found", http.StatusNotFound)
	}

	var appErr *apperror.AppError
	if !errors.As(err, &appErr) {
		return err
	}

	if appErr.Code == apperror.CodeValidationError && strings.EqualFold(strings.TrimSpace(appErr.Message), "cart item not found") {
		return apperror.New(apperror.CodeCartItemNotFound, "cart item not found", http.StatusNotFound)
	}

	return err
}

func parseCents(value string) (int64, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return 0, fmt.Errorf("empty amount")
	}

	sign := int64(1)
	if strings.HasPrefix(trimmed, "-") {
		sign = -1
		trimmed = strings.TrimPrefix(trimmed, "-")
	}

	parts := strings.Split(trimmed, ".")
	if len(parts) > 2 {
		return 0, fmt.Errorf("invalid amount format")
	}

	whole := int64(0)
	if parts[0] != "" {
		parsed, err := strconv.ParseInt(parts[0], 10, 64)
		if err != nil {
			return 0, err
		}
		whole = parsed
	}

	fraction := int64(0)
	if len(parts) == 2 {
		frac := parts[1]
		if len(frac) > 2 {
			return 0, fmt.Errorf("invalid fraction")
		}
		if len(frac) == 1 {
			frac += "0"
		}
		if frac != "" {
			parsed, err := strconv.ParseInt(frac, 10, 64)
			if err != nil {
				return 0, err
			}
			fraction = parsed
		}
	}

	return sign * (whole*100 + fraction), nil
}

func formatCents(cents int64) string {
	sign := ""
	if cents < 0 {
		sign = "-"
		cents *= -1
	}

	whole := cents / 100
	fraction := cents % 100
	return fmt.Sprintf("%s%d.%02d", sign, whole, fraction)
}
