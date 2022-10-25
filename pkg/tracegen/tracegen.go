package tracegen

import (
	"encoding/hex"
	"fmt"
	"math/rand"
	"time"
	"unsafe"

	"github.com/grafana/xk6-client-tracing/pkg/random"
	"go.opentelemetry.io/collector/model/pdata"
)

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

func GenerateResource(t TraceEntry, dest pdata.ResourceSpans) {
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
