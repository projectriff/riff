---
id: riff-function-tail
title: "riff function tail"
---
## riff function tail

watch build logs

### Synopsis

Stream build logs for a function until canceled. To cancel, press Ctl-c in the
shell or kill the process.

As new builds are started, the logs are displayed. To show historical logs use
--since.

```
riff function tail <name> [flags]
```

### Examples

```
riff function tail my-function
riff function tail my-function --since 1h
```

### Options

```
  -h, --help             help for tail
  -n, --namespace name   kubernetes namespace (defaulted from kube config)
      --since duration   time duration to start reading logs from
```

### Options inherited from parent commands

```
      --config file       config file (default is $HOME/.riff.yaml)
      --kubeconfig file   kubectl config file (default is $HOME/.kube/config)
      --no-color          disable color output in terminals
```

### SEE ALSO

* [riff function](riff_function.md)	 - functions built from source using function buildpacks

