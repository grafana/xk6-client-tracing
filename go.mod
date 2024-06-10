module github.com/grafana/xk6-client-tracing

go 1.20

require (
	github.com/grafana/sobek v0.0.0-20240607083612-4f0cd64f4e78
	github.com/open-telemetry/opentelemetry-collector-contrib/exporter/jaegerexporter v0.85.0
	github.com/pkg/errors v0.9.1
	github.com/stretchr/testify v1.9.0
	go.k6.io/k6 v0.51.1-0.20240610082146-1f01a9bc2365
	go.opentelemetry.io/collector/component v0.90.1
	go.opentelemetry.io/collector/config/configgrpc v0.90.1
	go.opentelemetry.io/collector/config/confighttp v0.90.1
	go.opentelemetry.io/collector/config/configopaque v0.90.0
	go.opentelemetry.io/collector/config/configtls v0.90.1
	go.opentelemetry.io/collector/exporter v0.90.0
	go.opentelemetry.io/collector/exporter/otlpexporter v0.90.1
	go.opentelemetry.io/collector/exporter/otlphttpexporter v0.90.1
	go.opentelemetry.io/collector/pdata v1.0.0
	go.opentelemetry.io/otel/metric v1.24.0
	go.opentelemetry.io/otel/trace v1.24.0
	go.uber.org/zap v1.26.0
)

require (
	cloud.google.com/go/compute/metadata v0.2.4-0.20230617002413-005d2dfb6b68 // indirect
	github.com/apache/thrift v0.19.0 // indirect
	github.com/cenkalti/backoff/v4 v4.2.1 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/dlclark/regexp2 v1.9.0 // indirect
	github.com/dop251/goja v0.0.0-20240516125602-ccbae20bcec2 // indirect
	github.com/evanw/esbuild v0.21.2 // indirect
	github.com/fatih/color v1.16.0 // indirect
	github.com/felixge/httpsnoop v1.0.4 // indirect
	github.com/fsnotify/fsnotify v1.7.0 // indirect
	github.com/go-logr/logr v1.4.1 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-sourcemap/sourcemap v2.1.4+incompatible // indirect
	github.com/gogo/googleapis v1.4.1 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/snappy v0.0.4 // indirect
	github.com/google/pprof v0.0.0-20230728192033-2ba5b33183c6 // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.19.0 // indirect
	github.com/hashicorp/go-version v1.6.0 // indirect
	github.com/jaegertracing/jaeger v1.41.0 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/klauspost/compress v1.17.7 // indirect
	github.com/knadh/koanf v1.5.0 // indirect
	github.com/knadh/koanf/v2 v2.0.1 // indirect
	github.com/mailru/easyjson v0.7.7 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/mitchellh/copystructure v1.2.0 // indirect
	github.com/mitchellh/mapstructure v1.5.1-0.20220423185008-bf980b35cac4 // indirect
	github.com/mitchellh/reflectwalk v1.0.2 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/mostynb/go-grpc-compression v1.2.2 // indirect
	github.com/mstoykov/atlas v0.0.0-20220811071828-388f114305dd // indirect
	github.com/nxadm/tail v1.4.8 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/coreinternal v0.85.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/pkg/translator/jaeger v0.85.0 // indirect
	github.com/opentracing/opentracing-go v1.2.0 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/rs/cors v1.10.1 // indirect
	github.com/serenize/snaker v0.0.0-20201027110005-a7ad2135616e // indirect
	github.com/sirupsen/logrus v1.9.3 // indirect
	github.com/spf13/afero v1.9.2 // indirect
	github.com/uber/jaeger-client-go v2.30.0+incompatible // indirect
	github.com/uber/jaeger-lib v2.4.1+incompatible // indirect
	go.opencensus.io v0.24.0 // indirect
	go.opentelemetry.io/collector v0.90.0 // indirect
	go.opentelemetry.io/collector/config/configauth v0.90.0 // indirect
	go.opentelemetry.io/collector/config/configcompression v0.90.0 // indirect
	go.opentelemetry.io/collector/config/confignet v0.90.0 // indirect
	go.opentelemetry.io/collector/config/configtelemetry v0.90.0 // indirect
	go.opentelemetry.io/collector/config/internal v0.90.0 // indirect
	go.opentelemetry.io/collector/confmap v0.90.0 // indirect
	go.opentelemetry.io/collector/consumer v0.90.0 // indirect
	go.opentelemetry.io/collector/extension v0.90.0 // indirect
	go.opentelemetry.io/collector/extension/auth v0.90.0 // indirect
	go.opentelemetry.io/collector/featuregate v1.0.0 // indirect
	go.opentelemetry.io/collector/semconv v0.85.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.46.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.46.1 // indirect
	go.opentelemetry.io/otel v1.24.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace v1.24.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.24.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp v1.24.0 // indirect
	go.opentelemetry.io/otel/sdk v1.24.0 // indirect
	go.opentelemetry.io/proto/otlp v1.1.0 // indirect
	go.uber.org/atomic v1.10.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	golang.org/x/net v0.26.0 // indirect
	golang.org/x/sys v0.21.0 // indirect
	golang.org/x/text v0.16.0 // indirect
	golang.org/x/time v0.5.0 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20240227224415-6ceb2ff114de // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240227224415-6ceb2ff114de // indirect
	google.golang.org/grpc v1.63.2 // indirect
	google.golang.org/protobuf v1.33.0 // indirect
	gopkg.in/guregu/null.v3 v3.3.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace cloud.google.com/go/compute/metadata => cloud.google.com/go/compute/metadata v0.2.3
