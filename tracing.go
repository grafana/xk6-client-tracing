package xk6_client_tracing

import (
	"context"
	"fmt"

	"go.k6.io/k6/js/common"
	"go.k6.io/k6/js/modules"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/component/componenttest"
	"go.opentelemetry.io/collector/config/configgrpc"
	"go.opentelemetry.io/collector/config/configtls"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/exporter/otlpexporter"
	"go.opentelemetry.io/collector/model/pdata"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

func init() {
	modules.Register("k6/x/tracing", new(ClientTracing))
}

type Config struct {
	Endpoint string `json:"url"`
}

type ClientTracing struct {
	exporter consumer.Traces
	cfg      *Config
}

func (c *ClientTracing) XClient(ctxPtr *context.Context, config Config) interface{} {
	if config.Endpoint == "" {
		config.Endpoint = "0.0.0.0:4317"
	}

	factory := otlpexporter.NewFactory()
	cfg := factory.CreateDefaultConfig().(*otlpexporter.Config)
	cfg.GRPCClientSettings = configgrpc.GRPCClientSettings{
		Endpoint: config.Endpoint,
		TLSSetting: configtls.TLSClientSetting{
			Insecure: true,
		},
	}

	exporter, err := factory.CreateTracesExporter(
		context.Background(),
		component.ExporterCreateSettings{
			TelemetrySettings: component.TelemetrySettings{
				Logger:         zap.NewNop(),
				TracerProvider: trace.NewNoopTracerProvider(),
				MeterProvider:  metric.NewNoopMeterProvider(),
			},
			BuildInfo: component.NewDefaultBuildInfo(),
		},
		cfg,
	)
	if err != nil {
		return err
	}
	exporter.Start(context.Background(), componenttest.NewNopHost())

	if err != nil {
		return fmt.Errorf("failed to create exporter: %v", err)
	}

	c.exporter = exporter
	c.cfg = &config

	rt := common.GetRuntime(*ctxPtr)
	return common.Bind(rt, c, ctxPtr)
}

func (c *ClientTracing) Send(ctx context.Context, spans []Span) error {
	resource := pdata.NewResource()

	traces := pdata.NewTraces()
	rspans := traces.ResourceSpans().AppendEmpty()
	resource.CopyTo(rspans.Resource())
	ispans := rspans.InstrumentationLibrarySpans().AppendEmpty()
	for _, span := range spans {
		span.construct().CopyTo(ispans.Spans().AppendEmpty())
	}

	err := c.exporter.ConsumeTraces(context.Background(), traces)
	if err != nil {
		return err
	}
	return nil
}
