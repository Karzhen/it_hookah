package handler

import (
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
	"github.com/karzhen/restaurant-lk/internal/utils/httpx"
)

type StockMovementHandler struct {
	service *service.StockMovementManager
	logger  *slog.Logger
}

func NewStockMovementHandler(service *service.StockMovementManager, logger *slog.Logger) *StockMovementHandler {
	return &StockMovementHandler{
		service: service,
		logger:  logger,
	}
}

func (h *StockMovementHandler) ListStockMovements(w http.ResponseWriter, r *http.Request) {
	filter, err := parseStockMovementFilter(r)
	if err != nil {
		httpx.WriteError(h.logger, w, err)
		return
	}

	items, err := h.service.ListMovements(r.Context(), filter)
	if err != nil {
		httpx.WriteError(h.logger, w, err)
		return
	}

	resp := make([]dto.StockMovementResponse, 0, len(items))
	for i := range items {
		resp = append(resp, toStockMovementResponse(items[i]))
	}

	httpx.WriteJSON(w, http.StatusOK, map[string]any{"items": resp})
}

func (h *StockMovementHandler) ListProductStockMovements(w http.ResponseWriter, r *http.Request) {
	productID, err := uuid.Parse(strings.TrimSpace(chi.URLParam(r, "id")))
	if err != nil {
		httpx.WriteError(h.logger, w, apperror.New(apperror.CodeValidationError, "invalid product id", http.StatusBadRequest))
		return
	}

	operation, limit, offset, err := parseProductStockMovementsQuery(r)
	if err != nil {
		httpx.WriteError(h.logger, w, err)
		return
	}

	items, err := h.service.ListProductMovements(r.Context(), productID, operation, limit, offset)
	if err != nil {
		httpx.WriteError(h.logger, w, err)
		return
	}

	resp := make([]dto.StockMovementResponse, 0, len(items))
	for i := range items {
		resp = append(resp, toStockMovementResponse(items[i]))
	}

	httpx.WriteJSON(w, http.StatusOK, map[string]any{"items": resp})
}

func toStockMovementResponse(item domain.StockMovement) dto.StockMovementResponse {
	var createdByUserID *string
	if item.CreatedByUserID != nil {
		value := item.CreatedByUserID.String()
		createdByUserID = &value
	}

	return dto.StockMovementResponse{
		ID:              item.ID.String(),
		ProductID:       item.ProductID.String(),
		ProductName:     item.ProductName,
		Operation:       string(item.Operation),
		Quantity:        item.Quantity,
		BeforeQuantity:  item.BeforeQuantity,
		AfterQuantity:   item.AfterQuantity,
		Reason:          item.Reason,
		CreatedByUserID: createdByUserID,
		CreatedAt:       item.CreatedAt.UTC().Format(time.RFC3339),
	}
}

func parseStockMovementFilter(r *http.Request) (domain.StockMovementFilter, error) {
	query := r.URL.Query()
	limit, offset, err := parseLimitOffset(query.Get("limit"), query.Get("offset"))
	if err != nil {
		return domain.StockMovementFilter{}, err
	}

	filter := domain.StockMovementFilter{
		Limit:  limit,
		Offset: offset,
	}

	if rawProductID := strings.TrimSpace(query.Get("product_id")); rawProductID != "" {
		productID, err := uuid.Parse(rawProductID)
		if err != nil {
			return domain.StockMovementFilter{}, apperror.New(apperror.CodeValidationError, "invalid product_id", http.StatusBadRequest)
		}
		filter.ProductID = &productID
	}

	if rawOperation := strings.TrimSpace(query.Get("operation")); rawOperation != "" {
		operation := domain.StockMovementOperation(strings.ToLower(rawOperation))
		filter.Operation = &operation
	}

	return filter, nil
}

func parseProductStockMovementsQuery(r *http.Request) (*domain.StockMovementOperation, int, int, error) {
	query := r.URL.Query()
	limit, offset, err := parseLimitOffset(query.Get("limit"), query.Get("offset"))
	if err != nil {
		return nil, 0, 0, err
	}

	var operation *domain.StockMovementOperation
	if rawOperation := strings.TrimSpace(query.Get("operation")); rawOperation != "" {
		op := domain.StockMovementOperation(strings.ToLower(rawOperation))
		operation = &op
	}

	return operation, limit, offset, nil
}

func parseLimitOffset(limitRaw, offsetRaw string) (int, int, error) {
	limit := 20
	if raw := strings.TrimSpace(limitRaw); raw != "" {
		value, err := strconv.Atoi(raw)
		if err != nil || value <= 0 {
			return 0, 0, apperror.New(apperror.CodeValidationError, "invalid limit", http.StatusBadRequest)
		}
		if value > 100 {
			value = 100
		}
		limit = value
	}

	offset := 0
	if raw := strings.TrimSpace(offsetRaw); raw != "" {
		value, err := strconv.Atoi(raw)
		if err != nil || value < 0 {
			return 0, 0, apperror.New(apperror.CodeValidationError, "invalid offset", http.StatusBadRequest)
		}
		offset = value
	}

	return limit, offset, nil
}
