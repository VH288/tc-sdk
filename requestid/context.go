package requestid

import "context"

const (
	HeaderRequestID     = "X-Request-ID"
	ContextKeyRequestID = "request_id"
)

func RequestIDFromContext(ctx context.Context) string {
	if v, ok := ctx.Value(ContextKeyRequestID).(string); ok {
		return v
	}
	return ""
}

func WithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, ContextKeyRequestID, requestID)
}
