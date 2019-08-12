---
id: riff-knative-adapter-create
title: "riff knative adapter create"
---
## riff knative adapter create

create an adapter to Knative Serving

### Synopsis

Create a new adapter by watching a build for the latest image, pushing those
images to a target Knative Service or Configuration.

No new Knative resources are created directly by the adapter, it only updates
the image for an existing resource.

```
riff knative adapter create <name> [flags]
```

### Examples

```
riff knative adapter create my-adapter --application-ref my-app --service-ref my-kservice
```

### Options

```
      --application-ref name     name of application to deploy
      --configuration-ref name   name of Knative configuration to update
      --container-ref name       name of container to deploy
      --dry-run                  print kubernetes resources to stdout rather than apply them to the cluster, messages normally on stdout will be sent to stderr
      --function-ref name        name of function to deploy
  -h, --help                     help for create
  -n, --namespace name           kubernetes namespace (defaulted from kube config)
      --service-ref name         name of Knative service to update
      --tail                     watch adapter logs
      --wait-timeout duration    duration to wait for the adapter to become ready when watching logs (default "10m")
```

### Options inherited from parent commands

```
      --config file        config file (default is $HOME/.riff.yaml)
      --kube-config file   kubectl config file (default is $HOME/.kube/config)
      --no-color           disable color output in terminals
```

### SEE ALSO

* [riff knative adapter](riff_knative_adapter.md)	 - adapters push built images to Knative

