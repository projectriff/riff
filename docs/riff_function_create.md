## riff function create

create a new function resource, with optional input binding

### Synopsis

create a new function resource, with optional input binding

```
riff function create [flags]
```

### Examples

```
  riff function create node square --image acme/square:1.0 --namespace joseph-ns
  riff function create java tweets-logger --image acme/tweets-logger:1.0.0 --input tweets --bus kafka
```

### Options

```
      --build string          TODO: build options?
      --bus name              the name of the bus to create the channel in.
      --cluster-bus name      the name of the cluster bus to create the channel in.
  -h, --help                  help for create
      --image name[:tag]      reference to an already built name[:tag] image that contains the function.
  -i, --input channel         name of the input channel to subscribe the function to.
  -n, --namespace namespace   the namespace to create resources in.
```

### Options inherited from parent commands

```
      --kubeconfig path   path to a kubeconfig. (default "~/.kube/config")
      --master address    the address of the Kubernetes API server. Overrides any value in kubeconfig.
```

### SEE ALSO

* [riff function](riff_function.md)	 - interact with function related resources

