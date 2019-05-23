## riff stream create

create a stream of messages

### Synopsis

create a stream of messages

```
riff stream create [flags]
```

### Examples

```
riff stream create --provider my-provider
```

### Options

```
  -h, --help               help for create
  -n, --namespace string   kubernetes namespace (defaulted from kube config)
      --provider string    stream provider
```

### Options inherited from parent commands

```
      --config string        config file (default is $HOME/.riff.yaml)
      --kube-config string   kubectl config file (default is $HOME/.kube/config)
      --no-color             disable color output in terminals
```

### SEE ALSO

* [riff stream](riff_stream.md)	 - stream of messages

