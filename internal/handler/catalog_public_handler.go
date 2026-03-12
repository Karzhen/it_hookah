package handler

import (
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/karzhen/restaurant-lk/internal/domain"
	"github.com/karzhen/restaurant-lk/internal/dto"
	"github.com/karzhen/restaurant-lk/internal/service"
	"github.com/karzhen/restaurant-lk/internal/utils/apperror"
	"github.com/karzhen/restaurant-lk/internal/utils/httpx"
)

type CatalogPublicHandler struct {
	service *service.CatalogService
	logger  *slog.Logger
}

func NewCatalogPublicHandler(service *service.CatalogService, logger *slog.Logger) *CatalogPublicHandler {
	return &CatalogPublicHandler{
		service: service,
		logger:  logger,
	}
}

func (h *CatalogPublicHandler) ListCategories(w http.ResponseWriter, r *http.Request) {
	items, err := h.service.ListPublicCategories(r.Context())
	if err != nil {
		httpx.WriteError(h.logger, w, err)
		return
	}

	resp := make([]dto.CategoryResponse, 0, len(items))
	for _, item := range items {
		resp = append(resp, toCategoryResponse(item))
	}

	httpx.WriteJSON(w, http.StatusOK, map[string]any{"items": resp})
}

func (h *CatalogPublicHandler) ListFlavors(w http.ResponseWriter, r *http.Request) {
	items, err := h.service.ListPublicFlavors(r.Context())
	if err != nil {
		httpx.WriteError(h.logger, w, err)
		return
	}

	resp := make([]dto.FlavorResponse, 0, len(items))
	for _, item := range items {
		resp = append(resp, toFlavorResponse(item))
	}

	httpx.WriteJSON(w, http.StatusOK, map[string]any{"items": resp})
}

func (h *CatalogPublicHandler) ListStrengths(w http.ResponseWriter, r *http.Request) {
	items, err := h.service.ListPublicStrengths(r.Context())
	if err != nil {
		httpx.WriteError(h.logger, w, err)
		return
	}

	resp := make([]dto.StrengthResponse, 0, len(items))
	for _, item := range items {
		resp = append(resp, toStrengthResponse(item))
	}

	httpx.WriteJSON(w, http.StatusOK, map[string]any{"items": resp})
}

func (h *CatalogPublicHandler) ListProducts(w http.ResponseWriter, r *http.Request) {
	filter, err := parseProductFilter(r, false)
	if err != nil {
		httpx.WriteError(h.logger, w, err)
		return
	}

	items, err := h.service.ListPublicProducts(r.Context(), filter)
	if err != nil {
		httpx.WriteError(h.logger, w, err)
		return
	}

	resp := make([]dto.ProductResponse, 0, len(items))
	for i := range items {
		item := items[i]
		resp = append(resp, toProductResponse(&item))
	}

	httpx.WriteJSON(w, http.StatusOK, map[string]any{"items": resp})
}

func (h *CatalogPublicHandler) GetProductByID(w http.ResponseWriter, r *http.Request) {
	productID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httpx.WriteError(h.logger, w, apperror.New(apperror.CodeValidationError, "invalid product id", http.StatusBadRequest))
		return
	}

	item, err := h.service.GetPublicProductByID(r.Context(), productID)
	if err != nil {
		httpx.WriteError(h.logger, w, err)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, toProductResponse(item))
}

func parseProductFilter(r *http.Request, adminView bool) (domain.ProductFilter, error) {
	query := r.URL.Query()
	filter := domain.ProductFilter{AdminView: adminView}

	if categoryCode := strings.TrimSpace(query.Get("category_code")); categoryCode != "" {
		filter.CategoryCode = &categoryCode
	}
	if search := strings.TrimSpace(query.Get("search")); search != "" {
		filter.Search = &search
	}
	if minPrice := strings.TrimSpace(query.Get("min_price")); minPrice != "" {
		if _, err := strconv.ParseFloat(minPrice, 64); err != nil {
			return domain.ProductFilter{}, apperror.New(apperror.CodeValidationError, "invalid min_price filter", http.StatusBadRequest)
		}
		filter.MinPrice = &minPrice
	}
	if maxPrice := strings.TrimSpace(query.Get("max_price")); maxPrice != "" {
		if _, err := strconv.ParseFloat(maxPrice, 64); err != nil {
			return domain.ProductFilter{}, apperror.New(apperror.CodeValidationError, "invalid max_price filter", http.StatusBadRequest)
		}
		filter.MaxPrice = &maxPrice
	}
	if inStockRaw := strings.TrimSpace(query.Get("in_stock")); inStockRaw != "" {
		inStock, err := strconv.ParseBool(inStockRaw)
		if err != nil {
			return domain.ProductFilter{}, apperror.New(apperror.CodeValidationError, "invalid in_stock filter", http.StatusBadRequest)
		}
		filter.InStock = &inStock
	}
	if strengthRaw := strings.TrimSpace(query.Get("strength_id")); strengthRaw != "" {
		parsed, err := uuid.Parse(strengthRaw)
		if err != nil {
			return domain.ProductFilter{}, apperror.New(apperror.CodeValidationError, "invalid strength_id filter", http.StatusBadRequest)
		}
		filter.StrengthID = &parsed
	}
	if flavorRaw := strings.TrimSpace(query.Get("flavor_id")); flavorRaw != "" {
		parsed, err := uuid.Parse(flavorRaw)
		if err != nil {
			return domain.ProductFilter{}, apperror.New(apperror.CodeValidationError, "invalid flavor_id filter", http.StatusBadRequest)
		}
		filter.FlavorID = &parsed
	}
	if adminView {
		if activeRaw := strings.TrimSpace(query.Get("is_active")); activeRaw != "" {
			parsed, err := strconv.ParseBool(activeRaw)
			if err != nil {
				return domain.ProductFilter{}, apperror.New(apperror.CodeValidationError, "invalid is_active filter", http.StatusBadRequest)
			}
			filter.IsActive = &parsed
		}
	}

	limit := 20
	if raw := strings.TrimSpace(query.Get("limit")); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil || parsed <= 0 {
			return domain.ProductFilter{}, apperror.New(apperror.CodeValidationError, "invalid limit", http.StatusBadRequest)
		}
		if parsed > 100 {
			parsed = 100
		}
		limit = parsed
	}
	filter.Limit = limit

	offset := 0
	if raw := strings.TrimSpace(query.Get("offset")); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil || parsed < 0 {
			return domain.ProductFilter{}, apperror.New(apperror.CodeValidationError, "invalid offset", http.StatusBadRequest)
		}
		offset = parsed
	}
	filter.Offset = offset

	return filter, nil
}
