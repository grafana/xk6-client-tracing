module github.com/grafana/xk6-client-tracing

go 1.16

require (
	github.com/HdrHistogram/hdrhistogram-go v1.1.2 // indirect
	github.com/mattn/go-isatty v0.0.14 // indirect
	github.com/onsi/gomega v1.13.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/exporter/jaegerexporter v0.38.0
	github.com/sirupsen/logrus v1.8.1
	github.com/stretchr/testify v1.7.1 // indirect
	go.k6.io/k6 v0.35.0
	go.opentelemetry.io/collector v0.38.0
	go.opentelemetry.io/collector/model v0.38.0
	go.opentelemetry.io/otel/metric v0.24.0
	go.opentelemetry.io/otel/trace v1.0.1
	go.uber.org/zap v1.19.1
	golang.org/x/sys v0.0.0-20210917161153-d61c044b1678 // indirect
	google.golang.org/genproto v0.0.0-20210917145530-b395a37504d4 // indirect
	google.golang.org/grpc v1.46.2 // indirect
	google.golang.org/protobuf v1.28.0 // indirect
)

replace github.com/spf13/afero => github.com/spf13/afero v1.1.2
