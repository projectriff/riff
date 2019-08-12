---
id: riff-application-tail
title: "riff application tail"
---
## riff application tail

watch build logs

### Synopsis

Stream build logs for an application until canceled. To cancel, press Ctl-c in
the shell or kill the process.

As new builds are started, the logs are displayed. To show historical logs use
--since.

```
riff application tail <name> [flags]
```

### Examples

```
riff application tail my-application
riff application tail my-application --since 1h
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

* [riff application](riff_application.md)	 - applications built from source using application buildpacks

