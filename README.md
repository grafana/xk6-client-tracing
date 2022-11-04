# xk6-client-tracing

> ### ⚠️ In Development
>
> This project is **in development** and changes a lot between commits. Use at your own risk.

This extension provides k6 with the required functionality required to load test distributed tracing backends.

## Getting started

To start using the k6 tracing extension, ensure you have the following prerequisites installed:

- Docker
- docker-compose
- make

### Build docker image

The docker image is compiled using a multi-stage Docker build and does not require further dependencies. 
To start the build process run:

```shell
make docker
```

After the command completed successfully the image `grafana/xk6-client-tracing:latest` is available.

### Run docker-compose example

> Note: before running the docker-compose example, make sure to complete the docker image build step above!

To run the example `cd` into the directory `examples/param` and run:

```shell
docker-compose up -d
```

In the example `k6-tracing` uses the script `param.js` to generate spans and sends them to the `otel-collector`.
The generated spans can be observed by inspecting the collector's logs:

```shell
docker-compose logs -f otel-collector
```

The example uses the OTLP gRPC exporter. 
If you want to use Jaeger gRPC, you can change `param.js` and use the following settings:

```javascript
const client = new tracing.Client({
    endpoint: "otel-collector:14250",
    exporter: "jaeger",
    insecure: true,
});
```

> Note: HTTP exporters aren't supported (yet)

### Build locally

Building the extension locally has additional prerequisites:

- [Go toolchain](https://go101.org/article/go-toolchain.html)
- Git

Furthermore, the build also requires [`xk6`](https://github.com/grafana/xk6) to compile k6 with the bundled tracing extension.
Run the following command to install `xk6`:

```shell
go install go.k6.io/xk6/cmd/xk6@latest
```

To build binary run:
```shell
make build
```

The build step produces the `k6-tracing` binary.
To test the binary you first need to change the endpoint in the client configuration in `examples/basic/param.js`:

```javascript
const client = new tracing.Client({
    endpoint: "localhost:4317",
    exporter: "otlp",
    insecure: true,
});
```

Once you've your new binary and configuration ready, you can run a local OTEL collector:
```bash
docker run --rm -p 13133:13133 -p 14250:14250 -p 14268:14268 \
      -p 55678-55679:55678-55679 -p 4317:4317 -p 9411:9411 \
      -v "${PWD}/examples/shared/collector-config.yaml":/collector-config.yaml \
      --name otelcol otel/opentelemetry-collector \
      --config collector-config.yaml
```

Once that's done, you can run a test like:
```
./k6-tracing run examples/basic/param.js
```

And see the generated spans in the OTEL collector logs!

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
