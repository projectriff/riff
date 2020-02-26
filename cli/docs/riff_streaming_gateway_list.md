---
id: riff-streaming-gateway-list
title: "riff streaming gateway list"
---
## riff streaming gateway list

table listing of gateways

### Synopsis

List gateways in a namespace or across all namespaces.

For detail regarding the status of a single gateway, run:

    riff streaming gateway status <gateway-name>

```
riff streaming gateway list [flags]
```

### Examples

```
riff streaming gateway list
riff streaming gateway list --all-namespaces
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

* [riff streaming gateway](riff_streaming_gateway.md)	 - (experimental) stream gateway

