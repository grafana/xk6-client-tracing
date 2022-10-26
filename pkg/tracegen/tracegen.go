package tracegen

import (
	"encoding/hex"
	"fmt"
	"math/rand"
	"time"
	"unsafe"

	"github.com/grafana/xk6-client-tracing/pkg/random"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/ptrace"
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

func GenerateResource(traceEntries []TraceEntry) ptrace.Traces {
	traceData := ptrace.NewTraces()

	resourceSpans := traceData.ResourceSpans()
	resourceSpans.EnsureCapacity(len(traceEntries))

	for _, te := range traceEntries {
		rspan := resourceSpans.AppendEmpty()
		serviceName := random.Service()
		if te.RandomServiceName {
			serviceName += "." + random.String(5)
		}
		rspan.Resource().Attributes().PutStr("k6", "true")
		rspan.Resource().Attributes().PutStr("service.name", serviceName)

		ilss := rspan.ScopeSpans()
		ilss.EnsureCapacity(1)
		ils := ilss.AppendEmpty()
		ils.Scope().SetName("k6")

		// Spans
		sps := ils.Spans()
		sps.EnsureCapacity(te.Spans.Count)
		for e := 0; e < te.Spans.Count; e++ {
			generateSpan(te, sps.AppendEmpty())
		}
	}

	return traceData
}

func generateSpan(t TraceEntry, dest ptrace.Span) {
	endTime := time.Now().Round(time.Second)
	startTime := endTime.Add(-time.Duration(rand.Intn(500)+10) * time.Millisecond)

	var b [16]byte
	traceID, _ := hex.DecodeString(t.ID)
	copy(b[:], traceID)

	spanName := random.Operation()
	if t.Spans.RandomName {
		spanName += "." + random.String(5)
	}

	span := ptrace.NewSpan()
	span.SetTraceID(b)
	span.SetSpanID(random.SpanID())
	span.SetParentSpanID(random.SpanID())
	span.SetName(spanName)
	span.SetKind(ptrace.SpanKindClient)
	span.SetStartTimestamp(pcommon.NewTimestampFromTime(startTime))
	span.SetEndTimestamp(pcommon.NewTimestampFromTime(endTime))
	span.TraceState().FromRaw("x:y")

	event := span.Events().AppendEmpty()
	event.SetName(random.K6String(12))
	event.SetTimestamp(pcommon.NewTimestampFromTime(startTime))
	event.Attributes().PutStr(random.K6String(12), random.K6String(12))

	status := span.Status()
	status.SetCode(1)
	status.SetMessage("OK")

	attrs := pcommon.NewMap()
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
		attrs.PutStr(rKey, rVal)

		size += int64(unsafe.Sizeof(rKey)) + int64(unsafe.Sizeof(rVal))
	}

	attrs.CopyTo(span.Attributes())
	span.CopyTo(dest)
}

func constructSpanAttributes(attributes map[string]interface{}, dst pcommon.Map) {
	attrs := pcommon.NewMap()
	for key, value := range attributes {
		if cast, ok := value.(int); ok {
			attrs.PutInt(key, int64(cast))
		} else if cast, ok := value.(int64); ok {
			attrs.PutInt(key, cast)
		} else {
			attrs.PutStr(key, fmt.Sprintf("%v", value))
		}
	}
	attrs.CopyTo(dst)
}
