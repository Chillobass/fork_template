package middleware_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"

	"github.com/Chillobass/fork_template/internal/http/domainerror"
	"github.com/Chillobass/fork_template/internal/http/middleware"
)

func TestRequestTelemetry(t *testing.T) {
	const pattern = "/api/v1/trips/{tripId}"
	for _, tc := range []struct {
		name    string
		handler http.HandlerFunc
		status  int
		level   string
	}{
		{name: "implicit success", handler: func(http.ResponseWriter, *http.Request) {}, status: http.StatusOK, level: "INFO"},
		{name: "client error", handler: func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		}, status: http.StatusNotFound, level: "INFO"},
		{name: "internal error", handler: func(w http.ResponseWriter, r *http.Request) {
			domainerror.Write(w, r, errors.New("connection refused"))
		}, status: http.StatusInternalServerError, level: "ERROR"},
		{name: "panic", handler: func(http.ResponseWriter, *http.Request) {
			panic("private panic value")
		}, status: http.StatusInternalServerError, level: "ERROR"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var logs bytes.Buffer
			spans := tracetest.NewSpanRecorder()
			reader := sdkmetric.NewManualReader()
			tracer := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(spans))
			meter := sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader))
			m, err := middleware.New(slog.New(slog.NewJSONHandler(&logs, nil)), tracer, meter.Meter("test"))
			if err != nil {
				t.Fatal(err)
			}
			router := chi.NewRouter()
			router.Get(pattern, tc.handler)
			rec := httptest.NewRecorder()
			m.Wrap(router).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/trips/example-id", nil))

			if rec.Code != tc.status {
				t.Fatalf("status = %d, want %d", rec.Code, tc.status)
			}
			var logEntry struct {
				Route  string
				Status int
				Level  string
				Error  string
				Stack  string
			}
			if err := json.Unmarshal(logs.Bytes(), &logEntry); err != nil {
				t.Fatalf("expected one request log, got %s: %v", logs.Bytes(), err)
			}
			if logEntry.Route != pattern || logEntry.Status != tc.status || logEntry.Level != tc.level {
				t.Fatalf("unexpected request log: %s", logs.Bytes())
			}
			if tc.name == "internal error" && logEntry.Error != "connection refused" {
				t.Fatalf("original error missing from log: %s", logs.Bytes())
			}
			if tc.name == "panic" && (logEntry.Stack == "" || !strings.Contains(rec.Body.String(), "internal_error")) {
				t.Fatalf("panic stack or problem missing: %s, %s", logs.Bytes(), rec.Body)
			}

			ended := spans.Ended()
			if len(ended) != 1 || ended[0].Name() != "GET "+pattern {
				t.Fatalf("unexpected spans: %v", ended)
			}
			var data metricdata.ResourceMetrics
			if err := reader.Collect(t.Context(), &data); err != nil {
				t.Fatal(err)
			}
			checkMetrics(t, data, pattern, tc.status)
		})
	}
}

func checkMetrics(t *testing.T, data metricdata.ResourceMetrics, pattern string, status int) {
	t.Helper()
	if len(data.ScopeMetrics) != 1 || len(data.ScopeMetrics[0].Metrics) != 2 {
		t.Fatalf("expected two instruments, got %+v", data.ScopeMetrics)
	}
	for _, metric := range data.ScopeMetrics[0].Metrics {
		switch points := metric.Data.(type) {
		case metricdata.Histogram[float64]:
			if len(points.DataPoints) != 1 {
				t.Fatalf("unexpected duration points: %+v", points)
			}
			point := points.DataPoints[0]
			route, _ := point.Attributes.Value("http.route")
			code, _ := point.Attributes.Value("http.response.status_code")
			if route.AsString() != pattern || code.AsInt64() != int64(status) || point.Count != 1 {
				t.Fatalf("unexpected duration point: %+v", point)
			}
		case metricdata.Sum[int64]:
			if len(points.DataPoints) != 1 || points.DataPoints[0].Value != 0 {
				t.Fatalf("active requests did not return to zero: %+v", points)
			}
		default:
			t.Fatalf("unexpected metric: %+v", metric)
		}
	}
}
