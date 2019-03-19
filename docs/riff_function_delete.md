## riff function delete

Delete existing functions

### Synopsis

Delete existing functions

```
riff function delete [flags]
```

### Examples

```
  riff function delete square --namespace joseph-ns
  riff function delete service-1 service-2
```

### Options

```
  -h, --help                  help for delete
  -n, --namespace namespace   the namespace of the service
```

### Options inherited from parent commands

```
      --kubeconfig path   the path of a kubeconfig (default "~/.kube/config")
      --master address    the address of the Kubernetes API server; overrides any value in kubeconfig
```

### SEE ALSO

* [riff function](riff_function.md)	 - Interact with function related resources

