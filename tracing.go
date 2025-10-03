package clienttracing

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"sync"

	"github.com/grafana/sobek"
	"go.k6.io/k6/js/common"
	"go.k6.io/k6/js/modules"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/component/componenttest"
	"go.opentelemetry.io/collector/config/configgrpc"
	"go.opentelemetry.io/collector/config/confighttp"
	"go.opentelemetry.io/collector/config/configopaque"
	"go.opentelemetry.io/collector/config/configtls"
	"go.opentelemetry.io/collector/exporter"
	"go.opentelemetry.io/collector/exporter/otlpexporter"
	"go.opentelemetry.io/collector/exporter/otlphttpexporter"
	"go.opentelemetry.io/collector/pdata/ptrace"
	metricnoop "go.opentelemetry.io/otel/metric/noop"
	tracenoop "go.opentelemetry.io/otel/trace/noop"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/grafana/xk6-client-tracing/pkg/tracegen"
	"github.com/grafana/xk6-client-tracing/pkg/util"
)

type exporterType string

const (
	exporterNone     exporterType = ""
	exporterOTLP     exporterType = "otlp"
	exporterOTLPHTTP exporterType = "otlphttp"
)

var (
	_ modules.Module   = &RootModule{}
	_ modules.Instance = &TracingModule{}
)

func init() {
	modules.Register("k6/x/tracing", new(RootModule))
}

type RootModule struct {
	sync.Mutex
}

func (r *RootModule) NewModuleInstance(vu modules.VU) modules.Instance {
	return &TracingModule{
		vu:                  vu,
		paramGenerators:     make(map[*sobek.Object]*tracegen.ParameterizedGenerator),
		templatedGenerators: make(map[*sobek.Object]*tracegen.TemplatedGenerator),
	}
}

type TracingModule struct {
	vu                  modules.VU
	client              *Client
	paramGenerators     map[*sobek.Object]*tracegen.ParameterizedGenerator
	templatedGenerators map[*sobek.Object]*tracegen.TemplatedGenerator
}

func (ct *TracingModule) Exports() modules.Exports {
	return modules.Exports{
		Named: map[string]interface{}{
			// constants
			"SEMANTICS_HTTP":     tracegen.SemanticsHTTP,
			"SEMANTICS_DB":       tracegen.SemanticsDB,
			"EXPORTER_OTLP":      exporterOTLP,
			"EXPORTER_OTLP_HTTP": exporterOTLPHTTP,
			// constructors
			"Client":                 ct.newClient,
			"ParameterizedGenerator": ct.newParameterizedGenerator,
			"TemplatedGenerator":     ct.newTemplatedGenerator,
		},
	}
}

func (ct *TracingModule) newClient(g sobek.ConstructorCall, rt *sobek.Runtime) *sobek.Object {
	var cfg ClientConfig
	err := rt.ExportTo(g.Argument(0), &cfg)
	if err != nil {
		common.Throw(rt, fmt.Errorf("unable to create client: constructor expects first argument to be ClientConfig: %w", err))
	}

	if ct.client == nil {
		ct.client, err = NewClient(&cfg, ct.vu)
		if err != nil {
			common.Throw(rt, fmt.Errorf("unable to create client: %w", err))
		}
	}

	return rt.ToValue(ct.client).ToObject(rt)
}

func (ct *TracingModule) newParameterizedGenerator(g sobek.ConstructorCall, rt *sobek.Runtime) *sobek.Object {
	paramVal := g.Argument(0)
	paramObj := paramVal.ToObject(rt)

	generator, found := ct.paramGenerators[paramObj]
	if !found {
		var param []*tracegen.TraceParams
		err := rt.ExportTo(paramVal, &param)
		if err != nil {
			common.Throw(rt, fmt.Errorf("the ParameterizedGenerator constructor expects first argument to be []TraceParams: %w", err))
		}

		generator = tracegen.NewParameterizedGenerator(param)
		ct.paramGenerators[paramObj] = generator
	}

	return rt.ToValue(generator).ToObject(rt)
}

