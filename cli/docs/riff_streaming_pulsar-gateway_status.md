---
id: riff-streaming-pulsar-gateway-status
title: "riff streaming pulsar-gateway status"
---
## riff streaming pulsar-gateway status

show pulsar gateway status

### Synopsis

Display status details for a Pulsar gateway.

The Ready condition is shown which should include a reason code and a
descriptive message when the status is not "True". The status for the condition
may be: "True", "False" or "Unknown". An "Unknown" status is common while the
pulsar gateway rollout is being processed.

```
riff streaming pulsar-gateway status <name> [flags]
```

### Examples

```
riff streamming pulsar-gateway status my-pulsar-gateway
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

* [riff streaming pulsar-gateway](riff_streaming_pulsar-gateway.md)	 - (experimental) pulsar stream gateway

