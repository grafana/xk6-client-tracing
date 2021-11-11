import { check, sleep } from 'k6';
import tracing from 'k6/x/tracing';

export let options = {
    vus: 10,
    duration: '10s',
};

const client = new tracing.Client({
    endpoint: "0.0.0.0:14250",
    exporter: "jaeger"
});

export default function () {
    let traces = {

    }
    client.send([{
        name: "Example",
        attributes: {
            "test": "test"
        },
        status: {
            code: 0,
            message: "ok"
        }
    }]);
    sleep(1)
}
