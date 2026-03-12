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

type TagHandler struct {
	service *service.TagManager
	logger  *slog.Logger
}

func NewTagHandler(service *service.TagManager, logger *slog.Logger) *TagHandler {
	return &TagHandler{
		service: service,
		logger:  logger,
	}
}

func (h *TagHandler) ListPublicTags(w http.ResponseWriter, r *http.Request) {
	limit, offset, err := parseTagPagination(r)
	if err != nil {
		httpx.WriteError(h.logger, w, err)
		return
	}

	items, err := h.service.ListPublicTags(r.Context(), limit, offset)
	if err != nil {
		httpx.WriteError(h.logger, w, err)
		return
	}

	resp := make([]dto.TagResponse, 0, len(items))
	for _, item := range items {
		resp = append(resp, toTagResponse(item))
	}

	httpx.WriteJSON(w, http.StatusOK, map[string]any{"items": resp})
}

func (h *TagHandler) ListAdminTags(w http.ResponseWriter, r *http.Request) {
	limit, offset, err := parseTagPagination(r)
	if err != nil {
		httpx.WriteError(h.logger, w, err)
		return
	}

	items, err := h.service.ListAdminTags(r.Context(), limit, offset)
	if err != nil {
		httpx.WriteError(h.logger, w, err)
		return
	}

	resp := make([]dto.TagResponse, 0, len(items))
	for _, item := range items {
		resp = append(resp, toTagResponse(item))
	}

	httpx.WriteJSON(w, http.StatusOK, map[string]any{"items": resp})
}

func (h *TagHandler) CreateTag(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateTagRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.WriteError(h.logger, w, err)
		return
	}

	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	item, err := h.service.CreateTag(r.Context(), &domain.Tag{
		Code:        req.Code,
		Name:        req.Name,
		Description: req.Description,
		IsActive:    isActive,
	})
	if err != nil {
		httpx.WriteError(h.logger, w, err)
		return
	}

	httpx.WriteJSON(w, http.StatusCreated, toTagResponse(*item))
}

func (h *TagHandler) UpdateTag(w http.ResponseWriter, r *http.Request) {
	tagID, err := uuid.Parse(strings.TrimSpace(chi.URLParam(r, "id")))
	if err != nil {
		httpx.WriteError(h.logger, w, apperror.New(apperror.CodeValidationError, "invalid tag id", http.StatusBadRequest))
		return
	}

	var req dto.UpdateTagRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.WriteError(h.logger, w, err)
		return
	}

	current, err := h.service.GetAdminTagByID(r.Context(), tagID)
	if err != nil {
		httpx.WriteError(h.logger, w, err)
		return
	}

	code := current.Code
	if req.Code != nil {
		code = *req.Code
	}

	name := current.Name
	if req.Name != nil {
		name = *req.Name
	}

	description := current.Description
	if req.Description != nil {
		description = req.Description
	}

	isActive := current.IsActive
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	item, err := h.service.UpdateTag(r.Context(), tagID, &domain.Tag{
		Code:        code,
		Name:        name,
		Description: description,
		IsActive:    isActive,
	})
	if err != nil {
		httpx.WriteError(h.logger, w, err)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, toTagResponse(*item))
}

func (h *TagHandler) DeleteTag(w http.ResponseWriter, r *http.Request) {
	tagID, err := uuid.Parse(strings.TrimSpace(chi.URLParam(r, "id")))
	if err != nil {
		httpx.WriteError(h.logger, w, apperror.New(apperror.CodeValidationError, "invalid tag id", http.StatusBadRequest))
		return
	}

	if err := h.service.DeactivateTag(r.Context(), tagID); err != nil {
		httpx.WriteError(h.logger, w, err)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, dto.MessageResponse{Message: "tag deactivated"})
}

func toTagResponse(item domain.Tag) dto.TagResponse {
	return dto.TagResponse{
		ID:          item.ID.String(),
		Code:        item.Code,
		Name:        item.Name,
		Description: item.Description,
		IsActive:    item.IsActive,
		CreatedAt:   item.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:   item.UpdatedAt.UTC().Format(time.RFC3339),
	}
}

func parseTagPagination(r *http.Request) (int, int, error) {
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
