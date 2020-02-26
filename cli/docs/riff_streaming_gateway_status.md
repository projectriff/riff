---
id: riff-streaming-gateway-status
title: "riff streaming gateway status"
---
## riff streaming gateway status

show gateway status

### Synopsis

Display status details for a gateway.

The Ready condition is shown which should include a reason code and a
descriptive message when the status is not "True". The status for the condition
may be: "True", "False" or "Unknown". An "Unknown" status is common while the
gateway rollout is being processed.

```
riff streaming gateway status <name> [flags]
```

### Examples

```
riff streamming gateway status my-gateway
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

* [riff streaming gateway](riff_streaming_gateway.md)	 - (experimental) stream gateway

