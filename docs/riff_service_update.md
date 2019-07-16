---
id: riff-service-update
title: "riff service update"
---
## riff service update

Create a new revision for a service, with updated attributes

### Synopsis

Create a new revision for a service, updating the application/function image and/or environment.

```
riff service update [flags]
```

### Examples

```
  riff service update square --image acme/square:1.1 --namespace joseph-ns
```

### Options

```
      --dry-run                don't create resources but print yaml representation on stdout
      --env stringArray        environment variable expressed in a 'key=value' format
      --env-from stringArray   environment variable created from a source reference; see command help for supported formats
  -h, --help                   help for update
      --image name[:tag]       the name[:tag] reference of an image containing the application/function
  -n, --namespace namespace    the namespace of the service
```

### Options inherited from parent commands

```
      --kubeconfig path   the path of a kubeconfig (default "~/.kube/config")
      --master address    the address of the Kubernetes API server; overrides any value in kubeconfig
```

### SEE ALSO

* [riff service](riff_service.md)	 - Interact with service related resources

