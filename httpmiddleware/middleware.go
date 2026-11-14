package httpmiddleware

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/VH288/tc-sdk/logger"
	"github.com/VH288/tc-sdk/requestid"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
)

type statusWriter struct {
	http.ResponseWriter
	status int
}

func (w *statusWriter) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqID := r.Header.Get(requestid.HeaderRequestID)
		if reqID == "" {
			reqID = uuid.NewString()
		}
		w.Header().Set(requestid.HeaderRequestID, reqID)
		ctx := requestid.WithRequestID(r.Context(), reqID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func Logging(base *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			sw := &statusWriter{ResponseWriter: w, status: http.StatusOK}

			reqID := requestid.RequestIDFromContext(r.Context())
			reqLogger := base.With(
				slog.String("request_id", reqID),
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
			)
			ctx := logger.WithContext(r.Context(), reqLogger)
			next.ServeHTTP(sw, r.WithContext(ctx))
			reqLogger.Info("request completed",
				slog.Int("status", sw.status),
				slog.Duration("latency", time.Since(start)),
			)
		})
	}
}

func Tracing(serviceName string) func(http.Handler) http.Handler {
	tracer := otel.Tracer(serviceName)
	propagator := otel.GetTextMapPropagator()

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := propagator.Extract(r.Context(), propagation.HeaderCarrier(r.Header))
			ctx, span := tracer.Start(ctx, r.Method+" "+r.URL.Path)
			defer span.End()

			span.SetAttributes(
				attribute.String("http.method", r.Method),
				attribute.String("http.route", r.URL.Path),
			)

			sw := &statusWriter{ResponseWriter: w, status: http.StatusOK}
			next.ServeHTTP(sw, r.WithContext(ctx))
			span.SetAttributes(attribute.Int("http.status_code", sw.status))
		})
	}
}
