---
title: "Introducing riff"
---

Welcome to the riff blog! With this first post, we are pleased to announce the 0.0.2 release of riff. The 0.0.2 release includes the code as shown in demos at SpringOne Platform, along with a handful of issues we were able to address over the past few days. 

<!--truncate-->

riff was introduced to the world during the Wednesday [keynote](/video/mark-fisher-at-springone-platform-2017/) at SpringOne Platform. Our goal was to describe riff in one concise story, with a slide for each step of the function lifecycle.
- developers write functions
- functions are packaged as containers
- sidecars connect functions with event brokers
- functions and topics are Kubernetes resources
- functions scale with events
- functions can process streams

Let’s walk through each of these in a little more detail.

### developers write functions

Functions can be written in any supported language, with the scope kept as small as possible. For example in javascript, it can be just a single function defined in a single file, such as “square.js” used in the demo:

```js
module.exports = (x) => x**2
```

For java it can be a single implementation of java.util.Function packaged into a JAR which only needs to include dependencies that are needed for the scope of the function.

### functions are packaged as containers

Code is packaged into a container by adding it as a layer in the container build. The base image is the language-specific function invoker. We provide invokers for shell, javascript, python, and java, but we will soon be documenting the process for creating new invoker images. We describe this as a form of Inversion of Control. It’s the Hollywood Principle: “don’t call us, we’ll call you”. It also looks a lot like the Template pattern.

### sidecars connect functions with event brokers

The sidecar is written in Go, and it can communicate with any of the language-specific function invokers. The Function Controller deploys the sidecar in the same pod as the function container. This is also a form of Inversion Control, essentially the Proxy pattern. The sidecar handles cross-cutting responsibilities so that the invokers can be as simple as possible. 

The sidecar handles all communication with the event broker, and it then sends and receives events over whatever protocol the particular invoker expects. We currently support HTTP, gRPC, and named pipes, but we are considering being more opinionated about that (see the Roadmap section for more detail). 

One way to send events to a topic that may have one or more functions subscribed is to use riff’s HTTP Gateway. The Gateway passes events to the topic whose name matches the final part of the URI path, with a base path of either /messages (fire-and-forget) or /requests (request/reply). We will also be exploring more sophisticated gateway options.

### functions and topics are Kubernetes resources

Kubernetes supports extensions via Custom Resource Definitions, and riff defines two: Function and Topic. Those resources are then described via YAML based on the model we define. This allows us to rely on Kubernetes for state management and the ability to “watch” for any additions, modifications, or deletions of those resources.

The Function Controller watches Function resources, and the Topic Controller watches Topic resources. When a new Function is added, the Function Controller will start managing its lifecycle including deployments (along with a sidecar) and scaling. The Topic Controller registers any new Topics with the underlying event broker (currently Kafka, but this will be pluggable).

### functions scale with events

In addition to watching Function resources, the Function Controller also monitors event activity. It scales a function from 0-1 as soon as any activity occurs on that function’s input topic. It then scales in the 1-N-1 range based on the lag between available and consumed events. The value of N is configurable via the “maxReplicas” property in function YAML, but the default value will correspond to the number of partitions on the function’s input topic. The Function Controller also scales functions to 0 when there is no activity for an “idle-timeout” period. That is also a configurable property in the YAML, but it’s default is 10 seconds.

### functions can process streams

One of the key objectives of riff is to provide first-class support for event stream processing. We do not want functions to be limited to handling a single request at a time. On the contrary, we want to expose the full capabilities of the underlying language and libraries when it comes to concurrency and streaming support. In the case of java, we support Reactor and the use of Flux types for input and output. In the SpringOne Platform demo, we used a windowing operation and emitted aggregated vote counts for 60-second windows that shifted every 2-seconds:

```java
public Flux<Map<String, Object>> windows(Flux<String> words) {
    return words.window(Duration.ofSeconds(60), Duration.ofSeconds(2))
                        .concatMap(w -> w.collect(VoteAggregate::new, VoteAggregate::sum)
                        .map(VoteAggregate::windowMap), Integer.MAX_VALUE);
}
```
## Roadmap

Here are some highlights of our plans for the upcoming iterations:

- Port the Function Controller to Go.
- Abstractions to make the event broker pluggable (not just Kafka).
- Explore the use of [rsocket](https://rsocket.io) between the Sidecar and Function Invoker.
- Prototype dynamically loaded functions using pools of pre-warmed invoker containers.
- Adjust our scaling algorithm and exposing more configurable properties.
- Add support for volumes to be defined in function configurations.