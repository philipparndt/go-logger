module github.com/philipparndt/go-logger/chi

go 1.24

require (
	github.com/go-chi/chi/v5 v5.2.5
	github.com/philipparndt/go-logger v0.0.0
)

require (
	github.com/go-logr/logr v1.4.3 // indirect
	k8s.io/klog/v2 v2.140.0 // indirect
)

replace github.com/philipparndt/go-logger => ../
