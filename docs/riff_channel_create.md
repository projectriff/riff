## riff channel create

create a new channel on a namespace or cluster bus

### Synopsis

create a new channel on a namespace or cluster bus

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
  -h, --help                  help for create
  -n, --namespace namespace   the namespace to create resources in.
```

### Options inherited from parent commands

```
      --kubeconfig path   path to a kubeconfig. (default "~/.kube/config")
      --master address    the address of the Kubernetes API server. Overrides any value in kubeconfig.
```

### SEE ALSO

* [riff channel](riff_channel.md)	 - interact with channel related resources

