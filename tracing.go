package xk6_client_tracing

import (
	"context"
	"encoding/base64"
	"github.com/pkg/errors"
	"go.opentelemetry.io/collector/pdata/ptrace"
	"os"

	"github.com/dop251/goja"
	"github.com/grafana/xk6-client-tracing/pkg/random"
	"github.com/grafana/xk6-client-tracing/pkg/tracegen"
	"github.com/open-telemetry/opentelemetry-collector-contrib/exporter/jaegerexporter"
	"go.k6.io/k6/js/common"
	"go.k6.io/k6/js/modules"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/component/componenttest"
	"go.opentelemetry.io/collector/config"
	"go.opentelemetry.io/collector/config/configgrpc"
	"go.opentelemetry.io/collector/config/configtls"
	"go.opentelemetry.io/collector/exporter/otlpexporter"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type exporterType string

const (
	exporterNone   exporterType = ""
	exporterOTLP   exporterType = "otlp"
	exporterJaeger exporterType = "jaeger"
)

var (
	_ modules.Module   = &RootModule{}
	_ modules.Instance = &TracingModule{}
)

func init() {
	modules.Register("k6/x/tracing", new(RootModule))
}

type RootModule struct{}

func (r *RootModule) NewModuleInstance(vu modules.VU) modules.Instance {
	return &TracingModule{
		vu: vu,
	}
}

type TracingModule struct {
	vu     modules.VU
	client *Client
}

func (ct *TracingModule) Exports() modules.Exports {
	return modules.Exports{
		Named: map[string]interface{}{
			// constants
			"EXPORTER_OTLP":   exporterOTLP,
			"EXPORTER_JAEGER": exporterJaeger,
			// constructors
			"Client":                 ct.newClient,
			"ParameterizedGenerator": ct.newParameterizedGenerator,
			// functions
			"generateRandomTraceID": ct.generateRandomTraceID,
		},
	}
}

func (ct *TracingModule) newClient(g goja.ConstructorCall, rt *goja.Runtime) *goja.Object {
	var cfg ClientConfig
	err := rt.ExportTo(g.Argument(0), &cfg)
	if err != nil {
		common.Throw(rt, errors.Wrap(err, "unable to create client: constructor expects first argument to be ClientConfig"))
	}

	if ct.client == nil {
		ct.client, err = NewClient(&cfg, ct.vu)
		if err != nil {
			common.Throw(rt, errors.Wrap(err, "unable to create client"))
		}
	}

	return rt.ToValue(ct.client).ToObject(rt)
}

func (ct *TracingModule) newParameterizedGenerator(g goja.ConstructorCall, rt *goja.Runtime) *goja.Object {
	var traceParams []*tracegen.TraceParams
	err := rt.ExportTo(g.Argument(0), &traceParams)
	if err != nil {
		common.Throw(rt, errors.Wrap(err, "the ParameterizedGenerator constructor expects first argument to be []TraceParams"))
	}
	generator := tracegen.NewParameterizedGenerator(traceParams)

	return rt.ToValue(generator).ToObject(rt)
}

func (ct *TracingModule) generateRandomTraceID() string {
	return random.TraceID().HexString()
}

type ClientConfig struct {
	Exporter       exporterType `json:"type"`
	Endpoint       string       `json:"url"`
	Insecure       bool         `json:"insecure"`
	Authentication struct {
		User     string `json:"user"`
		Password string `json:"password"`
	}
	Headers map[string]string `json:"headers"`
}

type Client struct {
	exporter component.TracesExporter
	vu       modules.VU
}

func NewClient(cfg *ClientConfig, vu modules.VU) (*Client, error) {
	if cfg.Endpoint == "" {
		cfg.Endpoint = "0.0.0.0:4317"
	}

	var (
		factory     component.ExporterFactory
		exporterCfg config.Exporter
	)

	switch cfg.Exporter {
	case exporterNone, exporterOTLP:
		factory = otlpexporter.NewFactory()
		exporterCfg = factory.CreateDefaultConfig()
		exporterCfg.(*otlpexporter.Config).GRPCClientSettings = configgrpc.GRPCClientSettings{
			Endpoint: cfg.Endpoint,
			TLSSetting: configtls.TLSClientSetting{
				Insecure: cfg.Insecure,
			},
			Headers: mergeMaps(map[string]string{
				"Authorization": "Basic " + base64.StdEncoding.EncodeToString([]byte(cfg.Authentication.User+":"+cfg.Authentication.Password)),
			}, cfg.Headers),
		}
	case exporterJaeger:
		factory = jaegerexporter.NewFactory()
		exporterCfg = factory.CreateDefaultConfig()
		exporterCfg.(*jaegerexporter.Config).GRPCClientSettings = configgrpc.GRPCClientSettings{
			Endpoint: cfg.Endpoint,
			TLSSetting: configtls.TLSClientSetting{
				Insecure: cfg.Insecure,
			},
			Headers: mergeMaps(map[string]string{
				"Authorization": "Basic " + base64.StdEncoding.EncodeToString([]byte(cfg.Authentication.User+":"+cfg.Authentication.Password)),
			}, cfg.Headers),
		}
	default:
		return nil, errors.Errorf("failed to init exporter: unknown exporter type %s", cfg.Exporter)
	}

	exporter, err := factory.CreateTracesExporter(
		context.Background(),
		component.ExporterCreateSettings{
			TelemetrySettings: component.TelemetrySettings{
				Logger:         zap.New(zapcore.NewCore(zapcore.NewJSONEncoder(zapcore.EncoderConfig{}), zapcore.AddSync(os.Stdout), zap.DebugLevel)),
				TracerProvider: trace.NewNoopTracerProvider(),
				MeterProvider:  metric.NewNoopMeterProvider(),
			},
			BuildInfo: component.NewDefaultBuildInfo(),
		},
		exporterCfg,
	)

	err = exporter.Start(vu.Context(), componenttest.NewNopHost())
	if err != nil {
		return nil, errors.Wrap(err, "failed to start exporter")
	}

	return &Client{
		exporter: exporter,
		vu:       vu,
	}, nil
}

func (c *Client) Push(traces ptrace.Traces) error {
	return c.exporter.ConsumeTraces(c.vu.Context(), traces)
}

func (c *Client) Shutdown() error {
	return c.exporter.Shutdown(c.vu.Context())
}

func mergeMaps(ms ...map[string]string) map[string]string {
	result := make(map[string]string)
	for _, m := range ms {
		for k, v := range m {
			result[k] = v
		}
	}
	return result
}
