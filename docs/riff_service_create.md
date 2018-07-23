## riff service create

Create a new service resource, with optional input binding

### Synopsis

Create a new service resource from a given image.
If an input channel and bus are specified, create the channel in the bus and subscribe the service to the channel.

```
riff service create [flags]
```

### Examples

```
  riff service create square --image acme/square:1.0 --namespace joseph-ns
  riff service create tweets-logger --image acme/tweets-logger:1.0.0 --input tweets --bus kafka
```

### Options

```
      --bus name              the name of the bus to create the channel in.
      --cluster-bus name      the name of the cluster bus to create the channel in.
      --dry-run               don't create resources but print yaml representation on stdout
  -h, --help                  help for create
      --image name[:tag]      the name[:tag] reference of an image containing the application/function
  -i, --input channel         name of the service's input channel, if any
  -n, --namespace namespace   the namespace of the service and any namespaced resources specified
```

### Options inherited from parent commands

```
      --kubeconfig path   the path of a kubeconfig (default "~/.kube/config")
      --master address    the address of the Kubernetes API server; overrides any value in kubeconfig
```

### SEE ALSO

* [riff service](riff_service.md)	 - Interact with service related resources

