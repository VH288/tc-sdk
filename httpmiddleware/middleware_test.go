package httpmiddleware

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/VH288/tc-sdk/logger"
	"github.com/VH288/tc-sdk/requestid"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"go.opentelemetry.io/otel/trace"
)

func TestRequestIDAndLogging(t *testing.T) {
	for _, supplied := range []string{"client-42", ""} {
		t.Run("header="+supplied, func(t *testing.T) {
			var logs bytes.Buffer
			log := logger.NewWithWriter(logger.Config{}, &logs)
			var seen string
			handler := RequestID(Logging(log)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				seen = requestid.RequestIDFromContext(r.Context())
				w.WriteHeader(http.StatusCreated)
			})))
			req := httptest.NewRequest(http.MethodPost, "/test/42?query=value", nil)
			req.Header.Set(requestid.HeaderRequestID, supplied)
			response := httptest.NewRecorder()
			handler.ServeHTTP(response, req)

			if supplied != "" && seen != supplied {
				t.Fatalf("handler request ID = %q, want %q", seen, supplied)
			}
			if supplied == "" {
				if _, err := uuid.Parse(seen); err != nil {
					t.Fatalf("generated request ID %q is not a UUID: %v", seen, err)
				}
			}
			if got := response.Header().Get(requestid.HeaderRequestID); got != seen {
				t.Fatalf("response request ID = %q, handler saw %q", got, seen)
			}
			if response.Code != http.StatusCreated {
				t.Fatalf("status = %d", response.Code)
			}
			var entry map[string]any
			if err := json.Unmarshal(logs.Bytes(), &entry); err != nil {
				t.Fatalf("decode log: %v", err)
			}
			for key, want := range map[string]any{"msg": "request completed", "request_id": seen, "method": "POST", "path": "/test/42", "status": float64(201)} {
				if entry[key] != want {
					t.Errorf("log %s = %v, want %v", key, entry[key], want)
				}
			}
			if _, ok := entry["latency"]; !ok {
				t.Error("completion log missing latency")
			}
		})
	}
}

func TestTracing(t *testing.T) {
	oldProvider := otel.GetTracerProvider()
	oldPropagator := otel.GetTextMapPropagator()
	t.Cleanup(func() {
		otel.SetTracerProvider(oldProvider)
		otel.SetTextMapPropagator(oldPropagator)
	})
	recorder := tracetest.NewSpanRecorder()
	provider := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(recorder))
	otel.SetTracerProvider(provider)
	otel.SetTextMapPropagator(propagation.TraceContext{})

	var valid bool
	handler := Tracing("test-service")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		valid = trace.SpanFromContext(r.Context()).SpanContext().IsValid()
		w.WriteHeader(http.StatusCreated)
	}))
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/test/42", nil))
	if !valid {
		t.Fatal("handler did not receive a valid span")
	}
	spans := recorder.Ended()
	if len(spans) != 1 {
		t.Fatalf("ended spans = %d, want 1", len(spans))
	}
	if spans[0].Name() != "GET /test/42" {
		t.Errorf("span name = %q", spans[0].Name())
	}
	attrs := make(map[attribute.Key]attribute.Value)
	for _, attr := range spans[0].Attributes() {
		attrs[attr.Key] = attr.Value
	}
	for key, want := range map[attribute.Key]attribute.Value{
		"http.method":      attribute.StringValue("GET"),
		"http.route":       attribute.StringValue("/test/42"),
		"http.status_code": attribute.IntValue(http.StatusCreated),
	} {
		if attrs[key] != want {
			t.Errorf("attribute %s = %v, want %v", key, attrs[key], want)
		}
	}
}
