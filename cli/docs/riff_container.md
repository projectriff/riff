---
id: riff-container
title: "riff container"
---
## riff container

containers resolve the latest image

### Synopsis

Containers are a mechanism to resolve and detect the latest container image.

The container resource is only responsible for resolving the latest image. The
container image may then be deployed to core or knative runtime.

### Options

```
  -h, --help   help for container
```

### Options inherited from parent commands

```
      --config file       config file (default is $HOME/.riff.yaml)
      --kubeconfig file   kubectl config file (default is $HOME/.kube/config)
      --no-color          disable color output in terminals
```

### SEE ALSO

* [riff](riff.md)	 - riff is for functions
* [riff container create](riff_container_create.md)	 - watch for new images in a repository
* [riff container delete](riff_container_delete.md)	 - delete container(s)
* [riff container list](riff_container_list.md)	 - table listing of containers
* [riff container status](riff_container_status.md)	 - show container status

