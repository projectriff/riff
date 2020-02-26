---
id: riff-doctor
title: "riff doctor"
---
## riff doctor

check riff's permissions

### Synopsis

The doctor checks that the current user has permission to access riff, and riff
related, resources in a namespace.

The doctor is not a tool for monitoring the health of the cluster or the riff
install.

```
riff doctor [flags]
```

### Examples

```
riff doctor
```

### Options

```
  -h, --help             help for doctor
  -n, --namespace name   kubernetes namespace (defaulted from kube config)
```

### Options inherited from parent commands

```
      --config file       config file (default is $HOME/.riff.yaml)
      --kubeconfig file   kubectl config file (default is $HOME/.kube/config)
      --no-color          disable color output in terminals
```

### SEE ALSO

* [riff](riff.md)	 - riff is for functions

