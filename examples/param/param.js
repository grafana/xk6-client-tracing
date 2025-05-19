import { sleep } from 'k6';
import tracing from 'k6/x/tracing';
import { randomIntBetween } from 'https://jslib.k6.io/k6-utils/1.2.0/index.js';

export let options = {
    vus: 1,
    duration: "20m",
};

const endpoint = __ENV.ENDPOINT || "otel-collector:4317"
const client = new tracing.Client({
    endpoint,
    exporter: tracing.EXPORTER_OTLP,
    tls: {
        insecure: true,
    }
});

export default function () {
    let pushSizeTraces = randomIntBetween(2, 3);
    let pushSizeSpans = 0;
    let t = [];
    for (let i = 0; i < pushSizeTraces; i++) {
        let c = randomIntBetween(5, 10)
        pushSizeSpans += c;

        t.push({
            random_service_name: false,
            count: 1,
            resource_size: 100,
            spans: {
                count: c,
                size: randomIntBetween(300, 1000),
                random_name: true,
                fixed_attrs: {
                    "test": "test",
                },
            }
        });
    }

    let gen = new tracing.ParameterizedGenerator(t)
    let traces = gen.traces()
    client.push(traces);

    console.log(`Pushed ${pushSizeSpans} spans from ${pushSizeTraces} different traces. Here is a random traceID: ${t[Math.floor(Math.random() * t.length)].id}`);
    sleep(15);
}

export function teardown() {
    client.shutdown();
}
