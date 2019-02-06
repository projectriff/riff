## riff system install

Install riff and Knative system components

### Synopsis

Install riff and Knative system components.

If an `istio-system` namespace isn't found, it will be created and Istio components will be installed. 
Use the `--node-port` flag when installing on Minikube and other clusters that don't support an external load balancer. 
Use the `--manifest` flag to specify the path or URL of a manifest file which provides the URLs of the kubernetes configuration files of the components to be installed.

```
riff system install [flags]
```

### Options

```
      --force             force the install of components without getting any prompts
  -h, --help              help for install
  -m, --manifest string   manifest of kubernetes configuration files to be applied; can be a named manifest (stable) or a path of a manifest file (default "stable")
      --node-port         whether to use NodePort instead of LoadBalancer for ingress gateways
```

### Options inherited from parent commands

```
      --kubeconfig path   the path of a kubeconfig (default "~/.kube/config")
      --master address    the address of the Kubernetes API server; overrides any value in kubeconfig
```

### SEE ALSO

* [riff system](riff_system.md)	 - Manage system related resources

