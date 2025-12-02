FROM golang:1.25.4-alpine AS xk6-client-tracing-build

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

# Pin k6 version to match go.mod
RUN xk6 build v1.4.2 --output ./k6-tracing \
    --with github.com/grafana/xk6-client-tracing=. 

FROM alpine:latest

COPY --from=xk6-client-tracing-build /opt/xk6-client-tracing/k6-tracing /k6-tracing
COPY ./examples/template/template.js /example-script.js

ENTRYPOINT [ "/k6-tracing" ]
CMD ["run", "/example-script.js"]
