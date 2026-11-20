package telemetry

import (
	"context"
	"log/slog"
	"os"

	"go.opentelemetry.io/otel/trace"
)

func NewLogger(service string, level slog.Level) *slog.Logger {
	json := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level})
	return slog.New(correlated{json}).With("service", service)
}

type correlated struct {
	slog.Handler
}

func (h correlated) Handle(ctx context.Context, record slog.Record) error {
	var traceID, spanID string
	if span := trace.SpanContextFromContext(ctx); span.IsValid() {
		traceID, spanID = span.TraceID().String(), span.SpanID().String()
	}
	record.AddAttrs(slog.String("trace_id", traceID), slog.String("span_id", spanID))
	return h.Handler.Handle(ctx, record)
}

func (h correlated) WithAttrs(attrs []slog.Attr) slog.Handler {
	return correlated{h.Handler.WithAttrs(attrs)}
}

func (h correlated) WithGroup(name string) slog.Handler {
	return correlated{h.Handler.WithGroup(name)}
}
