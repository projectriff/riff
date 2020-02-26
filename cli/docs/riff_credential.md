---
id: riff-credential
title: "riff credential"
---
## riff credential

credentials for container registries

### Synopsis

Credentials allow builds to push images to authenticated registries. If the
registry allows unauthenticated image pushes, credentials are not required
(while useful for local development environments, this is not recommended).

Credentials are defined by a hostname, username and password. These values are
specified explicitly or via shortcuts for Docker Hub and Google Container
Registry (GCR).

The credentials are saved as Kubernetes secrets and exposed to build pods.

To manage credentials, read and write access to Secrets is required for the
namespace. To manage the default image prefix, read and write access to the
'riff-build' ConfigMap is required for the namespace.

### Options

```
  -h, --help   help for credential
```

### Options inherited from parent commands

```
      --config file       config file (default is $HOME/.riff.yaml)
      --kubeconfig file   kubectl config file (default is $HOME/.kube/config)
      --no-color          disable color output in terminals
```

### SEE ALSO

* [riff](riff.md)	 - riff is for functions
* [riff credential apply](riff_credential_apply.md)	 - create or update credentials for a container registry
* [riff credential delete](riff_credential_delete.md)	 - delete credential(s)
* [riff credential list](riff_credential_list.md)	 - table listing of credentials

