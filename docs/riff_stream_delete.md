## riff stream delete

delete a stream of messages

### Synopsis

delete a stream of messages

```
riff stream delete [flags]
```

### Examples

```
riff stream delete my-stream
riff stream delete --all 
```

### Options

```
      --all                delete all streams within the namespace
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

* [riff stream](riff_stream.md)	 - stream of messages
