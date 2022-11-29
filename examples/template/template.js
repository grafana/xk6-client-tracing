import {sleep} from 'k6';
import tracing from 'k6/x/tracing';

export const options = {
    vus: 1,
    duration: "20m",
};

const endpoint = __ENV.ENDPOINT || "otel-collector:4317"
const client = new tracing.Client({
    endpoint,
    exporter: tracing.EXPORTER_OTLP,
    insecure: true,
});

const traceDefaults = {
    attributeSemantics: tracing.SEMANTICS_HTTP,
    attributes: {"one": "three"},
    randomAttributes: {count: 2, cardinality: 5}
}

const traceTemplates = [
    {
        defaults: traceDefaults,
        spans: [
            {service: "shop-backend", name: "list-articles", duration: {min: 200, max: 900}},
            {service: "shop-backend", name: "authenticate", duration: {min: 50, max: 100}},
            {service: "auth-service", name: "authenticate"},
            {service: "shop-backend", name: "fetch-articles", parentIdx: 0},
            {service: "article-service", name: "list-articles"},
            {service: "article-service", name: "select-articles", attributeSemantics: tracing.SEMANTICS_DB},
            {service: "postgres", name: "query-articles", attributeSemantics: tracing.SEMANTICS_DB, randomAttributes: {count: 5}},
        ]
    },
    {
        defaults: {
            attributeSemantics: tracing.SEMANTICS_HTTP,
        },
        spans: [
            {service: "shop-backend", name: "article-to-cart", duration: {min: 400, max: 1200}},
            {service: "shop-backend", name: "authenticate", duration: {min: 70, max: 200}},
            {service: "auth-service", name: "authenticate"},
            {service: "shop-backend", name: "get-article", parentIdx: 0},
            {service: "article-service", name: "get-article"},
            {service: "article-service", name: "select-articles", attributeSemantics: tracing.SEMANTICS_DB},
            {service: "postgres", name: "query-articles", attributeSemantics: tracing.SEMANTICS_DB, randomAttributes: {count: 2}},
            {service: "shop-backend", name: "place-articles", parentIdx: 0},
            {service: "cart-service", name: "place-articles", attributes: {"article.count": 1, "http.status_code": 201}},
            {service: "cart-service", name: "persist-cart"}
        ]
    },
    {
        defaults: traceDefaults,
        spans: [
            {service: "shop-backend", attributes: {"http.status_code": 403}},
            {service: "shop-backend", name: "authenticate"},
            {service: "auth-service", name: "authenticate", attributes: {"http.status_code": 403}},
        ]
    },
]

export default function () {
    traceTemplates.forEach(function (tmpl) {
        let gen = new tracing.TemplatedGenerator(tmpl)
        let traces = gen.traces()
        client.push(traces)
    });

    sleep(5);
}

export function teardown() {
    client.shutdown();
}
