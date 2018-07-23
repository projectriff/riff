## riff service subscribe

subscribe a service to an existing input channel

### Synopsis

subscribe a service to an existing input channel

```
riff service subscribe [flags]
```

### Examples

```
  riff service subscribe square --input numbers --namespace joseph-ns
```

### Options

```
  -h, --help                  help for subscribe
  -i, --input channel         name of the input channel to subscribe the service to.
  -n, --namespace namespace   the namespace to use when interacting with resources.
      --subscription string   name of the subscription (default SERVICE_NAME)
```

### Options inherited from parent commands

```
      --kubeconfig path   path to a kubeconfig. (default "~/.kube/config")
      --master address    the address of the Kubernetes API server. Overrides any value in kubeconfig.
```

### SEE ALSO

* [riff service](riff_service.md)	 - interact with service related resources

