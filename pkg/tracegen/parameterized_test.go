package tracegen

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/pdata/ptrace"
)

func TestParameterizedGenerator_Traces(t *testing.T) {
	// Create a simple trace parameter with minimal configuration
	traceParams := []*TraceParams{
		{
			ID:                "1234567890abcdef1234567890abcdef",
			Count:             1,
			RandomServiceName: false,
			ResourceSize:      2,
			Spans: SpanParams{
				Count:      2,
				Size:       2,
				RandomName: false,
				FixedAttrs: map[string]interface{}{
					"test.attr": "test.value",
				},
			},
		},
	}

	generator := NewParameterizedGenerator(traceParams)
	traces := generator.Traces()

	// Basic validation
	require.Equal(t, 1, traces.ResourceSpans().Len(), "Should have one resource span")

	// Validate resource span
	rs := traces.ResourceSpans().At(0)
	attrs := rs.Resource().Attributes()
	_, hasServiceName := attrs.Get(attrServiceName)
	_, hasK6 := attrs.Get("k6")
	assert.True(t, hasServiceName, "Should have service.name attribute")
	assert.True(t, hasK6, "Should have k6 attribute")
	k6Val, _ := attrs.Get("k6")
	assert.Equal(t, "true", k6Val.Str(), "k6 attribute should be true")

	// Validate scope spans
	require.Equal(t, 1, rs.ScopeSpans().Len(), "Should have one scope span")
	ils := rs.ScopeSpans().At(0)
	assert.Contains(t, ils.Scope().Name(), "k6-scope-name/", "Scope name should have prefix")
	assert.Contains(t, ils.Scope().Version(), "k6-scope-version:v", "Scope version should have prefix")

	// Validate spans
	require.Equal(t, 2, ils.Spans().Len(), "Should have two spans")

	// Validate first span (parent)
	span1 := ils.Spans().At(0)
	assert.Equal(t, "1234567890abcdef1234567890abcdef", span1.TraceID().String(), "TraceID should match")
	assert.Equal(t, ptrace.SpanKindClient, span1.Kind(), "Span kind should be client")
	span1Attrs := span1.Attributes()
	_, hasTestAttr := span1Attrs.Get("test.attr")
	assert.True(t, hasTestAttr, "Should have fixed attribute")
	testAttrVal, _ := span1Attrs.Get("test.attr")
	assert.Equal(t, "test.value", testAttrVal.Str(), "Fixed attribute value should match")
	assert.Equal(t, 1, span1.Events().Len(), "Should have one event")
	assert.Equal(t, 1, span1.Links().Len(), "Should have one link")

	// Validate second span (child)
	span2 := ils.Spans().At(1)
	assert.Equal(t, "1234567890abcdef1234567890abcdef", span2.TraceID().String(), "TraceID should match")
	assert.Equal(t, span1.SpanID(), span2.ParentSpanID(), "Parent span ID should match first span's ID")
	assert.Equal(t, ptrace.SpanKindClient, span2.Kind(), "Span kind should be client")
	span2Attrs := span2.Attributes()
	_, hasTestAttr2 := span2Attrs.Get("test.attr")
	assert.True(t, hasTestAttr2, "Should have fixed attribute")
	testAttrVal2, _ := span2Attrs.Get("test.attr")
	assert.Equal(t, "test.value", testAttrVal2.Str(), "Fixed attribute value should match")
	assert.Equal(t, 1, span2.Events().Len(), "Should have one event")
	assert.Equal(t, 1, span2.Links().Len(), "Should have one link")
}
