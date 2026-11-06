package config

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"time"
)

type Config struct {
	HTTP            HTTPConfig
	Database        DatabaseConfig
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

const (
	defaultHTTPReadTimeout       = 10 * time.Second
	defaultHTTPReadHeaderTimeout = 5 * time.Second
	defaultHTTPWriteTimeout      = 15 * time.Second
	defaultHTTPIdleTimeout       = 60 * time.Second
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
