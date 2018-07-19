## riff service status

display the status of a service

### Synopsis

display the status of a service

```
riff service status [flags]
```

### Examples

```
  riff service status square --namespace joseph-ns
```

### Options

```
  -h, --help                  help for status
  -n, --namespace namespace   the namespace to use when interacting with resources.
```

### Options inherited from parent commands

```
      --kubeconfig path   path to a kubeconfig. (default "~/.kube/config")
      --master address    the address of the Kubernetes API server. Overrides any value in kubeconfig.
```

### SEE ALSO

* [riff service](riff_service.md)	 - interact with service related resources

