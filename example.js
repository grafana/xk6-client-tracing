import { sleep } from 'k6';
import remote from 'k6/x/tracing';

export let options = {
    vus: 10,
    duration: '10s',
};

const client = new remote.Client({
    url: "test.test:6666"
});

export default function () {
    let spans = [
        {name: "trace0"},
        {name: "trace1"},
        {name: "trace2"},
        {name: "trace3"},
        {name: "trace4"},
    ];
    let res = client.test(spans);
    console.log(res)
    sleep(1)
}