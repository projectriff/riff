## riff system install

Install riff and Knative system components

### Synopsis

Install riff and Knative system components

If an 'istio-system' namespace isn't found, it will be created and Istio components will be installed.

Use the '--node-port' flag when installing on Minikube and other clusters that don't support an external load balancer.

Use the '--manifest' flag to specify the path of a manifest file which provides the URLs of the YAML definitions of the
components to be installed. The manifest file contents should be of the following form:

manifestVersion: 0.1
istio:
  - https://path/to/istio-release.yaml
knative:
  - https://path/to/serving-release.yaml
  - https://path/to/eventing-release.yaml
  - https://path/to/stub-bus-release.yaml


```
riff system install [flags]
```

### Examples

```
  riff system install
```

### Options

```
      --force             force the install of components without getting any prompts
  -h, --help              help for install
  -m, --manifest string   manifest of YAML files to be applied; can be a named manifest (stable or latest) or a file path of a manifest file (default "stable")
      --node-port         whether to use NodePort instead of LoadBalancer for ingress gateways
```

### Options inherited from parent commands

```
      --kubeconfig path   the path of a kubeconfig (default "~/.kube/config")
      --master address    the address of the Kubernetes API server; overrides any value in kubeconfig
```

### SEE ALSO

* [riff system](riff_system.md)	 - Manage system related resources

