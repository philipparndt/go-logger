package logger

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"sync"
)

// ColoredLogHandlerNoUptime is a slog.Handler that writes log-style output
// without the uptime field:
//
//	2024-01-15T15:04:05Z INFO [  1] message key=value
type ColoredLogHandlerNoUptime struct {
	level slog.Leveler
	out   io.Writer
	mu    sync.Mutex
	attrs []slog.Attr
}

func NewColoredLogHandlerNoUptime(level slog.Leveler, out io.Writer) *ColoredLogHandlerNoUptime {
	return &ColoredLogHandlerNoUptime{level: level, out: out}
}

func (h *ColoredLogHandlerNoUptime) Enabled(_ context.Context, l slog.Level) bool {
	return l >= h.level.Level()
}

func (h *ColoredLogHandlerNoUptime) Handle(ctx context.Context, record slog.Record) error {
	var buf bytes.Buffer

	// Timestamp
	buf.WriteString(formatTime(record.Time))
	buf.WriteByte(' ')

	// Level name (colored) with padding for alignment
	color := levelColor(record.Level)
	reset := resetColor()
	buf.WriteString(color)
	levelStr := levelName(record.Level)
	buf.WriteString(levelStr)
	buf.WriteString(reset)

	// Pad level to levelPadding characters for alignment
	padding := levelPadding - len(levelStr)
	for i := 0; i < padding; i++ {
		buf.WriteByte(' ')
	}

	// Goroutine ID in brackets in gray after level
	if showThreads {
		gid := goroutineID()
		if gid > 0 {
			buf.WriteString(" ")
			buf.WriteString(colorGray)
			_, _ = fmt.Fprintf(&buf, "[%3d]", gid)
			buf.WriteString(resetColor())
		}
	}

	// Message
	buf.WriteString(" ")
	buf.WriteString(record.Message)

	// Handler-level attrs and record-level attrs wrapped in gray
	hasAttrs := len(h.attrs) > 0
	record.Attrs(func(_ slog.Attr) bool {
		hasAttrs = true
		return true
	})

	// Context enrichment (e.g. trace ID)
	ctxKey, ctxVal := enrichFromContext(ctx)
	hasContextInfo := hasAttrs || ctxKey != ""

	if hasContextInfo {
		buf.WriteString(" ")
		buf.WriteString(colorGray)

		// Handler-level attrs (from WithAttrs)
		for _, a := range h.attrs {
			writeAttr(&buf, a)
			buf.WriteByte(' ')
		}

		// Record-level attrs
		record.Attrs(func(a slog.Attr) bool {
			writeAttr(&buf, a)
			buf.WriteByte(' ')
			return true
		})

		// Context enrichment
		if ctxKey != "" {
			_, _ = fmt.Fprintf(&buf, "%s=%q", ctxKey, ctxVal)
		} else {
			// Remove trailing space if we only have attrs
			buf.Truncate(buf.Len() - 1)
		}

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

func (h *ColoredLogHandlerNoUptime) WithAttrs(attrs []slog.Attr) slog.Handler {
	newAttrs := make([]slog.Attr, len(h.attrs)+len(attrs))
	copy(newAttrs, h.attrs)
	copy(newAttrs[len(h.attrs):], attrs)
	return &ColoredLogHandlerNoUptime{
		level: h.level,
		out:   h.out,
		attrs: newAttrs,
	}
}

func (h *ColoredLogHandlerNoUptime) WithGroup(_ string) slog.Handler {
	return h
}
