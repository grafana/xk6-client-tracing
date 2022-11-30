package tracegen

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/grafana/xk6-client-tracing/pkg/random"
	"github.com/grafana/xk6-client-tracing/pkg/util"
	"github.com/pkg/errors"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/ptrace"
)

// OTelSemantics describes a specific set of OpenTelemetry semantic conventions.
type OTelSemantics string

const (
	SemanticsHTTP OTelSemantics = "http"
	SemanticsDB   OTelSemantics = "db"

	defaultMinDuration                = time.Millisecond * 500
	defaultMaxDuration                = time.Millisecond * 800
	defaultRandomAttributeCardinality = 20
	randomAttributeKeySize            = 15
	randomAttributeValueSize          = 30
)

// Range represents and interval with the given upper and lower bound [Max, Min)
type Range struct {
	Min int64
	Max int64
}

// AttributeParams describe how random attributes should be created.
type AttributeParams struct {
	// Count the number of attributes to creat.
	Count int
	// Cardinality how many distinct values are created for each attribute.
	Cardinality *int
}

// SpanDefaults contains template parameters that are applied to all generated spans.
type SpanDefaults struct {
	// AttributeSemantics whether to create attributes that follow specific OpenTelemetry semantics.
	AttributeSemantics *OTelSemantics `js:"attributeSemantics"`
	// Attributes that are added to each span.
	Attributes map[string]interface{} `js:"attributes"`
	// RandomAttributes random attributes generated for each span.
	RandomAttributes *AttributeParams `js:"randomAttributes"`
}

// SpanTemplate parameters that define how a span is created.
type SpanTemplate struct {
	// Service is used to set the service.name attribute of the corresponding resource span.
	Service string `js:"service"`
	// Name represents the name of the span. If empty, the name will be randomly generated.
	Name *string `js:"name"`
	// ParentIDX defines the index of the parent span in TraceTemplate.Spans. ParentIDX must be smaller than the
	// own index. If empty, the parent is the span with the position directly before this span in TraceTemplate.Spans.
	ParentIDX *int `js:"parentIdx"`
	// Duration defines the interval for the generated span duration. If missing, a random duration is generated that
	// is shorter than the duration of the parent span.
	Duration *Range `js:"duration"`
	// AttributeSemantics can be set in order to generate attributes that follow a certain OpenTelemetry semantic
	// convention. So far only semantic convention for HTTP requests is supported.
	AttributeSemantics *OTelSemantics `js:"attributeSemantics"`
	// Attributes that are added to this span.
	Attributes map[string]interface{} `js:"attributes"`
	// RandomAttributes parameters to configure the creation of random attributes. If missing, no random attributes
	// are added to the span.
	RandomAttributes *AttributeParams `js:"randomAttributes"`
}

// TraceTemplate describes how all a trace and it's spans are generated.
type TraceTemplate struct {
	// Defaults parameters that are applied to each generated span.
	Defaults SpanDefaults `js:"defaults"`
	// Spans parameters for the individual spans of a trace.
	Spans []SpanTemplate `js:"spans"`
}

// NewTemplatedGenerator creates a new trace generator.
func NewTemplatedGenerator(template *TraceTemplate) (*TemplatedGenerator, error) {
	gen := &TemplatedGenerator{}
	err := gen.initialize(template)
	if err != nil {
		return nil, errors.Wrap(err, "fail to create new templated generator")
	}
	return gen, nil
}

// TemplatedGenerator a trace generator that creates randomized traces based on a given TraceTemplate.
// The generator interprets the template parameters such that realistically looking traces with consistent
// spans and attributes are generated.
type TemplatedGenerator struct {
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
	attributeSemantics *OTelSemantics
	attributes         map[string]interface{}
	randomAttributes   map[string][]interface{}
}

type internalResourceTemplate struct {
	service   string
	hostName  string
	hostIP    string
	transport string
	hostPort  int
}

