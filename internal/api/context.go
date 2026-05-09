package api

import (
	"context"
	"net/http"
)

type contextKey string

const userIDKey contextKey = "user_id"

func WithUserID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, userIDKey, id)
}

func UserIDFromContext(ctx context.Context) string {
	v, _ := ctx.Value(userIDKey).(string)
	return v
}

func UserIDFromRequest(r *http.Request) string {
	return UserIDFromContext(r.Context())
}
