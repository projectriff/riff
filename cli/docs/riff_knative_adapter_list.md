---
id: riff-knative-adapter-list
title: "riff knative adapter list"
---
## riff knative adapter list

table listing of adapters

### Synopsis

List adapters in a namespace or across all namespaces.

For detail regarding the status of a single adapter, run:

    riff knative adapter status <adapter-name>

```
riff knative adapter list [flags]
```

### Examples

```
riff knative adapter list
riff knative adapter list --all-namespaces
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

* [riff knative adapter](riff_knative_adapter.md)	 - adapters push built images to Knative

