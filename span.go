package xk6_client_tracing

import (
	"fmt"
	"time"

	"go.opentelemetry.io/collector/model/pdata"
)

type Span struct {
	Name       string
	Attributes map[string]interface{}
	Status     struct {
		Code    int
		Message string
	}
}

func (s Span) construct() pdata.Span {
	endTime := time.Now().Round(time.Second)
	startTime := endTime.Add(-90 * time.Second)
	spanAttributes := constructSpanAttributes(s.Attributes)

	span := pdata.NewSpan()
	span.SetTraceID(newTraceID())
	span.SetSpanID(newSegmentID())
	span.SetParentSpanID(newSegmentID())
	span.SetName(s.Name)
	span.SetKind(pdata.SpanKindClient)
	span.SetStartTimestamp(pdata.NewTimestampFromTime(startTime))
	span.SetEndTimestamp(pdata.NewTimestampFromTime(endTime))

	status := pdata.NewSpanStatus()
	status.SetCode(pdata.StatusCode(s.Status.Code))
	status.SetMessage(s.Status.Message)
	status.CopyTo(span.Status())

	spanAttributes.CopyTo(span.Attributes())
	return span
}

func constructSpanAttributes(attributes map[string]interface{}) pdata.AttributeMap {
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
	return attrs
}
