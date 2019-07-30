---
id: riff-subscription-create
title: "riff subscription create"
---
## riff subscription create

[DEPRECATED] Create a new subscription, binding a service to an input channel

### Synopsis

Create a new, optionally named subscription, binding a service to an input channel. The default name of the subscription is the provided subscriber name. The subscription can optionally be bound to an output channel.

```
riff subscription create [flags]
```

### Examples

```
  riff subscription create --channel tweets --subscriber tweets-logger
  riff subscription create my-subscription --channel tweets --subscriber tweets-logger
  riff subscription create --channel tweets --subscriber tweets-logger --reply logged-tweets
```

### Options

```
  -c, --channel string      the input channel of the subscription
  -h, --help                help for create
  -n, --namespace string    the namespace of the subscription
  -r, --reply string        the optional output channel of the subscription
  -s, --subscriber string   the subscriber of the subscription
```

### Options inherited from parent commands

```
      --kubeconfig path   the path of a kubeconfig (default "~/.kube/config")
      --master address    the address of the Kubernetes API server; overrides any value in kubeconfig
```

### SEE ALSO

* [riff subscription](riff_subscription.md)	 - [DEPRECATED] Interact with subscription-related resources

