---
id: riff-streaming-stream-delete
title: "riff streaming stream delete"
---
## riff streaming stream delete

delete stream(s)

### Synopsis

Delete one or more streams by name or all streams within a namespace.

Deleting a stream will prevent processors from reading and writing messages on
the stream. Existing messages in the stream may be preserved by the underlying
messaging middleware, depending on the implementation.

```
riff streaming stream delete <name(s)> [flags]
```

### Examples

```
riff streaming stream delete my-stream
riff streaming stream delete --all 
```

### Options

```
      --all              delete all streams within the namespace
  -h, --help             help for delete
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

