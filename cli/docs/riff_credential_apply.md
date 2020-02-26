---
id: riff-credential-apply
title: "riff credential apply"
---
## riff credential apply

create or update credentials for a container registry

### Synopsis

Create or update credentials for a container registry.

In addition to creating a credential, the default image prefix can be set by
specifying --set-default-image-prefix. The prefix is applied to builds in order
to skip needing to specify a fully qualified image repository.

The default image prefix depends on the repository and take the form:
- Docker Hub: docker.io/<docker-user-name>
- GCR: gcr.io/<google-cloud-project-id>

Other image prefix values may be defined by specifying --default-image-prefix.

While multiple credentials can be created in a single namespace, only a single
default image prefix can be set.

```
riff credential apply <name> [flags]
```

### Examples

```
riff credential apply my-docker-hub-creds --docker-hub my-docker-id
riff credential apply my-docker-hub-creds --docker-hub my-docker-id --set-default-image-prefix
riff credential apply my-gcr-creds --gcr path/to/token.json
riff credential apply my-gcr-creds --gcr path/to/token.json --set-default-image-prefix
riff credential apply my-registry-creds --registry http://registry.example.com --registry-user my-username
riff credential apply my-registry-creds --registry http://registry.example.com --registry-user my-username --default-image-prefix registry.example.com/my-username
```

### Options

```
      --default-image-prefix repository   default repository prefix for built images, implies --set-default-image-prefix
      --docker-hub username               Docker Hub username, the password must be provided via stdin
      --dry-run                           print kubernetes resources to stdout rather than apply them to the cluster, messages normally on stdout will be sent to stderr
      --gcr file                          path to Google Container Registry service account token file
  -h, --help                              help for apply
  -n, --namespace name                    kubernetes namespace (defaulted from kube config)
      --registry url                      registry url
      --registry-user username            username for a registry, the password must be provided via stdin
      --set-default-image-prefix          use this registry as the default for built images
```

### Options inherited from parent commands

```
      --config file       config file (default is $HOME/.riff.yaml)
      --kubeconfig file   kubectl config file (default is $HOME/.kube/config)
      --no-color          disable color output in terminals
```

### SEE ALSO

* [riff credential](riff_credential.md)	 - credentials for container registries

