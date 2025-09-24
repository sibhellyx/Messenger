package logger

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os"
)

type PrettyHandler struct {
	w          io.Writer
	timeFormat string
}

func NewPrettyHandler(w io.Writer) *PrettyHandler {
	return &PrettyHandler{
		w:          w,
		timeFormat: "2006-01-02 15:04:05",
	}
}

func (h *PrettyHandler) Handle(ctx context.Context, r slog.Record) error {
	level := r.Level.String()
	timeStr := r.Time.Format(h.timeFormat)
	msg := r.Message

	// colors for levelss
	switch r.Level {
	case slog.LevelDebug:
		level = "\033[36mDEBUG\033[0m" // Cyan
	case slog.LevelInfo:
		level = "\033[32mINFO\033[0m" // Green
	case slog.LevelWarn:
		level = "\033[33mWARN\033[0m" // Yellow
	case slog.LevelError:
		level = "\033[31mERROR\033[0m" // Red
	}

	// base print
	fmt.Fprintf(h.w, "%s [%s] %s", timeStr, level, msg)

	// atributes
	if r.NumAttrs() > 0 {
		r.Attrs(func(attr slog.Attr) bool {
			value, _ := json.Marshal(attr.Value.Any())
			fmt.Fprintf(h.w, " \033[90m%s=%s\033[0m", attr.Key, string(value))
			return true
		})
	}

	fmt.Fprintln(h.w)
	return nil
}

func (h *PrettyHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return true
}

func (h *PrettyHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return h
}

func (h *PrettyHandler) WithGroup(name string) slog.Handler {
	return h
}

func NewLogger(env string) *slog.Logger {
	if env == "development" {
		return slog.New(NewPrettyHandler(os.Stdout))
	}

	// Production - JSON format
	return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
}
