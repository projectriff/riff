---
id: riff-knative-deployer-tail
title: "riff knative deployer tail"
---
## riff knative deployer tail

watch deployer logs

### Synopsis

Stream runtime logs for a deployer until canceled. To cancel, press Ctl-c in the
shell or kill the process.

As new deployer pods are started, the logs are displayed. To show historical logs
use --since.

```
riff knative deployer tail <name> [flags]
```

### Examples

```
riff knative deployer tail my-deployer
riff knative deployer tail my-deployer --since 1h
```

### Options

```
  -h, --help             help for tail
  -n, --namespace name   kubernetes namespace (defaulted from kube config)
      --since duration   time duration to start reading logs from
```

### Options inherited from parent commands

```
      --config file        config file (default is $HOME/.riff.yaml)
      --kube-config file   kubectl config file (default is $HOME/.kube/config)
      --no-color           disable color output in terminals
```

### SEE ALSO

* [riff knative deployer](riff_knative_deployer.md)	 - deployers map HTTP requests to a workload

