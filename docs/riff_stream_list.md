## riff stream list

list streams in a namespace

### Synopsis

list streams in a namespace

```
riff stream list [flags]
```

### Examples

```
riff stream list
riff stream list --all-namespaces
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

* [riff stream](riff_stream.md)	 - stream of messages
