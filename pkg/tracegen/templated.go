package tracegen

import (
	"fmt"
	"github.com/grafana/xk6-client-tracing/pkg/random"
	"github.com/pkg/errors"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/ptrace"
	"strings"
	"time"
)

type OTelSemantics int

const (
	SemanticsNone OTelSemantics = iota
	SemanticsHTTP
	SemanticsDB

	serviceName = "service.name"
	spanKind    = "span.kind"

	defaultMinDuration = time.Millisecond * 500
	defaultMaxDuration = time.Millisecond * 800
)

type Range struct {
	Min int64
	Max int64
}

type AttributeParams struct {
	Count       int
	Cardinality *int
}

type SpanDefaults struct {
	AttributeSemantics OTelSemantics          `js:"attributeSemantics"`
	Attributes         map[string]interface{} `js:"attributes"`
	RandomAttributes   *AttributeParams       `js:"randomAttributes"`
}

type SpanTemplate struct {
	Service            string                 `js:"service"`
	Name               *string                `js:"name"`
	ParentIDX          *int                   `js:"parentIdx"`
	Duration           *Range                 `js:"duration"`
	AttributeSemantics OTelSemantics          `js:"attributeSemantics"`
	Attributes         map[string]interface{} `js:"attributes"`
	RandomAttributes   *AttributeParams       `js:"randomAttributes"`
}

type TraceTemplate struct {
	Defaults *SpanDefaults  `js:"defaults"`
	Spans    []SpanTemplate `js:"spans"`
}

func NewTemplatedGenerator(template *TraceTemplate) (*TemplatedGenerator, error) {
	fmt.Printf("NewTemplatedGenerator(template='%p')\n", template)
	gen := &TemplatedGenerator{}
	err := gen.initialize(template)
	if err != nil {
		return nil, errors.Wrap(err, "fail to create new templated generator")
	}
	return gen, nil
}

type TemplatedGenerator struct {
	//template *TraceTemplate
	randomAttributes map[string][]interface{}
	resources        map[string]*internalResourceTemplate
	spans            []*internalSpanTemplate
}

type internalSpanTemplate struct {
	idx                int
	resource           *internalResourceTemplate
	parent             *internalSpanTemplate
	name               string
	kind               ptrace.SpanKind
	duration           *Range
	attributeSemantics OTelSemantics
	attributes         map[string]interface{}
	randomAttributes   map[string][]interface{}
}

type internalResourceTemplate struct {
	service string
}

func (g *TemplatedGenerator) Traces() ptrace.Traces {
	var (
		traceID      = random.TraceID()
		traceData    = ptrace.NewTraces()
		resSpanSlice = traceData.ResourceSpans()
		resSpanMap   = map[string]ptrace.ResourceSpans{}
		spans        []ptrace.Span
	)

	var indexes [][2]int
	for i, t := range g.spans {
		pidx := -1
		if t.parent != nil {
			pidx = t.parent.idx
		}
		indexes = append(indexes, [2]int{i, pidx})
	}

	for _, tmpl := range g.spans {
		// get or generate the corresponding ResourceSpans
		resSpans, found := resSpanMap[tmpl.resource.service]
		if !found {
			resSpans = g.generateResourceSpans(resSpanSlice, tmpl.resource)
			resSpanMap[tmpl.resource.service] = resSpans
		}
		scopeSpans := resSpans.ScopeSpans().At(0)

		// generate new span
		var parent *ptrace.Span
		if tmpl.parent != nil {
			parent = &spans[tmpl.parent.idx]
		}
		spans = append(spans, g.generateSpan(scopeSpans, tmpl, parent, traceID))
	}

	return traceData
}

func (g *TemplatedGenerator) generateResourceSpans(resSpanSlice ptrace.ResourceSpansSlice, tmpl *internalResourceTemplate) ptrace.ResourceSpans {
	resSpans := resSpanSlice.AppendEmpty()
	resSpans.Resource().Attributes().PutStr("k6", "true")
	resSpans.Resource().Attributes().PutStr("service.name", tmpl.service)

	scopeSpans := resSpans.ScopeSpans().AppendEmpty()
	scopeSpans.Scope().SetName("k6")
	return resSpans
}

