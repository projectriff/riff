---
id: riff-container-list
title: "riff container list"
---
## riff container list

table listing of containers

### Synopsis

List containers in a namespace or across all namespaces.

For detail regarding the status of a single container, run:

    riff container status <container-name>

```
riff container list [flags]
```

### Examples

```
riff container list
riff container list --all-namespaces
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

* [riff container](riff_container.md)	 - containers resolve the latest image

