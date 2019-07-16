---
id: riff-service-delete
title: "riff service delete"
---
## riff service delete

Delete existing services

### Synopsis

Delete existing services

```
riff service delete [flags]
```

### Examples

```
  riff service delete square --namespace joseph-ns
  riff service delete service-1 service-2
```

### Options

```
  -h, --help                  help for delete
  -n, --namespace namespace   the namespace of the service
```

### Options inherited from parent commands

```
      --kubeconfig path   the path of a kubeconfig (default "~/.kube/config")
      --master address    the address of the Kubernetes API server; overrides any value in kubeconfig
```

### SEE ALSO

* [riff service](riff_service.md)	 - Interact with service related resources

