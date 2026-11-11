package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/exaring/otelpgx"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/noop"
	"go.opentelemetry.io/otel/trace"
)

type queryTracer struct {
	tracer   *otelpgx.Tracer
	duration metric.Float64Histogram
}

func NewTracer(provider trace.TracerProvider, meter metric.Meter) (pgx.QueryTracer, error) {
	duration, err := meter.Float64Histogram("db.client.operation.duration",
		metric.WithUnit("s"),
		metric.WithExplicitBucketBoundaries(0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5))
	if err != nil {
		return nil, fmt.Errorf("register database duration: %w", err)
	}
	return &queryTracer{
		tracer: otelpgx.NewTracer(
			otelpgx.WithTracerProvider(provider),
			// otelpgx lacks per-query collection labels on its histogram.
			otelpgx.WithMeterProvider(noop.NewMeterProvider()),
			otelpgx.WithDisableConnectionDetailsInAttributes(),
		),
		duration: duration,
	}, nil
}

type queryKey struct{}

type queryInfo struct {
	started    time.Time
	operation  attribute.KeyValue
	collection attribute.KeyValue
}

func (t *queryTracer) TraceQueryStart(ctx context.Context, conn *pgx.Conn, data pgx.TraceQueryStartData) context.Context {
	op := otelpgx.SQLOperationName(ctx, data.SQL)
	info := queryInfo{
		started:    time.Now(),
		operation:  attribute.String("db.operation.name", op),
		collection: attribute.String("db.collection.name", collection(data.SQL)),
	}
	ctx = t.tracer.TraceQueryStart(ctx, conn, data)
	trace.SpanFromContext(ctx).SetAttributes(info.operation, info.collection)
	return context.WithValue(ctx, queryKey{}, info)
}

func (t *queryTracer) TraceQueryEnd(ctx context.Context, conn *pgx.Conn, data pgx.TraceQueryEndData) {
	info := ctx.Value(queryKey{}).(queryInfo)
	attrs := make([]attribute.KeyValue, 0, 3)
	attrs = append(attrs, info.operation, info.collection)
	if data.Err != nil && !errors.Is(data.Err, pgx.ErrNoRows) {
		errorType := attribute.String("error.type", dbErrorType(data.Err))
		attrs = append(attrs, errorType)
		trace.SpanFromContext(ctx).SetAttributes(errorType)
	}
	t.duration.Record(ctx, time.Since(info.started).Seconds(), metric.WithAttributeSet(attribute.NewSet(attrs...)))
	t.tracer.TraceQueryEnd(ctx, conn, data)
}

func dbErrorType(err error) string {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code
	}
	return fmt.Sprintf("%T", err)
}

// The service emits single-table SQL. Keep only known table tokens to bound cardinality.
func collection(sql string) string {
	for token := range strings.FieldsSeq(sql) {
		switch table := strings.Trim(token, "(),\""); table {
		case "trips", "trip_status_history":
			return table
		}
	}
	return ""
}
