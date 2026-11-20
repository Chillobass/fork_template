package push

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/metric/noop"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

const (
	pushEndpointPath      = "/api/v1/push"
	kindRequestPosition   = "REQUEST_POSITION"
	maxResponseDrainBytes = 4 << 10
	defaultIdleConns      = 64
	defaultIdleTimeout    = 90 * time.Second
)

var (
	ErrClientStatus     = errors.New("push service rejected request with 4xx status")
	ErrServerStatus     = errors.New("push service failed with 5xx status")
	ErrUnexpectedStatus = errors.New("push service returned unexpected status")
)

type StatusError struct {
	StatusCode int
	RetryAfter string
	cause      error
}

func (e *StatusError) Error() string {
	if e.RetryAfter != "" {
		return fmt.Sprintf("push service status %d (Retry-After: %s)", e.StatusCode, e.RetryAfter)
	}
	return fmt.Sprintf("push service status %d", e.StatusCode)
}

func (e *StatusError) Unwrap() error {
	return e.cause
}

type Client struct {
	endpoint   string
	transport  *http.Transport
	httpClient *http.Client
}

func New(baseURL string, timeout time.Duration, tracerProvider trace.TracerProvider) *Client {
	transport := &http.Transport{
		DialContext: (&net.Dialer{
			Timeout:   timeout,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		MaxIdleConns:          defaultIdleConns,
		MaxIdleConnsPerHost:   defaultIdleConns,
		IdleConnTimeout:       defaultIdleTimeout,
		TLSHandshakeTimeout:   timeout,
		ResponseHeaderTimeout: timeout,
	}

	instrumented := otelhttp.NewTransport(
		transport,
		otelhttp.WithTracerProvider(tracerProvider),
		otelhttp.WithMeterProvider(noop.NewMeterProvider()),
		otelhttp.WithPropagators(propagation.TraceContext{}),
		otelhttp.WithSpanNameFormatter(func(_ string, r *http.Request) string {
			return "push " + r.Method + " " + pushEndpointPath
		}),
	)

	return &Client{
		endpoint:  strings.TrimRight(baseURL, "/") + pushEndpointPath,
		transport: transport,
		httpClient: &http.Client{
			Transport: instrumented,
			Timeout:   timeout,
		},
	}
}

type pushRequest struct {
	RequestID   string            `json:"request_id"`
	RecipientID string            `json:"recipient_id"`
	Kind        string            `json:"kind"`
	Data        map[string]string `json:"data"`
}

func (c *Client) RequestPosition(ctx context.Context, tripID, driverID uuid.UUID) error {
	payload, err := json.Marshal(pushRequest{
		RequestID:   uuid.NewString(),
		RecipientID: driverID.String(),
		Kind:        kindRequestPosition,
		Data: map[string]string{
			"trip_id": tripID.String(),
		},
	})
	if err != nil {
		return fmt.Errorf("marshal push request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpoint, bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("create push request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("send push request: %w", err)
	}
	defer func() {
		_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, maxResponseDrainBytes))
		_ = resp.Body.Close()
	}()

	switch {
	case resp.StatusCode == http.StatusAccepted:
		return nil
	case resp.StatusCode >= http.StatusBadRequest && resp.StatusCode < http.StatusInternalServerError:
		return &StatusError{
			StatusCode: resp.StatusCode,
			RetryAfter: resp.Header.Get("Retry-After"),
			cause:      ErrClientStatus,
		}
	case resp.StatusCode >= http.StatusInternalServerError:
		return &StatusError{
			StatusCode: resp.StatusCode,
			cause:      ErrServerStatus,
		}
	default:
		return &StatusError{
			StatusCode: resp.StatusCode,
			cause:      ErrUnexpectedStatus,
		}
	}
}

func (c *Client) Close() {
	c.transport.CloseIdleConnections()
}
