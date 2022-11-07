package tracegen

import (
	"go.opentelemetry.io/collector/pdata/ptrace"
)

// Generator creates traces to be used in k6 tests
type Generator interface {
	Traces() ptrace.Traces
}
