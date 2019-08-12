---
id: riff-doctor
title: "riff doctor"
---
## riff doctor

check riff's requirements are installed

### Synopsis

Check that riff is installed.

The doctor checks that necessary system components are installed and the user
has access to resources in the namespace.

The doctor is not a tool for monitoring the health of the cluster.

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
      --config file        config file (default is $HOME/.riff.yaml)
      --kube-config file   kubectl config file (default is $HOME/.kube/config)
      --no-color           disable color output in terminals
```

### SEE ALSO

* [riff](riff.md)	 - riff is for functions

