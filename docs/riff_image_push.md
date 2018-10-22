## riff image push

Push docker images to a registry

### Synopsis

Load, tag, and push the images in an image manifest to a registry, for later consumption by `riff system install`.

For details of image manifests, see `riff image relocate -h`.

NOTE: This command requires the `docker` command line tool, as well as a docker daemon.

SEE ALSO: To load and tag images, but not push them, use `riff image load`.

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
  -i, --images string   path of an image manifest
```

### SEE ALSO

* [riff image](riff_image.md)	 - Interact with docker images

