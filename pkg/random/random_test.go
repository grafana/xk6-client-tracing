package random

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/pdata/pcommon"
)

const (
	testRounds = 10
)

func TestSelectElement(t *testing.T) {
	var prev string
	var eqCount int

	for i := 0; i < testRounds; i++ {
		res := SelectElement(resources)
		if res == prev {
			eqCount++
		}

		assert.Contains(t, resources, res)
		prev = res
	}

	assert.Less(t, eqCount, 4, "too many equal selections")
}

func TestString(t *testing.T) {
	for n := 5; n <= 20; n += 5 {
		t.Run(fmt.Sprintf("length_%d", n), func(t *testing.T) {
			var prev string
			for i := 0; i < testRounds; i++ {
				s := String(n)
				assert.Len(t, s, n)
				assert.NotEqual(t, prev, s)
				prev = s
			}
		})
	}
}

func TestK6String(t *testing.T) {
	for n := 5; n <= 20; n += 5 {
		t.Run(fmt.Sprintf("length_%d", n), func(t *testing.T) {
			var prev string
			for i := 0; i < testRounds; i++ {
				s := K6String(n)
				assert.Len(t, s, n+3)
				assert.Equal(t, "k6.", s[:3])
				assert.NotEqual(t, prev, s)
				prev = s
			}
		})
	}
}

func TestIntBetween(t *testing.T) {
	const (
		min = 15
		max = 25
	)

	var prev, eqCount int
	for i := 0; i < testRounds; i++ {
		n := IntBetween(min, max)
		if n == prev {
			eqCount++
		}

		assert.GreaterOrEqual(t, n, min)
		assert.Less(t, n, max)
		prev = n
	}

	assert.Less(t, eqCount, 4, "too many equal random numbers")
}

func TestDBService(t *testing.T) {
	db := DBService()

	assert.Contains(t, dbNames, db)
}

func TestOperation(t *testing.T) {
	op := Operation()

	parts := strings.Split(op, "-")
	require.Equal(t, 2, len(parts))
	assert.Contains(t, operations, parts[0])
	assert.Contains(t, resources, parts[1])
}

func TestService(t *testing.T) {
	srv := Service()

	parts := strings.Split(srv, "-")
	assert.Contains(t, resources, parts[0])
	if len(parts) > 1 {
		assert.Contains(t, serviceSuffix, parts[1])
	}
}

func TestSpanID(t *testing.T) {
	var prev pcommon.SpanID
	for i := 0; i < testRounds; i++ {
		id := SpanID()
		assert.False(t, id.IsEmpty())
		assert.NotEqual(t, prev, id)
		prev = id
	}
}

func TestTraceID(t *testing.T) {
	var prev pcommon.TraceID
	for i := 0; i < testRounds; i++ {
		id := TraceID()
		assert.False(t, id.IsEmpty())
		assert.NotEqual(t, prev, id)
		prev = id
	}
}
