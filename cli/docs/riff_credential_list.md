---
id: riff-credential-list
title: "riff credential list"
---
## riff credential list

table listing of credentials

### Synopsis

List credentials in a namespace or across all namespaces.

```
riff credential list [flags]
```

### Examples

```
riff credential list
riff credential list --all-namespaces
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

* [riff credential](riff_credential.md)	 - credentials for container registries

