## riff processor delete

stop processing messages

### Synopsis


<todo>


```
riff processor delete [flags]
```

### Examples

```
riff processor delete my-processor
riff processor delete --all 
```

### Options

```
      --all                delete all processors within the namespace
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

* [riff processor](riff_processor.md)	 - processors apply functions to messages on streams

