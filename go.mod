module github.com/grafana/xk6-client-tracing

go 1.16

require (
	cloud.google.com/go v0.83.0 // indirect
	github.com/onsi/ginkgo v1.15.0 // indirect
	github.com/onsi/gomega v1.10.5 // indirect
	go.k6.io/k6 v0.34.1
	go.opentelemetry.io/collector v0.38.0
	go.opentelemetry.io/collector/model v0.38.0
	go.opentelemetry.io/otel/metric v0.24.0
	go.opentelemetry.io/otel/trace v1.0.1
	go.uber.org/zap v1.19.1
	golang.org/x/crypto v0.0.0-20210616213533-5ff15b29337e // indirect
)

replace github.com/spf13/afero => github.com/spf13/afero v1.1.2
