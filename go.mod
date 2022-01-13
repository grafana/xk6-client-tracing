module github.com/grafana/xk6-client-tracing

go 1.16

require (
	github.com/onsi/ginkgo v1.15.0 // indirect
	github.com/onsi/gomega v1.10.5 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/exporter/jaegerexporter v0.38.0
	github.com/sirupsen/logrus v1.8.1
	go.k6.io/k6 v0.35.0
	go.opentelemetry.io/collector v0.38.0
	go.opentelemetry.io/collector/model v0.38.0
	go.opentelemetry.io/otel/metric v0.24.0
	go.opentelemetry.io/otel/trace v1.0.1
	go.uber.org/zap v1.19.1
)

replace github.com/spf13/afero => github.com/spf13/afero v1.1.2
