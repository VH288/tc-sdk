package requestid

import (
	"context"
	"testing"
)

func TestRequestIDFromContext(t *testing.T) {
	ctx := context.Background()
	if got := RequestIDFromContext(ctx); got != "" {
		t.Fatalf("missing request ID = %q", got)
	}
	if got := RequestIDFromContext(context.WithValue(ctx, ContextKeyRequestID, 42)); got != "" {
		t.Fatalf("wrong-type request ID = %q", got)
	}
	if got := RequestIDFromContext(context.WithValue(ctx, "request_id", "existing")); got != "existing" {
		t.Fatalf("existing string key request ID = %q", got)
	}
	if got := RequestIDFromContext(WithRequestID(ctx, "client-42")); got != "client-42" {
		t.Fatalf("request ID = %q", got)
	}
}
