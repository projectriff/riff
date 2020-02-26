---
id: riff-streaming-kafka-gateway-delete
title: "riff streaming kafka-gateway delete"
---
## riff streaming kafka-gateway delete

delete kafka gateway(s)

### Synopsis

Delete one or more Kafka gateways by name or all Kafka gateways within a
namespace.

Deleting a Kafka gateway will disrupt all processors consuming streams managed
by the gateway. Existing messages in the stream may be preserved by the
underlying Kafka broker, depending on the implementation.

```
riff streaming kafka-gateway delete <name(s)> [flags]
```

### Examples

```
riff streaming kafka-gateway delete my-kafka-gateway
riff streaming kafka-gateway delete --all 
```

### Options

```
      --all              delete all kafka gateways within the namespace
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

* [riff streaming kafka-gateway](riff_streaming_kafka-gateway.md)	 - (experimental) kafka stream gateway

