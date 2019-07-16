---
id: riff-stream-status
title: "riff stream status"
---
## riff stream status

show stream status

### Synopsis

Display status details for a stream.

The Ready condition is shown which should include a reason code and a
descriptive message when the status is not "True". The status for the condition
may be: "True", "False" or "Unknown". An "Unknown" status is common while the
stream is being processed.

```
riff stream status [flags]
```

### Examples

```
riff stream status my-stream
```

### Options

```
  -h, --help             help for status
  -n, --namespace name   kubernetes namespace (defaulted from kube config)
```

### Options inherited from parent commands

```
      --config file        config file (default is $HOME/.riff.yaml)
      --kube-config file   kubectl config file (default is $HOME/.kube/config)
      --no-color           disable color output in terminals
```

### SEE ALSO

* [riff stream](riff_stream.md)	 - streams of messages

