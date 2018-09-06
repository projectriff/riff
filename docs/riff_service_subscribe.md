## riff service subscribe

Subscribe a service to an existing input or output channel

### Synopsis

Subscribe a service to an existing input or output channel

```
riff service subscribe [flags]
```

### Examples

```
  riff service subscribe square --input numbers --namespace joseph-ns
```

### Options

```
      --dry-run               don't create resources but print yaml representation on stdout
  -h, --help                  help for subscribe
  -i, --input channel         the name of an input channel for the service
  -n, --namespace namespace   the namespace of the subscription, channel, and service
  -o, --output channel        the name of an output channel for the service
      --subscription name     name of the subscription (default SERVICE_NAME)
```

### Options inherited from parent commands

```
      --kubeconfig path   the path of a kubeconfig (default "~/.kube/config")
      --master address    the address of the Kubernetes API server; overrides any value in kubeconfig
```

### SEE ALSO

* [riff service](riff_service.md)	 - Interact with service related resources

