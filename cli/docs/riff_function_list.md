---
id: riff-function-list
title: "riff function list"
---
## riff function list

table listing of functions

### Synopsis

List functions in a namespace or across all namespaces.

For detail regarding the status of a single function, run:

    riff function status <function-name>

```
riff function list [flags]
```

### Examples

```
riff function list
riff function list --all-namespaces
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

* [riff function](riff_function.md)	 - functions built from source using function buildpacks

