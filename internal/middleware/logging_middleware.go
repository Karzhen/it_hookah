package middleware

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/karzhen/restaurant-lk/internal/utils/ctxkeys"
	"github.com/karzhen/restaurant-lk/internal/utils/httpx"
)

type statusRecorder struct {
	http.ResponseWriter
	statusCode int
}

func (r *statusRecorder) WriteHeader(statusCode int) {
	r.statusCode = statusCode
	r.ResponseWriter.WriteHeader(statusCode)
}

func LoggingMiddleware(logger *slog.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			recorder := &statusRecorder{ResponseWriter: w, statusCode: http.StatusOK}

			next.ServeHTTP(recorder, r)

			requestID, _ := ctxkeys.RequestID(r.Context())
			logger.Info("http request",
				"request_id", requestID,
				"method", r.Method,
				"path", r.URL.Path,
				"status", recorder.statusCode,
				"duration_ms", time.Since(start).Milliseconds(),
				"ip", httpx.ClientIP(r),
			)
		})
	}
}
