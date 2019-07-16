---
id: riff-subscription-list
title: "riff subscription list"
---
## riff subscription list

[DEPRECATED] List existing subscriptions

### Synopsis

[DEPRECATED] List existing subscriptions

```
riff subscription list [flags]
```

### Examples

```
  riff subscription list
  riff subscription list --namespace joseph-ns
```

### Options

```
  -h, --help               help for list
  -n, --namespace string   the namespace of the subscriptions
  -o, --output string      the custom output format to use. Use 'dot' to output graphviz representation
```

### Options inherited from parent commands

```
      --kubeconfig path   the path of a kubeconfig (default "~/.kube/config")
      --master address    the address of the Kubernetes API server; overrides any value in kubeconfig
```

### SEE ALSO

* [riff subscription](riff_subscription.md)	 - [DEPRECATED] Interact with subscription-related resources