// Traces implements Generator for TemplatedGenerator
func (g *TemplatedGenerator) Traces() ptrace.Traces {
	var (
		traceID      = random.TraceID()
		traceData    = ptrace.NewTraces()
		resSpanSlice = traceData.ResourceSpans()
		resSpanMap   = map[string]ptrace.ResourceSpans{}
		spans        []ptrace.Span
	)

	randomTraceAttributes := make(map[string]interface{}, len(g.randomAttributes))
	for k, v := range g.randomAttributes {
		randomTraceAttributes[k] = random.SelectElement(v)
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
		s := g.generateSpan(scopeSpans, tmpl, parent, traceID)

		// attributes
		for k, v := range randomTraceAttributes {
			if _, found := s.Attributes().Get(k); !found {
				s.Attributes().PutEmpty(k).FromRaw(v)
			}
		}

		spans = append(spans, s)
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

	// add attributes
	for k, v := range tmpl.attributes {
		span.Attributes().PutEmpty(k).FromRaw(v)
	}

	for k, v := range tmpl.randomAttributes {
		span.Attributes().PutEmpty(k).FromRaw(random.SelectElement(v))
	}

	g.generateNetworkAttributes(tmpl, &span, parent)
	if tmpl.attributeSemantics != nil {
		switch *tmpl.attributeSemantics {
		case SemanticsHTTP:
			g.generateHTTPAttributes(tmpl, &span, parent)
		}
	}

	return span
}

func (g *TemplatedGenerator) generateNetworkAttributes(tmpl *internalSpanTemplate, span, parent *ptrace.Span) {
	if tmpl.kind == ptrace.SpanKindInternal {
		return
	}

	putIfNotExists(span.Attributes(), "net.transport", tmpl.resource.transport)
	putIfNotExists(span.Attributes(), "net.sock.family", "inet")
	if tmpl.kind == ptrace.SpanKindClient {
		putIfNotExists(span.Attributes(), "net.peer.port", random.Port())
	} else if tmpl.kind == ptrace.SpanKindServer {
		putIfNotExists(span.Attributes(), "net.sock.host.addr", tmpl.resource.hostIP)
		putIfNotExists(span.Attributes(), "net.host.name", tmpl.resource.hostName)
		putIfNotExists(span.Attributes(), "net.host.port", tmpl.resource.hostPort)

		if parent != nil && parent.Kind() == ptrace.SpanKindClient {
			ip, _ := span.Attributes().Get("net.sock.host.addr")
			putIfNotExists(parent.Attributes(), "net.sock.peer.addr", ip.Str())
			name, _ := span.Attributes().Get("net.host.name")
			putIfNotExists(parent.Attributes(), "net.peer.name", name.Str())
		}
	}
}

func (g *TemplatedGenerator) generateHTTPAttributes(tmpl *internalSpanTemplate, span, parent *ptrace.Span) {
	if tmpl.kind == ptrace.SpanKindInternal {
		return
	}
	parentAttr := pcommon.NewMap()
	if parent != nil {
		parentAttr = parent.Attributes()
	}

	putIfNotExists(span.Attributes(), "http.flavor", "1.1")

	if tmpl.kind == ptrace.SpanKindServer {
		var method string
		if m, found := span.Attributes().Get("http.method"); found {
			method = m.Str()
		} else if m, found = parentAttr.Get("http.method"); found {
			method = m.Str()
		} else {
			method = random.HTTPMethod()
			span.Attributes().PutStr("http.method", method)
		}

		var status int64
		if st, found := span.Attributes().Get("http.status_code"); found {
			status = st.Int()
		} else if st, found = parentAttr.Get("http.status_code"); found {
			status = st.Int()
		} else {
			status = random.HTTPStatusSuccess()
			span.Attributes().PutInt("http.status_code", status)
		}
		if status >= 500 {
			span.Status().SetCode(ptrace.StatusCodeError)
			span.Status().SetMessage(http.StatusText(int(status)))
		}

		var requestURL *url.URL
		if u, found := span.Attributes().Get("http.url"); found {
			requestURL, _ = url.ParseRequestURI(u.Str())
		} else if u, found = parentAttr.Get("http.url"); found {
			requestURL, _ = url.ParseRequestURI(u.Str())
		} else {
			requestURL, _ = url.ParseRequestURI(fmt.Sprintf("https://%s:%d/%s", tmpl.resource.hostName, tmpl.resource.hostPort, tmpl.name))
			span.Attributes().PutStr("http.url", requestURL.String())
		}
		span.Attributes().PutStr("http.scheme", requestURL.Scheme)
		span.Attributes().PutStr("http.target", requestURL.Path)

		putIfNotExists(span.Attributes(), "http.response_content_length", random.IntBetween(100_000, 1_000_000))
		if method == http.MethodPatch || method == http.MethodPost || method == http.MethodPut {
			putIfNotExists(span.Attributes(), "http.request_content_length", random.IntBetween(10_000, 100_000))
		}

		if parent != nil && parent.Kind() == ptrace.SpanKindClient {
			if status >= 400 {
				parent.Status().SetCode(ptrace.StatusCodeError)
				parent.Status().SetMessage(http.StatusText(int(status)))
			}
			putIfNotExists(parent.Attributes(), "http.method", method)
			putIfNotExists(parent.Attributes(), "http.status_code", status)
			putIfNotExists(parent.Attributes(), "http.url", requestURL.String())
			res, _ := span.Attributes().Get("http.response_content_length")
			putIfNotExists(parent.Attributes(), "http.response_content_length", res.Int())
			if req, found := span.Attributes().Get("http.request_content_length"); found {
				putIfNotExists(span.Attributes(), "http.request_content_length", req.Int())
			}
		}
	}
}

func (g *TemplatedGenerator) initialize(template *TraceTemplate) error {
	g.resources = map[string]*internalResourceTemplate{}
	g.randomAttributes = initializeRandomAttributes(template.Defaults.RandomAttributes)

	for i, tmpl := range template.Spans {
		// span templates must have a service
		if tmpl.Service == "" {
			return errors.New("trace template invalid: span template must have a name")
		}

		// get or generate the corresponding ResourceSpans
		_, found := g.resources[tmpl.Service]
		if !found {
			res := g.initializeResource(&tmpl)
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

		span, err := g.initializeSpan(i, parent, &template.Defaults, &tmpl, child)
		if err != nil {
			return err
		}
		g.spans = append(g.spans, span)
	}

	return nil
}

func (g *TemplatedGenerator) initializeResource(tmpl *SpanTemplate) *internalResourceTemplate {
	return &internalResourceTemplate{
		service:   tmpl.Service,
		hostName:  fmt.Sprintf("%s.local", tmpl.Service),
		hostIP:    random.IPAddr(),
		hostPort:  random.Port(),
		transport: "ip_tcp",
	}
}

func (g *TemplatedGenerator) initializeSpan(idx int, parent *internalSpanTemplate, defaults *SpanDefaults, tmpl, child *SpanTemplate) (*internalSpanTemplate, error) {
	res := g.resources[tmpl.Service]
	span := internalSpanTemplate{
		idx:      idx,
		parent:   parent,
		resource: res,
		duration: tmpl.Duration,
	}

	// apply defaults
	if tmpl.AttributeSemantics == nil {
		span.attributeSemantics = defaults.AttributeSemantics
	}
	span.attributes = util.MergeMaps(defaults.Attributes, tmpl.Attributes)

	// set span name
	if tmpl.Name != nil {
		span.name = *tmpl.Name
	} else {
		span.name = random.Operation()
	}

	kind, err := initializeSpanKind(parent, tmpl, child)
	if err != nil {
		return nil, err
	}
	span.kind = kind

	span.randomAttributes = initializeRandomAttributes(tmpl.RandomAttributes)

	return &span, nil
}

func initializeSpanKind(parent *internalSpanTemplate, tmpl, child *SpanTemplate) (ptrace.SpanKind, error) {
	var kind ptrace.SpanKind
	if k, found := tmpl.Attributes["span.kind"]; found {
		kindStr, ok := k.(string)
		if !ok {
			return ptrace.SpanKindUnspecified, errors.Errorf("attribute %s expected to be a string, but was %T", "span.kind", k)
		}
		kind = spanKindFromString(kindStr)
	} else {
		if parent == nil {
			if child == nil || tmpl.Service == child.Service {
				kind = ptrace.SpanKindServer
			} else {
				kind = ptrace.SpanKindClient
			}
		} else {
			parentService := parent.resource.service
			if tmpl.Service != parentService {
				kind = ptrace.SpanKindServer
			} else if child != nil && tmpl.Service != child.Service {
				kind = ptrace.SpanKindClient
			} else {
				kind = ptrace.SpanKindInternal
			}
		}
	}
	return kind, nil
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

func putIfNotExists(m pcommon.Map, k string, v interface{}) {
	if _, found := m.Get(k); !found {
		m.PutEmpty(k).FromRaw(v)
	}
}

func initializeRandomAttributes(attributeParams *AttributeParams) map[string][]interface{} {
	if attributeParams == nil {
		return map[string][]interface{}{}
	}

	if attributeParams.Cardinality == nil {
		tmp := defaultRandomAttributeCardinality
		attributeParams.Cardinality = &tmp
	}

	attributes := make(map[string][]interface{}, attributeParams.Count)
	for i := 0; i < attributeParams.Count; i++ {
		key := random.K6String(randomAttributeKeySize)
		values := make([]interface{}, 0, *attributeParams.Cardinality)
		for j := 0; j < *attributeParams.Cardinality; j++ {
			values = append(values, random.String(randomAttributeValueSize))
		}
		attributes[key] = values
	}

	return attributes
}
