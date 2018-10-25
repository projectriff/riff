## riff image relocate

Relocate docker image names to another registry

### Synopsis

Relocate either a single kubernetes configuration file or a riff manifest, its kubernetes configuration files, and an image manifest, so that image names refer to another (private or public) registry.

To relocate a single kubernetes configuration file, use the `--file` flag to specify the path or URL of the file. Use the `--output` flag to specify the path for the relocated file. If `--output` is an existing directory, the relocated file will be placed in that directory. Otherwise the relocated file will be written to the path specified in `--output`.

To relocate a manifest, use the `--manifest` flag to specify the path or URL of a manifest file which provides the paths or URLs of the kubernetes configuration files for riff components. Use the `--output` flag to specify the path of a directory to contain the relocated manifest, kubernetes configuration files, and image manifest. Any associated images are copied to the output directory.

Specify the registry hostname using the `--registry` flag, the user owning the images using the `--registry-user` flag, and the images to be mapped using the `--images` flag. The `--images` flag contains the path of an image manifest file, mapping image names to image ids, of the following form:

    manifestVersion: 0.1
    images:
    ...
      "docker.io/istio/proxyv2:1.0.1": "sha256:7ae1462913665ac77389087f43d3d3dda86b4a0883b1cafcd105a4eeb648498f"
    ...
      "gcr.io/knative-releases/github.com/knative/serving/cmd/autoscaler@sha256:76222399addc02454db9837ea3ff54bae29849168586051a9d0180daa2c1a805": "sha256:76222399addc02454db9837ea3ff54bae29849168586051a9d0180daa2c1a805"
    ... 



```
riff image relocate [flags]
```

### Examples

```
  riff image relocate --manifest=path/to/manifest.yaml --registry=hostname --registry-user=username --images=path/to/image-manifest.yaml --output=path/to/output/dir
  riff image relocate --file=path/to/file --registry=hostname --registry-user=username --images=path/to/image-manifest.yaml --output=path/to/output
```

### Options

```
  -f, --file string            path of a kubernetes configuration file
  -h, --help                   help for relocate
  -i, --images string          path of an image manifest
  -m, --manifest string        path of a riff manifest (default "manifest.yaml")
  -o, --output string          path to contain the output file(s)
  -r, --registry string        hostname for mapped images
  -u, --registry-user string   user name for mapped images
```

### SEE ALSO

* [riff image](riff_image.md)	 - Interact with docker images

