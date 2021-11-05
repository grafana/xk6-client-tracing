package xk6_client_tracing

import (
	"context"
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"time"

	"go.k6.io/k6/js/common"
	"go.k6.io/k6/js/modules"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/component/componenttest"
	"go.opentelemetry.io/collector/config/configgrpc"
	"go.opentelemetry.io/collector/config/configtls"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/exporter/otlpexporter"
	"go.opentelemetry.io/collector/model/pdata"
	conventions "go.opentelemetry.io/collector/model/semconv/v1.5.0"
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

type ClientTracing struct{}

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

	rt := common.GetRuntime(*ctxPtr)
	return common.Bind(rt, &Client{
		exporter: exporter,
		cfg:      &config,
	}, ctxPtr)
}

type Client struct {
	exporter consumer.Traces
	cfg      *Config
}

func (c *Client) SendFake(ctx context.Context) error {
	tc := constructFakeSpanData()
	err := c.exporter.ConsumeTraces(context.Background(), tc)
	if err != nil {
		return err
	}
	return nil
}

func constructFakeSpanData() pdata.Traces {
	resource := constructResource()

	traces := pdata.NewTraces()
	rspans := traces.ResourceSpans().AppendEmpty()
	resource.CopyTo(rspans.Resource())
	ispans := rspans.InstrumentationLibrarySpans().AppendEmpty()
	constructHTTPClientSpan().CopyTo(ispans.Spans().AppendEmpty())
	constructHTTPServerSpan().CopyTo(ispans.Spans().AppendEmpty())
	return traces
}

func constructResource() pdata.Resource {
	resource := pdata.NewResource()
	attrs := pdata.NewAttributeMap()
	attrs.InsertString(conventions.AttributeServiceName, "signup_aggregator")
	attrs.InsertString(conventions.AttributeContainerName, "signup_aggregator")
	attrs.InsertString(conventions.AttributeContainerImageName, "otel/signupaggregator")
	attrs.InsertString(conventions.AttributeContainerImageTag, "v1")
	attrs.InsertString(conventions.AttributeCloudProvider, conventions.AttributeCloudProviderAWS)
	attrs.InsertString(conventions.AttributeCloudAccountID, "999999998")
	attrs.InsertString(conventions.AttributeCloudRegion, "us-west-2")
	attrs.InsertString(conventions.AttributeCloudAvailabilityZone, "us-west-1b")
	attrs.CopyTo(resource.Attributes())
	return resource
}
func constructHTTPClientSpan() pdata.Span {
	attributes := make(map[string]interface{})
	attributes[conventions.AttributeHTTPMethod] = "GET"
	attributes[conventions.AttributeHTTPURL] = "https://api.example.com/users/junit"
	attributes[conventions.AttributeHTTPStatusCode] = 200
	endTime := time.Now().Round(time.Second)
	startTime := endTime.Add(-90 * time.Second)
	spanAttributes := constructSpanAttributes(attributes)

	span := pdata.NewSpan()
	span.SetTraceID(newTraceID())
	span.SetSpanID(newSegmentID())
	span.SetParentSpanID(newSegmentID())
	span.SetName("/users/junit")
	span.SetKind(pdata.SpanKindClient)
	span.SetStartTimestamp(pdata.NewTimestampFromTime(startTime))
	span.SetEndTimestamp(pdata.NewTimestampFromTime(endTime))

	status := pdata.NewSpanStatus()
	status.SetCode(0)
	status.SetMessage("OK")
	status.CopyTo(span.Status())

	spanAttributes.CopyTo(span.Attributes())
	return span
}
func constructHTTPServerSpan() pdata.Span {
	attributes := make(map[string]interface{})
	attributes[conventions.AttributeHTTPMethod] = "GET"
	attributes[conventions.AttributeHTTPURL] = "https://api.example.com/users/junit"
	attributes[conventions.AttributeHTTPClientIP] = "192.168.15.32"
	attributes[conventions.AttributeHTTPStatusCode] = 200
	endTime := time.Now().Round(time.Second)
	startTime := endTime.Add(-90 * time.Second)
	spanAttributes := constructSpanAttributes(attributes)

	span := pdata.NewSpan()
	span.SetTraceID(newTraceID())
	span.SetSpanID(newSegmentID())
	span.SetParentSpanID(newSegmentID())
	span.SetName("/users/junit")
	span.SetKind(pdata.SpanKindServer)
	span.SetStartTimestamp(pdata.NewTimestampFromTime(startTime))
	span.SetEndTimestamp(pdata.NewTimestampFromTime(endTime))

	status := pdata.NewSpanStatus()
	status.SetCode(0)
	status.SetMessage("OK")
	status.CopyTo(span.Status())

	spanAttributes.CopyTo(span.Attributes())
	return span
}

func constructSpanAttributes(attributes map[string]interface{}) pdata.AttributeMap {
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
	return attrs
}

func newTraceID() pdata.TraceID {
	var r [16]byte
	epoch := time.Now().Unix()
	binary.BigEndian.PutUint32(r[0:4], uint32(epoch))
	_, err := rand.Read(r[4:])
	if err != nil {
		panic(err)
	}
	return pdata.NewTraceID(r)
}

func newSegmentID() pdata.SpanID {
	var r [8]byte
	_, err := rand.Read(r[:])
	if err != nil {
		panic(err)
	}
	return pdata.NewSpanID(r)
}
