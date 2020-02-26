---
id: riff-streaming-pulsar-gateway-create
title: "riff streaming pulsar-gateway create"
---
## riff streaming pulsar-gateway create

create a pulsar gateway of messages

### Synopsis

Creates a Pulsar gateway within a namespace.

The gateway is configured with a Pulsar service URL.

```
riff streaming pulsar-gateway create <name> [flags]
```

### Examples

```
riff streaming pulsar-gateway create my-pulsar-gateway --service-url pulsar://localhost:6650
```

### Options

```
      --dry-run                 print kubernetes resources to stdout rather than apply them to the cluster, messages normally on stdout will be sent to stderr
  -h, --help                    help for create
  -n, --namespace name          kubernetes namespace (defaulted from kube config)
      --service-url url         url of the pulsar service
      --tail                    watch creation progress
      --wait-timeout duration   duration to wait for the gateway to become ready when watching progress (default 1m0s)
```

### Options inherited from parent commands

```
      --config file       config file (default is $HOME/.riff.yaml)
      --kubeconfig file   kubectl config file (default is $HOME/.kube/config)
      --no-color          disable color output in terminals
```

### SEE ALSO

* [riff streaming pulsar-gateway](riff_streaming_pulsar-gateway.md)	 - (experimental) pulsar stream gateway

