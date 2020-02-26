---
id: riff-streaming-kafka-gateway-create
title: "riff streaming kafka-gateway create"
---
## riff streaming kafka-gateway create

create a kafka gateway of messages

### Synopsis

Creates a Kafka gateway within a namespace.

The gateway is configured with the address of the Kafka broker.

```
riff streaming kafka-gateway create <name> [flags]
```

### Examples

```
riff streaming kafka-gateway create my-kafka-gateway --bootstrap-servers kafka.local:9092
```

### Options

```
      --bootstrap-servers address   address of the kafka broker
      --dry-run                     print kubernetes resources to stdout rather than apply them to the cluster, messages normally on stdout will be sent to stderr
  -h, --help                        help for create
  -n, --namespace name              kubernetes namespace (defaulted from kube config)
      --tail                        watch creation progress
      --wait-timeout duration       duration to wait for the gateway to become ready when watching progress (default 1m0s)
```

### Options inherited from parent commands

```
      --config file       config file (default is $HOME/.riff.yaml)
      --kubeconfig file   kubectl config file (default is $HOME/.kube/config)
      --no-color          disable color output in terminals
```

### SEE ALSO

* [riff streaming kafka-gateway](riff_streaming_kafka-gateway.md)	 - (experimental) kafka stream gateway

