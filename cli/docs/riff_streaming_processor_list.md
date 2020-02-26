---
id: riff-streaming-processor-list
title: "riff streaming processor list"
---
## riff streaming processor list

table listing of processors

### Synopsis

List processors in a namespace or across all namespaces.

For detail regarding the status of a single processor, run:

    riff processor status <processor-name>

```
riff streaming processor list [flags]
```

### Examples

```
riff streaming processor list
riff streaming processor list --all-namespaces
```

### Options

```
      --all-namespaces   use all kubernetes namespaces
  -h, --help             help for list
  -n, --namespace name   kubernetes namespace (defaulted from kube config)
```

### Options inherited from parent commands

```
      --config file       config file (default is $HOME/.riff.yaml)
      --kubeconfig file   kubectl config file (default is $HOME/.kube/config)
      --no-color          disable color output in terminals
```

### SEE ALSO

* [riff streaming processor](riff_streaming_processor.md)	 - (experimental) processors apply functions to messages on streams

