module github.com/grafana/xk6-client-tracing

go 1.23.0

toolchain go1.24.0

tool (
	golang.org/x/tools/cmd/goimports
	gotest.tools/gotestsum
)

require (
	github.com/grafana/sobek v0.0.0-20250219104821-ed22af7a8d6c
	github.com/stretchr/testify v1.10.0
	go.k6.io/k6 v0.57.0
	go.opentelemetry.io/collector/component v1.31.0
	go.opentelemetry.io/collector/component/componenttest v0.120.0
	go.opentelemetry.io/collector/config/configgrpc v0.120.0
	go.opentelemetry.io/collector/config/confighttp v0.120.0
	go.opentelemetry.io/collector/config/configopaque v1.26.0
	go.opentelemetry.io/collector/config/configtls v1.26.0
	go.opentelemetry.io/collector/exporter v0.120.0
	go.opentelemetry.io/collector/exporter/otlpexporter v0.120.0
	go.opentelemetry.io/collector/exporter/otlphttpexporter v0.120.0
	go.opentelemetry.io/collector/pdata v1.31.0
	go.opentelemetry.io/otel/metric v1.35.0
	go.opentelemetry.io/otel/trace v1.35.0
	go.uber.org/zap v1.27.0
)

require (
	github.com/bitfield/gotestdox v0.2.2 // indirect
	github.com/cenkalti/backoff/v4 v4.3.0 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/dlclark/regexp2 v1.11.5 // indirect
	github.com/dnephin/pflag v1.0.7 // indirect
	github.com/evanw/esbuild v0.25.0 // indirect
	github.com/fatih/color v1.18.0 // indirect
	github.com/felixge/httpsnoop v1.0.4 // indirect
	github.com/fsnotify/fsnotify v1.8.0 // indirect
	github.com/go-logr/logr v1.4.2 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-sourcemap/sourcemap v2.1.4+incompatible // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/snappy v0.0.4 // indirect
	github.com/google/pprof v0.0.0-20250302191652-9094ed2288e7 // indirect
	github.com/google/shlex v0.0.0-20191202100458-e7afc7fbc510 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.26.1 // indirect
	github.com/hashicorp/go-version v1.7.0 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/klauspost/compress v1.18.0 // indirect
	github.com/mailru/easyjson v0.9.0 // indirect
	github.com/mattn/go-colorable v0.1.14 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/mostynb/go-grpc-compression v1.2.3 // indirect
	github.com/mstoykov/atlas v0.0.0-20220811071828-388f114305dd // indirect
	github.com/onsi/ginkgo v1.16.5 // indirect
	github.com/onsi/gomega v1.33.0 // indirect
	github.com/pierrec/lz4/v4 v4.1.22 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/rs/cors v1.11.1 // indirect
	github.com/serenize/snaker v0.0.0-20201027110005-a7ad2135616e // indirect
	github.com/sirupsen/logrus v1.9.3 // indirect
	github.com/spf13/afero v1.12.0 // indirect
	go.opentelemetry.io/auto/sdk v1.1.0 // indirect
	go.opentelemetry.io/collector v0.120.0 // indirect
	go.opentelemetry.io/collector/client v1.26.0 // indirect
	go.opentelemetry.io/collector/config/configauth v0.120.0 // indirect
	go.opentelemetry.io/collector/config/configcompression v1.26.0 // indirect
	go.opentelemetry.io/collector/config/confignet v1.26.0 // indirect
	go.opentelemetry.io/collector/config/configretry v1.26.0 // indirect
	go.opentelemetry.io/collector/consumer v1.26.0 // indirect
	go.opentelemetry.io/collector/consumer/consumererror v0.120.0 // indirect
	go.opentelemetry.io/collector/consumer/consumererror/xconsumererror v0.120.0 // indirect
	go.opentelemetry.io/collector/consumer/xconsumer v0.120.0 // indirect
	go.opentelemetry.io/collector/exporter/exporterhelper/xexporterhelper v0.120.0 // indirect
	go.opentelemetry.io/collector/exporter/xexporter v0.120.0 // indirect
	go.opentelemetry.io/collector/extension v0.120.0 // indirect
	go.opentelemetry.io/collector/extension/auth v0.120.0 // indirect
	go.opentelemetry.io/collector/extension/xextension v0.120.0 // indirect
	go.opentelemetry.io/collector/featuregate v1.31.0 // indirect
	go.opentelemetry.io/collector/internal/telemetry v0.125.0 // indirect
	go.opentelemetry.io/collector/pdata/pprofile v0.120.0 // indirect
	go.opentelemetry.io/collector/pipeline v0.125.0 // indirect
	go.opentelemetry.io/collector/pipeline/xpipeline v0.120.0 // indirect
	go.opentelemetry.io/contrib/bridges/otelzap v0.10.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.59.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.59.0 // indirect
	go.opentelemetry.io/otel v1.35.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace v1.34.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.34.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp v1.34.0 // indirect
	go.opentelemetry.io/otel/log v0.11.0 // indirect
	go.opentelemetry.io/otel/sdk v1.35.0 // indirect
	go.opentelemetry.io/otel/sdk/metric v1.35.0 // indirect
	go.opentelemetry.io/proto/otlp v1.5.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	golang.org/x/mod v0.17.0 // indirect
	golang.org/x/net v0.39.0 // indirect
	golang.org/x/sync v0.13.0 // indirect
	golang.org/x/sys v0.32.0 // indirect
	golang.org/x/term v0.31.0 // indirect
	golang.org/x/text v0.24.0 // indirect
	golang.org/x/time v0.10.0 // indirect
	golang.org/x/tools v0.21.1-0.20240508182429-e35e4ccd0d2d // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20250303144028-a0af3efb3deb // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250303144028-a0af3efb3deb // indirect
	google.golang.org/grpc v1.72.0 // indirect
	google.golang.org/protobuf v1.36.6 // indirect
	gopkg.in/guregu/null.v3 v3.5.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	gotest.tools/gotestsum v1.12.0 // indirect
)
