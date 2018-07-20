## riff service invoke

Invoke a service.

### Synopsis

Invoke a service by shelling out to curl.

The curl command is printed so it can be copied and extended.

Additional curl arguments and flags may be specified after a double dash (--).

```
riff service invoke [flags]
```

### Examples

```
  riff service invoke square --namespace joseph-ns
  riff service invoke square -- --include
```

### Options

```
  -h, --help                  help for invoke
  -n, --namespace namespace   the namespace to use when interacting with resources.
```

### Options inherited from parent commands

```
      --kubeconfig path   path to a kubeconfig. (default "~/.kube/config")
      --master address    the address of the Kubernetes API server. Overrides any value in kubeconfig.
```

### SEE ALSO

* [riff service](riff_service.md)	 - interact with service related resources

