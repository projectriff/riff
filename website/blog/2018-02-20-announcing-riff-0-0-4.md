---
title: "Announcing riff 0.0.4"
---

We are happy to announce a new release of riff. Thank you, once again, everyone
who contributed to this effort. Here are some of the highlights.

<!--truncate-->

## a new riff CLI written in go
The [riff CLI](https://github.com/projectriff/riff/tree/master/riff-cli) is now a go binary, available to download from our GitHub [releases](https://github.com/projectriff/riff/releases) page or, if you have a go development environment, you can use 'go get' to build and install into $GOPATH/bin as follows.

```
go get github.com/projectriff/riff
```

The new CLI provides a lot more infomation about what it's doing. E.g.
```
~/riff/riff/samples/shell/echo (master)$ riff create
Initializing ~/riff/riff/samples/shell/echo/echo-topics.yaml
Initializing ~/riff/riff/samples/shell/echo/echo-function.yaml
Initializing ~/riff/riff/samples/shell/echo/Dockerfile
Building image ...
```

A `--dry-run` option displays the content of the Dockerfile and k8s resource .yaml files without
generating them.

NOTE that some of the riff [CLI configuration](https://github.com/projectriff/riff/blob/master/Getting-Started.adoc#riff-cli-configuration) options have changed.

## gRPC under the hood
For this iteration we have introduced a gRPC interface between the function sidecar
container and each of the function invokers.

While HTTP works well for invoking functions one event at a time, it was not designed to serve as a protocol for bidirectional streams which don't follow a strict request/reply pattern.

gRPC will allow us to extend streaming semantics, which already exist for [Java functions](https://github.com/projectriff/java-function-invoker/tree/master/src/test/java/io/projectriff/functions) using the reactive Flux interface, to functions written in JavaScript and other languages.

Function code provided by users should not be impacted by the underlying changes between sidecar and invoker, in fact, we expect the invoker protocol to be hidden from function developers in future releases.

### gRPC proto
Here is the gRPC function.proto definition which we are using for the 0.0.4 release. Please note that we are still experimenting with this and other streaming protocols. Developers who are experimenting with writing new invokers should expect changes to this layer in future releases. 

```protobuf
syntax = "proto3";

package function;

message Message {
	message HeaderValue {
		repeated string values = 1;
	}

	bytes payload = 1;
	map<string, HeaderValue> headers = 2;
}

service MessageFunction {
  rpc Call(stream Message) returns (stream Message) {}
}
```

## go shell invoker
The [shell invoker](https://github.com/projectriff/shell-function-invoker) is now a go binary executable, which executes commands directly rather than running them from inside a shell. This allows the shell invoker to connect to the sidecar via gRPC just like other languages.

NOTE that the new shell invoker requires a shebang for shell scripts, and uses stdin instead of a command line parameter for inputs from events. The echo sample has been modified to use the 'cat' utility which simply copies stdin to stdout. Here is the `echo.sh` script from the sample.

```sh
#!/bin/sh
cat
```

Alternatively, you could specify the `cat` command directly in the Dockerfile - no shell script required!

```docker
FROM projectriff/shell-function-invoker:0.0.4
ENV FUNCTION_URI cat
```

## node invoker
The [node invoker](https://github.com/projectriff/node-function-invoker) just keeps getting better, with HTTP and gRPC protocols both implemented in this release.

For functions which need to manage connections or perform other kinds of one time setup/teardown, the invoker now calls `$init` on startup and `$destroy` before terminating the function. These functions can return promises as well. The node invoker has been fixed to respond promptly to termination signals from Kubernetes.

To help you create functions with npm dependencies, the CLI will now recognize a `package.json` file, and generate a Dockerfile which copies the whole directory into the image and installs dependencies during the build.

E.g. `riff init node --dry-run -a package.json` will produce:

```docker
FROM projectriff/node-function-invoker:0.0.4
ENV FUNCTION_URI /functions/
COPY . ${FUNCTION_URI}
RUN (cd ${FUNCTION_URI} && npm install --production)
```

Finally, the HTTP gateway will now respond with a 500 status when node functions throw an error.

## java invoker
Our java invoker is still the leader in terms of supporting streaming (Flux) as well as request/response functions.

In this release, the [java invoker](https://github.com/projectriff/java-function-invoker/commit/60d675c48817cc75f17af76178afe588a5cd8b42) is talking to the sidecar over the same gRPC interface described above. Support for the `pipes` protocol has been removed, so please modify your function yaml to use `grpc` if you upgrade to 0.0.4 and rebuild your functions with the new invoker.

## separate Kafka install
Starting with the 0.0.4 release, we have separated the Kafka installation from the riff Helm chart. You can still use the single-node Kafka chart provided by riff, or the dynamically scalable, multi-node [incubator/kafka](https://github.com/kubernetes/charts/tree/master/incubator/kafka) service. For details see the [Getting Started](https://github.com/projectriff/riff/blob/master/Getting-Started.adoc) guide on GitHub or the [docs](/docs) on this site.

## next steps

In addition to our pursuit of improved interfaces for event-streaming, we are continuing to invest in
making our architecure more pluggable so that riff can connect to other message brokers. 

We are also targeting better function and topic management in the function controller,
and smoother function scaling using event metrics and lifecycle information from our sidecars.
