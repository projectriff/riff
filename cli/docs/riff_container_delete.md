---
id: riff-container-delete
title: "riff container delete"
---
## riff container delete

delete container(s)

### Synopsis

Delete one or more containers by name or all containers within a namespace.

Deleting a container prevents resolution of new images.

```
riff container delete <name(s)> [flags]
```

### Examples

```
riff container delete my-container
riff container delete --all
```

### Options

```
      --all              delete all containers within the namespace
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

* [riff container](riff_container.md)	 - containers resolve the latest image

