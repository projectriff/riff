---
id: riff-knative-deployer
title: "riff knative deployer"
---
## riff knative deployer

deployers map HTTP requests to a workload

### Synopsis

Deployers can be created for a build reference or image. Build based deployers
continuously watch for the latest built image and will deploy new images. If the
underlying build resource is deleted, the deployer will continue to run, but will
no longer self update. Image based deployers must be manually updated to trigger
roll out of an updated image.

Users wishing to perform checks on built images before deploying them can
provide their own external process to watch the build resource for new images
and only update the deployer image once those checks pass.

The hostname to access the deployer is available in the deployer listing.

### Options

```
  -h, --help   help for deployer
```

### Options inherited from parent commands

```
      --config file        config file (default is $HOME/.riff.yaml)
      --kube-config file   kubectl config file (default is $HOME/.kube/config)
      --no-color           disable color output in terminals
```

### SEE ALSO

* [riff knative](riff_knative.md)	 - Knative runtime for riff workloads
* [riff knative deployer create](riff_knative_deployer_create.md)	 - create a deployer to map HTTP requests to a workload
* [riff knative deployer delete](riff_knative_deployer_delete.md)	 - delete deployer(s)
* [riff knative deployer list](riff_knative_deployer_list.md)	 - table listing of deployers
* [riff knative deployer status](riff_knative_deployer_status.md)	 - show knative deployer status
* [riff knative deployer tail](riff_knative_deployer_tail.md)	 - watch deployer logs

