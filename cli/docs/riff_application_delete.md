---
id: riff-application-delete
title: "riff application delete"
---
## riff application delete

delete application(s)

### Synopsis

Delete one or more applications by name or all applications within a namespace.

Deleting an application prevents new builds while preserving built images in the
registry.

```
riff application delete <name(s)> [flags]
```

### Examples

```
riff application delete my-application
riff application delete --all
```

### Options

```
      --all              delete all applications within the namespace
  -h, --help             help for delete
  -n, --namespace name   kubernetes namespace (defaulted from kube config)
```

### Options inherited from parent commands

```
      --config file       config file (default is $HOME/.riff.yaml)
      --kubeconfig file   kubectl config file (default is $HOME/.kube/config)
      --no-color          disable color output in terminals
```

### SEE ALSO

* [riff application](riff_application.md)	 - applications built from source using application buildpacks

