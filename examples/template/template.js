import {sleep} from 'k6';
import tracing from 'k6/x/tracing';
import { randomIntBetween } from 'https://jslib.k6.io/k6-utils/1.2.0/index.js';

export const options = {
    vus: 1,
    duration: "20m",
};

const endpoint = __ENV.ENDPOINT || "otel-collector:4317"
const orgid = __ENV.TEMPO_X_SCOPE_ORGID || "k6-test"
const client = new tracing.Client({
    endpoint,
    exporter: tracing.EXPORTER_OTLP,
    tls: {
        insecure: true,
    },
    headers: {
        "X-Scope-Orgid": orgid
    }
});

const traceDefaults = {
    attributeSemantics: tracing.SEMANTICS_HTTP,
    attributes: {"one": "three"},
    randomAttributes: {count: 2, cardinality: 5},
    randomEvents: {count: 0.1, exceptionCount: 0.2, randomAttributes: {count: 6, cardinality: 20}},
    resource: { randomAttributes: {count: 3} },
}

const traceTemplates = [
    {
        defaults: traceDefaults,
        spans: [
            {service: "shop-backend", name: "list-articles", duration: {min: 200, max: 900}, resource: { attributes: {"namespace": "shop"} }},
            {service: "shop-backend", name: "authenticate", duration: {min: 50, max: 100}, resource: { randomAttributes: {count: 4} }},
            {service: "auth-service", name: "authenticate", resource: { randomAttributes: {count: 2}, attributes: {"namespace": "auth"} }},
            {service: "shop-backend", name: "fetch-articles", parentIdx: 0},
            {
                service: "article-service",
                name: "list-articles",
                links: [{attributes: {"link-type": "parent-child"}, randomAttributes: {count: 2, cardinality: 5}}],
                resource: { attributes: {"namespace": "shop" }}
            },
            {service: "article-service", name: "select-articles", attributeSemantics: tracing.SEMANTICS_DB},
            {service: "postgres", name: "query-articles", attributeSemantics: tracing.SEMANTICS_DB, randomAttributes: {count: 5}, resource: { attributes: {"namespace": "db"} }},
        ]
    },
    {
        defaults: {
            attributes: {"numbers": ["one", "two", "three"]},
            attributeSemantics: tracing.SEMANTICS_HTTP,
            randomEvents: {count: 2, randomAttributes: {count: 3, cardinality: 10}},
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
            {service: "shop-backend", name: "authenticate", attributes: {"http.request.header.accept": ["application/json"]}},
            {
                service: "auth-service",
                name: "authenticate",
                attributes: {"http.status_code": 403},
                randomEvents: {count: 0.5, exceptionCount: 2, randomAttributes: {count: 5, cardinality: 5}}
            },
        ]
    },
    {
        defaults: traceDefaults,
        spans: [
            {service: "shop-backend"},
            {service: "shop-backend", name: "authenticate", attributes: {"http.request.header.accept": ["application/json"]}},
            {service: "auth-service", name: "authenticate"},
            {
                service: "cart-service",
                name: "checkout",
                randomEvents: {count: 0.5, exceptionCount: 2, exceptionOnError: true, randomAttributes: {count: 5, cardinality: 5}}
            },
            {
                service: "billing-service",
                name: "payment",
                randomLinks: {count: 0.5, randomAttributes: {count: 3, cardinality: 10}},
                randomEvents: {exceptionOnError: true, randomAttributes: {count: 4}}}
        ]
    },
]

export default function () {
    const templateIndex = randomIntBetween(0, traceTemplates.length-1)
    const gen = new tracing.TemplatedGenerator(traceTemplates[templateIndex])
    client.push(gen.traces())

    sleep(randomIntBetween(1, 5));
}

export function teardown() {
    client.shutdown();
}