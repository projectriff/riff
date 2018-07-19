## riff function status

display the status of a function

### Synopsis

display the status conditions of a function's service

```
riff function status [flags]
```

### Examples

```
  riff function status square --namespace joseph-ns
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

* [riff function](riff_function.md)	 - interact with function related resources

