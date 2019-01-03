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
      --dockerhub string   dockerhub username for authentication; password will be read from stdin
      --gcr string         path to a file containing Google Container Registry credentials
  -h, --help               help for init
  -m, --manifest string    manifest of kubernetes configuration files to be applied; can be a named manifest (latest, stable) or a path of a manifest file (default "stable")
      --no-secret          no secret required for the image registry
  -s, --secret secret      the name of a secret containing credentials for the image registry (default "push-credentials")
```

### Options inherited from parent commands

```
      --kubeconfig path   the path of a kubeconfig (default "~/.kube/config")
      --master address    the address of the Kubernetes API server; overrides any value in kubeconfig
```

### SEE ALSO

* [riff namespace](riff_namespace.md)	 - Manage namespaces used for riff resources

