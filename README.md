# xk6-client-tracing

This extension provides k6 with the required functionality required to load test distributed tracing backends.

> :warning: This extension is in development and is not yet ready for use.

## Getting started  

To start using k6 with the extension ensure you have the prerequisites:

- [Go toolchain](https://go101.org/article/go-toolchain.html)
- Git

Then:

1. Install `xk6`:
  ```shell
  go install go.k6.io/xk6/cmd/xk6@latest
  ```

2. Build the binary:
```shell
xk6 build v0.35.0 \
  --with github.com/grafana/xk6-client-tracing="$PWD/../xk6-client-tracing"
```

Once you've your new binary ready, you can run a local OTEL collector:
```bash
docker run --rm -p 13133:13133 -p 14250:14250 -p 14268:14268 \
      -p 55678-55679:55678-55679 -p 4317:4317 -p 8888:8888 -p 9411:9411 \
      -v "${PWD}/collector-config.yaml":/collector-config.yaml \
      --name otelcol otel/opentelemetry-collector \
      --config collector-config.yaml
```

Once that's done, you can run a test like:
```
./k6 run examples/basic.js
```

And see your spans on the OTEL collector logs!