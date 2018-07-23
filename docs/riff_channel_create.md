## riff channel create

Create a new channel on a bus or a cluster bus

### Synopsis

Create a new channel on a bus or a cluster bus

```
riff channel create [flags]
```

### Examples

```
  riff channel create tweets --bus kafka --namespace steve-ns
  riff channel create orders --cluster-bus global-rabbit
```

### Options

```
      --bus name              the name of a bus for the channel
      --cluster-bus name      the name of a cluster bus for the channel
  -f, --force                 whether to force writing of files if they already exist
  -h, --help                  help for create
  -n, --namespace namespace   the namespace of resource names
  -w, --write                 whether to write yaml files for created resources
```

### Options inherited from parent commands

```
      --kubeconfig path   the path of a kubeconfig (default "~/.kube/config")
      --master address    the address of the Kubernetes API server; overrides any value in kubeconfig
```

### SEE ALSO

* [riff channel](riff_channel.md)	 - Interact with channel related resources

