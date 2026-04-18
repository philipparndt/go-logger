module github.com/philipparndt/go-logger/otel

go 1.24

require (
	github.com/philipparndt/go-logger v0.0.0
	go.opentelemetry.io/otel/trace v1.38.0
)

require go.opentelemetry.io/otel v1.38.0 // indirect

replace github.com/philipparndt/go-logger => ../
