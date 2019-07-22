---
title: "riff 0.0.5"
---

We are happy to announce the 0.0.5 release of riff. A huge thanks to all
who contributed to this effort. Here are some of the highlights.

<!--truncate-->

## go invoker
A new [Go function invoker](https://github.com/projectriff/go-function-invoker) provides support for functions built as [Go plugins](https://golang.org/pkg/plugin/). This makes it possible to compile Go functions into standalone packages, loaded at runtime into the Go function invoker. E.g.

```go
package main

// functions that don't return error are also supported
func Encode(in string) (string, error) { 
	result := []rune(in)
	for i, c := range result {
		switch {
		case 'a' <= c && c <= 'z':
			result[i] = 'z' - (c - 'a')
		case 'A' <= c && c <= 'Z':
			result[i] = 'Z' - (c - 'A')
		}
	}
	return string(result), nil
}
```

The Go invoker uses the gRPC protocol to connect to the riff sidecar. It calls an exported function inside the Go plugin for each request in the input stream, sending the return value to the output stream.

## http request header whitelisting
We've introduced the ability to whitelist HTTP request headers. In this release, the whitelist is applied globally to all requests at the http-gateway.

The primary use case motivating this work was webhooks -- shoutout here to [Kelsey Hightower](https://youtu.be/MLq_CWqQauA?t=4918)!

Github signs webhook requests with the `X-Hub-Signature header`. To pass this header through to all riff function invokers, add the following parameter to your helm install

```
--set httpGateway.httpHeadersWhitelist=X-Hub-Signature
```

Whitelisted header values are passed to java functions via the java function invoker. We will be adding message header support for other language invokers soon.

## streaming node functions
The [node function invoker](https://github.com/projectriff/node-function-invoker) now has experimental support for streaming functions. 

Setting `$interactionModel` to `node-streams` will cause the function to be invoked with two arguments, an input [Readable Stream](https://nodejs.org/dist/latest-v8.x/docs/api/stream.html#stream_class_stream_readable) and an output [Writeable Stream](https://nodejs.org/dist/latest-v8.x/docs/api/stream.html#stream_class_stream_writable).

```js
// echo.js
module.exports = (input, output) => {
    input.pipe(output);
};
module.exports.$interactionModel = 'node-streams';
```
Npm packages that work with Node Streams, like [mississippi](https://github.com/maxogden/mississippi), can be used to manipulate the streams.

## improved Python support
A [Python 3 function invoker](https://github.com/projectriff/python3-function-invoker) has been added and the [Python 2 function invoker]() has been enhanced. Both support the gRPC protocol and use the Content-Type header to convert input messages to str or dict. 

Python 3 is now the default when using the riff CLI. The Python 2 invoker may be deprecated and dropped from future releases unless there is a compelling reason to maintain it. 

```python
import collections

def concat(vals):
    '''
    :param vals: expects a dict
    :return: a singleton dict whose value is concatenated keys and values
    '''
    od = collections.OrderedDict(sorted(vals.items()))
    result = ''
    for k, v in od.items():
        result = result + k + v
    return {'result': result}
```

## mono-repo and new CI
The core components of riff now live in a single [git repo](https://github.com/projectriff/riff), making it easier to build and test while managing dependencies across components. A single unified Makefile is provided at the root of the repo.

We have moved our test pipeline to travis-ci which is publicly visible and supports nice integration with GitHub PRs. Check it out at [https://travis-ci.org/projectriff](https://travis-ci.org/projectriff). 

The ci pipeline automatically publishes a "snapshot" version of the helm chart for new builds on the master branch.

```
$ helm search projectriff/riff -l 
NAME            	VERSION       	DESCRIPTION                                  
projectriff/riff	0.0.6-snapshot	riff is for functions - a FaaS for Kubernetes
projectriff/riff	0.0.5         	riff is for functions - a FaaS for Kubernetes
```

## miscellaneous enhancements

- `riff publish` will now work without specifying `--namespace` even if riff is running in another namespace like riff-system. The `--namespace` option has been deprecated

- `riff create` or `riff init` will take an `--invoker-version` parameter to override the default version tag on the invoker base image in the Dockerfile; this used to be `--riff-version`.

- The shell-function-invoker has been renamed [command-function-invoker](https://github.com/projectriff/command-function-invoker).

## in the pipeline for 0.0.6
- autoscaling using sidecar metrics
- invoker Custom Resources
- event source functions
- message event payloads with headers for node and other invokers
- iteration planning in public GitHub projects
