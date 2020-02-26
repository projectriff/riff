---
id: riff-streaming-inmemory-gateway-delete
title: "riff streaming inmemory-gateway delete"
---
## riff streaming inmemory-gateway delete

delete in-memory gateway(s)

### Synopsis

Delete one or more in-memory gateways by name or all in-memory gateways within
a namespace.

Deleting a in-memory gateway will disrupt all processors consuming streams
managed by the gateway. Existing messages in the stream may be preserved by the
underlying in-memory broker, depending on the implementation.

```
riff streaming inmemory-gateway delete <name(s)> [flags]
```

### Examples

```
riff streaming inmemory-gateway delete my-inmemory-gateway
riff streaming inmemory-gateway delete --all 
```

### Options

```
      --all              delete all inmemory gateways within the namespace
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

* [riff streaming inmemory-gateway](riff_streaming_inmemory-gateway.md)	 - (experimental) in-memory stream gateway

