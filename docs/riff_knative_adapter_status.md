---
id: riff-knative-adapter-status
title: "riff knative adapter status"
---
## riff knative adapter status

show knative adapter status

### Synopsis

Display status details for a adapter.

The Ready condition is shown which should include a reason code and a
descriptive message when the status is not "True". The status for the condition
may be: "True", "False" or "Unknown". An "Unknown" status is common while the
adapter roll out is processed.

```
riff knative adapter status <name> [flags]
```

### Examples

```
riff knative adapter status my-adapter
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

* [riff knative adapter](riff_knative_adapter.md)	 - adapters push built images to Knative

