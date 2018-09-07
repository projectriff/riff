## riff service unsubscribe

Unsubscribe a service from an existing subscription

### Synopsis

Unsubscribe a service from an existing subscription

```
riff service unsubscribe [flags]
```

### Examples

```
  riff service unsubscribe subscription --namespace joseph-ns
```

### Options

```
  -h, --help               help for unsubscribe
  -n, --namespace string   the namespace of the subscription
```

### Options inherited from parent commands

```
      --kubeconfig path   the path of a kubeconfig (default "~/.kube/config")
      --master address    the address of the Kubernetes API server; overrides any value in kubeconfig
```

### SEE ALSO

* [riff service](riff_service.md)	 - Interact with service related resources

