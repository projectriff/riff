---
id: riff-application-list
title: "riff application list"
---
## riff application list

table listing of applications

### Synopsis

List applications in a namespace or across all namespaces.

For detail regarding the status of a single application, run:

    riff application status <application-name>

```
riff application list [flags]
```

### Examples

```
riff application list
riff application list --all-namespaces
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

* [riff application](riff_application.md)	 - applications built from source using application buildpacks

