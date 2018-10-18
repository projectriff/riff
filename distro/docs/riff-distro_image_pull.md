## riff-distro image pull

Pull all docker images referenced in a distribution image-manifest and write them to disk

### Synopsis

Pull the set of images identified by the provided image manifest from remote registries, in preparation of an offline distribution tarball.

NOTE: This command requires the `docker` command line tool, as well as a (local) docker daemon and will load and tag the images using that daemon.

```
riff-distro image pull [flags]
```

### Examples

```
  riff-distro image pull --images=riff-distro-xx/image-manifest.yaml
```

### Options

```
  -c, --continue           whether to continue if an image doesn't have the same digest as stated in the image manifest; fail otherwise
  -h, --help               help for pull
  -i, --images string      path of an image manifest of image names to be pulled
  -o, --output directory   output directory for both the new manifest and images; defaults to rewriting the manifest in place with a sibling images/ directory
```

### SEE ALSO

* [riff-distro image](riff-distro_image.md)	 - Interact with docker images

