## riff credentials delete

Delete specified credentials

### Synopsis

Delete specified credentials

```
riff credentials delete [flags]
```

### Examples

```
  riff credentials delete secret1 secret2
  riff credentials delete --namespace joseph-ns secret
```

### Options

```
  -h, --help                  help for delete
  -n, --namespace namespace   the namespace of the credentials to be deleted
```

### Options inherited from parent commands

```
      --kubeconfig path   the path of a kubeconfig (default "~/.kube/config")
      --master address    the address of the Kubernetes API server; overrides any value in kubeconfig
```

### SEE ALSO

* [riff credentials](riff_credentials.md)	 - Interact with credentials related resources

