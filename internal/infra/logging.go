package infra

import (
	"context"
	"log/slog"
	"os"
	"time"

	"base-service/config"
)

var counter uint64

type LogEntry struct {
	Level      string    `json:"level"`
	Timestamp  time.Time `json:"timestamp"`
	Message    string    `json:"message"`
	Error      string    `json:"error,omitempty"`
	Caller     string    `json:"caller,omitempty"`
	Additional any       `json:"additional,omitempty"`
}

type logHandler struct {
	slog.Handler
	additionalAttrs []slog.Attr
}

func InitLogger(cf config.Config) *slog.Logger {
	cfg := cf.Log
	programLevel := new(slog.LevelVar)
	programLevel.Set(cfg.LogLevel)

	handlerOptions := &slog.HandlerOptions{
		AddSource: cfg.AddSource,
		Level:     programLevel,
	}

	var baseHandler slog.Handler
	if cfg.JSONOutput {
		baseHandler = slog.NewJSONHandler(os.Stdout, handlerOptions)
	} else {
		baseHandler = slog.NewTextHandler(os.Stdout, handlerOptions)
	}

	handler := &logHandler{
		Handler: baseHandler,
		additionalAttrs: []slog.Attr{
			slog.String("environment", cf.Profile),
			slog.Time("boot_time", time.Now()),
		},
	}

	logger := slog.New(handler)
	slog.SetDefault(logger)

	return logger
}

func (h *logHandler) Handle(ctx context.Context, r slog.Record) error {
	for _, attr := range h.additionalAttrs {
		r.AddAttrs(attr)
	}

	return h.Handler.Handle(ctx, r)
}

func (h *logHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &logHandler{
		Handler:         h.Handler.WithAttrs(attrs),
		additionalAttrs: h.additionalAttrs,
	}
}
