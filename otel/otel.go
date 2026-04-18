// Package otel provides OpenTelemetry trace ID extraction for go-logger handlers.
//
// Usage:
//
//	import (
//	    logger "github.com/philipparndt/go-logger"
//	    loggerOtel "github.com/philipparndt/go-logger/otel"
//	)
//
//	// Call once at startup after logger.Init()
//	loggerOtel.Enable()
//
//	// Now trace IDs from context will appear in log output
//	ctx := ... // context with OTel span
//	logger.Info(ctx, "request handled", "method", "GET")
package otel

import (
	"context"

	logger "github.com/philipparndt/go-logger"
	"go.opentelemetry.io/otel/trace"
)

func init() {
	Enable()
}

// Enable registers the OpenTelemetry trace ID extractor with the logger.
// Called automatically on import, but can be called explicitly if needed.
func Enable() {
	logger.SetContextEnricher(func(ctx context.Context) (string, string) {
		spanCtx := trace.SpanContextFromContext(ctx)
		if spanCtx.HasTraceID() {
			return "trace_id", spanCtx.TraceID().String()
		}
		return "", ""
	})
}

// Disable removes the OpenTelemetry trace ID extractor from the logger.
func Disable() {
	logger.SetContextEnricher(nil)
}
