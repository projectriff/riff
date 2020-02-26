---
id: riff-streaming-inmemory-gateway-status
title: "riff streaming inmemory-gateway status"
---
## riff streaming inmemory-gateway status

show inmemory gateway status

### Synopsis

Display status details for an in-memory gateway.

The Ready condition is shown which should include a reason code and a
descriptive message when the status is not "True". The status for the condition
may be: "True", "False" or "Unknown". An "Unknown" status is common while the
in-memory gateway rollout is being processed.

```
riff streaming inmemory-gateway status <name> [flags]
```

### Examples

```
riff streamming inmemory-gateway status my-inmemory-gateway
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

* [riff streaming inmemory-gateway](riff_streaming_inmemory-gateway.md)	 - (experimental) in-memory stream gateway

