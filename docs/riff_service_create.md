## riff service create

Create a new service resource, with optional input binding

### Synopsis

Create a new service resource, with optional input binding

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
      --bus name              the name of a bus for the channel
      --cluster-bus name      the name of a cluster bus for the channel
  -f, --force                 whether to force writing of files if they already exist.
  -h, --help                  help for create
      --image name[:tag]      the name[:tag] reference of an image containing the application/function
  -i, --input channel         name of the service's input channel, if any
  -n, --namespace namespace   the namespace of resource names
  -w, --write                 whether to write yaml files for created resources.
```

### Options inherited from parent commands

```
      --kubeconfig path   the path of a kubeconfig (default "~/.kube/config")
      --master address    the address of the Kubernetes API server; overrides any value in kubeconfig
```

### SEE ALSO

* [riff service](riff_service.md)	 - Interact with service related resources

