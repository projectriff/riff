---
title: "Announcing riff v0.0.6"
---

The 0.0.6 release of riff is now available. A big thank you, once again, to everyone
who contributed on this effort.

<!--truncate-->

## riff installation

We recommend [installing](/docs/) riff v0.0.6 on a fresh Kubernetes cluster. The latest Helm chart has an option to deploy riff together with Kafka. If you played with previous riff releases, remember to install the [latest riff CLI](https://github.com/projectriff/riff/releases) as well.

## api version change
If you have existing functions and topics, and prefer not to regenerate the yaml for these using the latest CLI, you will need to change the `apiVersion` in the yaml to `projectriff.io/v1alpha1`. E.g.

```yaml
---
apiVersion: projectriff.io/v1alpha1
kind: Function
metadata:
  name: square
spec:
  container:
    image: projectriff/square:0.0.2
  input: numbers
  protocol: grpc
```

Dockerfiles should not require any changes for this release.


## installable invokers
Starting in v0.0.6, riff [invokers](/invokers/) are installable Kubernetes resources.

This invoker separation is the first step toward future enhancements such as invoker-specific configuration, validation, and, dynamic loading of functions into pre-warmed invoker containers.

The yaml file for an invoker can come from a file on disk or from a URL. This allows users to add new invokers without changes to the CLI. The riff CLI has been extended with `riff invokers` sub-commands.

```
$ riff invokers --help
Manage invokers in the cluster

Usage:
  riff invokers [command]

Available Commands:
  apply       Install or update an invoker in the cluster
  delete      Remove an invoker from the cluster
  list        List invokers in the cluster
```

To install the latest available invokers, run the following CLI commands or refer to the [invoker docs](/invokers/).

```bash
riff invokers apply -f https://github.com/projectriff/command-function-invoker/raw/v0.0.6/command-invoker.yaml
riff invokers apply -f https://github.com/projectriff/go-function-invoker/raw/v0.0.2/go-invoker.yaml
riff invokers apply -f https://github.com/projectriff/java-function-invoker/raw/v0.0.5-sr.1/java-invoker.yaml
riff invokers apply -f https://github.com/projectriff/node-function-invoker/raw/v0.0.6/node-invoker.yaml
riff invokers apply -f https://github.com/projectriff/python2-function-invoker/raw/v0.0.6/python2-invoker.yaml
riff invokers apply -f https://github.com/projectriff/python3-function-invoker/raw/v0.0.6/python3-invoker.yaml
```

To generate a Dockerfile and yaml resources for an invoker, specify the invoker with 'riff init' or 'riff create' E.g.

```bash
riff create node
```

## node 'message' type

JavaScript functions that need to interact with headers can now opt to receive and/or produce messages. A message is an object that contains both headers and a payload. Message headers are a map with case-insensitive keys and multiple string values.

Since JavaScript and Node have no built-in type for messages or headers, riff uses the [@projectriff/message](https://github.com/projectriff/node-message/) npm module. To use messages, functions should install the `@projectriff/message` package:

```bash
npm install --save @projectriff/message
```

### receiving messages

```js
const { Message } = require('@projectriff/message');

// a function that accepts a message, which is an instance of Message
module.exports = message => {
    const authorization = message.headers.getValue('Authorization');
    ...
};

// tell the invoker the function wants to receive messages
module.exports.$argumentType = 'message';

// tell the invoker to produce this particular type of message
Message.install();
```

### producing messages

```js
const { Message } = require('@projectriff/message');

const instanceId = Math.round(Math.random() * 10000);
let invocationCount = 0;

// a function that produces a Message
module.exports = name => {
    return Message.builder()
        .addHeader('X-Riff-Instance', instanceId)
        .addHeader('X-Riff-Count', invocationCount++)
        .payload(`Hello ${name}!`)
        .build();
};

// even if the function receives payloads, it can still produce a message
module.exports.$argumentType = 'payload';
```


## http-gateway topic validation
We’ve updated the http-gateway to validate topics and return an HTTP 404 error when it receives a message or request on an unknown topic endpoint.

```
$ riff publish --input nosuchrifftopic --data "404 From Message"
Posting to http://192.168.39.148:32508/messages/nosuchrifftopic
could not find Riff topic 'nosuchrifftopic'

riff publish --input nosuchrifftopic --data "404 From Request" --reply
Posting to http://192.168.39.148:32508/requests/nosuchrifftopic
could not find Riff topic 'nosuchrifftopic'
```

Note that, for now, Functions will run only in the 'default' k8s namespace.

## foundations for new autoscaler

Prior to v0.0.6, the function-controller scaled up the number of replicas of a function pod in response to changes in producer and consumer offsets in the input topic. The 0.0.6 autoscaler reproduces this behaviour but, instead of using offsets, uses the queue length together with production and consumption metrics from the topic. This is a step towards enabling riff to support message brokers other than Kafka. We also factored out the autoscaler subcomponent in the code, and introduced a workload simulator to measure the autoscaler behaviour as a baseline for future improvements.

The simulations below show three busy periods separated by periods of inactivity:
1. writes increase and then decrease in a step function
2. writes vary like a sine wave
3. writes ramp up and down linearly.

The graphs show writes in black, queue length in light blue, and the number of replicas in red. Although the number of replicas would normally be limited by the number of partitions or by user configuration, the simulator does not apply a limit so that the autoscaler behaviour is easy to see.

The first graph shows what the behaviour would be if replicas started up instantaneously. The queue length stays under control and the number of instances is fairly small at all times. However, there is high frequency “noise” in the number of replicas, so the smoothing needs improving.

![graph with instant startup](/img/graph1.png)

The second graph shows a more realistic scenario in which each replica takes a while to start up. When there is a sudden increase in workload, the queue length builds dramatically while replicas are starting, and the autoscaler over-reacts - another area for improvement.

![graph with slower startup](/img/graph2.png)
 