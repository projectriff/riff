---
id: riff-streaming-stream-list
title: "riff streaming stream list"
---
## riff streaming stream list

table listing of streams

### Synopsis

List streams in a namespace or across all namespaces.

For detail regarding the status of a single stream, run:

    riff streaming stream status <stream-name>

```
riff streaming stream list [flags]
```

### Examples

```
riff streaming stream list
riff streaming stream list --all-namespaces
```

### Options

```
      --all-namespaces   use all kubernetes namespaces
  -h, --help             help for list
  -n, --namespace name   kubernetes namespace (defaulted from kube config)
```

### Options inherited from parent commands

```
      --config file       config file (default is $HOME/.riff.yaml)
      --kubeconfig file   kubectl config file (default is $HOME/.kube/config)
      --no-color          disable color output in terminals
```

### SEE ALSO

* [riff streaming stream](riff_streaming_stream.md)	 - (experimental) streams of messages

