package xk6_client_tracing

import (
	"context"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"math/rand"
	"os"
	"time"
	"unsafe"

	"github.com/dop251/goja"
	"github.com/grafana/xk6-client-tracing/pkg/random"
	"github.com/open-telemetry/opentelemetry-collector-contrib/exporter/jaegerexporter"
	log "github.com/sirupsen/logrus"
	"go.k6.io/k6/js/common"
	"go.k6.io/k6/js/modules"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/component/componenttest"
	"go.opentelemetry.io/collector/config"
	"go.opentelemetry.io/collector/config/configgrpc"
	"go.opentelemetry.io/collector/config/configtls"
	"go.opentelemetry.io/collector/exporter/otlpexporter"
	"go.opentelemetry.io/collector/model/pdata"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func init() {
	modules.Register("k6/x/tracing", new(tracingClientModule))
}

type ClientTracing struct {
	vu modules.VU
}

type tracingClientModule struct{}

var _ modules.Module = &tracingClientModule{}

func (r *tracingClientModule) NewModuleInstance(vu modules.VU) modules.Instance {
	return &ClientTracing{vu: vu}
}

func (ct *ClientTracing) Exports() modules.Exports {
	return modules.Exports{
		Named: map[string]interface{}{
			"Client":                ct.xclient,
			"generateRandomTraceID": ct.generateRandomTraceID,
		},
	}
}

type exporter string

const (
	noExporter exporter = ""
	// todo: add http
	otlpExporter exporter = "otlp"
	// todo: add thrift, http
	jaegerExporter exporter = "jaeger"
)

type Client struct {
	exporter component.TracesExporter
	cfg      *Config
	vu       modules.VU
}

type Config struct {
	Exporter       exporter `json:"type"`
	Endpoint       string   `json:"url"`
	Insecure       bool     `json:"insecure"`
	Authentication struct {
		User     string `json:"user"`
		Password string `json:"password"`
	}
	Headers map[string]string `json:"headers"`
}

func (ct *ClientTracing) xclient(g goja.ConstructorCall) *goja.Object {
	var cfg Config
	rt := ct.vu.Runtime()
	err := rt.ExportTo(g.Argument(0), &cfg)
	if err != nil {
		common.Throw(rt, fmt.Errorf("Client constructor expects first argument to be Config"))
	}

	if cfg.Endpoint == "" {
		cfg.Endpoint = "0.0.0.0:4317"
	}

	var (
		factory     component.ExporterFactory
		exporterCfg config.Exporter
	)
	switch cfg.Exporter {
	case noExporter, otlpExporter:
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
	case jaegerExporter:
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
		log.Fatal(fmt.Errorf("failed to init exporter: unknown exporter type %s", cfg.Exporter))
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
	if err != nil {
		log.Fatal(err)
	}
	_ = exporter.Start(context.Background(), componenttest.NewNopHost())

	if err != nil {
		log.Fatal(fmt.Errorf("failed to create exporter: %v", err))
	}

	return rt.ToValue(&Client{
		exporter: exporter,
		cfg:      &cfg,
		vu:       ct.vu,
	}).ToObject(rt)
}

func (ct *ClientTracing) generateRandomTraceID() string {
	return random.TraceID().HexString()
}

type TraceEntry struct {
	ID                string     `json:"id"`
	RandomServiceName bool       `json:"random_service_name"`
	Spans             SpansEntry `json:"spans"`
}

type SpansEntry struct {
	Count      int                    `json:"count"`
	Size       int                    `json:"size"`
	RandomName bool                   `json:"random_name"`
	FixedAttrs map[string]interface{} `json:"fixed_attrs"`
}

func (c *Client) Push(te []TraceEntry) error {
	traceData := pdata.NewTraces()

	rss := traceData.ResourceSpans()
	rss.EnsureCapacity(len(te))

	for _, t := range te {
		generateResource(t, rss.AppendEmpty())
	}

	err := c.exporter.ConsumeTraces(context.Background(), traceData)
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) PushDebug(te []TraceEntry) error {
	for _, t := range te {
		log.Info("Pushing traceID=", t.ID, " spans=", t.Spans.Count, " size=", t.Spans.Size)
	}
	return c.Push(te)
}

func generateResource(t TraceEntry, dest pdata.ResourceSpans) {
	serviceName := random.Service()
	if t.RandomServiceName {
		serviceName += "." + random.String(5)
	}
	dest.Resource().Attributes().InsertString("k6", "true")
	dest.Resource().Attributes().InsertString("service.name", serviceName)

	ilss := dest.InstrumentationLibrarySpans()
	ilss.EnsureCapacity(1)
	ils := ilss.AppendEmpty()
	ils.InstrumentationLibrary().SetName("k6")

	// Spans
	sps := ils.Spans()
	sps.EnsureCapacity(t.Spans.Count)
	for e := 0; e < t.Spans.Count; e++ {
		generateSpan(t, sps.AppendEmpty())
	}
}

func generateSpan(t TraceEntry, dest pdata.Span) {
	endTime := time.Now().Round(time.Second)
	startTime := endTime.Add(-time.Duration(rand.Intn(500)+10) * time.Millisecond)

	var b [16]byte
	traceID, _ := hex.DecodeString(t.ID)
	copy(b[:], traceID)

	spanName := random.Operation()
	if t.Spans.RandomName {
		spanName += "." + random.String(5)
	}

	span := pdata.NewSpan()
	span.SetTraceID(pdata.NewTraceID(b))
	span.SetSpanID(random.SpanID())
	span.SetParentSpanID(random.SpanID())
	span.SetName(spanName)
	span.SetKind(pdata.SpanKindClient)
	span.SetStartTimestamp(pdata.NewTimestampFromTime(startTime))
	span.SetEndTimestamp(pdata.NewTimestampFromTime(endTime))
	span.SetTraceState("x:y")

	event := span.Events().AppendEmpty()
	event.SetName(random.K6String(12))
	event.SetTimestamp(pdata.NewTimestampFromTime(startTime))
	event.Attributes().InsertString(random.K6String(12), random.K6String(12))

	status := span.Status()
	status.SetCode(1)
	status.SetMessage("OK")

	attrs := pdata.NewAttributeMap()

	if len(t.Spans.FixedAttrs) > 0 {
		constructSpanAttributes(t.Spans.FixedAttrs, attrs)
	}

	// Fill the span with some random data
	var size int64
	for {
		if size >= int64(t.Spans.Size) {
			break
		}

		rKey := random.K6String(rand.Intn(15))
		rVal := random.K6String(rand.Intn(15))
		attrs.InsertString(rKey, rVal)

		size += int64(unsafe.Sizeof(rKey)) + int64(unsafe.Sizeof(rVal))
	}

	attrs.CopyTo(span.Attributes())
	span.CopyTo(dest)
}

func (c *Client) Shutdown() error {
	return c.exporter.Shutdown(context.Background())
}

func constructSpanAttributes(attributes map[string]interface{}, dst pdata.AttributeMap) {
	attrs := pdata.NewAttributeMap()
	for key, value := range attributes {
		if cast, ok := value.(int); ok {
			attrs.InsertInt(key, int64(cast))
		} else if cast, ok := value.(int64); ok {
			attrs.InsertInt(key, cast)
		} else {
			attrs.InsertString(key, fmt.Sprintf("%v", value))
		}
	}
	attrs.CopyTo(dst)
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
