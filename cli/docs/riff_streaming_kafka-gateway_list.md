---
id: riff-streaming-kafka-gateway-list
title: "riff streaming kafka-gateway list"
---
## riff streaming kafka-gateway list

table listing of kafka gateways

### Synopsis

List Kafka gateways in a namespace or across all namespaces.

For detail regarding the status of a single Kafka gateway, run:

    riff streaming kafka-gateway status <kafka-gateway-name>

```
riff streaming kafka-gateway list [flags]
```

### Examples

```
riff streaming kafka-gateway list
riff streaming kafka-gateway list --all-namespaces
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

* [riff streaming kafka-gateway](riff_streaming_kafka-gateway.md)	 - (experimental) kafka stream gateway

