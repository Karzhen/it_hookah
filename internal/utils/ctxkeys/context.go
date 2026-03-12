package ctxkeys

import (
	"context"

	"github.com/google/uuid"
)

type contextKey string

const (
	userIDKey    contextKey = "user_id"
	roleKey      contextKey = "role"
	requestIDKey contextKey = "request_id"
)

func WithUserID(ctx context.Context, userID uuid.UUID) context.Context {
	return context.WithValue(ctx, userIDKey, userID)
}

func UserID(ctx context.Context) (uuid.UUID, bool) {
	value, ok := ctx.Value(userIDKey).(uuid.UUID)
	return value, ok
}

func WithRole(ctx context.Context, role string) context.Context {
	return context.WithValue(ctx, roleKey, role)
}

func Role(ctx context.Context) (string, bool) {
	value, ok := ctx.Value(roleKey).(string)
	return value, ok
}

func WithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, requestIDKey, requestID)
}

func RequestID(ctx context.Context) (string, bool) {
	value, ok := ctx.Value(requestIDKey).(string)
	return value, ok
}
