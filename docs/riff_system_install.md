## riff system install

Install riff and Knative system components

### Synopsis

Install riff and Knative system components

If an 'istio-system' namespace isn't found, it will be created and Istio components will be installed.

Use the '--node-port' flag when installing on Minikube and other clusters that don't support an external load balancer.

Use the '--manifest' flag to specify the path of a manifest file which provides the URLs of the YAML definitions of the
components to be installed. The manifest file contents should be of the following form:

```yaml
manifestVersion: 0.1
istio:
- https://path/to/istio-release.yaml
knative:
- https://path/to/serving-release.yaml
- https://path/to/eventing-release.yaml
- https://path/to/stub-bus-release.yaml
namespace:
- https://path/to/riff-buildtemplate-release.yaml
```

To map Docker image names to images in a (private or public) registry, specify the registry hostname using the
'--registry' flag, the user owning the images using the '--registry-user' flag, and a complete list of the images to be
mapped using the '--images' flag. The '--images' flag contains the file path of an image manifest file with contents of
the following form:
```yaml
manifestVersion: 0.1
images:
...
- docker.io/istio/sidecar_injector
...
- gcr.io/knative-releases/github.com/knative/serving/cmd/autoscaler@sha256:76222399addc02454db9837ea3ff54bae29849168586051a9d0180daa2c1a805
...
```


```
riff system install [flags]
```

### Examples

```
  riff system install
```

### Options

```
      --force                  force the install of components without getting any prompts
  -h, --help                   help for install
      --images string          file path of an image manifest of images to be mapped
  -m, --manifest string        manifest of YAML files to be applied; can be a named manifest (stable or latest) or a file path of a manifest file (default "stable")
      --node-port              whether to use NodePort instead of LoadBalancer for ingress gateways
      --registry string        hostname of a Docker registry containing mapped images
      --registry-user string   user owning mapped images
```

### Options inherited from parent commands

```
      --kubeconfig path   the path of a kubeconfig (default "~/.kube/config")
      --master address    the address of the Kubernetes API server; overrides any value in kubeconfig
```

### SEE ALSO

* [riff system](riff_system.md)	 - Manage system related resources

