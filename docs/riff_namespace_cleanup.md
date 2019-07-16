---
id: riff-namespace-cleanup
title: "riff namespace cleanup"
---
## riff namespace cleanup

cleans up riff resources in the namespace

### Synopsis

cleans up riff resources in the namespace and the namespace itself if "--remove-ns" is set

```
riff namespace cleanup [flags]
```

### Examples

```
  riff namespace cleanup my-ns
  riff namespace cleanup my-ns --remove-ns
```

### Options

```
  -h, --help        help for cleanup
      --remove-ns   removes the (non-default) namespace as well
```

### Options inherited from parent commands

```
      --kubeconfig path   the path of a kubeconfig (default "~/.kube/config")
      --master address    the address of the Kubernetes API server; overrides any value in kubeconfig
```

### SEE ALSO

* [riff namespace](riff_namespace.md)	 - Manage namespaces used for riff resources

