FROM golang:1.25.4-alpine@sha256:d3f0cf7723f3429e3f9ed846243970b20a2de7bae6a5b66fc5914e228d831bbb AS xk6-client-tracing-build

RUN apk add --no-cache \
    build-base \
    gcc \
    git \
    make

RUN go install go.k6.io/xk6/cmd/xk6@latest

WORKDIR /opt/xk6-client-tracing
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN make build

FROM alpine:latest@sha256:51183f2cfa6320055da30872f211093f9ff1d3cf06f39a0bdb212314c5dc7375

COPY --from=xk6-client-tracing-build /opt/xk6-client-tracing/k6-tracing /k6-tracing
COPY ./examples/template/template.js /example-script.js

ENTRYPOINT [ "/k6-tracing" ]
CMD ["run", "/example-script.js"]
