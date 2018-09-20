## riff image load

Load and tag docker images

### Synopsis

Load the set of images identified by the provided image manifest into a docker daemon.

NOTE: This command requires the `docker` command line tool, as well as a (local) docker daemon.

SEE ALSO: To load, tag, and push images, use `riff image push`.

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
  -i, --images string   path of an image manifest of image names to be loaded
```

### SEE ALSO

* [riff image](riff_image.md)	 - Interact with docker images

