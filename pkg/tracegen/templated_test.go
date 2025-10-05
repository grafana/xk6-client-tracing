package tracegen

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/ptrace"
)

const testRounds = 5

func TestTemplatedGenerator_Traces(t *testing.T) {
	attributeSemantics := []OTelSemantics{SemanticsHTTP, SemanticsDB}
	template := TraceTemplate{
		Defaults: SpanDefaults{
			Attributes:       map[string]interface{}{"fixed.attr": "some-value"},
			RandomAttributes: &AttributeParams{Count: 3},
		},
		Spans: []SpanTemplate{
			{Service: "test-service", Name: ptr("perform-test"), RandomAttributes: &AttributeParams{Count: 2}},
			{Service: "test-service"},
			{Service: "test-service", Name: ptr("get_test_data")},
			{Service: "test-data", Name: ptr("list_test_data"), Attributes: map[string]interface{}{attrHTTPStatusCode: 400}},
			{Service: "test-forced-semantic", AttributeSemantics: &attributeSemantics[0]},
		},
	}

	for _, semantics := range attributeSemantics {
		template.Defaults.AttributeSemantics = &semantics
		gen, err := NewTemplatedGenerator(&template)
		assert.NoError(t, err)

		for range testRounds {
			count := 0
			for i, span := range iterSpans(gen.Traces()) {
				count++
				requireAttributeCountGreaterOrEqual(t, span.Attributes(), 3, "k6.")
				spanTemplate := template.Spans[i]
				if spanTemplate.Name != nil {
					assert.Equal(t, *template.Spans[i].Name, span.Name())
				}
				if span.Kind() != ptrace.SpanKindInternal {
					requireAttributeCountGreaterOrEqual(t, span.Attributes(), 3, "net.")
					if *template.Defaults.AttributeSemantics == SemanticsHTTP {
						requireAttributeCountGreaterOrEqual(t, span.Attributes(), 3, "http.")
					}
					if spanTemplate.AttributeSemantics != nil && *spanTemplate.AttributeSemantics == SemanticsHTTP {
						requireAttributeCountGreaterOrEqual(t, span.Attributes(), 3, "http.")
					}
				}
			}
			assert.Equal(t, len(template.Spans), count, "unexpected number of spans")
		}
	}
}

func TestTemplatedGenerator_Resource(t *testing.T) {
	template := TraceTemplate{
		Defaults: SpanDefaults{
			Attributes: map[string]interface{}{"span-attr": "val-01"},
			Resource:   &ResourceTemplate{RandomAttributes: &AttributeParams{Count: 2}},
		},
		Spans: []SpanTemplate{
			{Service: "test-service-a", Name: ptr("action-a-a"), Resource: &ResourceTemplate{
				Attributes:       map[string]interface{}{"res-attr-01": "res-val-01"},
				RandomAttributes: &AttributeParams{Count: 5},
			}},
			{Service: "test-service-a", Name: ptr("action-a-b"), Resource: &ResourceTemplate{
				Attributes: map[string]interface{}{"res-attr-02": "res-val-02"},
			}},
			{Service: "test-service-b", Name: ptr("action-b-a"), Resource: &ResourceTemplate{
				Attributes: map[string]interface{}{"res-attr-03": "res-val-03"},
			}},
			{Service: "test-service-b", Name: ptr("action-b-b")},
		},
	}

	gen, err := NewTemplatedGenerator(&template)
	require.NoError(t, err)

	for range testRounds {
		for _, res := range iterResources(gen.Traces()) {
			srv, found := res.Attributes().Get(attrServiceName)
			require.True(t, found, "service.name not found")

			switch srv.Str() {
			case "test-service-a":
				requireAttributeCountEqual(t, res.Attributes(), 5, "k6.")
				requireAttributeEqual(t, res.Attributes(), "res-attr-01", "res-val-01")
				requireAttributeEqual(t, res.Attributes(), "res-attr-02", "res-val-02")
			case "test-service-b":
				requireAttributeCountEqual(t, res.Attributes(), 3, "k6.")
				requireAttributeEqual(t, res.Attributes(), "res-attr-03", "res-val-03")
			default:
				require.Fail(t, "unexpected service name %s", srv.Str())
			}
		}
	}
}

