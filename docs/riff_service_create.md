## riff service create

create a new service resource, with optional input binding

### Synopsis

create a new service resource, with optional input binding

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
  -f, --force                 force writing of files if they already exist.
  -h, --help                  help for create
      --image name[:tag]      reference to an already built name[:tag] image that contains the application/function.
  -i, --input channel         name of the input channel to subscribe the service to.
  -n, --namespace namespace   the namespace to use when interacting with resources.
  -w, --write                 whether to write yaml files for created resources.
```

### Options inherited from parent commands

```
      --kubeconfig path   path to a kubeconfig. (default "~/.kube/config")
      --master address    the address of the Kubernetes API server. Overrides any value in kubeconfig.
```

### SEE ALSO

* [riff service](riff_service.md)	 - interact with service related resources

