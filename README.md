# xk6-client-tracing

> ### ⚠️ In Development
>
> This project is **in development** and changes a lot between commits. Use at your own risk.

This extension provides k6 with the required functionality required to load test distributed tracing backends.

## Usage

Generating traces and sending them to an agent or backend requires two things: a client and a trace generator.
Generators have a method called `traces()` that can be used to generate traces.
The client provides a method `push()` which receives the generated traces as first parameter and sends them to the configured collector.

Creating a client requires a client configuration:

```javascript
const config = {
    endpoint: "localhost:4317",
    exporter: tracing.EXPORTER_OTLP,
};
let client = new tracing.Client(config);
```

The configuration is an object with the following schema:

```javascript
{
    // The endpoint to which the traces are sent in the form of <hostname>:<port>
    endpoint: string,
    // The exporter protocol used for sending the traces: tracing.EXPORTER_OTLP or tracing.EXPORTER_JAEGER
    exporter: string,
    // Credentials used for authentication (optional)
    authentication: { user: string, password: string },
    // Additional headers sent by the client (optional)
    headers: { string : string }
    // TLS configuration
    tls: {
        // Whether insecure connections are allowed (optional, default: false)
        insecure: boolean,
        // Enable TLS but skip verification (optional, default: false)
        insecure_skip_verify: boolean,
        // The server name requested by the client (optional)
        server_name: string,
        // The path to the CA certificate file (optional)
        ca_file: string,
        // The path to the certificate file (optional)
        cert_file: string,
        // The path to the key file (optional)
        key_file: string,
    },
}
```

There are two different types of generators which are described in the following sections.

### Parameterized trace generator

This generator creates traces consisting of completely randomized spans.
The spans contain a configurable number of random attributes with randomly assigned values.
The main purpose of this generator is to create a large amount of spans with few lines of code.

An example can be found in [./examples/param](./examples/param).

### Templated trace generator

This generator creates realistically looking and traces that contain spans with span name, span kind, and attributes.
The trace is generated from a template configurations that describes how each should be generated.

The following listing creates a generator that creates traces with a single span:

```javascript
const template = {
    spans: [
        {service: "article-service", name: "get-articles", attributes: {"http.request.method": "GET"}}
    ]
};
let gen = new tracing.TemplatedGenerator(template);
client.push(gen.traces());
```

The generated span will have the name `get-articles`. 
The generator will further assign a span kind as well as some commonly used attributes.
There will also be a corresponding resource span with the respective `service.name` attribute.

The template has the following schema:

```javascript
{
    // The defaults can be used to configure parameters that are applied to all spans (optional)
    defaults: {
        // Fixed attributes that are added to every generated span (optional)
        attributes: { string : any },
        // attributeSemantics can be set in order to generate attributes that follow a certain OpenTelemetry 
        // semantic convention. For example tracing.SEMANTICS_HTTP (optional)
        attributeSemantics: string,
        // Parameters to configure the creation of random attributes. If missing, no random attributes
        // are added to the spans (optional)
        randomAttributes: { 
            // The number of random attributes to generate
            count: int,
            // The number of distinct values to generate for each attribute (optional, default: 50)
            cardinality: int
        }
        // Default resource attributes for all resources in the trace (optional)
        resource: {
            // Fixed attributs that are added to each resource (optional)
            attributes: { string : any },
            // Parameters to configure the creation of random resource attributes (optional)
            randomAttributes: {
                // The number of random attributes to generate
                count: int,
                // The number of distinct values to generate for each attribute (optional, default: 50)
                cardinality: int
            }
        }
    },
    // Templates for the individual spans
    spans: [
        {
            // Is used to set the service.name attribute of the corresponding resource span
            service: string,
            // The name of the span. If empty, the name will be randomly generated (optional)
            name: string,
            // The index of the parent span in `spans`. The index must be smaller than the
            // own index. If empty, the parent is the span with the position directly before 
            // this span in `spans` (optional)
            parentIdx: int,
            // The interval for the generated span duration. If missing, a random duration is 
            // generated that is shorter than the duration of the parent span (optional)
            duration: { min: int, max: int },
            // Fixed attributes that are added to this (optional)
            attributes: { string : any },
            // attributeSemantics can be set in order to generate attributes that follow a certain OpenTelemetry 
            // semantic convention. For example tracing.SEMANTICS_HTTP (optional)
            attributeSemantics: string,
            // Parameters to configure the creation of random attributes. If missing, no random attributes
            // are added to the span (optional)
            randomAttributes: {
                // The number of random attributes to generate
                count: int,
                // The number of distinct values to generate for each attribute (optional, default: 50)
                cardinality: int
            },
            // Additional attributes for the resource associated with this span. Resource attribute definitions
            // of different spans with the same service name will me merged into a singe resource (optional)
            resource: {
                // Fixed attributs that are added to the resource (optional)
                attributes: { string : any },
                // Parameters to configure the creation of random resource attributes (optional)
                randomAttributes: {
                    // The number of random attributes to generate
                    count: int,
                    // The number of distinct values to generate for each attribute (optional, default: 50)
                    cardinality: int
                }
            }
        },
        ...
    ] 
}
```

An example with a templated generator can be found in [./examples/template](./examples/template).

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
To test the binary you first need to change the endpoint in the client configuration to:

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
