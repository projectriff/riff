## riff image load

Load and tag docker images

### Synopsis

Load the images in an image manifest into a docker daemon and tag them.

For details of image manifests, see `riff image relocate -h`.

NOTE: This command requires the `docker` command line tool, as well as a docker daemon.

SEE ALSO: To load, tag, and push images to a registry, use `riff image push`.

```
riff image load [flags]
```

### Examples

```
  riff image load --images=riff-distro-xx/image-manifest.yaml
```

### Options

```
  -h, --help            help for load
  -i, --images string   path of an image manifest
```

### SEE ALSO

* [riff image](riff_image.md)	 - Interact with docker images

