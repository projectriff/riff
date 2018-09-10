## riff service create

Create a new service resource, with optional input binding

### Synopsis

Create a new service resource from a given image.

If an input channel and bus are specified, create the channel in the bus and subscribe the service to the channel.

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
  riff service create tweets-logger --image acme/tweets-logger:1.0.0 --input tweets --bus kafka
```

### Options

```
      --bus name               the name of the bus to create the channel in.
      --cluster-bus name       the name of the cluster bus to create the channel in.
      --dry-run                don't create resources but print yaml representation on stdout
      --env stringArray        environment variable expressed in a 'key=value' format
      --env-from stringArray   environment variable created from a source reference; see command help for supported formats
  -h, --help                   help for create
      --image name[:tag]       the name[:tag] reference of an image containing the application/function
  -i, --input channel          name of the service's input channel, if any
  -n, --namespace namespace    the namespace of the service and any namespaced resources specified
  -o, --output channel         name of the service's output channel, if any
```

### Options inherited from parent commands

```
      --kubeconfig path   the path of a kubeconfig (default "~/.kube/config")
      --master address    the address of the Kubernetes API server; overrides any value in kubeconfig
```

### SEE ALSO

* [riff service](riff_service.md)	 - Interact with service related resources

