package push_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"go.uber.org/goleak"

	"github.com/Chillobass/fork_template/internal/gateway/push"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m)
}

func TestRequestPosition(t *testing.T) {
	tripID := uuid.New()
	driverID := uuid.New()

	for _, tc := range []struct {
		name       string
		status     int
		retryAfter string
		wantErr    error
	}{
		{name: "accepted", status: http.StatusAccepted},
		{name: "bad request", status: http.StatusBadRequest, wantErr: push.ErrClientStatus},
		{name: "rate limited", status: http.StatusTooManyRequests, retryAfter: "5", wantErr: push.ErrClientStatus},
		{name: "service unavailable", status: http.StatusServiceUnavailable, wantErr: push.ErrServerStatus},
		{name: "unexpected 200", status: http.StatusOK, wantErr: push.ErrUnexpectedStatus},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var (
				gotPath        string
				gotTraceparent string
				gotBody        struct {
					RequestID   string            `json:"request_id"`
					RecipientID string            `json:"recipient_id"`
					Kind        string            `json:"kind"`
					Data        map[string]string `json:"data"`
				}
			)

			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				gotPath = r.URL.Path
				gotTraceparent = r.Header.Get("Traceparent")
				if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
					t.Errorf("decode body: %v", err)
				}
				if tc.retryAfter != "" {
					w.Header().Set("Retry-After", tc.retryAfter)
				}
				w.WriteHeader(tc.status)
				if tc.status == http.StatusAccepted {
					_, _ = w.Write([]byte(`{"message_id":"m1","accepted_at":"2026-09-01T10:00:00Z"}`))
				}
			}))
			defer srv.Close()

			spans := tracetest.NewSpanRecorder()
			tp := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(spans))
			client := push.New(srv.URL, time.Second, tp)
			defer client.Close()

			err := client.RequestPosition(t.Context(), tripID, driverID)
			if !errors.Is(err, tc.wantErr) {
				t.Fatalf("error = %v, want %v", err, tc.wantErr)
			}
			if gotPath != "/api/v1/push" || gotBody.Kind != "REQUEST_POSITION" ||
				gotBody.RecipientID != driverID.String() || gotBody.Data["trip_id"] != tripID.String() ||
				gotBody.RequestID == "" || gotTraceparent == "" {
				t.Fatalf("unexpected request: path=%q traceparent=%q body=%+v", gotPath, gotTraceparent, gotBody)
			}

			if tc.wantErr != nil {
				var statusErr *push.StatusError
				if !errors.As(err, &statusErr) || statusErr.StatusCode != tc.status || statusErr.RetryAfter != tc.retryAfter {
					t.Fatalf("unexpected status error: %v", err)
				}
			}

			if len(spans.Ended()) != 1 {
				t.Fatalf("expected 1 push span, got %d", len(spans.Ended()))
			}
		})
	}
}

func TestRequestPositionTransportError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	srvURL := srv.URL
	srv.Close()

	tp := sdktrace.NewTracerProvider()
	client := push.New(srvURL, 200*time.Millisecond, tp)
	defer client.Close()

	err := client.RequestPosition(t.Context(), uuid.New(), uuid.New())
	if err == nil {
		t.Fatal("expected transport error, got nil")
	}
	var statusErr *push.StatusError
	if errors.As(err, &statusErr) {
		t.Fatalf("transport error must not be StatusError: %v", err)
	}
}
