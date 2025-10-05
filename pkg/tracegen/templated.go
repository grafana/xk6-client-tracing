package tracegen

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/grafana/xk6-client-tracing/pkg/random"
	"github.com/grafana/xk6-client-tracing/pkg/util"
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
	// Count the number of attributes to create.
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
	// Random events generated for each span
	RandomEvents *EventParams `js:"randomEvents"`
	// Random links generated for each span
	RandomLinks *LinkParams `js:"randomLinks"`
	// Resource controls the default attributes for all resources.
	Resource *ResourceTemplate `js:"resource"`
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
	// List of events for the span with specific parameters
	Events []Event `js:"events"`
	// List of links for the span with specific parameters
	Links []Link `js:"links"`
	// Generate random events for the span
	RandomEvents *EventParams `js:"randomEvents"`
	// Generate random links for the span
	RandomLinks *LinkParams `js:"randomLinks"`
	// Resource controls the attributes generated for the resource. Spans with the same Service will have the same
	// resource. Multiple resource definitions will be merged.
	Resource *ResourceTemplate `js:"resource"`
}

type ResourceTemplate struct {
	// Attributes that are added to this resource.
	Attributes map[string]interface{} `js:"attributes"`
	// RandomAttributes parameters to configure the creation of random attributes. If missing, no random attributes
	// are added to the resource.
	RandomAttributes *AttributeParams `js:"randomAttributes"`
}

// TraceTemplate describes how all a trace and it's spans are generated.
type TraceTemplate struct {
	// Defaults parameters that are applied to each generated span.
	Defaults SpanDefaults `js:"defaults"`
	// Spans parameters for the individual spans of a trace.
	Spans []SpanTemplate `js:"spans"`
}

type Link struct {
	// Attributes for this link
	Attributes map[string]interface{} `js:"attributes"`
	// Generate random attributes for this link
	RandomAttributes *AttributeParams `js:"randomAttributes"`
}

type Event struct {
	// Name of event
	Name string `js:"name"`
	// Attributes for this event
	Attributes map[string]interface{} `js:"attributes"`
	// Generate random attributes for this event
	RandomAttributes *AttributeParams `js:"randomAttributes"`
}

type LinkParams struct {
	// Count of random links per each span (default: 1)
	Count float32 `js:"count"`
	// Generate random attributes for this link
	RandomAttributes *AttributeParams `js:"randomAttributes"`
}

type EventParams struct {
	// Count of random events per each span
	Count float32 `js:"count"`
	// ExceptionCount indicates how many exception events to add to the span
	ExceptionCount float32 `js:"exceptionCount"`
	// ExceptionOnError generates exceptions if status code of the span is >= 400
	ExceptionOnError bool `js:"exceptionOnError"`
	// Generate random attributes for this event
	RandomAttributes *AttributeParams `js:"randomAttributes"`
}

// NewTemplatedGenerator creates a new trace generator.
func NewTemplatedGenerator(template *TraceTemplate) (*TemplatedGenerator, error) {
	gen := &TemplatedGenerator{}
	err := gen.initialize(template)
	if err != nil {
		return nil, fmt.Errorf("fail to create new templated generator: %w", err)
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
	events             []internalEventTemplate
	links              []internalLinkTemplate
}

type internalResourceTemplate struct {
	service          string
	hostName         string
	hostIP           string
	transport        string
	hostPort         int
	attributes       map[string]interface{}
	randomAttributes map[string][]interface{}
}

type internalLinkTemplate struct {
	rate             float32
	attributes       map[string]interface{}
	randomAttributes map[string][]interface{}
}

type internalEventTemplate struct {
	rate             float32
	exceptionOnError bool
	name             string
	attributes       map[string]interface{}
	randomAttributes map[string][]interface{}
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
				_ = s.Attributes().PutEmpty(k).FromRaw(v)
			}
		}

		spans = append(spans, s)
	}

	return traceData
}

