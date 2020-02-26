---
id: riff-streaming-processor-tail
title: "riff streaming processor tail"
---
## riff streaming processor tail

watch processor logs

### Synopsis

Stream runtime logs for a processor until canceled. To cancel, press Ctl-c in
the shell or kill the process.

As new processor pods are started, the logs are displayed. To show historical
logs use --since.

```
riff streaming processor tail <name> [flags]
```

### Examples

```
riff streaming processor tail my-processor
riff streaming processor tail my-processor --since 1h
```

### Options

```
  -h, --help             help for tail
  -n, --namespace name   kubernetes namespace (defaulted from kube config)
```

### Options inherited from parent commands

```
      --config file       config file (default is $HOME/.riff.yaml)
      --kubeconfig file   kubectl config file (default is $HOME/.kube/config)
      --no-color          disable color output in terminals
```

### SEE ALSO

* [riff streaming processor](riff_streaming_processor.md)	 - (experimental) processors apply functions to messages on streams

