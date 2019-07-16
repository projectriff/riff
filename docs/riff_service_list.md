---
id: riff-service-list
title: "riff service list"
---
## riff service list

List service resources

### Synopsis

List service resources

```
riff service list [flags]
```

### Examples

```
  riff service list
  riff service list --namespace joseph-ns
```

### Options

```
  -h, --help                  help for list
  -n, --namespace namespace   the namespace of the services to be listed
```

### Options inherited from parent commands

```
      --kubeconfig path   the path of a kubeconfig (default "~/.kube/config")
      --master address    the address of the Kubernetes API server; overrides any value in kubeconfig
```

### SEE ALSO

* [riff service](riff_service.md)	 - Interact with service related resources

