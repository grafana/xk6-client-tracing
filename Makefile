BINARY     ?= k6-tracing
IMAGE      ?= ghcr.io/grafana/xk6-client-tracing
IMAGE_TAG  ?= latest

GO_MODULE      := $(shell head -n1 go.mod | cut -d' ' -f2)
GO_TEST_OPTS   := -race -count=1 -cover
GO_LINT_OPTS   := --config ./golangci.yml
XK6_BUILD_OPTS := --output ./$(BINARY)

.PHONY: build
build:
	xk6 build $(XK6_BUILD_OPTS) --with $(GO_MODULE)=.

.PHONY: test
test:
	go test $(GO_TEST_OPTS) ./...

.PHONY: lint
lint:
	golangci-lint run $(GO_LINT_OPTS) ./...

.PHONY: docker
docker:
	docker build . -t $(IMAGE):$(IMAGE_TAG)

.PHONY: clean
clean:
	go clean -cache -testcache
	docker rmi -f $(IMAGE)
