module github.com/grafana/xk6-client-tracing

go 1.25

tool (
	golang.org/x/tools/cmd/goimports
	gotest.tools/gotestsum
)

require (
	github.com/grafana/sobek v0.0.0-20250723111835-dd8a13f0d439
	github.com/stretchr/testify v1.11.1
	go.k6.io/k6 v1.2.3
	go.opentelemetry.io/collector/component v1.42.0
	go.opentelemetry.io/collector/component/componenttest v0.136.0
	go.opentelemetry.io/collector/config/configgrpc v0.136.0
	go.opentelemetry.io/collector/config/confighttp v0.136.0
	go.opentelemetry.io/collector/config/configopaque v1.42.0
	go.opentelemetry.io/collector/config/configtls v1.42.0
	go.opentelemetry.io/collector/exporter v1.42.0
	go.opentelemetry.io/collector/exporter/otlpexporter v0.136.0
	go.opentelemetry.io/collector/exporter/otlphttpexporter v0.136.0
	go.opentelemetry.io/collector/pdata v1.42.0
	go.opentelemetry.io/otel/metric v1.38.0
	go.opentelemetry.io/otel/trace v1.38.0
	go.uber.org/zap v1.27.0
)

require (
	github.com/Masterminds/semver/v3 v3.4.0 // indirect
	github.com/bitfield/gotestdox v0.2.2 // indirect
	github.com/cenkalti/backoff/v5 v5.0.3 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/dlclark/regexp2 v1.11.5 // indirect
	github.com/dnephin/pflag v1.0.7 // indirect
	github.com/evanw/esbuild v0.25.10 // indirect
	github.com/fatih/color v1.18.0 // indirect
	github.com/felixge/httpsnoop v1.0.4 // indirect
	github.com/foxboron/go-tpm-keyfiles v0.0.0-20250903184740-5d135037bd4d // indirect
	github.com/fsnotify/fsnotify v1.9.0 // indirect
	github.com/go-logr/logr v1.4.3 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-sourcemap/sourcemap v2.1.4+incompatible // indirect
	github.com/go-viper/mapstructure/v2 v2.4.0 // indirect
	github.com/gobwas/glob v0.2.3 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/snappy v1.0.0 // indirect
	github.com/google/go-tpm v0.9.6 // indirect
	github.com/google/pprof v0.0.0-20250923004556-9e5a51aed1e8 // indirect
	github.com/google/shlex v0.0.0-20191202100458-e7afc7fbc510 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.27.2 // indirect
	github.com/hashicorp/go-version v1.7.0 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/klauspost/compress v1.18.0 // indirect
	github.com/knadh/koanf/maps v0.1.2 // indirect
	github.com/knadh/koanf/providers/confmap v1.0.0 // indirect
	github.com/knadh/koanf/v2 v2.3.0 // indirect
	github.com/mailru/easyjson v0.9.1 // indirect
	github.com/mattn/go-colorable v0.1.14 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/mitchellh/copystructure v1.2.0 // indirect
	github.com/mitchellh/reflectwalk v1.0.2 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.3-0.20250322232337-35a7c28c31ee // indirect
	github.com/mostynb/go-grpc-compression v1.2.3 // indirect
	github.com/mstoykov/atlas v0.0.0-20220811071828-388f114305dd // indirect
	github.com/onsi/ginkgo v1.16.5 // indirect
	github.com/onsi/gomega v1.33.0 // indirect
	github.com/pierrec/lz4/v4 v4.1.22 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/rs/cors v1.11.1 // indirect
	github.com/serenize/snaker v0.0.0-20201027110005-a7ad2135616e // indirect
	github.com/sirupsen/logrus v1.9.3 // indirect
	github.com/spf13/afero v1.15.0 // indirect
	go.opentelemetry.io/auto/sdk v1.2.1 // indirect
	go.opentelemetry.io/collector v0.136.0 // indirect
	go.opentelemetry.io/collector/client v1.42.0 // indirect
	go.opentelemetry.io/collector/config/configauth v0.136.0 // indirect
	go.opentelemetry.io/collector/config/configcompression v1.42.0 // indirect
	go.opentelemetry.io/collector/config/configmiddleware v1.42.0 // indirect
	go.opentelemetry.io/collector/config/confignet v1.42.0 // indirect
	go.opentelemetry.io/collector/config/configoptional v0.136.0 // indirect
	go.opentelemetry.io/collector/config/configretry v1.42.0 // indirect
	go.opentelemetry.io/collector/confmap v1.42.0 // indirect
	go.opentelemetry.io/collector/confmap/xconfmap v0.136.0 // indirect
	go.opentelemetry.io/collector/consumer v1.42.0 // indirect
	go.opentelemetry.io/collector/consumer/consumererror v0.136.0 // indirect
	go.opentelemetry.io/collector/consumer/consumererror/xconsumererror v0.136.0 // indirect
	go.opentelemetry.io/collector/consumer/xconsumer v0.136.0 // indirect
	go.opentelemetry.io/collector/exporter/exporterhelper v0.136.0 // indirect
	go.opentelemetry.io/collector/exporter/exporterhelper/xexporterhelper v0.136.0 // indirect
	go.opentelemetry.io/collector/exporter/xexporter v0.136.0 // indirect
	go.opentelemetry.io/collector/extension v1.42.0 // indirect
	go.opentelemetry.io/collector/extension/extensionauth v1.42.0 // indirect
	go.opentelemetry.io/collector/extension/extensionmiddleware v0.136.0 // indirect
	go.opentelemetry.io/collector/extension/xextension v0.136.0 // indirect
	go.opentelemetry.io/collector/featuregate v1.42.0 // indirect
	go.opentelemetry.io/collector/internal/telemetry v0.136.0 // indirect
	go.opentelemetry.io/collector/pdata/pprofile v0.136.0 // indirect
	go.opentelemetry.io/collector/pdata/xpdata v0.136.0 // indirect
	go.opentelemetry.io/collector/pipeline v1.42.0 // indirect
	go.opentelemetry.io/collector/pipeline/xpipeline v0.136.0 // indirect
	go.opentelemetry.io/contrib/bridges/otelzap v0.13.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.63.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.63.0 // indirect
	go.opentelemetry.io/otel v1.38.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace v1.38.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.38.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp v1.38.0 // indirect
	go.opentelemetry.io/otel/log v0.14.0 // indirect
	go.opentelemetry.io/otel/sdk v1.38.0 // indirect
	go.opentelemetry.io/otel/sdk/metric v1.38.0 // indirect
	go.opentelemetry.io/proto/otlp v1.8.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	go.yaml.in/yaml/v3 v3.0.4 // indirect
	golang.org/x/crypto v0.42.0 // indirect
	golang.org/x/mod v0.27.0 // indirect
	golang.org/x/net v0.44.0 // indirect
	golang.org/x/sync v0.17.0 // indirect
	golang.org/x/sys v0.36.0 // indirect
	golang.org/x/term v0.35.0 // indirect
	golang.org/x/text v0.29.0 // indirect
	golang.org/x/time v0.13.0 // indirect
	golang.org/x/tools v0.36.0 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20250929231259-57b25ae835d4 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250929231259-57b25ae835d4 // indirect
	google.golang.org/grpc v1.75.1 // indirect
	google.golang.org/protobuf v1.36.9 // indirect
	gopkg.in/guregu/null.v3 v3.5.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	gotest.tools/gotestsum v1.13.0 // indirect
)
