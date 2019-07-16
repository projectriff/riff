---
id: riff-processor-list
title: "riff processor list"
---
## riff processor list

table listing of processors

### Synopsis

List processors in a namespace or across all namespaces.

For detail regarding the status of a single processor, run:

	riff processor status <processor-name>

```
riff processor list [flags]
```

### Examples

```
riff processor list
riff processor list --all-namespaces
```

### Options

```
      --all-namespaces   use all kubernetes namespaces
  -h, --help             help for list
  -n, --namespace name   kubernetes namespace (defaulted from kube config)
```

### Options inherited from parent commands

```
      --config file        config file (default is $HOME/.riff.yaml)
      --kube-config file   kubectl config file (default is $HOME/.kube/config)
      --no-color           disable color output in terminals
```

### SEE ALSO

* [riff processor](riff_processor.md)	 - processors apply functions to messages on streams

