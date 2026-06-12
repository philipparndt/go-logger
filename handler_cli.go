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

// ColoredCLIHandler is a slog.Handler for CLI tools. It writes the uptime in
// seconds followed by the message in the color of the level:
//
//	[   0] message key=value
//
// With showLevel enabled, the level name is printed before the message:
//
//	[   0] INFO  message key=value
type ColoredCLIHandler struct {
	level     slog.Leveler
	out       io.Writer
	showLevel bool
	mu        sync.Mutex
	attrs     []slog.Attr
}

func NewColoredCLIHandler(level slog.Leveler, out io.Writer, showLevel bool) *ColoredCLIHandler {
	return &ColoredCLIHandler{level: level, out: out, showLevel: showLevel}
}

func (h *ColoredCLIHandler) Enabled(_ context.Context, l slog.Level) bool {
	return l >= h.level.Level()
}

func (h *ColoredCLIHandler) Handle(ctx context.Context, record slog.Record) error {
	var buf bytes.Buffer

	// Uptime in seconds, in gray brackets
	uptime := int64(time.Since(startTime).Seconds())
	buf.WriteString(colorGray)
	_, _ = fmt.Fprintf(&buf, "[%4d]", uptime)
	buf.WriteString(resetColor())

	color := levelColor(record.Level)
	reset := resetColor()

	// Level name (colored) with padding for alignment
	if h.showLevel {
		buf.WriteByte(' ')
		buf.WriteString(color)
		levelStr := levelName(record.Level)
		buf.WriteString(levelStr)
		buf.WriteString(reset)

		// Pad level to levelPadding characters for alignment
		padding := levelPadding - len(levelStr)
		for i := 0; i < padding; i++ {
			buf.WriteByte(' ')
		}
	}

	// Message in the color of the level
	buf.WriteByte(' ')
	buf.WriteString(color)
	buf.WriteString(record.Message)
	buf.WriteString(reset)

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

func (h *ColoredCLIHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	newAttrs := make([]slog.Attr, len(h.attrs)+len(attrs))
	copy(newAttrs, h.attrs)
	copy(newAttrs[len(h.attrs):], attrs)
	return &ColoredCLIHandler{
		level:     h.level,
		out:       h.out,
		showLevel: h.showLevel,
		attrs:     newAttrs,
	}
}

func (h *ColoredCLIHandler) WithGroup(_ string) slog.Handler {
	return h
}
