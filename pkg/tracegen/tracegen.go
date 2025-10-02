package tracegen

import (
	"go.opentelemetry.io/collector/pdata/ptrace"
)

const (
	attrServiceName                     = "service.name"
	attrHTTPStatusCode                  = "http.response.status_code"
	attrHTTPStatusCodeOld               = "http.status_code"
	attrHTTPMethod                      = "http.request.method"
	attrHTTPMethodOld                   = "http.method"
	attrHTTPRequestHeaderAccept         = "http.request.header.accept"
	attrHTTPRequestHeaderContentLength  = "http.request.header.content-length"
	attrHTTPResponseHeaderContentLength = "http.response.header.content-length"
	attrHTTPResponseHeaderContentType   = "http.response.header.content-type"
	attrURL                             = "url.full"
	attrURLScheme                       = "url.schema"
	attrURLTarget                       = "url.target"
)

// Generator creates traces to be used in k6 tests
type Generator interface {
	Traces() ptrace.Traces
}
