---
id: riff-namespace-init
title: "riff namespace init"
---
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
      --docker-hub string      Docker ID for authenticating with Docker Hub; password will be read from stdin
      --gcr string             path to a file containing Google Container Registry credentials
  -h, --help                   help for init
      --image-prefix string    image prefix to use for commands that would otherwise require an --image argument. If not set, this value will be derived for Docker Hub and GCR
  -m, --manifest string        manifest of kubernetes configuration files to be applied; can be a named manifest (latest, nightly, stable) or a path of a manifest file (default "stable")
      --no-secret              no secret required for the image registry
      --registry string        registry server host, scheme must be "http" or "https" (default "https")
      --registry-user string   registry username; password will be read from stdin
  -s, --secret secret          the name of a secret containing credentials for the image registry (default "push-credentials")
```

### Options inherited from parent commands

```
      --kubeconfig path   the path of a kubeconfig (default "~/.kube/config")
      --master address    the address of the Kubernetes API server; overrides any value in kubeconfig
```

### SEE ALSO

* [riff namespace](riff_namespace.md)	 - Manage namespaces used for riff resources

