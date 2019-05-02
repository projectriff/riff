## riff credentials set

create or update secret and bind it to the riff service account

### Synopsis

create or update secret and bind it to the riff service account

```
riff credentials set [flags]
```

### Examples

```
  riff credentials set build-secret --namespace default --docker-hub johndoe
```

### Options

```
      --docker-hub string      Docker ID for authenticating with Docker Hub; password will be read from stdin
      --gcr string             path to a file containing Google Container Registry credentials
  -h, --help                   help for set
      --namespace namespace    the namespace of the credentials to be added
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

* [riff credentials](riff_credentials.md)	 - Interact with credentials related resources

