## riff service list

List service resources.

### Synopsis

List service resources.

```
riff service list [flags]
```

### Examples

```
  riff service list
  riff service list --namespace joseph-ns
```

### Options

```
  -h, --help                  help for list
  -n, --namespace namespace   the namespace to use when interacting with resources.
```

### Options inherited from parent commands

```
      --kubeconfig path   path to a kubeconfig. (default "~/.kube/config")
      --master address    the address of the Kubernetes API server. Overrides any value in kubeconfig.
```

### SEE ALSO

* [riff service](riff_service.md)	 - interact with service related resources

