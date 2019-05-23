## riff stream delete

delete stream(s)

### Synopsis

<todo>

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
      --all              delete all streams within the namespace
  -h, --help             help for delete
  -n, --namespace name   kubernetes namespace (defaulted from kube config)
```

### Options inherited from parent commands

```
      --config file        config file (default is $HOME/.riff.yaml)
      --kube-config file   kubectl config file (default is $HOME/.kube/config)
      --no-color           disable color output in terminals
```

### SEE ALSO

* [riff stream](riff_stream.md)	 - streams of messages

