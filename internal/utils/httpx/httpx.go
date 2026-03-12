package httpx

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net"
	"net/http"
	"strings"

	"github.com/karzhen/restaurant-lk/internal/dto"
	"github.com/karzhen/restaurant-lk/internal/utils/apperror"
)

func DecodeJSON(r *http.Request, dst any) error {
	decoder := json.NewDecoder(io.LimitReader(r.Body, 1<<20))
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(dst); err != nil {
		return apperror.New(apperror.CodeValidationError, "invalid request body", http.StatusBadRequest)
	}

	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		return apperror.New(apperror.CodeValidationError, "request body must contain a single JSON object", http.StatusBadRequest)
	}

	return nil
}

func WriteJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func WriteError(logger *slog.Logger, w http.ResponseWriter, err error) {
	var appErr *apperror.AppError
	if !errors.As(err, &appErr) {
		logger.Error("internal error", "error", err)
		appErr = apperror.New(apperror.CodeInternal, "internal server error", http.StatusInternalServerError)
	} else if appErr.Status >= 500 && appErr.Err != nil {
		logger.Error("internal error", "error", appErr.Err)
	}

	WriteJSON(w, appErr.Status, dto.ErrorResponse{
		Error: dto.ErrorBody{
			Code:    appErr.Code,
			Message: appErr.Message,
		},
	})
}

func ClientIP(r *http.Request) string {
	if xff := strings.TrimSpace(r.Header.Get("X-Forwarded-For")); xff != "" {
		parts := strings.Split(xff, ",")
		if len(parts) > 0 {
			return strings.TrimSpace(parts[0])
		}
	}

	if xrip := strings.TrimSpace(r.Header.Get("X-Real-IP")); xrip != "" {
		return xrip
	}

	host, _, err := net.SplitHostPort(strings.TrimSpace(r.RemoteAddr))
	if err != nil {
		return strings.TrimSpace(r.RemoteAddr)
	}

	return host
}
