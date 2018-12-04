## riff channel create

Create a new channel

### Synopsis

Create a new channel

```
riff channel create [flags]
```

### Examples

```
  riff channel create tweets --cluster-provisioner kafka --namespace steve-ns
  riff channel create orders --cluster-provisioner global-rabbit
```

### Options

```
      --cluster-provisioner name   the name of the cluster channel provisioner to provision the channel with.
      --dry-run                    don't create resources but print yaml representation on stdout
  -h, --help                       help for create
  -n, --namespace namespace        the namespace of the channel
```

### Options inherited from parent commands

```
      --kubeconfig path   the path of a kubeconfig (default "~/.kube/config")
      --master address    the address of the Kubernetes API server; overrides any value in kubeconfig
```

### SEE ALSO

* [riff channel](riff_channel.md)	 - Interact with channel related resources

