module github.com/grafana/xk6-client-tracing

go 1.16

require (
	github.com/gogo/protobuf v1.3.2
	github.com/grafana/tempo v1.0.1
	go.k6.io/k6 v0.34.1
	go.opentelemetry.io/collector v0.38.0
	go.opentelemetry.io/collector/model v0.38.0
)

replace (
	k8s.io/api => k8s.io/api v0.21.0
	k8s.io/client-go => k8s.io/client-go v0.21.0
)
