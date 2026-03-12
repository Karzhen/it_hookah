package handler

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/karzhen/restaurant-lk/internal/domain"
	"github.com/karzhen/restaurant-lk/internal/dto"
	"github.com/karzhen/restaurant-lk/internal/service"
	"github.com/karzhen/restaurant-lk/internal/utils/apperror"
	"github.com/karzhen/restaurant-lk/internal/utils/ctxkeys"
	"github.com/karzhen/restaurant-lk/internal/utils/httpx"
)

type UserHandler struct {
	userService *service.UserService
	logger      *slog.Logger
}

func NewUserHandler(userService *service.UserService, logger *slog.Logger) *UserHandler {
	return &UserHandler{
		userService: userService,
		logger:      logger,
	}
}

func (h *UserHandler) Me(w http.ResponseWriter, r *http.Request) {
	userID, ok := ctxkeys.UserID(r.Context())
	if !ok {
		httpx.WriteError(h.logger, w, apperror.New(apperror.CodeUnauthorized, "unauthorized", http.StatusUnauthorized))
		return
	}

	user, err := h.userService.GetMe(r.Context(), userID)
	if err != nil {
		httpx.WriteError(h.logger, w, err)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, toUserProfileResponse(user))
}

func (h *UserHandler) UpdateMe(w http.ResponseWriter, r *http.Request) {
	userID, ok := ctxkeys.UserID(r.Context())
	if !ok {
		httpx.WriteError(h.logger, w, apperror.New(apperror.CodeUnauthorized, "unauthorized", http.StatusUnauthorized))
		return
	}

	var req dto.UpdateMeRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.WriteError(h.logger, w, err)
		return
	}

	updated, err := h.userService.UpdateMe(r.Context(), userID, req)
	if err != nil {
		httpx.WriteError(h.logger, w, err)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, toUserProfileResponse(updated))
}

func toUserProfileResponse(user *domain.User) dto.UserProfileResponse {
	return dto.UserProfileResponse{
		ID:         user.ID.String(),
		Email:      user.Email,
		FirstName:  user.FirstName,
		LastName:   user.LastName,
		MiddleName: user.MiddleName,
		Phone:      user.Phone,
		Age:        user.Age,
		Role:       user.Role,
		IsActive:   user.IsActive,
		CreatedAt:  user.CreatedAt.UTC().Format(time.RFC3339),
	}
}
