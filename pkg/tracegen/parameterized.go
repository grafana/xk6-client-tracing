package tracegen

import (
	"encoding/hex"
	"fmt"
	"strconv"
	"time"
	"unsafe"

	"github.com/grafana/xk6-client-tracing/pkg/random"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/ptrace"
)

const (
	defaultTraceCount   = 1
	defaultSpanCount    = 10
	defaultSpanSize     = 1000
	defaultResourceSize = 0
)

type TraceParams struct {
	ID                string     `json:"id"`
	ParentID          string     `json:"parent_id"`
	RandomServiceName bool       `json:"random_service_name"`
	ResourceSize      int        `json:"resource_size"`
	Count             int        `json:"count"`
	Spans             SpanParams `json:"spans"`
}

type SpanParams struct {
	Count      int                    `json:"count"`
	Size       int                    `json:"size"`
	RandomName bool                   `json:"random_name"`
	FixedAttrs map[string]interface{} `json:"fixed_attrs"`
}

func (tp *TraceParams) setDefaults() {
	if tp.Count == 0 {
		tp.Count = defaultTraceCount
	}
	if tp.ResourceSize <= 0 {
		tp.ResourceSize = defaultResourceSize
	}
	if tp.Spans.Count == 0 {
		tp.Spans.Count = defaultSpanCount
	}
	if tp.Spans.Size <= 0 {
		tp.Spans.Size = defaultSpanSize
	}
}

func NewParameterizedGenerator(traceParams []*TraceParams) *ParameterizedGenerator {
	for _, tp := range traceParams {
		tp.setDefaults()
	}

	return &ParameterizedGenerator{
		traceParams: traceParams,
	}
}

type ParameterizedGenerator struct {
	traceParams []*TraceParams
}

func (g *ParameterizedGenerator) Traces() ptrace.Traces {
	traceData := ptrace.NewTraces()

	resourceSpans := traceData.ResourceSpans()
	resourceSpans.EnsureCapacity(len(g.traceParams))

	for _, te := range g.traceParams {
		rspan := resourceSpans.AppendEmpty()
		serviceName := random.Service()
		if te.RandomServiceName {
			serviceName += "." + random.String(5)
		}
		resourceAttributes := g.constructAttributes(te.ResourceSize)
		resourceAttributes.CopyTo(rspan.Resource().Attributes())
		rspan.Resource().Attributes().PutStr("k6", "true")
		rspan.Resource().Attributes().PutStr(attrServiceName, serviceName)

		ilss := rspan.ScopeSpans()
		ilss.EnsureCapacity(1)
		ils := ilss.AppendEmpty()
		ils.Scope().SetName("k6-scope-name/" + random.String(15))
		ils.Scope().SetVersion("k6-scope-version:v" + strconv.Itoa(random.IntBetween(0, 99)) + "." + strconv.Itoa(random.IntBetween(0, 99)))

		for range te.Count {
			// Randomize traceID every time if we're generating multiple traces
			if te.ID == "" || te.Count > 1 {
				te.ID = random.TraceID().String()
				te.ParentID = ""
			}

			// Spans
			sps := ils.Spans()
			sps.EnsureCapacity(te.Spans.Count)
			for e := range te.Spans.Count {
				if e == 0 {
					g.generateSpan(te, sps.AppendEmpty())
					idxSpan := sps.At(0)
					te.ParentID = idxSpan.SpanID().String()
				} else {
					g.generateSpan(te, sps.AppendEmpty())
				}
			}
		}
	}

	return traceData
}

func (g *ParameterizedGenerator) generateSpan(t *TraceParams, dest ptrace.Span) {
	endTime := time.Now().Round(time.Second)
	startTime := endTime.Add(-time.Duration(random.IntN(500)+10) * time.Millisecond)

	var traceID pcommon.TraceID
	b, _ := hex.DecodeString(t.ID)
	copy(traceID[:], b)

	spanName := random.Operation()
	if t.Spans.RandomName {
		spanName += "." + random.String(5)
	}

	span := ptrace.NewSpan()
	span.SetTraceID(traceID)
	if t.ParentID != "" {
		var parentID pcommon.SpanID
		p, _ := hex.DecodeString(t.ParentID)
		copy(parentID[:], p)
		span.SetParentSpanID(parentID)
	}
	span.SetSpanID(random.SpanID())
	span.SetName(spanName)
	span.SetKind(ptrace.SpanKindClient)
	span.SetStartTimestamp(pcommon.NewTimestampFromTime(startTime))
	span.SetEndTimestamp(pcommon.NewTimestampFromTime(endTime))
	span.TraceState().FromRaw("ot=x:y")

	event := span.Events().AppendEmpty()
	event.SetName(random.K6String(12))
	event.SetTimestamp(pcommon.NewTimestampFromTime(startTime))
	event.Attributes().PutStr(random.K6String(5), random.K6String(12))

	link := span.Links().AppendEmpty()
	link.SetTraceID(traceID)
	link.SetSpanID(random.SpanID())
	link.Attributes().PutStr(random.K6String(12), random.K6String(12))

	status := span.Status()
	status.SetCode(1)
	status.SetMessage("OK")

	attrs := g.constructAttributes(t.Spans.Size)
	g.constructSpanAttributes(t.Spans.FixedAttrs, attrs)

	attrs.CopyTo(span.Attributes())
	span.CopyTo(dest)
}

func (g *ParameterizedGenerator) constructSpanAttributes(attributes map[string]interface{}, dst pcommon.Map) {
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

func (g *ParameterizedGenerator) constructAttributes(size int) pcommon.Map {
	attrs := pcommon.NewMap()

	// Fill the span with some random data
	var currentSize int64
	for currentSize < int64(size) {
		rKey := random.K6String(random.IntN(15) + 1)
		rVal := random.K6String(random.IntN(15) + 1)
		attrs.PutStr(rKey, rVal)

		currentSize += int64(unsafe.Sizeof(rKey)) + int64(unsafe.Sizeof(rVal))
	}

	return attrs
}
