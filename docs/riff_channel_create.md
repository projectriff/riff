---
id: riff-channel-create
title: "riff channel create"
---
## riff channel create

[DEPRECATED] Create a new channel

### Synopsis

[DEPRECATED] Create a new channel

```
riff channel create [flags]
```

### Examples

```
  riff channel create tweets --cluster-provisioner kafka --namespace steve-ns
  riff channel create orders
```

### Options

```
      --cluster-provisioner name   the name of the cluster channel provisioner to provision the channel with. Uses the cluster's default provisioner if not specified.
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

* [riff channel](riff_channel.md)	 - [DEPRECATED] Interact with channel related resources

