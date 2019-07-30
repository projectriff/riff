---
id: riff-service-create
title: "riff service create"
---
## riff service create

Create a new service resource

### Synopsis

Create a new service resource from a given image.

If `--env-from` is specified the source reference can be `configMapKeyRef` to select a key from a ConfigMap or `secretKeyRef` to select a key from a Secret. The following formats are supported:

    --env-from configMapKeyRef:{config-map-name}:{key-to-select}
    --env-from secretKeyRef:{secret-name}:{key-to-select}


```
riff service create [flags]
```

### Examples

```
  riff service create square --image acme/square:1.0 --namespace joseph-ns
  riff service create greeter --image acme/greeter:1.0 --env FOO=bar --env MESSAGE=Hello
  riff service create tweets-logger --image acme/tweets-logger:1.0.0
```

### Options

```
      --dry-run                don't create resources but print yaml representation on stdout
      --env stringArray        environment variable expressed in a 'key=value' format
      --env-from stringArray   environment variable created from a source reference; see command help for supported formats
  -h, --help                   help for create
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

