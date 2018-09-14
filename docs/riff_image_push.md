## riff image push

Push (relocated) docker image names to another registry

### Synopsis

TODO

```
riff image push [flags]
```

### Examples

```
  riff image push --images=/path/to/image/manifest
```

### Options

```
  -h, --help            help for push
  -i, --images string   path of an image manifest of image names to be pushed
```

### Options inherited from parent commands

```
      --kubeconfig path   the path of a kubeconfig (default "~/.kube/config")
      --master address    the address of the Kubernetes API server; overrides any value in kubeconfig
```

### SEE ALSO

* [riff image](riff_image.md)	 - Interact with docker images

