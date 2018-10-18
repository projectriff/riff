## riff-distro image list

List some or all of the images for a riff manifest

### Synopsis

Search a riff manifest and associated kubernetes configuration files for image names and create an image manifest listing the images.

It does not guarantee to find all referenced images and so the resultant image manifest needs to be validated, for example by manual inspection or testing.

NOTE: This command requires the `docker` command line tool to check the images.

```
riff-distro image list [flags]
```

### Examples

```
  riff-distro image list --manifest=path/to/manifest.yaml --images=path/for/image-manifest.yaml
```

### Options

```
      --force             overwrite the image manifest if it already exists
  -h, --help              help for list
  -i, --images string     path of the image manifest to be created; defaults to 'image-manifest.yaml' relative to the manifest
  -m, --manifest string   manifest to be searched; can be a named manifest (stable or latest) or a path of a manifest file (default "stable")
      --no-check          skips checking the images, thus not omitting the ones unknown to docker
```

### SEE ALSO

* [riff-distro image](riff-distro_image.md)	 - Interact with docker images

