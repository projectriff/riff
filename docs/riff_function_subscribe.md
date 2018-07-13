## riff function subscribe

subscribe a function to an existing input channel

### Synopsis

subscribe a function to an existing input channel

```
riff function subscribe [flags]
```

### Examples

```
  riff function subscribe square --input numbers --namespace joseph-ns
```

### Options

```
  -h, --help                  help for subscribe
  -i, --input channel         name of the input channel to subscribe the function to.
  -n, --namespace namespace   the namespace to create resources in.
```

### Options inherited from parent commands

```
      --kubeconfig path   path to a kubeconfig. (default "~/.kube/config")
      --master address    the address of the Kubernetes API server. Overrides any value in kubeconfig.
```

### SEE ALSO

* [riff function](riff_function.md)	 - interact with function related resources

