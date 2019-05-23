## riff handler delete

delete an http request handler

### Synopsis


<todo>


```
riff handler delete [flags]
```

### Examples

```
riff handler delete my-handler
riff handler delete --all 
```

### Options

```
      --all                delete all handlers within the namespace
  -h, --help               help for delete
  -n, --namespace string   kubernetes namespace (defaulted from kube config)
```

### Options inherited from parent commands

```
      --config string        config file (default is $HOME/.riff.yaml)
      --kube-config string   kubectl config file (default is $HOME/.kube/config)
      --no-color             disable color output in terminals
```

### SEE ALSO

* [riff handler](riff_handler.md)	 - handlers map HTTP requests to applications, functions or images

