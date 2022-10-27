package tracegen

import (
	"go.opentelemetry.io/collector/pdata/ptrace"
)

type Generator interface {
	Traces() ptrace.Traces
}
