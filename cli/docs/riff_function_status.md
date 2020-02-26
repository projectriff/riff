---
id: riff-function-status
title: "riff function status"
---
## riff function status

show function status

### Synopsis

Display status details for a function.

The Ready condition is shown which should include a reason code and a
descriptive message when the status is not "True". The status for the condition
may be: "True", "False" or "Unknown". An "Unknown" status is common while the
function roll out is processed or a build is in progress.

```
riff function status <name> [flags]
```

### Examples

```
riff function status my-function
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

* [riff function](riff_function.md)	 - functions built from source using function buildpacks

