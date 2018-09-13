## riff image push

Push (relocated) docker image names to an image registry

### Synopsis

Push the set of images identified by the provided image manifest into a remote registry, for later consumption by `riff system install`.

NOTE: This command requires the `docker` command line tool, as well as a (local) docker daemon and will load and tag the images using that daemon.

```
riff image push [flags]
```

### Examples

```
  riff image push --images=riff-distro-xx/image-manifest.yaml
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

