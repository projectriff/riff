---
id: riff-streaming-stream-create
title: "riff streaming stream create"
---
## riff streaming stream create

create a stream of messages

### Synopsis

Create a stream resource within a namespace and provision a stream in the
underlying message broker via the referenced stream gateway.

The created stream can then be referenced as an input or an output of a given
function when creating a streaming processor.

```
riff streaming stream create <name> [flags]
```

### Examples

```
riff streaming stream create my-stream --gateway my-gateway
```

### Options

```
      --content-type MIME type   MIME type for message payloads accepted by the stream
      --dry-run                  print kubernetes resources to stdout rather than apply them to the cluster, messages normally on stdout will be sent to stderr
      --gateway name             name of stream gateway
  -h, --help                     help for create
  -n, --namespace name           kubernetes namespace (defaulted from kube config)
      --tail                     watch provisioning progress
      --wait-timeout duration    duration to wait for the stream to become ready when watching progress (default 10s)
```

### Options inherited from parent commands

```
      --config file       config file (default is $HOME/.riff.yaml)
      --kubeconfig file   kubectl config file (default is $HOME/.kube/config)
      --no-color          disable color output in terminals
```

### SEE ALSO

* [riff streaming stream](riff_streaming_stream.md)	 - (experimental) streams of messages

