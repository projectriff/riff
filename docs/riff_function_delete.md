---
id: riff-function-delete
title: "riff function delete"
---
## riff function delete

delete function(s)

### Synopsis

Delete one or more functions by name or all functions within a namespace.

Deleting a function prevents new builds while preserving built images in the
registry.

```
riff function delete <name(s)> [flags]
```

### Examples

```
riff function delete my-function
riff function delete --all 
```

### Options

```
      --all              delete all functions within the namespace
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

* [riff function](riff_function.md)	 - functions built from source using function buildpacks

