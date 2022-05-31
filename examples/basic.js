import { sleep } from 'k6';
import tracing from 'k6/x/tracing';
import { randomIntBetween } from 'https://jslib.k6.io/k6-utils/1.2.0/index.js';
import { SharedArray } from 'k6/data';

export let options = {
    vus: 1,
    duration: "5m",
};

const traceIDs = new SharedArray('traceIDs', function () {
    let toret = [];
    for (let i = 0; i < 10; i++) {
        toret.push(tracing.generateRandomTraceID());
    }
    return toret;
});

const client = new tracing.Client({
    endpoint: "0.0.0.0:4317",
    exporter: "otlp",
    insecure: true,
});

export default function () {
    let pushSizeTraces = randomIntBetween(2,3);
    let pushSizeSpans = 0;
    let t = [];
    for (let i = 0; i < pushSizeTraces; i++) {
        let c = randomIntBetween(5,10)
        pushSizeSpans += c;

        t.push({
            id: traceIDs[Math.floor(Math.random() * traceIDs.length)],
            random_service_name: false,
            spans: {
                count: c,
                size: randomIntBetween(300,1000),
                random_name: true,
            }
        });
    }
    client.push(t);
    console.log(`Pushed ${pushSizeSpans} spans from ${pushSizeTraces} different traces. Here is a random traceID: ${t[Math.floor(Math.random() * t.length)].id}`);
    sleep(15);
}

export function teardown() {
    client.shutdown();
}