---
id: riff-core-deployer-delete
title: "riff core deployer delete"
---
## riff core deployer delete

delete deployer(s)

### Synopsis

Delete one or more deployers by name or all deployers within a namespace.

```
riff core deployer delete <name(s)> [flags]
```

### Examples

```
riff core deployer delete my-deployer
riff core deployer delete --all
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

* [riff core deployer](riff_core_deployer.md)	 - deployers deploy a workload

