---
id: riff-core-deployer
title: "riff core deployer"
---
## riff core deployer

deployers deploy a workload

### Synopsis

Deployers can be created for a build or an image. Build based deployers
continuously watch for the latest image and will deploy new images. If the
underlying build is deleted, the deployer will continue to run, but will no
longer self update. Image based deployers must be manually updated to trigger
roll out of an updated image.

Users wishing to perform checks on built images before deploying them can
provide their own external process to watch the build for new images and only
update the deployer image once those checks pass.

The service to access the deployer is available in the deployer listing.

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

* [riff core](riff_core.md)	 - core runtime for riff workloads
* [riff core deployer create](riff_core_deployer_create.md)	 - create a deployer to deploy a workload
* [riff core deployer delete](riff_core_deployer_delete.md)	 - delete deployer(s)
* [riff core deployer list](riff_core_deployer_list.md)	 - table listing of deployers
* [riff core deployer status](riff_core_deployer_status.md)	 - show core deployer status
* [riff core deployer tail](riff_core_deployer_tail.md)	 - watch deployer logs

