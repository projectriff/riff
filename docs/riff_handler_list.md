## riff handler list

list http request handlers in a namespace

### Synopsis

list http request handlers in a namespace

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

* [riff handler](riff_handler.md)	 - handle http requests with an application, function or image

