package handler

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/karzhen/restaurant-lk/internal/dto"
	"github.com/karzhen/restaurant-lk/internal/service"
	"github.com/karzhen/restaurant-lk/internal/utils/apperror"
	"github.com/karzhen/restaurant-lk/internal/utils/httpx"
)

type CatalogAdminHandler struct {
	service *service.CatalogService
	logger  *slog.Logger
}

func NewCatalogAdminHandler(service *service.CatalogService, logger *slog.Logger) *CatalogAdminHandler {
	return &CatalogAdminHandler{service: service, logger: logger}
}

func (h *CatalogAdminHandler) ListCategories(w http.ResponseWriter, r *http.Request) {
	items, err := h.service.ListAdminCategories(r.Context())
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

func (h *CatalogAdminHandler) CreateCategory(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateCategoryRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.WriteError(h.logger, w, err)
		return
	}

	item, err := h.service.CreateCategory(r.Context(), req)
	if err != nil {
		httpx.WriteError(h.logger, w, err)
		return
	}

	httpx.WriteJSON(w, http.StatusCreated, toCategoryResponse(*item))
}

func (h *CatalogAdminHandler) UpdateCategory(w http.ResponseWriter, r *http.Request) {
	categoryID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httpx.WriteError(h.logger, w, apperror.New(apperror.CodeValidationError, "invalid category id", http.StatusBadRequest))
		return
	}

	var req dto.UpdateCategoryRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.WriteError(h.logger, w, err)
		return
	}

	item, err := h.service.UpdateCategory(r.Context(), categoryID, req)
	if err != nil {
		httpx.WriteError(h.logger, w, err)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, toCategoryResponse(*item))
}

func (h *CatalogAdminHandler) DeleteCategory(w http.ResponseWriter, r *http.Request) {
	categoryID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httpx.WriteError(h.logger, w, apperror.New(apperror.CodeValidationError, "invalid category id", http.StatusBadRequest))
		return
	}

	if err := h.service.DeactivateCategory(r.Context(), categoryID); err != nil {
		httpx.WriteError(h.logger, w, err)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, dto.MessageResponse{Message: "category deactivated"})
}

func (h *CatalogAdminHandler) ListFlavors(w http.ResponseWriter, r *http.Request) {
	items, err := h.service.ListAdminFlavors(r.Context())
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

func (h *CatalogAdminHandler) CreateFlavor(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateFlavorRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.WriteError(h.logger, w, err)
		return
	}

	item, err := h.service.CreateFlavor(r.Context(), req)
	if err != nil {
		httpx.WriteError(h.logger, w, err)
		return
	}

	httpx.WriteJSON(w, http.StatusCreated, toFlavorResponse(*item))
}

func (h *CatalogAdminHandler) UpdateFlavor(w http.ResponseWriter, r *http.Request) {
	flavorID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httpx.WriteError(h.logger, w, apperror.New(apperror.CodeValidationError, "invalid flavor id", http.StatusBadRequest))
		return
	}

	var req dto.UpdateFlavorRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.WriteError(h.logger, w, err)
		return
	}

	item, err := h.service.UpdateFlavor(r.Context(), flavorID, req)
	if err != nil {
		httpx.WriteError(h.logger, w, err)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, toFlavorResponse(*item))
}

func (h *CatalogAdminHandler) DeleteFlavor(w http.ResponseWriter, r *http.Request) {
	flavorID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httpx.WriteError(h.logger, w, apperror.New(apperror.CodeValidationError, "invalid flavor id", http.StatusBadRequest))
		return
	}

	if err := h.service.DeactivateFlavor(r.Context(), flavorID); err != nil {
		httpx.WriteError(h.logger, w, err)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, dto.MessageResponse{Message: "flavor deactivated"})
}

func (h *CatalogAdminHandler) ListStrengths(w http.ResponseWriter, r *http.Request) {
	items, err := h.service.ListAdminStrengths(r.Context())
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

func (h *CatalogAdminHandler) CreateStrength(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateStrengthRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.WriteError(h.logger, w, err)
		return
	}

	item, err := h.service.CreateStrength(r.Context(), req)
	if err != nil {
		httpx.WriteError(h.logger, w, err)
		return
	}

	httpx.WriteJSON(w, http.StatusCreated, toStrengthResponse(*item))
}

func (h *CatalogAdminHandler) UpdateStrength(w http.ResponseWriter, r *http.Request) {
	strengthID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httpx.WriteError(h.logger, w, apperror.New(apperror.CodeValidationError, "invalid strength id", http.StatusBadRequest))
		return
	}

	var req dto.UpdateStrengthRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.WriteError(h.logger, w, err)
		return
	}

	item, err := h.service.UpdateStrength(r.Context(), strengthID, req)
	if err != nil {
		httpx.WriteError(h.logger, w, err)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, toStrengthResponse(*item))
}

func (h *CatalogAdminHandler) DeleteStrength(w http.ResponseWriter, r *http.Request) {
	strengthID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httpx.WriteError(h.logger, w, apperror.New(apperror.CodeValidationError, "invalid strength id", http.StatusBadRequest))
		return
	}

	if err := h.service.DeactivateStrength(r.Context(), strengthID); err != nil {
		httpx.WriteError(h.logger, w, err)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, dto.MessageResponse{Message: "strength deactivated"})
}

func (h *CatalogAdminHandler) ListProducts(w http.ResponseWriter, r *http.Request) {
	filter, err := parseProductFilter(r, true)
	if err != nil {
		httpx.WriteError(h.logger, w, err)
		return
	}

	items, err := h.service.ListAdminProducts(r.Context(), filter)
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

func (h *CatalogAdminHandler) CreateProduct(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateProductRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.WriteError(h.logger, w, err)
		return
	}

	item, err := h.service.CreateProduct(r.Context(), req)
	if err != nil {
		httpx.WriteError(h.logger, w, err)
		return
	}

	httpx.WriteJSON(w, http.StatusCreated, toProductResponse(item))
}

func (h *CatalogAdminHandler) UpdateProduct(w http.ResponseWriter, r *http.Request) {
	productID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httpx.WriteError(h.logger, w, apperror.New(apperror.CodeValidationError, "invalid product id", http.StatusBadRequest))
		return
	}

	var req dto.UpdateProductRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.WriteError(h.logger, w, err)
		return
	}

	item, err := h.service.UpdateProduct(r.Context(), productID, req)
	if err != nil {
		httpx.WriteError(h.logger, w, err)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, toProductResponse(item))
}

func (h *CatalogAdminHandler) DeleteProduct(w http.ResponseWriter, r *http.Request) {
	productID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httpx.WriteError(h.logger, w, apperror.New(apperror.CodeValidationError, "invalid product id", http.StatusBadRequest))
		return
	}

	if err := h.service.DeactivateProduct(r.Context(), productID); err != nil {
		httpx.WriteError(h.logger, w, err)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, dto.MessageResponse{Message: "product deactivated"})
}

func (h *CatalogAdminHandler) UpdateProductStock(w http.ResponseWriter, r *http.Request) {
	productID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httpx.WriteError(h.logger, w, apperror.New(apperror.CodeValidationError, "invalid product id", http.StatusBadRequest))
		return
	}

	var req dto.UpdateStockRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.WriteError(h.logger, w, err)
		return
	}

	item, err := h.service.UpdateProductStock(r.Context(), productID, req)
	if err != nil {
		httpx.WriteError(h.logger, w, err)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, toProductResponse(item))
}
