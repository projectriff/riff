---
id: riff-core-deployer-list
title: "riff core deployer list"
---
## riff core deployer list

table listing of deployers

### Synopsis

List deployers in a namespace or across all namespaces.

For detail regarding the status of a single deployer, run:

    riff core deployer status <deployer-name>

```
riff core deployer list [flags]
```

### Examples

```
riff core deployer list
riff core deployer list --all-namespaces
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

* [riff core deployer](riff_core_deployer.md)	 - deployers deploy a workload

