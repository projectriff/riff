## riff image relocate

Relocate docker image names to another registry

### Synopsis

Relocate either a single YAML file or a riff manifest and its YAML files so that image names refer to
another (private or public) registry.

To transform a single YAML file, use the '--yaml' flag to specify the path or URL of the YAML file. Use the '--output'
flag to specify the path to contain the transformed YAML file. If '--output' is an existing directory, the YAML file
will be created in that directory. Otherwise the YAML file will be written to the path specified in '--output'.

To transform a manifest, use the '--manifest' flag to specify the path of a manifest file which provides the paths or
URLs of the YAML definitions of riff components. Use the '--output' flag to specify the path of a directory to contain
the transformed manifest and YAML definitions.

Specify the registry hostname using the '--registry' flag, the user owning the images using the '--registry-user' flag,
and a complete list of the images to be mapped using the '--images' flag. The '--images' flag contains the path of animage manifest file with contents of the following form:

    manifestVersion: 0.1
    images:
    ...
    - istio/sidecar_injector
    ...
    - gcr.io/knative-releases/github.com/knative/serving/cmd/autoscaler@sha256:76222399addc02454db9837ea3ff54bae29849168586051a9d0180daa2c1a805
    ...
    


```
riff image relocate [flags]
```

### Examples

```
  riff image relocate --manifest=/path/to/manifest --registry=hostname --user=username --images=/path/to/image/manifest --output=/path/to/output/dir
  riff image relocate --yaml=/path/to/yaml/file --registry=hostname --user=username --images=/path/to/image/manifest --output=/path/to/output
```

### Options

```
  -h, --help                   help for relocate
  -i, --images string          file path of an image manifest of image names to be mapped
  -m, --manifest string        file path of a riff manifest file
  -o, --output string          file path to contain the output file(s)
  -r, --registry string        hostname for mapped images
  -u, --registry-user string   user name for mapped images
  -y, --yaml string            file path of a YAML file
```

### Options inherited from parent commands

```
      --kubeconfig path   the path of a kubeconfig (default "~/.kube/config")
      --master address    the address of the Kubernetes API server; overrides any value in kubeconfig
```

### SEE ALSO

* [riff image](riff_image.md)	 - Interact with docker images

