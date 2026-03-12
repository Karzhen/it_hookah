package handler

import (
	"log/slog"
	"net/http"

	"github.com/karzhen/restaurant-lk/internal/dto"
	"github.com/karzhen/restaurant-lk/internal/service"
	"github.com/karzhen/restaurant-lk/internal/utils/apperror"
	"github.com/karzhen/restaurant-lk/internal/utils/ctxkeys"
	"github.com/karzhen/restaurant-lk/internal/utils/httpx"
)

type AuthHandler struct {
	authService *service.AuthService
	logger      *slog.Logger
}

func NewAuthHandler(authService *service.AuthService, logger *slog.Logger) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		logger:      logger,
	}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req dto.RegisterRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.WriteError(h.logger, w, err)
		return
	}

	user, err := h.authService.Register(r.Context(), req)
	if err != nil {
		httpx.WriteError(h.logger, w, err)
		return
	}

	resp := dto.RegisterResponse{
		Message: "user registered successfully",
		User:    toUserProfileResponse(user),
	}

	httpx.WriteJSON(w, http.StatusCreated, resp)
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req dto.LoginRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.WriteError(h.logger, w, err)
		return
	}

	result, err := h.authService.Login(r.Context(), req, r.UserAgent(), httpx.ClientIP(r))
	if err != nil {
		httpx.WriteError(h.logger, w, err)
		return
	}

	resp := dto.TokenResponse{
		AccessToken:  result.AccessToken,
		RefreshToken: result.RefreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    result.ExpiresIn,
		User: &dto.AuthUserResponse{
			ID:        result.User.ID.String(),
			Email:     result.User.Email,
			FirstName: result.User.FirstName,
			LastName:  result.User.LastName,
			Role:      result.User.Role,
		},
	}

	httpx.WriteJSON(w, http.StatusOK, resp)
}

func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	var req dto.RefreshRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.WriteError(h.logger, w, err)
		return
	}

	result, err := h.authService.Refresh(r.Context(), req.RefreshToken, r.UserAgent(), httpx.ClientIP(r))
	if err != nil {
		httpx.WriteError(h.logger, w, err)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, dto.TokenResponse{
		AccessToken:  result.AccessToken,
		RefreshToken: result.RefreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    result.ExpiresIn,
	})
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	var req dto.LogoutRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.WriteError(h.logger, w, err)
		return
	}

	if err := h.authService.Logout(r.Context(), req.RefreshToken); err != nil {
		httpx.WriteError(h.logger, w, err)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, dto.MessageResponse{Message: "logged out successfully"})
}

func (h *AuthHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	userID, ok := ctxkeys.UserID(r.Context())
	if !ok {
		httpx.WriteError(h.logger, w, apperror.New(apperror.CodeUnauthorized, "unauthorized", http.StatusUnauthorized))
		return
	}

	var req dto.ChangePasswordRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.WriteError(h.logger, w, err)
		return
	}

	if err := h.authService.ChangePassword(r.Context(), userID, req.OldPassword, req.NewPassword); err != nil {
		httpx.WriteError(h.logger, w, err)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, dto.MessageResponse{Message: "password changed successfully"})
}
