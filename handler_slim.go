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

// ColoredSlimHandler is a slog.Handler that writes slim-style output:
// Arguments are printed as an array in gray:
//
//	2024-01-15T15:04:05Z [INFO] message [arg1 arg2 arg3]
type ColoredSlimHandler struct {
	level slog.Leveler
	out   io.Writer
	mu    sync.Mutex
	attrs []slog.Attr
}

func NewColoredSlimHandler(level slog.Leveler, out io.Writer) *ColoredSlimHandler {
	return &ColoredSlimHandler{level: level, out: out}
}

func (h *ColoredSlimHandler) Enabled(_ context.Context, l slog.Level) bool {
	return l >= h.level.Level()
}

func (h *ColoredSlimHandler) Handle(ctx context.Context, record slog.Record) error {
	var buf bytes.Buffer

	// Timestamp
	buf.WriteString(record.Time.UTC().Format(time.RFC3339))
	buf.WriteByte(' ')

	// Colored [LEVEL]
	color := levelColor(record.Level)
	reset := resetColor()
	buf.WriteString(color)
	buf.WriteByte('[')
	buf.WriteString(levelName(record.Level))
	buf.WriteByte(']')
	buf.WriteString(reset)
	buf.WriteByte(' ')

	// Message
	buf.WriteString(record.Message)

	// Collect all args into a slice
	args := make([]any, 0, len(h.attrs)*2+8)

	extractValue := func(v slog.Value) any {
		v = v.Resolve()
		return v.Any()
	}

	// Handler-level attrs
	for _, a := range h.attrs {
		args = append(args, a.Key)
		args = append(args, extractValue(a.Value))
	}

	// Record-level attrs
	record.Attrs(func(a slog.Attr) bool {
		args = append(args, a.Key)
		args = append(args, extractValue(a.Value))
		return true
	})

	// Goroutine ID
	if showThreads {
		gid := goroutineID()
		if gid > 0 {
			args = append(args, "gid")
			args = append(args, gid)
		}
	}

	// Context enrichment (e.g. trace ID)
	if ctxKey, ctxVal := enrichFromContext(ctx); ctxKey != "" {
		args = append(args, ctxKey)
		args = append(args, ctxVal)
	}

	// Print all args as array in gray if any exist
	if len(args) > 0 {
		buf.WriteByte(' ')
		buf.WriteString(colorGray)
		buf.WriteByte('[')
		for i, arg := range args {
			if i > 0 {
				buf.WriteByte(' ')
			}
			_, _ = fmt.Fprint(&buf, arg)
		}
		buf.WriteByte(']')
		buf.WriteString(resetColor())
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

func (h *ColoredSlimHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	newAttrs := make([]slog.Attr, len(h.attrs)+len(attrs))
	copy(newAttrs, h.attrs)
	copy(newAttrs[len(h.attrs):], attrs)
	return &ColoredSlimHandler{
		level: h.level,
		out:   h.out,
		attrs: newAttrs,
	}
}

func (h *ColoredSlimHandler) WithGroup(_ string) slog.Handler {
	return h
}
