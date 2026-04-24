module github.com/philipparndt/go-logger/otel

go 1.24.0

require (
	github.com/philipparndt/go-logger v0.0.0
	go.opentelemetry.io/otel/trace v1.41.0
)

require (
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/go-logr/logr v1.4.3 // indirect
	go.opentelemetry.io/otel v1.41.0 // indirect
	k8s.io/klog/v2 v2.140.0 // indirect
)

replace github.com/philipparndt/go-logger => ../
