# xk6-client-tracing

> ### ⚠️ In Development
>
> This project is **in development** and changes a lot between commits. Use at your own risk.

This extension provides k6 with the required functionality required to load test distributed tracing backends.

## Getting started  

To start using k6 with the extension, ensure you have the prerequisites:

- [Go toolchain](https://go101.org/article/go-toolchain.html)
- Git

Then:

1. Install `xk6`:
```shell
go install go.k6.io/xk6/cmd/xk6@latest
```

2. Build the binary:
```shell
xk6 build --with github.com/grafana/xk6-client-tracing@latest
```

Once you've your new binary ready, you can run a local OTEL collector:
```bash
docker run --rm -p 13133:13133 -p 14250:14250 -p 14268:14268 \
      -p 55678-55679:55678-55679 -p 4317:4317 -p 9411:9411 \
      -v "${PWD}/collector-config.yaml":/collector-config.yaml \
      --name otelcol otel/opentelemetry-collector \
      --config collector-config.yaml
```

Once that's done, you can run a test like:
```
./k6 run examples/basic.js
```

And see your spans on the OTEL collector logs!

The example uses the OTLP gRPC exporter. If you want to use Jaeger gRPC, you can use these settings:
```javascript
const client = new tracing.Client({
    endpoint: "0.0.0.0:14250",
    exporter: "jaeger",
    insecure: true,
});
```

> Note: HTTP exporters aren't supported (yet)

## Using the extension with Grafana Cloud

You can do that, by using the OTLP exporter and setting the required auth credentials:
```javascript
const client = new tracing.Client({
    endpoint: "you-tempo-endpoint:443"
    exporter: "otlp",
    insecure: false,
    authentication: {
        user: "tenant-id",
        password: "api-token"
    }
});
```


