package random

import (
	crand "crypto/rand"
	"math/big"
	"math/rand"
	"net/http"

	"go.opentelemetry.io/collector/model/pdata"
)

var (
	letters             = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	httpStatusesSuccess = []int{200, 201, 202, 204}
	httpStatusesError   = []int{400, 401, 403, 404, 405, 406, 408, 409, 410, 411, 412, 413, 414, 415, 417, 428, 427, 500, 501, 502}
	httpMethods         = []string{http.MethodGet, http.MethodDelete, http.MethodPost, http.MethodPut, http.MethodPatch}
	operations          = []string{"get", "list", "query", "search", "set", "add", "create", "update", "send", "remove", "delete"}
	serviceSuffix       = []string{"", "", "service", "backend", "api", "proxy", "engine"}
	dbNames             = []string{"redis", "mysql", "postgres", "memcached", "mongodb", "elasticsearch"}
	resources           = []string{
		"order", "payment", "customer", "product", "stock", "inventory",
		"shipping", "billing", "checkout", "cart", "search", "analytics"}
)

func init() {
	seed, _ := crand.Int(crand.Reader, big.NewInt(int64(^uint64(0)>>1)))
	rand.Seed(seed.Int64())
}

func selectElem[T comparable](elements []T) T {
	return elements[rand.Intn(len(elements))]
}

func String(n int) string {
	s := make([]rune, n)
	for i := range s {
		s[i] = selectElem(letters)
	}
	return string(s)
}

func K6String(n int) string {
	return "k6." + String(n)
}

func HTTPStatusSuccess() int {
	return selectElem(httpStatusesSuccess)
}

func HTTPStatusErr() int {
	return selectElem(httpStatusesError)
}

func HTTPMethod() string {
	return selectElem(httpMethods)
}

func DBService() string {
	return selectElem(dbNames)
}

func Service() string {
	resource := selectElem(resources)
	return ServiceForResource(resource)
}

func ServiceForResource(resource string) string {
	name := resource
	suffix := selectElem(serviceSuffix)
	if suffix != "" {
		name = name + "-" + suffix
	}
	return name
}

func Operation() string {
	resource := selectElem(resources)
	return OperationForResource(resource)
}

func OperationForResource(resource string) string {
	op := selectElem(operations)
	return op + "-" + resource
}

func TraceID() pdata.TraceID {
	var b [16]byte
	_, _ = rand.Read(b[:]) // always returns nil error
	return pdata.NewTraceID(b)
}

func SpanID() pdata.SpanID {
	var b [8]byte
	_, _ = rand.Read(b[:]) // always returns nil error
	return pdata.NewSpanID(b)
}
