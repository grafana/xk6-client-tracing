import { check, sleep } from 'k6';
import tracing from 'k6/x/tracing';

export let options = {
    vus: 10,
    duration: '10s',
};

const client = new tracing.Client({
    endpoint: "0.0.0.0:4317"
});

export default function () {
    client.sendFake();
    sleep(1)
}
