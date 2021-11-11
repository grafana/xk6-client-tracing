package xk6_client_tracing

import (
	"crypto/rand"
	"encoding/binary"
	"time"

	"go.opentelemetry.io/collector/model/pdata"
)

func newTraceID() pdata.TraceID {
	var r [16]byte
	epoch := time.Now().Unix()
	binary.BigEndian.PutUint32(r[0:4], uint32(epoch))
	_, err := rand.Read(r[4:])
	if err != nil {
		panic(err)
	}
	return pdata.NewTraceID(r)
}

func newSegmentID() pdata.SpanID {
	var r [8]byte
	_, err := rand.Read(r[:])
	if err != nil {
		panic(err)
	}
	return pdata.NewSpanID(r)
}
