package config

import (
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	HTTP            HTTPConfig
	Database        DatabaseConfig
	Positions       PositionsConfig
	Push            PushConfig
	LogLevel        slog.Level
	ShutdownTimeout time.Duration
}

type HTTPConfig struct {
	Addr              string
	ReadTimeout       time.Duration
	ReadHeaderTimeout time.Duration
	WriteTimeout      time.Duration
	IdleTimeout       time.Duration
}

type DatabaseConfig struct {
	URL             string
	MaxConns        int32
	MinConns        int32
	MaxConnLifetime time.Duration
	ConnectTimeout  time.Duration
	QueryTimeout    time.Duration
}

type PositionsConfig struct {
	CheckInterval time.Duration
	StaleAfter    time.Duration
	Workers       int
	QueueSize     int
	ScanBatchSize int
}

type PushConfig struct {
	HTTPURL        string
	RequestTimeout time.Duration
}

const (
	defaultHTTPReadTimeout       = 10 * time.Second
	defaultHTTPReadHeaderTimeout = 5 * time.Second
	defaultHTTPWriteTimeout      = 15 * time.Second
	defaultHTTPIdleTimeout       = 60 * time.Second

	defaultPositionCheckInterval = 30 * time.Second
	defaultPositionStaleAfter    = 60 * time.Second
	defaultPositionWorkers       = 4
	defaultPositionQueueSize     = 100
	defaultPositionScanBatchSize = 100
	defaultPushRequestTimeout    = 1 * time.Second
)

func Load() (Config, error) {
	r := &envReader{}

	cfg := Config{
		HTTP: HTTPConfig{
			Addr:              r.String("HTTP_ADDR"),
			ReadTimeout:       r.DurationOr("HTTP_READ_TIMEOUT", defaultHTTPReadTimeout),
			ReadHeaderTimeout: r.DurationOr("HTTP_READ_HEADER_TIMEOUT", defaultHTTPReadHeaderTimeout),
			WriteTimeout:      r.DurationOr("HTTP_WRITE_TIMEOUT", defaultHTTPWriteTimeout),
			IdleTimeout:       r.DurationOr("HTTP_IDLE_TIMEOUT", defaultHTTPIdleTimeout),
		},
		Database: DatabaseConfig{
			URL:             r.String("DATABASE_URL"),
			MaxConns:        r.Int32("DATABASE_MAX_CONNS"),
			MinConns:        r.Int32("DATABASE_MIN_CONNS"),
			MaxConnLifetime: r.Duration("DATABASE_MAX_CONN_LIFETIME"),
			ConnectTimeout:  r.Duration("DATABASE_CONNECT_TIMEOUT"),
			QueryTimeout:    r.Duration("DATABASE_QUERY_TIMEOUT"),
		},
		Positions: PositionsConfig{
			CheckInterval: r.DurationOr("POSITION_CHECK_INTERVAL", defaultPositionCheckInterval),
			StaleAfter:    r.DurationOr("POSITION_STALE_AFTER", defaultPositionStaleAfter),
			Workers:       r.IntOr("POSITION_WORKERS", defaultPositionWorkers),
			QueueSize:     r.IntOr("POSITION_QUEUE_SIZE", defaultPositionQueueSize),
			ScanBatchSize: r.IntOr("POSITION_SCAN_BATCH_SIZE", defaultPositionScanBatchSize),
		},
		Push: PushConfig{
			HTTPURL:        r.HTTPURL("PUSH_HTTP_URL"),
			RequestTimeout: r.DurationOr("PUSH_REQUEST_TIMEOUT", defaultPushRequestTimeout),
		},
		LogLevel:        r.LogLevel("LOG_LEVEL"),
		ShutdownTimeout: r.Duration("SHUTDOWN_TIMEOUT"),
	}

	if cfg.Database.MinConns > cfg.Database.MaxConns {
		r.addError("DATABASE_MIN_CONNS", errors.New("must not be greater than DATABASE_MAX_CONNS"))
	}

	if len(r.errs) > 0 {
		return Config{}, fmt.Errorf("invalid configuration: %w", errors.Join(r.errs...))
	}

	return cfg, nil
}

type envReader struct {
	errs []error
}

func (r *envReader) addError(name string, err error) {
	r.errs = append(r.errs, fmt.Errorf("%s: %w", name, err))
}

func (r *envReader) required(name string) (string, bool) {
	value, ok := os.LookupEnv(name)
	if !ok || value == "" {
		r.addError(name, errors.New("is required but not set"))
		return "", false
	}

	return value, true
}

func (r *envReader) String(name string) string {
	value, _ := r.required(name)
	return value
}

func (r *envReader) HTTPURL(name string) string {
	value, ok := r.required(name)
	if !ok {
		return ""
	}

	u, err := url.Parse(value)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
		r.addError(name, fmt.Errorf("must be a valid http(s) URL, got %q", value))
		return ""
	}

	return strings.TrimRight(value, "/")
}

func (r *envReader) Int32(name string) int32 {
	value, ok := r.required(name)
	if !ok {
		return 0
	}

	n, err := strconv.ParseInt(value, 10, 32)
	if err != nil {
		r.addError(name, fmt.Errorf("must be an integer, got %q", value))
		return 0
	}
	if n <= 0 {
		r.addError(name, fmt.Errorf("must be positive, got %d", n))
		return 0
	}

	return int32(n)
}

func (r *envReader) IntOr(name string, def int) int {
	value, ok := os.LookupEnv(name)
	if !ok || value == "" {
		return def
	}

	n, err := strconv.Atoi(value)
	if err != nil {
		r.addError(name, fmt.Errorf("must be an integer, got %q", value))
		return 0
	}
	if n <= 0 {
		r.addError(name, fmt.Errorf("must be positive, got %d", n))
		return 0
	}

	return n
}

func (r *envReader) Duration(name string) time.Duration {
	value, ok := r.required(name)
	if !ok {
		return 0
	}

	return r.parseDuration(name, value)
}

func (r *envReader) DurationOr(name string, def time.Duration) time.Duration {
	value, ok := os.LookupEnv(name)
	if !ok || value == "" {
		return def
	}

	return r.parseDuration(name, value)
}

func (r *envReader) parseDuration(name, value string) time.Duration {
	d, err := time.ParseDuration(value)
	if err != nil {
		r.addError(name, fmt.Errorf("must be a duration like 5s or 30m, got %q", value))
		return 0
	}
	if d <= 0 {
		r.addError(name, fmt.Errorf("must be positive, got %s", d))
		return 0
	}

	return d
}

func (r *envReader) LogLevel(name string) slog.Level {
	value, ok := r.required(name)
	if !ok {
		return slog.LevelInfo
	}

	var level slog.Level
	if err := level.UnmarshalText([]byte(value)); err != nil {
		r.addError(name, fmt.Errorf("must be one of debug, info, warn, error, got %q", value))
		return slog.LevelInfo
	}

	return level
}
