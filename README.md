# xk6-client-tracing

This extension provides k6 with the required functionality required to load test distributed tracing backends.

> :warning: This extension is in development and is not yet ready for use.

## Dev docs

To build the extension locally:
```
xk6 build master \
  --with github.com/grafana/xk6-client-tracing="$PWD/../xk6-client-tracing"
```

**Reminder:** We can remove the master requirement once a new k6 version has been released (see https://github.com/grafana/k6/pull/2216).

To run a local OTEL Collector:
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