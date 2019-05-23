## riff credential apply

create or update credentials for a container registry

### Synopsis


<todo>


```
riff credential apply [flags]
```

### Examples

```
riff credential apply my-docker-hub-creds --docker-hub my-docker-id
riff credential apply my-gcr-creds --gcr path/to/token.json
riff credential apply my-registry-creds --registry http://registry.example.com --registry-user my-username
```

### Options

```
      --docker-hub username      Docker Hub username, the password must be provided via stdin
      --gcr file                 path to Google Container Registry service account token file
  -h, --help                     help for apply
  -n, --namespace name           kubernetes namespace (defaulted from kube config)
      --registry url             registry url
      --registry-user username   username for a registry, the password must be provided via stdin
```

### Options inherited from parent commands

```
      --config file        config file (default is $HOME/.riff.yaml)
      --kube-config file   kubectl config file (default is $HOME/.kube/config)
      --no-color           disable color output in terminals
```

### SEE ALSO

* [riff credential](riff_credential.md)	 - credentials for container registries

