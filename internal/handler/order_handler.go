package handler

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/karzhen/restaurant-lk/internal/domain"
	"github.com/karzhen/restaurant-lk/internal/dto"
	"github.com/karzhen/restaurant-lk/internal/service"
	"github.com/karzhen/restaurant-lk/internal/utils/apperror"
	"github.com/karzhen/restaurant-lk/internal/utils/ctxkeys"
	"github.com/karzhen/restaurant-lk/internal/utils/httpx"
)

type orderUserProvider interface {
	GetByID(ctx context.Context, userID uuid.UUID) (*domain.User, error)
}

type OrderHandler struct {
	service *service.OrderManager
	users   orderUserProvider
	logger  *slog.Logger
}

func NewOrderHandler(service *service.OrderManager, users orderUserProvider, logger *slog.Logger) *OrderHandler {
	return &OrderHandler{
		service: service,
		users:   users,
		logger:  logger,
	}
}

func (h *OrderHandler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	userID, ok := ctxkeys.UserID(r.Context())
	if !ok {
		httpx.WriteError(h.logger, w, apperror.New(apperror.CodeUnauthorized, "unauthorized", http.StatusUnauthorized))
		return
	}

	order, err := h.service.Checkout(r.Context(), userID)
	if err != nil {
		httpx.WriteError(h.logger, w, err)
		return
	}

	httpx.WriteJSON(w, http.StatusCreated, toCreateOrderResponse(order))
}

func (h *OrderHandler) ListMyOrders(w http.ResponseWriter, r *http.Request) {
	userID, ok := ctxkeys.UserID(r.Context())
	if !ok {
		httpx.WriteError(h.logger, w, apperror.New(apperror.CodeUnauthorized, "unauthorized", http.StatusUnauthorized))
		return
	}

	orders, err := h.service.ListUserOrders(r.Context(), userID)
	if err != nil {
		httpx.WriteError(h.logger, w, err)
		return
	}

	resp := make([]dto.OrderResponse, 0, len(orders))
	for i := range orders {
		resp = append(resp, toOrderResponse(&orders[i]))
	}

	httpx.WriteJSON(w, http.StatusOK, map[string]any{"items": resp})
}

func (h *OrderHandler) GetMyOrderByID(w http.ResponseWriter, r *http.Request) {
	userID, ok := ctxkeys.UserID(r.Context())
	if !ok {
		httpx.WriteError(h.logger, w, apperror.New(apperror.CodeUnauthorized, "unauthorized", http.StatusUnauthorized))
		return
	}

	orderID, err := uuid.Parse(strings.TrimSpace(chi.URLParam(r, "id")))
	if err != nil {
		httpx.WriteError(h.logger, w, apperror.New(apperror.CodeValidationError, "invalid order id", http.StatusBadRequest))
		return
	}

	order, err := h.service.GetOrderByID(r.Context(), userID, orderID)
	if err != nil {
		httpx.WriteError(h.logger, w, err)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, toOrderResponse(order))
}

func (h *OrderHandler) ListAllOrders(w http.ResponseWriter, r *http.Request) {
	role, ok := ctxkeys.Role(r.Context())
	if !ok {
		httpx.WriteError(h.logger, w, apperror.New(apperror.CodeUnauthorized, "unauthorized", http.StatusUnauthorized))
		return
	}

	limit, offset, err := parseOrderPagination(r)
	if err != nil {
		httpx.WriteError(h.logger, w, err)
		return
	}

	orders, err := h.service.ListAllOrders(r.Context(), role, limit, offset)
	if err != nil {
		httpx.WriteError(h.logger, w, err)
		return
	}

	userCache := make(map[uuid.UUID]dto.OrderUserSummary, len(orders))
	resp := make([]dto.AdminOrderResponse, 0, len(orders))
	for i := range orders {
		order := &orders[i]

		summary, err := h.resolveOrderUserSummary(r.Context(), order.UserID, userCache)
		if err != nil {
			httpx.WriteError(h.logger, w, err)
			return
		}

		resp = append(resp, toAdminOrderResponse(order, summary))
	}

	httpx.WriteJSON(w, http.StatusOK, map[string]any{"items": resp})
}

