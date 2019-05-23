## riff processor list

list processors in a namespace

### Synopsis

list processors in a namespace

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
      --all-namespaces     use all kubernetes namespaces
  -h, --help               help for list
  -n, --namespace string   kubernetes namespace (defaulted from kube config)
```

### Options inherited from parent commands

```
      --config string        config file (default is $HOME/.riff.yaml)
      --kube-config string   kubectl config file (default is $HOME/.kube/config)
      --no-color             disable color output in terminals
```

### SEE ALSO

* [riff processor](riff_processor.md)	 - process messages with a function