func (ct *TracingModule) newTemplatedGenerator(g sobek.ConstructorCall, rt *sobek.Runtime) *sobek.Object {
	tmplVal := g.Argument(0)
	tmplObj := tmplVal.ToObject(rt)

	generator, found := ct.templatedGenerators[tmplObj]
	if !found {
		var tmpl tracegen.TraceTemplate
		err := rt.ExportTo(tmplVal, &tmpl)
		if err != nil {
			common.Throw(rt, fmt.Errorf("the TemplatedGenerator constructor expects first argument to be TraceTemplate: %w", err))
		}

		generator, err = tracegen.NewTemplatedGenerator(&tmpl)
		if err != nil {
			common.Throw(rt, fmt.Errorf("unable to generate TemplatedGenerator: %w", err))
		}

		ct.templatedGenerators[tmplObj] = generator
	}

	return rt.ToValue(generator).ToObject(rt)
}

type TLSClientConfig struct {
	Insecure           bool   `js:"insecure"`
	InsecureSkipVerify bool   `js:"insecure_skip_verify"`
	ServerName         string `js:"server_name"`
	CAFile             string `js:"ca_file"`
	CertFile           string `js:"cert_file"`
	KeyFile            string `js:"key_file"`
}

type ClientConfig struct {
	Exporter       exporterType    `js:"exporter"`
	Endpoint       string          `js:"endpoint"`
	TLS            TLSClientConfig `js:"tls"`
	Authentication struct {
		User     string `js:"user"`
		Password string `js:"password"`
	}
	Headers map[string]configopaque.String `js:"headers"`
}

type Client struct {
	exporter exporter.Traces
	vu       modules.VU
}

func NewClient(cfg *ClientConfig, vu modules.VU) (*Client, error) {
	if cfg.Endpoint == "" {
		cfg.Endpoint = "0.0.0.0:4317"
	}

	var (
		factory     exporter.Factory
		exporterCfg component.Config
	)

	tlsConfig := configtls.ClientConfig{
		Insecure:           cfg.TLS.Insecure,
		InsecureSkipVerify: cfg.TLS.InsecureSkipVerify,
		ServerName:         cfg.TLS.ServerName,
		Config: configtls.Config{
			CAFile:   cfg.TLS.CAFile,
			CertFile: cfg.TLS.CertFile,
			KeyFile:  cfg.TLS.KeyFile,
		},
	}

	switch cfg.Exporter {
	case exporterNone, exporterOTLP:
		factory = otlpexporter.NewFactory()
		exporterCfg = factory.CreateDefaultConfig()
		exporterCfg.(*otlpexporter.Config).ClientConfig = configgrpc.ClientConfig{
			Endpoint: cfg.Endpoint,
			TLS:      tlsConfig,
			Headers: util.MergeMaps(map[string]configopaque.String{
				"Authorization": authorizationHeader(cfg.Authentication.User, cfg.Authentication.Password),
			}, cfg.Headers),
		}
	case exporterOTLPHTTP:
		factory = otlphttpexporter.NewFactory()
		exporterCfg = factory.CreateDefaultConfig()
		exporterCfg.(*otlphttpexporter.Config).ClientConfig = confighttp.ClientConfig{
			Endpoint: cfg.Endpoint,
			TLS:      tlsConfig,
			Headers: util.MergeMaps(map[string]configopaque.String{
				"Authorization": authorizationHeader(cfg.Authentication.User, cfg.Authentication.Password),
			}, cfg.Headers),
		}
	default:
		return nil, fmt.Errorf("failed to init exporter: unknown exporter type %s", cfg.Exporter)
	}

	exporter, err := factory.CreateTraces(
		context.Background(),
		exporter.Settings{
			ID: component.NewID(factory.Type()),
			TelemetrySettings: component.TelemetrySettings{
				Logger:         zap.New(zapcore.NewCore(zapcore.NewJSONEncoder(zapcore.EncoderConfig{}), zapcore.AddSync(os.Stdout), zap.InfoLevel)),
				TracerProvider: tracenoop.NewTracerProvider(),
				MeterProvider:  metricnoop.NewMeterProvider(),
			},
			BuildInfo: component.NewDefaultBuildInfo(),
		},
		exporterCfg,
	)
	if err != nil {
		return nil, fmt.Errorf("failed create exporter: %w", err)
	}

	err = exporter.Start(vu.Context(), componenttest.NewNopHost())
	if err != nil {
		return nil, fmt.Errorf("failed to start exporter: %w", err)
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

func authorizationHeader(user, password string) configopaque.String {
	return configopaque.String("Basic " + base64.StdEncoding.EncodeToString([]byte(user+":"+password)))
}
