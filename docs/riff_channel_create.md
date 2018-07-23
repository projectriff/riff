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
      --bus name              the name of the bus to create the channel in.
      --cluster-bus name      the name of the cluster bus to create the channel in.
      --dry-run               don't create resources but print yaml representation on stdout
  -h, --help                  help for create
  -n, --namespace namespace   the namespace of the channel and any non-cluster bus
```

### Options inherited from parent commands

```
      --kubeconfig path   the path of a kubeconfig (default "~/.kube/config")
      --master address    the address of the Kubernetes API server; overrides any value in kubeconfig
```

### SEE ALSO

* [riff channel](riff_channel.md)	 - Interact with channel related resources

