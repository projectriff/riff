---
id: riff-streaming-inmemory-gateway-create
title: "riff streaming inmemory-gateway create"
---
## riff streaming inmemory-gateway create

create an in-memory gateway of messages

### Synopsis

Creates an in-memory gateway within a namespace.

```
riff streaming inmemory-gateway create <name> [flags]
```

### Examples

```
riff streaming inmemory-gateway create my-inmemory-gateway
```

### Options

```
      --dry-run                 print kubernetes resources to stdout rather than apply them to the cluster, messages normally on stdout will be sent to stderr
  -h, --help                    help for create
  -n, --namespace name          kubernetes namespace (defaulted from kube config)
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

* [riff streaming inmemory-gateway](riff_streaming_inmemory-gateway.md)	 - (experimental) in-memory stream gateway

