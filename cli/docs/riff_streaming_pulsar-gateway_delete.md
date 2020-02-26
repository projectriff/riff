---
id: riff-streaming-pulsar-gateway-delete
title: "riff streaming pulsar-gateway delete"
---
## riff streaming pulsar-gateway delete

delete pulsar gateway(s)

### Synopsis

Delete one or more Pulsar gateways by name or all Pulsar gateways within a
namespace.

Deleting a Pulsar gateway will disrupt all processors consuming streams managed
by the gateway. Existing messages in the stream may be preserved by the
underlying pulsar broker, depending on the implementation.

```
riff streaming pulsar-gateway delete <name(s)> [flags]
```

### Examples

```
riff streaming pulsar-gateway delete my-pulsar-gateway
riff streaming pulsar-gateway delete --all 
```

### Options

```
      --all              delete all pulsar gateways within the namespace
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

* [riff streaming pulsar-gateway](riff_streaming_pulsar-gateway.md)	 - (experimental) pulsar stream gateway

