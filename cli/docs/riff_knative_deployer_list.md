---
id: riff-knative-deployer-list
title: "riff knative deployer list"
---
## riff knative deployer list

table listing of deployers

### Synopsis

List deployers in a namespace or across all namespaces.

For detail regarding the status of a single deployer, run:

    riff knative deployer status <deployer-name>

```
riff knative deployer list [flags]
```

### Examples

```
riff knative deployer list
riff knative deployer list --all-namespaces
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

* [riff knative deployer](riff_knative_deployer.md)	 - deployers map HTTP requests to a workload

