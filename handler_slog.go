package logger

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"sync"
	"time"
)

// ColoredSlogHandler is a slog.Handler that writes structured slog-style output:
//
//	time=2024-01-15T15:04:05Z level=INFO msg="message" key=value
type ColoredSlogHandler struct {
	level slog.Leveler
	out   io.Writer
	mu    sync.Mutex
	attrs []slog.Attr
}

func NewColoredSlogHandler(level slog.Leveler, out io.Writer) *ColoredSlogHandler {
	return &ColoredSlogHandler{level: level, out: out}
}

func (h *ColoredSlogHandler) Enabled(_ context.Context, l slog.Level) bool {
	return l >= h.level.Level()
}

func (h *ColoredSlogHandler) Handle(ctx context.Context, record slog.Record) error {
	var buf bytes.Buffer

	color := levelColor(record.Level)
	reset := resetColor()

	// time=...
	_, _ = fmt.Fprintf(&buf, "time=%s ", record.Time.UTC().Format(time.RFC3339))

	// level=... (colored)
	_, _ = fmt.Fprintf(&buf, "%slevel=%s%s ", color, levelName(record.Level), reset)

	// msg=...
	_, _ = fmt.Fprintf(&buf, "msg=%q", record.Message)

	// Handler-level attrs
	for _, a := range h.attrs {
		buf.WriteByte(' ')
		writeAttr(&buf, a)
	}

	// Record-level attrs
	record.Attrs(func(a slog.Attr) bool {
		buf.WriteByte(' ')
		writeAttr(&buf, a)
		return true
	})

	// Goroutine ID
	if showThreads {
		gid := goroutineID()
		if gid > 0 {
			_, _ = fmt.Fprintf(&buf, " gid=%d", gid)
		}
	}

	// Context enrichment (e.g. trace ID)
	if ctxKey, ctxVal := enrichFromContext(ctx); ctxKey != "" {
		_, _ = fmt.Fprintf(&buf, " %s=%q", ctxKey, ctxVal)
	}

	buf.WriteByte('\n')

	h.mu.Lock()
	defer h.mu.Unlock()
	_, err := h.out.Write(buf.Bytes())
	if err != nil {
		return fmt.Errorf("write log: %w", err)
	}
	return nil
}

func (h *ColoredSlogHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	newAttrs := make([]slog.Attr, len(h.attrs)+len(attrs))
	copy(newAttrs, h.attrs)
	copy(newAttrs[len(h.attrs):], attrs)
	return &ColoredSlogHandler{
		level: h.level,
		out:   h.out,
		attrs: newAttrs,
	}
}

func (h *ColoredSlogHandler) WithGroup(_ string) slog.Handler {
	return h
}
