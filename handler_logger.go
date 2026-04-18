package logger

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"runtime"
	"sync"
	"time"
)

const (
	levelPadding       = 5
	stackBufSize       = 64
	goroutineIDSScanfN = 1
)

// ColoredLogHandler is a slog.Handler that writes log-style output:
//
//	2024-01-15T15:04:05Z INFO [1] message key=value
type ColoredLogHandler struct {
	level slog.Leveler
	out   io.Writer
	mu    sync.Mutex
	attrs []slog.Attr
}

func NewColoredLogHandler(level slog.Leveler, out io.Writer) *ColoredLogHandler {
	return &ColoredLogHandler{level: level, out: out}
}

func (h *ColoredLogHandler) Enabled(_ context.Context, l slog.Level) bool {
	return l >= h.level.Level()
}

func (h *ColoredLogHandler) Handle(ctx context.Context, record slog.Record) error {
	var buf bytes.Buffer

	// Timestamp
	buf.WriteString(record.Time.UTC().Format(time.RFC3339))
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

func (h *ColoredLogHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	newAttrs := make([]slog.Attr, len(h.attrs)+len(attrs))
	copy(newAttrs, h.attrs)
	copy(newAttrs[len(h.attrs):], attrs)
	return &ColoredLogHandler{
		level: h.level,
		out:   h.out,
		attrs: newAttrs,
	}
}

func (h *ColoredLogHandler) WithGroup(_ string) slog.Handler {
	return h
}

// ---- shared helpers ----

// writeAttr writes a single slog.Attr as key=value.
// String values are quoted; others use their default fmt.
func writeAttr(buf *bytes.Buffer, attr slog.Attr) {
	attr.Value = attr.Value.Resolve()
	if attr.Equal(slog.Attr{}) {
		return
	}
	buf.WriteString(attr.Key)
	buf.WriteByte('=')
	v := attr.Value
	if v.Kind() == slog.KindString {
		_, _ = fmt.Fprintf(buf, "%q", v.String())
	} else {
		buf.WriteString(v.String())
	}
}

// goroutineID returns the ID of the current goroutine.
func goroutineID() uint64 {
	b := make([]byte, stackBufSize)
	runtime.Stack(b, false)
	var id uint64
	_, _ = fmt.Sscanf(string(b), "goroutine %d", &id)
	return id
}
