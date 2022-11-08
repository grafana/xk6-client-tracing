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
