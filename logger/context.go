package logger

import (
	"context"
	"log/slog"
)

type ctxLoggerKey struct{}

func WithContext(ctx context.Context, log *slog.Logger) context.Context {
	if log == nil {
		return ctx
	}
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithValue(ctx, ctxLoggerKey{}, log)
}

func FromContext(ctx context.Context) *slog.Logger {
	if ctx == nil {
		return slog.Default()
	}
	if v, ok := ctx.Value(ctxLoggerKey{}).(*slog.Logger); ok && v != nil {
		return v
	}
	return slog.Default()
}
