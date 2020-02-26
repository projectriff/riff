---
id: riff-knative-adapter
title: "riff knative adapter"
---
## riff knative adapter

adapters push built images to Knative

### Synopsis

The Knative runtime adapter updates a Knative Service or Configuration with the
latest image from a riff build. As the build produces new images, they will be
rolled out automatically to the target Knative resource.

No new Knative resources are created directly by the adapter, it only updates
the image for an existing resource.

### Options

```
  -h, --help   help for adapter
```

### Options inherited from parent commands

```
      --config file       config file (default is $HOME/.riff.yaml)
      --kubeconfig file   kubectl config file (default is $HOME/.kube/config)
      --no-color          disable color output in terminals
```

### SEE ALSO

* [riff knative](riff_knative.md)	 - Knative runtime for riff workloads
* [riff knative adapter create](riff_knative_adapter_create.md)	 - create an adapter to Knative Serving
* [riff knative adapter delete](riff_knative_adapter_delete.md)	 - delete adapter(s)
* [riff knative adapter list](riff_knative_adapter_list.md)	 - table listing of adapters
* [riff knative adapter status](riff_knative_adapter_status.md)	 - show knative adapter status

