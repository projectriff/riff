---
id: riff-streaming-pulsar-gateway
title: "riff streaming pulsar-gateway"
---
## riff streaming pulsar-gateway

(experimental) pulsar stream gateway

### Synopsis

The Pulsar gateway encapsulates the address of a streaming gateway and a Pulsar
provisioner instance.

The Pulsar provisioner is responsible for resolving topic addresses in a Pulsar
cluster. The streaming gateway coordinates and standardizes reads and writes to
a Pulsar broker.

### Options

```
  -h, --help   help for pulsar-gateway
```

### Options inherited from parent commands

```
      --config file       config file (default is $HOME/.riff.yaml)
      --kubeconfig file   kubectl config file (default is $HOME/.kube/config)
      --no-color          disable color output in terminals
```

### SEE ALSO

* [riff streaming](riff_streaming.md)	 - (experimental) streaming runtime for riff functions
* [riff streaming pulsar-gateway create](riff_streaming_pulsar-gateway_create.md)	 - create a pulsar gateway of messages
* [riff streaming pulsar-gateway delete](riff_streaming_pulsar-gateway_delete.md)	 - delete pulsar gateway(s)
* [riff streaming pulsar-gateway list](riff_streaming_pulsar-gateway_list.md)	 - table listing of pulsar gateways
* [riff streaming pulsar-gateway status](riff_streaming_pulsar-gateway_status.md)	 - show pulsar gateway status

