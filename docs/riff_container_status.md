---
id: riff-container-status
title: "riff container status"
---
## riff container status

show container status

### Synopsis

Display status details for a container.

The Ready condition is shown which should include a reason code and a
descriptive message when the status is not "True". The status for the condition
may be: "True", "False" or "Unknown". An "Unknown" status is common while the
container is processed or a build is in progress.

```
riff container status <name> [flags]
```

### Examples

```
riff container status my-container
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

* [riff container](riff_container.md)	 - containers resolve the latest image

