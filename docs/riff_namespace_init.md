## riff namespace init

initialize riff resources in the namespace

### Synopsis

initialize riff resources in the namespace

```
riff namespace init [flags]
```

### Examples

```
  riff namespace init default --secret build-secret
```

### Options

```
  -h, --help              help for init
  -m, --manifest string   manifest of YAML files to be applied; can be a named manifest (stable or latest) or a file path of a manifest file (default "stable")
  -s, --secret secret     the name of a secret containing credentials for the image registry
```

### Options inherited from parent commands

```
      --kubeconfig path   the path of a kubeconfig (default "~/.kube/config")
      --master address    the address of the Kubernetes API server; overrides any value in kubeconfig
```

### SEE ALSO

* [riff namespace](riff_namespace.md)	 - Manage namespaces used for riff resources