func (g *TemplatedGenerator) generateSpan(scopeSpans ptrace.ScopeSpans, tmpl *internalSpanTemplate, parent *ptrace.Span, traceID pcommon.TraceID) ptrace.Span {
	span := scopeSpans.Spans().AppendEmpty()

	span.SetTraceID(traceID)
	span.SetSpanID(random.SpanID())
	if parent != nil {
		span.SetParentSpanID(parent.SpanID())
	}
	span.SetName(tmpl.name)
	span.SetKind(tmpl.kind)

	// set start and end time
	var start time.Time
	var duration time.Duration
	if parent == nil {
		start = time.Now().Add(-5 * time.Second)
		if tmpl.duration == nil {
			duration = random.Duration(defaultMinDuration, defaultMaxDuration)
		}
	} else {
		pStart := parent.StartTimestamp().AsTime()
		pDuration := parent.EndTimestamp().AsTime().Sub(pStart)
		start = pStart.Add(random.Duration(pDuration/20, pDuration/10))
		if tmpl.duration == nil {
			duration = random.Duration(pDuration/2, pDuration-pDuration/10)
		}
	}
	if tmpl.duration != nil {
		duration = random.Duration(time.Duration(tmpl.duration.Min)*time.Millisecond, time.Duration(tmpl.duration.Max)*time.Millisecond)
	}
	end := start.Add(duration)
	span.SetStartTimestamp(pcommon.NewTimestampFromTime(start))
	span.SetEndTimestamp(pcommon.NewTimestampFromTime(end))

	return span
}

func (g *TemplatedGenerator) initialize(template *TraceTemplate) error {
	g.resources = map[string]*internalResourceTemplate{}
	g.randomAttributes = map[string][]interface{}{}

	for i, tmpl := range template.Spans {
		// span templates must have a service
		if tmpl.Service == "" {
			return errors.New("trace template invalid: span template must have a name")
		}

		// get or generate the corresponding ResourceSpans
		res, found := g.resources[tmpl.Service]
		if !found {
			res = &internalResourceTemplate{
				service: tmpl.Service,
			}
			g.resources[tmpl.Service] = res
		}

		// span template parent index must reference a previous span
		if tmpl.ParentIDX != nil && (*tmpl.ParentIDX >= i || *tmpl.ParentIDX < 0) {
			return errors.New("trace template invalid: span index must be greater than span parent index")
		}

		// initialize span using information from the parent span, the template and child template
		parentIdx := i - 1
		if tmpl.ParentIDX != nil {
			parentIdx = *tmpl.ParentIDX
		}
		var parent *internalSpanTemplate
		if parentIdx >= 0 {
			parent = g.spans[parentIdx]
		}

		var child *SpanTemplate
		for j := i + 1; j < len(template.Spans); j++ {
			n := template.Spans[j]
			if n.ParentIDX == nil || *n.ParentIDX == i {
				child = &n
				break
			}
		}

		span, err := g.initializeSpan(i, parent, &tmpl, child)
		if err != nil {
			return err
		}
		g.spans = append(g.spans, span)
	}

	return nil
}

func (g *TemplatedGenerator) initializeSpan(idx int, parent *internalSpanTemplate, tmpl, child *SpanTemplate) (*internalSpanTemplate, error) {
	res, _ := g.resources[tmpl.Service]
	span := internalSpanTemplate{
		idx:                idx,
		parent:             parent,
		resource:           res,
		duration:           tmpl.Duration,
		attributeSemantics: tmpl.AttributeSemantics,
		attributes:         tmpl.Attributes,
	}

	// set span name
	if tmpl.Name != nil {
		span.name = *tmpl.Name
	} else {
		span.name = random.Operation()
	}

	// set span kind
	if kind, found := tmpl.Attributes[spanKind]; found {
		kindStr, ok := kind.(string)
		if !ok {
			return nil, errors.Errorf("attribute %s expected to be a string, but was %T", spanKind, kind)
		}
		span.kind = spanKindFromString(kindStr)
	} else {
		if parent == nil {
			if child == nil || tmpl.Service == child.Service {
				span.kind = ptrace.SpanKindServer
			} else {
				span.kind = ptrace.SpanKindClient
			}
		} else {
			parentService := parent.resource.service
			if tmpl.Service != parentService {
				span.kind = ptrace.SpanKindServer
			} else if child != nil && tmpl.Service != child.Service {
				span.kind = ptrace.SpanKindClient
			} else {
				span.kind = ptrace.SpanKindInternal
			}
		}
	}

	return &span, nil
}

func spanKindFromString(s string) ptrace.SpanKind {
	s = strings.ToLower(s)
	s = strings.TrimPrefix(s, "span_kind_")
	switch s {
	case "internal":
		return ptrace.SpanKindInternal
	case "server":
		return ptrace.SpanKindServer
	case "client":
		return ptrace.SpanKindClient
	case "producer":
		return ptrace.SpanKindProducer
	case "consumer":
		return ptrace.SpanKindConsumer
	default:
		return ptrace.SpanKindUnspecified
	}
}
