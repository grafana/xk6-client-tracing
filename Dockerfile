FROM golang:alpine AS xk6-client-tracing-build

RUN apk add --no-cache \
    build-base \
    gcc \
    git \
    make

RUN go install go.k6.io/xk6/cmd/xk6@latest \
    && wget -O- -nv https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s v1.50.0 \
    && golangci-lint --version

WORKDIR /opt/xk6-client-tracing
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN make build

FROM alpine:latest

COPY --from=xk6-client-tracing-build /opt/xk6-client-tracing/k6-tracing /k6-tracing
ENTRYPOINT [ "/k6-tracing" ]
