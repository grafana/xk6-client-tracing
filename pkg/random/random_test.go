package random

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testRounds = 10
)

func TestDBService(t *testing.T) {
	var prev string
	var eqCount int

	for i := 0; i < testRounds; i++ {
		srv := DBService()
		if srv == prev {
			eqCount++
		}

		assert.Contains(t, dbNames, srv)
		prev = srv
	}

	assert.Less(t, eqCount, 3, "too many equal random db names")
}

func TestOperation(t *testing.T) {
	var prev string
	var eqCount int

	for i := 0; i < testRounds; i++ {
		op := Operation()
		if op == prev {
			eqCount++
		}

		parts := strings.Split(op, "-")
		require.Equal(t, 2, len(parts))
		assert.Contains(t, operations, parts[0])
		assert.Contains(t, resources, parts[1])
		prev = op
	}

	assert.Less(t, eqCount, 3, "too many equal random operation names")
}

func TestService(t *testing.T) {
	var prev string
	var eqCount int

	for i := 0; i < testRounds; i++ {
		srv := Service()
		if srv == prev {
			eqCount++
		}

		parts := strings.Split(srv, "-")
		assert.Contains(t, resources, parts[0])
		if len(parts) > 1 {
			assert.Contains(t, serviceSuffix, parts[1])
		}
		prev = srv
	}

	assert.Less(t, eqCount, 3, "too many equal random service names")
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

func TestSpanID(t *testing.T) {
	var prev [8]byte
	for i := 0; i < testRounds; i++ {
		id := SpanID()
		assert.False(t, id.IsEmpty())
		assert.NotEqual(t, prev, id.Bytes())
		prev = id.Bytes()
	}
}

func TestTraceID(t *testing.T) {
	var prev [16]byte
	for i := 0; i < testRounds; i++ {
		id := TraceID()
		assert.False(t, id.IsEmpty())
		assert.NotEqual(t, prev, id.Bytes())
		prev = id.Bytes()
	}
}
