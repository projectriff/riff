---
id: riff-knative-adapter-delete
title: "riff knative adapter delete"
---
## riff knative adapter delete

delete adapter(s)

### Synopsis

Delete one or more adapters by name or all adapters within a namespace.

```
riff knative adapter delete <name(s)> [flags]
```

### Examples

```
riff knative adapter delete my-adapter
riff knative adapter delete --all
```

### Options

```
      --all              delete all adapters within the namespace
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

* [riff knative adapter](riff_knative_adapter.md)	 - adapters push built images to Knative

