package middleware

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"runtime/debug"
	"strconv"
	"time"

	"github.com/felixge/httpsnoop"
	"github.com/go-chi/chi/v5"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/noop"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"

	"github.com/Chillobass/fork_template/internal/http/domainerror"
)

type HTTP struct {
	logger   *slog.Logger
	tracer   trace.TracerProvider
	duration metric.Float64Histogram
	active   metric.Int64UpDownCounter
}

func New(logger *slog.Logger, tracer trace.TracerProvider, meter metric.Meter) (*HTTP, error) {
	duration, err := meter.Float64Histogram("http.server.request.duration",
		metric.WithUnit("s"),
		metric.WithExplicitBucketBoundaries(0.005, 0.01, 0.025, 0.05, 0.075, 0.1, 0.25, 0.5, 1, 2.5, 5, 10))
	if err != nil {
		return nil, fmt.Errorf("register HTTP duration: %w", err)
	}
	active, err := meter.Int64UpDownCounter("http.server.active_requests", metric.WithUnit("{request}"))
	if err != nil {
		return nil, fmt.Errorf("register active HTTP requests: %w", err)
	}
	return &HTTP{logger: logger, tracer: tracer, duration: duration, active: active}, nil
}

type requestKey struct{}

type observation struct {
	route   string
	started time.Time
	status  int
	ctx     context.Context
	failure error
}

func (o *observation) statusCode() int {
	if o.status != 0 {
		return o.status
	}
	return http.StatusInternalServerError
}

func (m *HTTP) Wrap(router *chi.Mux) http.Handler {
	return m.recover(router, m.tracing(m.logging(m.metrics(router))))
}

func (m *HTTP) recover(router chi.Routes, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		o := &observation{route: routePattern(router, r), started: time.Now(), ctx: r.Context()}
		ctx := context.WithValue(r.Context(), requestKey{}, o)
		ctx = domainerror.WithReport(ctx, func(err error) { o.failure = err })

		wrapped := httpsnoop.Wrap(w, httpsnoop.Hooks{
			WriteHeader: func(next httpsnoop.WriteHeaderFunc) httpsnoop.WriteHeaderFunc {
				return func(code int) {
					if o.status == 0 && (code >= 200 || code == http.StatusSwitchingProtocols) {
						o.status = code
					}
					next(code)
				}
			},
			Write: func(next httpsnoop.WriteFunc) httpsnoop.WriteFunc {
				return func(body []byte) (int, error) {
					o.setOK()
					return next(body)
				}
			},
			ReadFrom: func(next httpsnoop.ReadFromFunc) httpsnoop.ReadFromFunc {
				return func(reader io.Reader) (int64, error) {
					o.setOK()
					return next(reader)
				}
			},
			Flush: func(next httpsnoop.FlushFunc) httpsnoop.FlushFunc {
				return func() {
					o.setOK()
					next()
				}
			},
		})
		defer func() {
			p := recover()
			if p == nil {
				return
			}
			if p == http.ErrAbortHandler {
				panic(p)
			}
			m.logRequest(r, o, slog.LevelError, "panic_type", fmt.Sprintf("%T", p), "stack", string(debug.Stack()))
			if o.status != 0 {
				panic(http.ErrAbortHandler)
			}
			domainerror.Write(w, r, fmt.Errorf("panic in HTTP handler"))
		}()
		next.ServeHTTP(wrapped, r.WithContext(ctx))
	})
}

func (m *HTTP) tracing(next http.Handler) http.Handler {
	traced := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		o := observed(r)
		o.ctx = r.Context()
		span := trace.SpanFromContext(o.ctx)
		span.SetAttributes(attribute.String("http.route", o.route))
		defer func() {
			status := o.statusCode()
			span.SetAttributes(attribute.Int("http.response.status_code", status))
			if status >= http.StatusBadRequest {
				span.SetStatus(codes.Error, "")
				span.SetAttributes(attribute.String("error.type", strconv.Itoa(status)))
			}
		}()
		next.ServeHTTP(w, r)
	})
	return otelhttp.NewHandler(traced, "http.request",
		otelhttp.WithTracerProvider(m.tracer),
		otelhttp.WithMeterProvider(noop.NewMeterProvider()),
		otelhttp.WithPropagators(propagation.TraceContext{}),
		otelhttp.WithSpanNameFormatter(func(_ string, r *http.Request) string {
			return r.Method + " " + observed(r).route
		}),
	)
}

func (m *HTTP) logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
		o := observed(r)
		level := slog.LevelInfo
		if o.statusCode() >= http.StatusInternalServerError {
			level = slog.LevelError
		}
		if o.failure != nil {
			m.logRequest(r, o, level, "error", o.failure)
			return
		}
		m.logRequest(r, o, level)
	})
}

func (m *HTTP) logRequest(r *http.Request, o *observation, level slog.Level, details ...any) {
	fields := make([]any, 0, 8+len(details))
	fields = append(fields, "method", r.Method, "route", o.route, "status", o.statusCode(),
		"duration_ms", time.Since(o.started).Milliseconds())
	m.logger.Log(o.ctx, level, "request handled", append(fields, details...)...)
}

func (m *HTTP) metrics(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		o := observed(r)
		ctx := r.Context()
		method := attribute.String("http.request.method", r.Method)
		route := attribute.String("http.route", o.route)
		activeLabels := metric.WithAttributeSet(attribute.NewSet(method, route))
		m.active.Add(ctx, 1, activeLabels)
		defer func() {
			m.active.Add(ctx, -1, activeLabels)
			m.duration.Record(ctx, time.Since(o.started).Seconds(),
				metric.WithAttributeSet(attribute.NewSet(method, route, attribute.Int("http.response.status_code", o.statusCode()))))
		}()
		next.ServeHTTP(w, r)
		o.setOK()
	})
}

func (o *observation) setOK() {
	if o.status == 0 {
		o.status = http.StatusOK
	}
}

func observed(r *http.Request) *observation {
	return r.Context().Value(requestKey{}).(*observation)
}

func routePattern(router chi.Routes, r *http.Request) string {
	path := r.URL.RawPath
	if path == "" {
		path = r.URL.Path
	}
	if pattern := router.Find(chi.NewRouteContext(), r.Method, path); pattern != "" {
		return pattern
	}
	return "<unmatched>"
}
