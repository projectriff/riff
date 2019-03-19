## riff application delete

Delete existing applications

### Synopsis

Delete existing applications

```
riff application delete [flags]
```

### Examples

```
  riff application delete square --namespace joseph-ns
  riff application delete service-1 service-2
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

* [riff application](riff_application.md)	 - Interact with application related resources

