---
id: riff-service-status
title: "riff service status"
---
## riff service status

Display the status of a service

### Synopsis

Display the status of a service

```
riff service status [flags]
```

### Examples

```
  riff service status square --namespace joseph-ns
```

### Options

```
  -h, --help                  help for status
  -n, --namespace namespace   the namespace of the service
```

### Options inherited from parent commands

```
      --kubeconfig path   the path of a kubeconfig (default "~/.kube/config")
      --master address    the address of the Kubernetes API server; overrides any value in kubeconfig
```

### SEE ALSO

* [riff service](riff_service.md)	 - Interact with service related resources

