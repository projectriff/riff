---
id: riff-handler-list
title: "riff handler list"
---
## riff handler list

table listing of handlers

### Synopsis

List handlers in a namespace or across all namespaces.

For detail regarding the status of a single handler, run:

	riff handler status <handler-name>

```
riff handler list [flags]
```

### Examples

```
riff handler list
riff handler list --all-namespaces
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

* [riff handler](riff_handler.md)	 - handlers map HTTP requests to applications, functions or images

