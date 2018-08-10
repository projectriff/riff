## riff system install

Install riff and Knative system components

### Synopsis

Install riff and Knative system components

If an 'istio-system' namespace isn't found then the it will be created and Istio components will be installed.

Use the '--node-port' flag when installing on Minikube and other clusters that don't support an external load balancer.'


```
riff system install [flags]
```

### Examples

```
  riff system install
```

### Options

```
      --force        force the install of components without getting any prompts
  -h, --help         help for install
      --latest       use the latest nightly build snapshot releases for Knative components
      --monitoring   install Prometheus and Grafana monitoring components
      --node-port    whether to use NodePort instead of LoadBalancer for ingress gateways
      --tracing      install Zipkin tracing components
```

### Options inherited from parent commands

```
      --kubeconfig path   the path of a kubeconfig (default "~/.kube/config")
      --master address    the address of the Kubernetes API server; overrides any value in kubeconfig
```

### SEE ALSO

* [riff system](riff_system.md)	 - Manage system related resources

