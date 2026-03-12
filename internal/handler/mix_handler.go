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
	"github.com/karzhen/restaurant-lk/internal/repository"
	"github.com/karzhen/restaurant-lk/internal/service"
	"github.com/karzhen/restaurant-lk/internal/utils/apperror"
	"github.com/karzhen/restaurant-lk/internal/utils/httpx"
)

type mixProductReader interface {
	GetProductByID(ctx context.Context, id uuid.UUID, adminView bool) (*domain.Product, error)
}

type MixHandler struct {
	service  *service.MixManager
	products mixProductReader
	logger   *slog.Logger
}

func NewMixHandler(service *service.MixManager, products mixProductReader, logger *slog.Logger) *MixHandler {
	return &MixHandler{
		service:  service,
		products: products,
		logger:   logger,
	}
}

func (h *MixHandler) ListPublicMixes(w http.ResponseWriter, r *http.Request) {
	limit, offset, err := parseMixPagination(r)
	if err != nil {
		httpx.WriteError(h.logger, w, err)
		return
	}

	mixes, err := h.service.ListPublicMixes(r.Context(), limit, offset)
	if err != nil {
		httpx.WriteError(h.logger, w, err)
		return
	}

	resp := make([]dto.MixResponse, 0, len(mixes))
	for i := range mixes {
		mixResp, err := h.buildMixResponse(r.Context(), &mixes[i])
		if err != nil {
			httpx.WriteError(h.logger, w, err)
			return
		}
		resp = append(resp, mixResp)
	}

	httpx.WriteJSON(w, http.StatusOK, map[string]any{"items": resp})
}

func (h *MixHandler) GetPublicMixByID(w http.ResponseWriter, r *http.Request) {
	mixID, err := uuid.Parse(strings.TrimSpace(chi.URLParam(r, "id")))
	if err != nil {
		httpx.WriteError(h.logger, w, apperror.New(apperror.CodeValidationError, "invalid mix id", http.StatusBadRequest))
		return
	}

	mix, err := h.service.GetPublicMixByID(r.Context(), mixID)
	if err != nil {
		httpx.WriteError(h.logger, w, err)
		return
	}

	resp, err := h.buildMixResponse(r.Context(), mix)
	if err != nil {
		httpx.WriteError(h.logger, w, err)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, resp)
}

func (h *MixHandler) ListAdminMixes(w http.ResponseWriter, r *http.Request) {
	limit, offset, err := parseMixPagination(r)
	if err != nil {
		httpx.WriteError(h.logger, w, err)
		return
	}

	mixes, err := h.service.ListAdminMixes(r.Context(), limit, offset)
	if err != nil {
		httpx.WriteError(h.logger, w, err)
		return
	}

	resp := make([]dto.MixResponse, 0, len(mixes))
	for i := range mixes {
		mixResp, err := h.buildMixResponse(r.Context(), &mixes[i])
		if err != nil {
			httpx.WriteError(h.logger, w, err)
			return
		}
		resp = append(resp, mixResp)
	}

	httpx.WriteJSON(w, http.StatusOK, map[string]any{"items": resp})
}

func (h *MixHandler) CreateMix(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateMixRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.WriteError(h.logger, w, err)
		return
	}

	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	items, err := parseMixItemInputs(req.Items)
	if err != nil {
		httpx.WriteError(h.logger, w, err)
		return
	}

	created, err := h.service.CreateMix(r.Context(), &domain.Mix{
		Name:               req.Name,
		Description:        req.Description,
		FinalStrengthLabel: req.FinalStrengthLabel,
		IsActive:           isActive,
	}, items)
	if err != nil {
		httpx.WriteError(h.logger, w, err)
		return
	}

	resp, err := h.buildMixResponse(r.Context(), created)
	if err != nil {
		httpx.WriteError(h.logger, w, err)
		return
	}

	httpx.WriteJSON(w, http.StatusCreated, resp)
}

