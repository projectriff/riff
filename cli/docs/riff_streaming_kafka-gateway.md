---
id: riff-streaming-kafka-gateway
title: "riff streaming kafka-gateway"
---
## riff streaming kafka-gateway

(experimental) kafka stream gateway

### Synopsis

The Kafka gateway encapsulates the address of a streaming gateway and a Kafka
provisioner instance.

The Kafka provisioner is responsible for creating topics in a Kafka cluster. The
streaming gateway coordinates and standardizes reads and writes to a Kafka
broker.

### Options

```
  -h, --help   help for kafka-gateway
```

### Options inherited from parent commands

```
      --config file       config file (default is $HOME/.riff.yaml)
      --kubeconfig file   kubectl config file (default is $HOME/.kube/config)
      --no-color          disable color output in terminals
```

### SEE ALSO

* [riff streaming](riff_streaming.md)	 - (experimental) streaming runtime for riff functions
* [riff streaming kafka-gateway create](riff_streaming_kafka-gateway_create.md)	 - create a kafka gateway of messages
* [riff streaming kafka-gateway delete](riff_streaming_kafka-gateway_delete.md)	 - delete kafka gateway(s)
* [riff streaming kafka-gateway list](riff_streaming_kafka-gateway_list.md)	 - table listing of kafka gateways
* [riff streaming kafka-gateway status](riff_streaming_kafka-gateway_status.md)	 - show kafka gateway status

