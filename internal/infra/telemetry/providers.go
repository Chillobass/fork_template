package telemetry

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"runtime/debug"
	"strings"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

type Providers struct {
	Traces  *sdktrace.TracerProvider
	Metrics *sdkmetric.MeterProvider
}

func New(ctx context.Context, service string, logger *slog.Logger) Providers {
	otel.SetErrorHandler(otel.ErrorHandlerFunc(func(err error) {
		logger.WarnContext(ctx, "telemetry export or collection failed", "error", err)
	}))

	res, err := resource.New(ctx,
		resource.WithAttributes(
			attribute.String("service.name", service),
			attribute.String("service.namespace", "tripgo"),
			attribute.String("service.version", buildVersion()),
			attribute.String("deployment.environment", "local"),
		),
		resource.WithFromEnv(),
	)
	if err != nil {
		logger.WarnContext(ctx, "some telemetry resource attributes were unavailable", "error", err)
	}

	traces := []sdktrace.TracerProviderOption{sdktrace.WithResource(res)}
	metrics := []sdkmetric.Option{sdkmetric.WithResource(res)}

	protocol := os.Getenv("OTEL_EXPORTER_OTLP_PROTOCOL")
	switch protocol {
	case "", "http/protobuf", "grpc":
		traceExporter, metricExporter := exporters(ctx, protocol == "grpc", logger)
		if traceExporter != nil {
			traces = append(traces, sdktrace.WithBatcher(traceExporter))
		}
		if metricExporter != nil {
			metrics = append(metrics, sdkmetric.WithReader(sdkmetric.NewPeriodicReader(metricExporter)))
		}
	default:
		logger.WarnContext(ctx, "unsupported OTLP protocol, export disabled", "protocol", protocol)
	}

	return Providers{
		Traces:  sdktrace.NewTracerProvider(traces...),
		Metrics: sdkmetric.NewMeterProvider(metrics...),
	}
}

func exporters(ctx context.Context, grpc bool, logger *slog.Logger) (sdktrace.SpanExporter, sdkmetric.Exporter) {
	var (
		traces    sdktrace.SpanExporter
		metrics   sdkmetric.Exporter
		traceErr  error
		metricErr error
	)
	if grpc {
		traces, traceErr = otlptracegrpc.New(ctx)
		metrics, metricErr = otlpmetricgrpc.New(ctx)
	} else {
		traces, traceErr = otlptracehttp.New(ctx)
		metrics, metricErr = otlpmetrichttp.New(ctx)
	}
	if traceErr != nil {
		logger.WarnContext(ctx, "could not start trace exporter", "error", traceErr)
		traces = nil
	}
	if metricErr != nil {
		logger.WarnContext(ctx, "could not start metric exporter", "error", metricErr)
		metrics = nil
	}
	return traces, metrics
}

func (p Providers) Shutdown(ctx context.Context) error {
	return errors.Join(p.Metrics.Shutdown(ctx), p.Traces.Shutdown(ctx))
}

func buildVersion() string {
	if info, ok := debug.ReadBuildInfo(); ok {
		if info.Main.Version != "" && info.Main.Version != "(devel)" {
			return info.Main.Version
		}
		for _, setting := range info.Settings {
			if setting.Key == "vcs.revision" {
				return strings.TrimSpace(setting.Value)
			}
		}
	}
	return "dev"
}