func (g *TemplatedGenerator) generateResourceSpans(resSpanSlice ptrace.ResourceSpansSlice, tmpl *internalResourceTemplate) ptrace.ResourceSpans {
	resSpans := resSpanSlice.AppendEmpty()
	resSpans.Resource().Attributes().PutStr("k6", "true")
	resSpans.Resource().Attributes().PutStr(attrServiceName, tmpl.service)

	for k, v := range tmpl.attributes {
		_ = resSpans.Resource().Attributes().PutEmpty(k).FromRaw(v)
	}
	for k, v := range tmpl.randomAttributes {
		_ = resSpans.Resource().Attributes().PutEmpty(k).FromRaw(random.SelectElement(v))
	}

	scopeSpans := resSpans.ScopeSpans().AppendEmpty()
	scopeSpans.Scope().SetName("k6-scope-name/" + random.String(15))
	scopeSpans.Scope().SetVersion("k6-scope-version:v" + strconv.Itoa(random.IntBetween(0, 99)) + "." + strconv.Itoa(random.IntBetween(0, 99)))
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
		_ = span.Attributes().PutEmpty(k).FromRaw(v)
	}

	for k, v := range tmpl.randomAttributes {
		_ = span.Attributes().PutEmpty(k).FromRaw(random.SelectElement(v))
	}

	g.generateNetworkAttributes(tmpl, &span, parent)
	if tmpl.attributeSemantics != nil && *tmpl.attributeSemantics == SemanticsHTTP {
		g.generateHTTPAttributes(tmpl, &span, parent)
	}

	// generate events
	var hasError bool
	if st, found := getHTTPStatusCode(span.Attributes()); found {
		hasError = st >= 400
	}

	span.Events().EnsureCapacity(len(tmpl.events))
	for _, e := range tmpl.events {
		if e.rate > 0 && random.Float32() > e.rate {
			continue
		}
		if e.exceptionOnError && !hasError {
			continue
		}

		event := span.Events().AppendEmpty()
		event.Attributes().EnsureCapacity(len(e.attributes) + len(e.randomAttributes))

		event.SetName(e.name)
		eventTime := start.Add(random.Duration(0, duration))
		event.SetTimestamp(pcommon.NewTimestampFromTime(eventTime))

		for k, v := range e.attributes {
			_ = event.Attributes().PutEmpty(k).FromRaw(v)
		}
		for k, v := range e.randomAttributes {
			_ = event.Attributes().PutEmpty(k).FromRaw(random.SelectElement(v))
		}
	}

	// generate links
	span.Links().EnsureCapacity(len(tmpl.links))
	for _, l := range tmpl.links {
		if l.rate > 0 && random.Float32() > l.rate {
			continue
		}

		link := span.Links().AppendEmpty()
		link.Attributes().EnsureCapacity(len(l.attributes) + len(l.randomAttributes))
		for k, v := range l.randomAttributes {
			_ = link.Attributes().PutEmpty(k).FromRaw(random.SelectElement(v))
		}
		for k, v := range l.attributes {
			_ = link.Attributes().PutEmpty(k).FromRaw(v)
		}

		// default to linking to parent span if exist
		// TODO: support linking to other existing spans
		if parent != nil {
			link.SetTraceID(traceID)
			link.SetSpanID(parent.SpanID())
		} else {
			link.SetTraceID(random.TraceID())
			link.SetSpanID(random.SpanID())
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
	switch tmpl.kind {
	case ptrace.SpanKindClient:
		putIfNotExists(span.Attributes(), "net.peer.port", random.Port())
	case ptrace.SpanKindServer:
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

	putIfNotExists(span.Attributes(), "network.protocol.name", "http")
	putIfNotExists(span.Attributes(), "network.protocol.version", "1.1")

	if tmpl.kind == ptrace.SpanKindServer {
		var method string
		if m, found := getHTTPMethod(span.Attributes()); found {
			method = m
		} else {
			method = random.HTTPMethod()
			span.Attributes().PutStr(attrHTTPMethod, method)
		}

		var contentType []any
		if ct, found := span.Attributes().Get(attrHTTPResponseHeaderContentType); found {
			contentType = ct.Slice().AsRaw()
		} else {
			contentType = random.HTTPContentType()
			_ = span.Attributes().PutEmptySlice(attrHTTPResponseHeaderContentType).FromRaw(contentType)
		}

		var status int64
		if st, found := getHTTPStatusCode(span.Attributes()); found {
			status = st
		} else {
			status = random.HTTPStatusSuccess()
			span.Attributes().PutInt(attrHTTPStatusCode, status)
		}
		if status >= 500 {
			span.Status().SetCode(ptrace.StatusCodeError)
			span.Status().SetMessage(http.StatusText(int(status)))
		}

		var requestURL *url.URL
		if u, found := span.Attributes().Get(attrURL); found {
			requestURL, _ = url.ParseRequestURI(u.Str())
		} else if u, found = parentAttr.Get(attrURL); found {
			requestURL, _ = url.ParseRequestURI(u.Str())
		} else {
			requestURL, _ = url.ParseRequestURI(fmt.Sprintf("https://%s:%d/%s", tmpl.resource.hostName, tmpl.resource.hostPort, tmpl.name))
			span.Attributes().PutStr(attrURL, requestURL.String())
		}
		span.Attributes().PutStr(attrURLScheme, requestURL.Scheme)
		span.Attributes().PutStr(attrURLTarget, requestURL.Path)

		putIfNotExists(span.Attributes(), attrHTTPResponseHeaderContentLength, []any{random.IntBetween(100_000, 1_000_000)})
		if method == http.MethodPatch || method == http.MethodPost || method == http.MethodPut {
			putIfNotExists(span.Attributes(), attrHTTPRequestHeaderContentLength, []any{random.IntBetween(10_000, 100_000)})
		}

		if parent != nil && parent.Kind() == ptrace.SpanKindClient {
			if status >= 400 {
				parent.Status().SetCode(ptrace.StatusCodeError)
				parent.Status().SetMessage(http.StatusText(int(status)))
			}
			putIfNotExists(parent.Attributes(), attrHTTPMethod, method)
			putIfNotExists(parent.Attributes(), attrHTTPRequestHeaderAccept, contentType)
			putIfNotExists(parent.Attributes(), attrHTTPStatusCode, status)
			putIfNotExists(parent.Attributes(), attrURL, requestURL.String())
			res, _ := span.Attributes().Get(attrHTTPResponseHeaderContentLength)
			putIfNotExists(span.Attributes(), attrHTTPResponseHeaderContentLength, res.AsRaw())
			if req, found := span.Attributes().Get(attrHTTPRequestHeaderContentLength); found {
				putIfNotExists(span.Attributes(), attrHTTPRequestHeaderContentLength, req.AsRaw())
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
		res, found := g.resources[tmpl.Service]
		if !found {
			res = g.initializeResource(&tmpl, &template.Defaults)
			g.resources[tmpl.Service] = res
		} else {
			g.amendInitializedResource(res, &tmpl)
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

func (g *TemplatedGenerator) initializeResource(tmpl *SpanTemplate, defaults *SpanDefaults) *internalResourceTemplate {
	res := internalResourceTemplate{
		service:   tmpl.Service,
		hostName:  fmt.Sprintf("%s.local", tmpl.Service),
		hostIP:    random.IPAddr(),
		hostPort:  random.Port(),
		transport: "ip_tcp",
	}

	// use defaults if no resource attributes are set
	if tmpl.Resource == nil {
		tmpl.Resource = defaults.Resource
	}

	if tmpl.Resource != nil {
		res.randomAttributes = initializeRandomAttributes(tmpl.Resource.RandomAttributes)
		res.attributes = tmpl.Resource.Attributes
	}

	return &res
}

func (g *TemplatedGenerator) amendInitializedResource(res *internalResourceTemplate, tmpl *SpanTemplate) {
	if tmpl.Resource == nil {
		return
	}

	if tmpl.Resource.RandomAttributes != nil {
		randAttr := initializeRandomAttributes(tmpl.Resource.RandomAttributes)
		res.randomAttributes = util.MergeMaps(res.randomAttributes, randAttr)
	}
	if tmpl.Resource.Attributes != nil {
		res.attributes = util.MergeMaps(res.attributes, tmpl.Resource.Attributes)
	}
}

func (g *TemplatedGenerator) initializeSpan(idx int, parent *internalSpanTemplate, defaults *SpanDefaults, tmpl, child *SpanTemplate) (*internalSpanTemplate, error) {
	res := g.resources[tmpl.Service]
	span := internalSpanTemplate{
		idx:                idx,
		parent:             parent,
		resource:           res,
		duration:           tmpl.Duration,
		attributeSemantics: tmpl.AttributeSemantics,
	}

	// apply defaults
	if span.attributeSemantics == nil {
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

	// initialize links for span
	span.links = g.initializeLinks(tmpl.Links, tmpl.RandomLinks, defaults.RandomLinks)

	// initialize events for the span
	span.events = g.initializeEvents(tmpl.Events, tmpl.RandomEvents, defaults.RandomEvents)

	return &span, nil
}

func initializeSpanKind(parent *internalSpanTemplate, tmpl, child *SpanTemplate) (ptrace.SpanKind, error) {
	var kind ptrace.SpanKind
	if k, found := tmpl.Attributes["span.kind"]; found {
		kindStr, ok := k.(string)
		if !ok {
			return ptrace.SpanKindUnspecified, fmt.Errorf("attribute span.kind expected to be a string, but was %T", k)
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
		_ = m.PutEmpty(k).FromRaw(v)
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

func (g *TemplatedGenerator) initializeEvents(tmplEvents []Event, randomEvents, defaultRandomEvents *EventParams) []internalEventTemplate {
	internalEvents := make([]internalEventTemplate, 0, len(tmplEvents))
	for _, e := range tmplEvents {
		event := internalEventTemplate{
			name:             e.Name,
			attributes:       e.Attributes,
			randomAttributes: initializeRandomAttributes(e.RandomAttributes),
		}
		internalEvents = append(internalEvents, event)
	}

	if randomEvents == nil {
		if defaultRandomEvents == nil {
			return internalEvents
		}
		randomEvents = defaultRandomEvents
	}

	// normal random events
	if randomEvents.Count == 0 { // default count is 1
		randomEvents.Count = 1
	}
	if randomEvents.Count < 1 {
		event := internalEventTemplate{
			rate:             randomEvents.Count,
			name:             random.EventName(),
			randomAttributes: initializeRandomAttributes(randomEvents.RandomAttributes),
		}
		internalEvents = append(internalEvents, event)
	} else {
		for i := 0; i < int(randomEvents.Count); i++ {
			event := internalEventTemplate{
				name:             random.EventName(),
				randomAttributes: initializeRandomAttributes(randomEvents.RandomAttributes),
			}
			internalEvents = append(internalEvents, event)
		}
	}

	// random exception events
	if randomEvents.ExceptionCount == 0 && randomEvents.ExceptionOnError {
		randomEvents.ExceptionCount = 1 // default exception count is 1, if ExceptionOnError is true
	}
	if randomEvents.ExceptionCount < 1 {
		event := internalEventTemplate{
			rate: randomEvents.ExceptionCount,
			name: "exception",
			attributes: map[string]interface{}{
				"exception.escape":     false,
				"exception.message":    generateRandomExceptionMsg(),
				"exception.stacktrace": generateRandomExceptionStackTrace(),
				"exception.type":       "error.type_" + random.K6String(10),
			},
			randomAttributes: initializeRandomAttributes(randomEvents.RandomAttributes),
			exceptionOnError: randomEvents.ExceptionOnError,
		}
		internalEvents = append(internalEvents, event)
	} else {
		for i := 0; i < int(randomEvents.ExceptionCount); i++ {
			event := internalEventTemplate{
				name: "exception",
				attributes: map[string]interface{}{
					"exception.escape":     false,
					"exception.message":    generateRandomExceptionMsg(),
					"exception.stacktrace": generateRandomExceptionStackTrace(),
					"exception.type":       "error.type_" + random.K6String(10),
				},
				randomAttributes: initializeRandomAttributes(randomEvents.RandomAttributes),
				exceptionOnError: randomEvents.ExceptionOnError,
			}
			internalEvents = append(internalEvents, event)
		}
	}

	return internalEvents
}

func generateRandomExceptionMsg() string {
	return "error: " + random.K6String(20)
}

func generateRandomExceptionStackTrace() string {
	var (
		panics    = []string{"runtime error: index out of range", "runtime error: can't divide by 0"}
		functions = []string{"main.main()", "trace.makespan()", "account.login()", "payment.collect()"}
	)

	return "panic: " + random.SelectElement(panics) + "\n" + random.SelectElement(functions)
}

func (g *TemplatedGenerator) initializeLinks(linkTemplates []Link, randomLinks, defaultRandomLinks *LinkParams) []internalLinkTemplate {
	internalLinks := make([]internalLinkTemplate, 0, len(linkTemplates))

	for _, lt := range linkTemplates {
		link := internalLinkTemplate{
			attributes:       lt.Attributes,
			randomAttributes: initializeRandomAttributes(lt.RandomAttributes),
		}
		internalLinks = append(internalLinks, link)
	}

	if randomLinks == nil {
		if defaultRandomLinks == nil {
			return internalLinks
		}
		randomLinks = defaultRandomLinks
	}
	if randomLinks.Count == 0 { // default count is 1
		randomLinks.Count = 1
	}

	if randomLinks.Count < 1 {
		link := internalLinkTemplate{
			rate:             randomLinks.Count,
			randomAttributes: initializeRandomAttributes(randomLinks.RandomAttributes),
		}
		internalLinks = append(internalLinks, link)
	} else {
		for i := 0; i < int(randomLinks.Count); i++ {
			link := internalLinkTemplate{
				randomAttributes: initializeRandomAttributes(randomLinks.RandomAttributes),
			}
			internalLinks = append(internalLinks, link)
		}
	}

	return internalLinks
}

func getHTTPStatusCode(attributes pcommon.Map) (int64, bool) {
	st, found := attributes.Get(attrHTTPStatusCode)
	if found {
		return st.Int(), found
	}
	st, found = attributes.Get(attrHTTPStatusCodeOld)
	if found {
		return st.Int(), found
	}
	return 0, false
}

func getHTTPMethod(attributes pcommon.Map) (string, bool) {
	m, found := attributes.Get(attrHTTPMethod)
	if found {
		return m.Str(), found
	}
	m, found = attributes.Get(attrHTTPMethodOld)
	if found {
		return m.Str(), found
	}
	return "", false
}
