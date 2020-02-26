---
id: riff-knative
title: "riff knative"
---
## riff knative

Knative runtime for riff workloads

### Synopsis

The Knative runtime uses Knative Configuration and Route resources to deploy
a workload. Knative provides both a zero-to-n autoscaler and managed ingress.

### Options

```
  -h, --help   help for knative
```

### Options inherited from parent commands

```
      --config file       config file (default is $HOME/.riff.yaml)
      --kubeconfig file   kubectl config file (default is $HOME/.kube/config)
      --no-color          disable color output in terminals
```

### SEE ALSO

* [riff](riff.md)	 - riff is for functions
* [riff knative adapter](riff_knative_adapter.md)	 - adapters push built images to Knative
* [riff knative deployer](riff_knative_deployer.md)	 - deployers map HTTP requests to a workload