func TestTemplatedGenerator_EventsLinks(t *testing.T) {
	attributeSemantics := []OTelSemantics{SemanticsHTTP}
	template := TraceTemplate{
		Defaults: SpanDefaults{
			Attributes:       map[string]interface{}{"fixed.attr": "some-value"},
			RandomAttributes: &AttributeParams{Count: 3},
			RandomLinks:      &LinkParams{Count: 0.5, RandomAttributes: &AttributeParams{Count: 3}},
			RandomEvents:     &EventParams{ExceptionOnError: true, Count: 0.5, RandomAttributes: &AttributeParams{Count: 3}},
		},
		Spans: []SpanTemplate{
			// do not change order of the first one
			{Service: "test-service", Name: ptr("only_default")},
			{Service: "test-service", Name: ptr("default_and_template"), Events: []Event{{Name: "event-name", RandomAttributes: &AttributeParams{Count: 2}}}, Links: []Link{{Attributes: map[string]interface{}{"link-attr-key": "link-attr-value"}}}},
			{Service: "test-service", Name: ptr("default_and_random"), RandomEvents: &EventParams{Count: 2, RandomAttributes: &AttributeParams{Count: 1}}, RandomLinks: &LinkParams{Count: 2, RandomAttributes: &AttributeParams{Count: 1}}},
			{Service: "test-service", Name: ptr("default_template_random"), Events: []Event{{Name: "event-name", RandomAttributes: &AttributeParams{Count: 2}}}, Links: []Link{{Attributes: map[string]interface{}{"link-attr-key": "link-attr-value"}}}, RandomEvents: &EventParams{Count: 2, RandomAttributes: &AttributeParams{Count: 1}}, RandomLinks: &LinkParams{Count: 2, RandomAttributes: &AttributeParams{Count: 1}}},
			{Service: "test-service", Name: ptr("default_generate_on_error"), Attributes: map[string]interface{}{attrHTTPStatusCode: 400}},
		},
	}

	for _, semantics := range attributeSemantics {
		template.Defaults.AttributeSemantics = &semantics
		gen, err := NewTemplatedGenerator(&template)
		assert.NoError(t, err)

		for range testRounds {
			count := 0
			for _, span := range iterSpans(gen.Traces()) {
				count++
				events := span.Events()
				links := span.Links()
				checkEventsLinksLength := func(expectedTemplate, expectedRandom int, spanName string) {
					expected := expectedTemplate + expectedRandom
					// because default rate is 0.5
					assert.GreaterOrEqual(t, events.Len(), expected, "test name: %s events", spanName)
					assert.GreaterOrEqual(t, links.Len(), expected, "test name: %s links", spanName)
					assert.LessOrEqual(t, events.Len(), expected+1, "test name: %s events", spanName)
					assert.LessOrEqual(t, links.Len(), expected+1, "test name: %s links", spanName)
				}

				checkLinks := func() {
					for i := 0; i < links.Len(); i++ {
						link := links.At(i)
						assert.Equal(t, span.TraceID(), link.TraceID())
						assert.Equal(t, span.ParentSpanID(), link.SpanID())
					}
				}

				switch span.Name() {
				case "only_default":
					checkEventsLinksLength(0, 0, span.Name())
					if events.Len() > 0 {
						// check default event with 3 random attributes
						event := events.At(0)
						assert.Equal(t, 3, len(event.Attributes().AsRaw()))
					}
					if links.Len() > 0 {
						// check default link with 3 random attributes
						// and not matching trace id and parent span id because this is
						// the first span, there is no previous span
						link := links.At(0)
						assert.Equal(t, 3, len(link.Attributes().AsRaw()))
						assert.NotEqual(t, span.TraceID(), link.TraceID())
						assert.NotEqual(t, span.ParentSpanID(), link.SpanID())
					}
				case "default_and_template":
					checkEventsLinksLength(1, 0, span.Name())
					checkLinks()
				case "default_and_random":
					checkEventsLinksLength(0, 2, span.Name())
					checkLinks()
				case "default_template_random":
					checkEventsLinksLength(1, 2, span.Name())
					checkLinks()
				case "default_generate_on_error":
					// there should be at least one event
					assert.GreaterOrEqual(t, events.Len(), 0, "test name: %s events", "default generate on error")
					found := false
					for i := 0; i < events.Len(); i++ {
						event := events.At(i)
						if event.Name() == "exception" {
							found = true
							assert.NotNil(t, event.Attributes().AsRaw()["exception.escape"])
							assert.NotNil(t, event.Attributes().AsRaw()["exception.message"])
							assert.NotNil(t, event.Attributes().AsRaw()["exception.stacktrace"])
							assert.NotNil(t, event.Attributes().AsRaw()["exception.type"])
						}
					}
					assert.True(t, found, "exception event not found")
				}
			}
			assert.Equal(t, len(template.Spans), count, "unexpected number of spans")
		}
	}
}

func iterSpans(traces ptrace.Traces) func(func(i int, e ptrace.Span) bool) {
	count := 0
	return func(f func(i int, e ptrace.Span) bool) {
		var elem ptrace.Span
		for i := 0; i < traces.ResourceSpans().Len(); i++ {
			rs := traces.ResourceSpans().At(i)
			for j := 0; j < rs.ScopeSpans().Len(); j++ {
				ss := rs.ScopeSpans().At(j)
				for k := 0; k < ss.Spans().Len(); k++ {
					elem = ss.Spans().At(k)
					if !f(count, elem) {
						return
					}
					count++
				}
			}
		}
	}
}

func iterResources(traces ptrace.Traces) func(func(i int, e pcommon.Resource) bool) {
	return func(f func(i int, e pcommon.Resource) bool) {
		var elem pcommon.Resource
		for i := 0; i < traces.ResourceSpans().Len(); i++ {
			rs := traces.ResourceSpans().At(i)
			elem = rs.Resource()
			if !f(i, elem) {
				return
			}
		}
	}
}

func requireAttributeCountGreaterOrEqual(t *testing.T, attributes pcommon.Map, compare int, prefixes ...string) {
	t.Helper()
	count := countAttributes(attributes, prefixes...)
	require.GreaterOrEqual(t, count, compare, "expected at least %d attributes, got %d", compare, count)
}

func requireAttributeCountEqual(t *testing.T, attributes pcommon.Map, expected int, prefixes ...string) {
	t.Helper()
	count := countAttributes(attributes, prefixes...)
	require.GreaterOrEqual(t, expected, count, "expected at least %d attributes, got %d", expected, count)
}

func requireAttributeEqual(t *testing.T, attributes pcommon.Map, key string, expected any) {
	t.Helper()
	val, found := attributes.Get(key)
	require.True(t, found, "attribute %s not found", key)
	require.Equal(t, expected, val.AsRaw(), "value %v expected for attribute %s but was %v", expected, key, val.AsRaw())
}

func countAttributes(attributes pcommon.Map, prefixes ...string) int {
	var count int
	attributes.Range(func(k string, _ pcommon.Value) bool {
		if len(prefixes) == 0 {
			count++
			return true
		}

		for _, prefix := range prefixes {
			if strings.HasPrefix(k, prefix) {
				count++
			}
		}
		return true
	})
	return count
}

func ptr[T any](v T) *T {
	return &v
}
