package random

import (
	crand "crypto/rand"
	"encoding/binary"
	"fmt"
	"math/rand/v2"
	"net/http"
	"sync"
	"time"

	"go.opentelemetry.io/collector/pdata/pcommon"
)

var (
	letters             = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	httpStatusesSuccess = []int64{200, 201, 202, 204}
	httpStatusesError   = []int64{400, 401, 403, 404, 405, 406, 408, 409, 410, 411, 412, 413, 414, 415, 417, 428, 427, 500, 501, 502}
	httpMethods         = []string{http.MethodGet, http.MethodDelete, http.MethodPost, http.MethodPut, http.MethodPatch}
	httpContentTypes    = []string{"application/json", "application/xml", "application/x-www-form-urlencoded", "text/plain", "text/html"}
	operations          = []string{"get", "list", "query", "search", "set", "add", "create", "update", "send", "remove", "delete"}
	serviceSuffix       = []string{"", "", "service", "backend", "api", "proxy", "engine"}
	dbNames             = []string{"redis", "mysql", "postgres", "memcached", "mongodb", "elasticsearch"}
	resources           = []string{
		"order", "payment", "customer", "product", "stock", "inventory",
		"shipping", "billing", "checkout", "cart", "search", "analytics"}

	// rnd contains rand.Rand instance protected by a mutex
	rnd = struct {
		sync.Mutex
		*rand.Rand
	}{}
)

func init() {
	var seed [32]byte
	_, err := crand.Read(seed[:])
	if err != nil {
		panic(err)
	}
	rnd.Rand = rand.New(rand.NewChaCha8(seed))
}

func Float32() float32 {
	rnd.Lock()
	defer rnd.Unlock()
	return rnd.Float32()
}

func IntN(n int) int {
	rnd.Lock()
	defer rnd.Unlock()
	return rnd.IntN(n)
}

func SelectElement[T any](elements []T) T {
	rnd.Lock()
	defer rnd.Unlock()
	return elements[rnd.IntN(len(elements))]
}

func String(n int) string {
	s := make([]rune, n)
	for i := range s {
		s[i] = SelectElement(letters)
	}
	return string(s)
}

func K6String(n int) string {
	return "k6." + String(n)
}

func IntBetween(min, max int) int {
	rnd.Lock()
	defer rnd.Unlock()
	n := rnd.IntN(max - min)
	return min + n
}

func Duration(min, max time.Duration) time.Duration {
	rnd.Lock()
	defer rnd.Unlock()
	n := rnd.Int64N(int64(max) - int64(min))
	return min + time.Duration(n)
}

func IPAddr() string {
	rnd.Lock()
	defer rnd.Unlock()
	return fmt.Sprintf("192.168.%d.%d", rnd.IntN(255), rnd.IntN(255))
}

func Port() int {
	return IntBetween(8000, 9000)
}

func HTTPStatusSuccess() int64 {
	return SelectElement(httpStatusesSuccess)
}

func HTTPStatusErr() int64 {
	return SelectElement(httpStatusesError)
}

func HTTPMethod() string {
	return SelectElement(httpMethods)
}

func HTTPContentType() []any {
	return []any{SelectElement(httpContentTypes)}
}

func DBService() string {
	return SelectElement(dbNames)
}

func Service() string {
	resource := SelectElement(resources)
	return ServiceForResource(resource)
}

func ServiceForResource(resource string) string {
	name := resource
	suffix := SelectElement(serviceSuffix)
	if suffix != "" {
		name = name + "-" + suffix
	}
	return name
}

func Operation() string {
	resource := SelectElement(resources)
	return OperationForResource(resource)
}

func OperationForResource(resource string) string {
	op := SelectElement(operations)
	return op + "-" + resource
}

func TraceID() pcommon.TraceID {
	rnd.Lock()
	defer rnd.Unlock()

	var b [16]byte
	binary.BigEndian.PutUint64(b[:8], rnd.Uint64())
	binary.BigEndian.PutUint64(b[8:], rnd.Uint64())
	return b
}

func SpanID() pcommon.SpanID {
	rnd.Lock()
	defer rnd.Unlock()

	var b [8]byte
	binary.BigEndian.PutUint64(b[:], rnd.Uint64())
	return b
}

func EventName() string {
	return "event_k6." + String(10)
}
