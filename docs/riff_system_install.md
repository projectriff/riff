---
id: riff-system-install
title: "riff system install"
---
## riff system install

Install riff and Knative system components

### Synopsis

Install riff and Knative system components.

If an `istio-system` namespace isn't found, it will be created and Istio components will be installed. 
Use the `--node-port` flag when installing on Minikube and other clusters that don't support an external load balancer. 
Use the `--manifest` flag to specify the path or URL of a manifest file which provides the URLs of the kubernetes configuration files of the components to be installed. The manifest file contents should be of the following form:

    manifestVersion: 0.1
    istio:
    - https://path/to/istio-release.yaml
    knative:
    - https://path/to/build-release.yaml
    - https://path/to/serving-release.yaml
    - https://path/to/eventing-release.yaml
    namespace:
    - https://path/to/buildtemplate-release.yaml

Note: relative file paths or http/https URLs may be used in the manifest.

```
riff system install [flags]
```

### Options

```
      --force             force the install of components without getting any prompts
  -h, --help              help for install
  -m, --manifest string   manifest of kubernetes configuration files to be applied; can be a named manifest (latest, nightly, stable) or a path of a manifest file (default "stable")
      --node-port         whether to use NodePort instead of LoadBalancer for ingress gateways
```

### Options inherited from parent commands

```
      --kubeconfig path   the path of a kubeconfig (default "~/.kube/config")
      --master address    the address of the Kubernetes API server; overrides any value in kubeconfig
```

### SEE ALSO

* [riff system](riff_system.md)	 - Manage system related resources