func (h *MixHandler) UpdateMix(w http.ResponseWriter, r *http.Request) {
	mixID, err := uuid.Parse(strings.TrimSpace(chi.URLParam(r, "id")))
	if err != nil {
		httpx.WriteError(h.logger, w, apperror.New(apperror.CodeValidationError, "invalid mix id", http.StatusBadRequest))
		return
	}

	var req dto.UpdateMixRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.WriteError(h.logger, w, err)
		return
	}

	current, err := h.service.GetAdminMixByID(r.Context(), mixID)
	if err != nil {
		httpx.WriteError(h.logger, w, err)
		return
	}

	name := current.Name
	if req.Name != nil {
		name = *req.Name
	}

	description := current.Description
	if req.Description != nil {
		description = req.Description
	}

	finalStrengthLabel := current.FinalStrengthLabel
	if req.FinalStrengthLabel != nil {
		finalStrengthLabel = req.FinalStrengthLabel
	}

	isActive := current.IsActive
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	items := current.Items
	if req.Items != nil {
		parsedItems, err := parseMixItemInputs(*req.Items)
		if err != nil {
			httpx.WriteError(h.logger, w, err)
			return
		}
		items = parsedItems
	}

	updated, err := h.service.UpdateMix(r.Context(), mixID, &domain.Mix{
		Name:               name,
		Description:        description,
		FinalStrengthLabel: finalStrengthLabel,
		IsActive:           isActive,
	}, items)
	if err != nil {
		httpx.WriteError(h.logger, w, err)
		return
	}

	resp, err := h.buildMixResponse(r.Context(), updated)
	if err != nil {
		httpx.WriteError(h.logger, w, err)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, resp)
}

func (h *MixHandler) DeleteMix(w http.ResponseWriter, r *http.Request) {
	mixID, err := uuid.Parse(strings.TrimSpace(chi.URLParam(r, "id")))
	if err != nil {
		httpx.WriteError(h.logger, w, apperror.New(apperror.CodeValidationError, "invalid mix id", http.StatusBadRequest))
		return
	}

	if err := h.service.DeactivateMix(r.Context(), mixID); err != nil {
		httpx.WriteError(h.logger, w, err)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, dto.MessageResponse{Message: "mix deactivated"})
}

func (h *MixHandler) buildMixResponse(ctx context.Context, mix *domain.Mix) (dto.MixResponse, error) {
	items := make([]dto.MixItemResponse, 0, len(mix.Items))
	for _, item := range mix.Items {
		product, err := h.products.GetProductByID(ctx, item.ProductID, true)
		if err != nil {
			if errors.Is(err, repository.ErrNotFound) {
				return dto.MixResponse{}, apperror.New(apperror.CodeTobaccoProductNotFound, "tobacco product not found", http.StatusNotFound)
			}
			return dto.MixResponse{}, apperror.Wrap(apperror.CodeInternal, "failed to fetch mix item product", http.StatusInternalServerError, err)
		}

		items = append(items, dto.MixItemResponse{
			ProductID:   item.ProductID.String(),
			ProductName: product.Name,
			Percent:     item.Percent,
		})
	}
	tags := make([]dto.MixTagResponse, 0, len(mix.Tags))
	for _, tag := range mix.Tags {
		tags = append(tags, dto.MixTagResponse{
			ID:   tag.ID.String(),
			Code: tag.Code,
			Name: tag.Name,
		})
	}

	return dto.MixResponse{
		ID:                 mix.ID.String(),
		Name:               mix.Name,
		Description:        mix.Description,
		FinalStrengthLabel: mix.FinalStrengthLabel,
		IsActive:           mix.IsActive,
		Items:              items,
		Tags:               tags,
		CreatedAt:          mix.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:          mix.UpdatedAt.UTC().Format(time.RFC3339),
	}, nil
}

func parseMixItemInputs(inputs []dto.MixItemInput) ([]domain.MixItem, error) {
	items := make([]domain.MixItem, 0, len(inputs))
	for _, input := range inputs {
		productID, err := uuid.Parse(strings.TrimSpace(input.ProductID))
		if err != nil {
			return nil, apperror.New(apperror.CodeValidationError, "invalid mix item product_id", http.StatusBadRequest)
		}

		items = append(items, domain.MixItem{
			ProductID: productID,
			Percent:   input.Percent,
		})
	}
	return items, nil
}

func parseMixPagination(r *http.Request) (int, int, error) {
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
