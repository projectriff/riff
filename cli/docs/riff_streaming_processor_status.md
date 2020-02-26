---
id: riff-streaming-processor-status
title: "riff streaming processor status"
---
## riff streaming processor status

show processor status

### Synopsis

Display status details for a processor.

The Ready condition is shown which should include a reason code and a
descriptive message when the status is not "True". The status for the condition
may be: "True", "False" or "Unknown". An "Unknown" status is common while the
processor rollout is processed.

```
riff streaming processor status <name> [flags]
```

### Examples

```
riff streaming processor status my-processor
```

### Options

```
  -h, --help             help for status
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

