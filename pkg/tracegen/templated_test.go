package tracegen

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/ptrace"
)

const testRounds = 5

func TestTemplatedGenerator_Traces(t *testing.T) {
	attributeSemantics := []OTelSemantics{SemanticsHTTP}
	template := TraceTemplate{
		Defaults: SpanDefaults{
			Attributes:       map[string]interface{}{"fixed.attr": "some-value"},
			RandomAttributes: &AttributeParams{Count: 3},
		},
		Spans: []SpanTemplate{
			{Service: "test-service", Name: ptr("perform-test"), RandomAttributes: &AttributeParams{Count: 2}},
			{Service: "test-service"},
			{Service: "test-service", Name: ptr("get_test_data")},
			{Service: "test-data", Name: ptr("list_test_data"), Attributes: map[string]interface{}{"http.status_code": 400}},
		},
	}

	for _, semantics := range attributeSemantics {
		template.Defaults.AttributeSemantics = &semantics
		gen, err := NewTemplatedGenerator(&template)
		assert.NoError(t, err)

		for i := 0; i < testRounds; i++ {
			traces := gen.Traces()
			spans := collectSpansFromTrace(traces)

			assert.Len(t, spans, len(template.Spans))
			for i, span := range spans {
				assert.GreaterOrEqual(t, attributesWithPrefix(span, "k6."), template.Defaults.RandomAttributes.Count)
				if template.Spans[i].Name != nil {
					assert.Equal(t, *template.Spans[i].Name, span.Name())
				}
				if span.Kind() != ptrace.SpanKindInternal {
					assert.GreaterOrEqual(t, attributesWithPrefix(span, "net."), 3)
					if *template.Defaults.AttributeSemantics == SemanticsHTTP {
						assert.GreaterOrEqual(t, attributesWithPrefix(span, "http."), 5)
					}
				}
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
			{Service: "test-service", Name: ptr("default_generate_on_error"), Attributes: map[string]interface{}{"http.status_code": 400}},
		},
	}

	for _, semantics := range attributeSemantics {
		template.Defaults.AttributeSemantics = &semantics
		gen, err := NewTemplatedGenerator(&template)
		assert.NoError(t, err)

		for i := 0; i < testRounds; i++ {
			traces := gen.Traces()
			spans := collectSpansFromTrace(traces)

			assert.Len(t, spans, len(template.Spans))
			for _, span := range spans {
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
		}
	}
}

func attributesWithPrefix(span ptrace.Span, prefix string) int {
	var count int
	span.Attributes().Range(func(k string, _ pcommon.Value) bool {
		if strings.HasPrefix(k, prefix) {
			count++
		}
		return true
	})
	return count
}

func collectSpansFromTrace(traces ptrace.Traces) []ptrace.Span {
	var spans []ptrace.Span
	for i := 0; i < traces.ResourceSpans().Len(); i++ {
		rs := traces.ResourceSpans().At(i)
		for j := 0; j < rs.ScopeSpans().Len(); j++ {
			ss := rs.ScopeSpans().At(j)
			for k := 0; k < ss.Spans().Len(); k++ {
				spans = append(spans, ss.Spans().At(k))
			}
		}
	}
	return spans
}

func ptr[T any](v T) *T {
	return &v
}
