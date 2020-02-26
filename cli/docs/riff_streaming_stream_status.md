---
id: riff-streaming-stream-status
title: "riff streaming stream status"
---
## riff streaming stream status

show stream status

### Synopsis

Display status details for a stream.

The Ready condition is shown which should include a reason code and a
descriptive message when the status is not "True". The status for the condition
may be: "True", "False" or "Unknown". An "Unknown" status is common while the
stream rollout is being processed.

```
riff streaming stream status <name> [flags]
```

### Examples

```
riff streaming stream status my-stream
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

* [riff streaming stream](riff_streaming_stream.md)	 - (experimental) streams of messages

