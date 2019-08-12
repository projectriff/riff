---
id: riff-knative-deployer-status
title: "riff knative deployer status"
---
## riff knative deployer status

show knative deployer status

### Synopsis

Display status details for a deployer.

The Ready condition is shown which should include a reason code and a
descriptive message when the status is not "True". The status for the condition
may be: "True", "False" or "Unknown". An "Unknown" status is common while the
deployer roll out is processed.

```
riff knative deployer status <name> [flags]
```

### Examples

```
riff knative deployer status my-deployer
```

### Options

```
  -h, --help             help for status
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

