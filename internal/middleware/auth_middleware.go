package middleware

import (
	"context"
	"log/slog"
	"net/http"
	"strings"

	"github.com/google/uuid"

	"github.com/karzhen/restaurant-lk/internal/auth"
	"github.com/karzhen/restaurant-lk/internal/domain"
	"github.com/karzhen/restaurant-lk/internal/utils/apperror"
	"github.com/karzhen/restaurant-lk/internal/utils/ctxkeys"
	"github.com/karzhen/restaurant-lk/internal/utils/httpx"
)

type UserProvider interface {
	GetByID(ctx context.Context, userID uuid.UUID) (*domain.User, error)
}

type AuthMiddleware struct {
	jwtManager *auth.JWTManager
	users      UserProvider
	logger     *slog.Logger
}

func NewAuthMiddleware(jwtManager *auth.JWTManager, users UserProvider, logger *slog.Logger) *AuthMiddleware {
	return &AuthMiddleware{
		jwtManager: jwtManager,
		users:      users,
		logger:     logger,
	}
}

func (m *AuthMiddleware) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := bearerToken(r.Header.Get("Authorization"))
		if token == "" {
			httpx.WriteError(m.logger, w, apperror.New(apperror.CodeUnauthorized, "unauthorized", http.StatusUnauthorized))
			return
		}

		claims, err := m.jwtManager.Parse(token)
		if err != nil {
			httpx.WriteError(m.logger, w, apperror.New(apperror.CodeUnauthorized, "unauthorized", http.StatusUnauthorized))
			return
		}

		userID, err := uuid.Parse(claims.UserID)
		if err != nil {
			httpx.WriteError(m.logger, w, apperror.New(apperror.CodeUnauthorized, "unauthorized", http.StatusUnauthorized))
			return
		}

		user, err := m.users.GetByID(r.Context(), userID)
		if err != nil || !user.IsActive {
			httpx.WriteError(m.logger, w, apperror.New(apperror.CodeUnauthorized, "unauthorized", http.StatusUnauthorized))
			return
		}

		ctx := ctxkeys.WithUserID(r.Context(), user.ID)
		ctx = ctxkeys.WithRole(ctx, claims.Role)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (m *AuthMiddleware) RequireRole(role string) func(next http.Handler) http.Handler {
	requiredRole := strings.ToLower(strings.TrimSpace(role))
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			currentRole, ok := ctxkeys.Role(r.Context())
			if !ok || strings.TrimSpace(currentRole) == "" {
				httpx.WriteError(m.logger, w, apperror.New(apperror.CodeUnauthorized, "unauthorized", http.StatusUnauthorized))
				return
			}

			if !strings.EqualFold(currentRole, requiredRole) {
				httpx.WriteError(m.logger, w, apperror.New(apperror.CodeForbidden, "insufficient permissions", http.StatusForbidden))
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func bearerToken(value string) string {
	parts := strings.SplitN(strings.TrimSpace(value), " ", 2)
	if len(parts) != 2 {
		return ""
	}
	if !strings.EqualFold(parts[0], "Bearer") {
		return ""
	}
	return strings.TrimSpace(parts[1])
}