func (h *OrderHandler) UpdateOrderStatus(w http.ResponseWriter, r *http.Request) {
	role, ok := ctxkeys.Role(r.Context())
	if !ok {
		httpx.WriteError(h.logger, w, apperror.New(apperror.CodeUnauthorized, "unauthorized", http.StatusUnauthorized))
		return
	}

	orderID, err := uuid.Parse(strings.TrimSpace(chi.URLParam(r, "id")))
	if err != nil {
		httpx.WriteError(h.logger, w, apperror.New(apperror.CodeValidationError, "invalid order id", http.StatusBadRequest))
		return
	}

	var req dto.UpdateOrderStatusRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.WriteError(h.logger, w, err)
		return
	}

	order, err := h.service.ChangeOrderStatus(r.Context(), role, orderID, domain.OrderStatus(req.Status))
	if err != nil {
		httpx.WriteError(h.logger, w, err)
		return
	}

	summary, err := h.resolveOrderUserSummary(r.Context(), order.UserID, map[uuid.UUID]dto.OrderUserSummary{})
	if err != nil {
		httpx.WriteError(h.logger, w, err)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, toAdminOrderResponse(order, summary))
}

func (h *OrderHandler) resolveOrderUserSummary(
	ctx context.Context,
	userID uuid.UUID,
	cache map[uuid.UUID]dto.OrderUserSummary,
) (dto.OrderUserSummary, error) {
	if cached, ok := cache[userID]; ok {
		return cached, nil
	}

	user, err := h.users.GetByID(ctx, userID)
	if err != nil {
		var appErr *apperror.AppError
		if errors.As(err, &appErr) {
			return dto.OrderUserSummary{}, appErr
		}
		return dto.OrderUserSummary{}, apperror.Wrap(apperror.CodeInternal, "failed to fetch order user", http.StatusInternalServerError, err)
	}

	summary := dto.OrderUserSummary{
		ID:        user.ID.String(),
		Email:     user.Email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Role:      user.Role,
	}
	cache[userID] = summary

	return summary, nil
}

func parseOrderPagination(r *http.Request) (int, int, error) {
	query := r.URL.Query()

	limit := 20
	if raw := strings.TrimSpace(query.Get("limit")); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil || parsed <= 0 {
			return 0, 0, apperror.New(apperror.CodeValidationError, "invalid limit", http.StatusBadRequest)
		}
		if parsed > 100 {
			parsed = 100
		}
		limit = parsed
	}

	offset := 0
	if raw := strings.TrimSpace(query.Get("offset")); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil || parsed < 0 {
			return 0, 0, apperror.New(apperror.CodeValidationError, "invalid offset", http.StatusBadRequest)
		}
		offset = parsed
	}

	return limit, offset, nil
}

func toCreateOrderResponse(order *domain.Order) dto.CreateOrderResponse {
	items := make([]dto.OrderItemResponse, 0, len(order.Items))
	for i := range order.Items {
		items = append(items, toOrderItemResponse(&order.Items[i]))
	}

	return dto.CreateOrderResponse{
		ID:          order.ID.String(),
		Status:      string(order.Status),
		TotalAmount: order.TotalAmount,
		Items:       items,
		CreatedAt:   order.CreatedAt.UTC().Format(time.RFC3339),
	}
}

func toOrderResponse(order *domain.Order) dto.OrderResponse {
	items := make([]dto.OrderItemResponse, 0, len(order.Items))
	for i := range order.Items {
		items = append(items, toOrderItemResponse(&order.Items[i]))
	}

	return dto.OrderResponse{
		ID:          order.ID.String(),
		UserID:      order.UserID.String(),
		Status:      string(order.Status),
		TotalAmount: order.TotalAmount,
		Items:       items,
		CreatedAt:   order.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:   order.UpdatedAt.UTC().Format(time.RFC3339),
	}
}

func toAdminOrderResponse(order *domain.Order, user dto.OrderUserSummary) dto.AdminOrderResponse {
	items := make([]dto.OrderItemResponse, 0, len(order.Items))
	for i := range order.Items {
		items = append(items, toOrderItemResponse(&order.Items[i]))
	}

	return dto.AdminOrderResponse{
		ID:          order.ID.String(),
		Status:      string(order.Status),
		TotalAmount: order.TotalAmount,
		Items:       items,
		User:        user,
		CreatedAt:   order.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:   order.UpdatedAt.UTC().Format(time.RFC3339),
	}
}

func toOrderItemResponse(item *domain.OrderItem) dto.OrderItemResponse {
	return dto.OrderItemResponse{
		ID:              item.ID.String(),
		ProductID:       item.ProductID.String(),
		ProductNameSnap: item.ProductNameSnap,
		UnitPriceSnap:   item.UnitPriceSnap,
		Quantity:        item.Quantity,
		Subtotal:        item.Subtotal,
		CreatedAt:       item.CreatedAt.UTC().Format(time.RFC3339),
	}
}
