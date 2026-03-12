package middleware

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/karzhen/restaurant-lk/internal/utils/apperror"
	"github.com/karzhen/restaurant-lk/internal/utils/httpx"
)

func RecoveryMiddleware(logger *slog.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rec := recover(); rec != nil {
					logger.Error("panic recovered", "panic", rec)
					httpx.WriteError(logger, w, apperror.Wrap(
						apperror.CodeInternal,
						"internal server error",
						http.StatusInternalServerError,
						fmt.Errorf("panic: %v", rec),
					))
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}
