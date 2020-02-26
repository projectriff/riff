---
id: riff-core-deployer-status
title: "riff core deployer status"
---
## riff core deployer status

show core deployer status

### Synopsis

Display status details for a deployer.

The Ready condition is shown which should include a reason code and a
descriptive message when the status is not "True". The status for the condition
may be: "True", "False" or "Unknown". An "Unknown" status is common while the
deployer roll out is processed.

```
riff core deployer status <name> [flags]
```

### Examples

```
riff core deployer status my-deployer
```

### Options

```
  -h, --help             help for status
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

