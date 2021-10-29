module github.com/grafana/xk6-client-tracing

go 1.16

require (
	github.com/gogo/protobuf v1.3.2
	github.com/grafana/tempo v1.1.0
	go.k6.io/k6 v0.34.1
	go.opentelemetry.io/collector v0.38.0
	go.opentelemetry.io/collector/model v0.38.0
)

replace github.com/spf13/afero => github.com/spf13/afero v1.1.2
