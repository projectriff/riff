---
id: riff-knative-deployer-delete
title: "riff knative deployer delete"
---
## riff knative deployer delete

delete deployer(s)

### Synopsis

Delete one or more deployers by name or all deployers within a namespace.

New HTTP requests addressed to the deployer will fail. A new deployer created with
the same name will start to receive new HTTP requests addressed to the same
deployer.

```
riff knative deployer delete <name(s)> [flags]
```

### Examples

```
riff knative deployer delete my-deployer
riff knative deployer delete --all
```

### Options

```
      --all              delete all deployers within the namespace
  -h, --help             help for delete
  -n, --namespace name   kubernetes namespace (defaulted from kube config)
```

### Options inherited from parent commands

```
      --config file        config file (default is $HOME/.riff.yaml)
      --kube-config file   kubectl config file (default is $HOME/.kube/config)
      --no-color           disable color output in terminals
```

### SEE ALSO

* [riff knative deployer](riff_knative_deployer.md)	 - deployers map HTTP requests to a workload

