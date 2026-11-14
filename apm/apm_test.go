package apm_test

import (
	"context"
	"io"
	"log/slog"
	"net"
	"testing"
	"time"

	"github.com/VH288/tc-sdk/apm"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	collector "go.opentelemetry.io/proto/otlp/collector/trace/v1"
	"google.golang.org/grpc"
)

func restoreGlobals(t *testing.T) {
	t.Helper()
	provider, propagator := otel.GetTracerProvider(), otel.GetTextMapPropagator()
	t.Cleanup(func() { otel.SetTracerProvider(provider); otel.SetTextMapPropagator(propagator) })
}

func TestTracerRecordsScope(t *testing.T) {
	restoreGlobals(t)
	recorder := tracetest.NewSpanRecorder()
	provider := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(recorder))
	t.Cleanup(func() { _ = provider.Shutdown(context.Background()) })
	otel.SetTracerProvider(provider)
	_, span := apm.Tracer("test-service").Start(context.Background(), "test.span")
	span.End()
	spans := recorder.Ended()
	require.Len(t, spans, 1)
	require.Equal(t, "test.span", spans[0].Name())
	require.Equal(t, "test-service", spans[0].InstrumentationScope().Name)
}

type traceCollector struct {
	collector.UnimplementedTraceServiceServer
	requests chan *collector.ExportTraceServiceRequest
}

func (c *traceCollector) Export(_ context.Context, req *collector.ExportTraceServiceRequest) (*collector.ExportTraceServiceResponse, error) {
	c.requests <- req
	return &collector.ExportTraceServiceResponse{}, nil
}

func TestInitExportsSpans(t *testing.T) {
	restoreGlobals(t)
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	t.Cleanup(func() { _ = listener.Close() })
	server := grpc.NewServer()
	receiver := &traceCollector{requests: make(chan *collector.ExportTraceServiceRequest, 1)}
	collector.RegisterTraceServiceServer(server, receiver)
	go func() { _ = server.Serve(listener) }()
	t.Cleanup(server.Stop)
	provider, err := apm.Init(context.Background(), apm.Config{TempoEndpoint: listener.Addr().String(), ServiceName: "sdk-test", TraceSampleRatio: 1}, slog.New(slog.NewTextHandler(io.Discard, nil)))
	require.NoError(t, err)
	_, span := apm.Tracer("test-service").Start(context.Background(), "test.span")
	span.End()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	require.NoError(t, provider.Shutdown(ctx))
	select {
	case req := <-receiver.requests:
		require.Len(t, req.ResourceSpans, 1)
		resource := req.ResourceSpans[0]
		found := false
		for _, attr := range resource.Resource.Attributes {
			if attr.Key == "service.name" && attr.Value.GetStringValue() == "sdk-test" {
				found = true
			}
		}
		require.True(t, found, "service.name missing")
		require.Len(t, resource.ScopeSpans, 1)
		require.Len(t, resource.ScopeSpans[0].Spans, 1)
		require.Equal(t, "test.span", resource.ScopeSpans[0].Spans[0].Name)
	case <-ctx.Done():
		t.Fatal("collector did not receive span")
	}
}

func TestNilProviderShutdown(t *testing.T) {
	var provider *apm.Provider
	require.NoError(t, provider.Shutdown(context.Background()))
}
