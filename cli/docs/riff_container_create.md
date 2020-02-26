---
id: riff-container-create
title: "riff container create"
---
## riff container create

watch for new images in a repository

### Synopsis

Create a container to watch for the latest image. There is no build performed
for containers.

```
riff container create <name> [flags]
```

### Examples

```
riff container create my-app --image registry.example.com/image
```

### Options

```
      --dry-run                 print kubernetes resources to stdout rather than apply them to the cluster, messages normally on stdout will be sent to stderr
  -h, --help                    help for create
      --image repository        repository where the built images are pushed (default "_")
  -n, --namespace name          kubernetes namespace (defaulted from kube config)
      --tail                    watch build logs
      --wait-timeout duration   duration to wait for the container to become ready when watching logs (default "10m")
```

### Options inherited from parent commands

```
      --config file       config file (default is $HOME/.riff.yaml)
      --kubeconfig file   kubectl config file (default is $HOME/.kube/config)
      --no-color          disable color output in terminals
```

### SEE ALSO

* [riff container](riff_container.md)	 - containers resolve the latest image

