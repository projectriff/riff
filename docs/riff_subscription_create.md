## riff subscription create

Create a new subscription, binding a service to an input channel

### Synopsis

Create a new, optionally named subscription, binding a service to an input channel.
The default name of the subscription is the provided service name.
The service can optionally be bound to an output channel.

```
riff subscription create [flags]
```

### Examples

```
  riff subscription create --from tweets --processor tweets-logger
  riff subscription create my-subscription --from tweets --processor tweets-logger
  riff subscription create --from tweets --processor tweets-logger --to logged-tweets
```

### Options

```
  -i, --from string        the input channel the service binds to
  -h, --help               help for create
  -n, --namespace string   the namespace of the subscription
  -s, --processor string   the subscriber registered in the subscription
  -o, --to string          the optional output channel the service binds to
```

### Options inherited from parent commands

```
      --kubeconfig path   the path of a kubeconfig (default "~/.kube/config")
      --master address    the address of the Kubernetes API server; overrides any value in kubeconfig
```

### SEE ALSO

* [riff subscription](riff_subscription.md)	 - Interact with subscription-related resources

