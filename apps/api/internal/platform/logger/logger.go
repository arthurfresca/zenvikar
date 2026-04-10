package logger

import (
	"context"
	"log/slog"
	"os"
	"strings"

	"go.opentelemetry.io/contrib/bridges/otelslog"
	sdklog "go.opentelemetry.io/otel/sdk/log"
)

// New returns a configured *slog.Logger with a JSON handler.
// The log level is read from the LOG_LEVEL environment variable.
// Supported values: "debug", "info", "warn", "error". Defaults to "info".
func New() *slog.Logger {
	level := parseLevel(os.Getenv("LOG_LEVEL"))
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: level,
	})
	return slog.New(handler)
}

// NewWithOTel returns a *slog.Logger that writes to both stdout (JSON) and
// the OTel log provider. Logs appear in the terminal and in Loki via the
// OTel Collector.
func NewWithOTel(lp *sdklog.LoggerProvider) *slog.Logger {
	level := parseLevel(os.Getenv("LOG_LEVEL"))

	jsonHandler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: level,
	})

	otelHandler := otelslog.NewHandler("zenvikar-api", otelslog.WithLoggerProvider(lp))

	return slog.New(&fanoutHandler{handlers: []slog.Handler{jsonHandler, otelHandler}})
}

// fanoutHandler sends log records to multiple slog handlers.
type fanoutHandler struct {
	handlers []slog.Handler
}

func (h *fanoutHandler) Enabled(ctx context.Context, level slog.Level) bool {
	for _, handler := range h.handlers {
		if handler.Enabled(ctx, level) {
			return true
		}
	}
	return false
}

func (h *fanoutHandler) Handle(ctx context.Context, r slog.Record) error {
	for _, handler := range h.handlers {
		_ = handler.Handle(ctx, r)
	}
	return nil
}

func (h *fanoutHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	newHandlers := make([]slog.Handler, len(h.handlers))
	for i, handler := range h.handlers {
		newHandlers[i] = handler.WithAttrs(attrs)
	}
	return &fanoutHandler{handlers: newHandlers}
}

func (h *fanoutHandler) WithGroup(name string) slog.Handler {
	newHandlers := make([]slog.Handler, len(h.handlers))
	for i, handler := range h.handlers {
		newHandlers[i] = handler.WithGroup(name)
	}
	return &fanoutHandler{handlers: newHandlers}
}

// parseLevel converts a string log level to a slog.Level.
// Returns slog.LevelInfo for unrecognized values.
func parseLevel(s string) slog.Level {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "debug":
		return slog.LevelDebug
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
