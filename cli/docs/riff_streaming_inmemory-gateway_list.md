---
id: riff-streaming-inmemory-gateway-list
title: "riff streaming inmemory-gateway list"
---
## riff streaming inmemory-gateway list

table listing of in-memory gateways

### Synopsis

List in-memory gateways in a namespace or across all namespaces.

For detail regarding the status of a single in-memory gateway, run:

    riff streaming inmemory-gateway status <inmemory-gateway-name>

```
riff streaming inmemory-gateway list [flags]
```

### Examples

```
riff streaming inmemory-gateway list
riff streaming inmemory-gateway list --all-namespaces
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

* [riff streaming inmemory-gateway](riff_streaming_inmemory-gateway.md)	 - (experimental) in-memory stream gateway

