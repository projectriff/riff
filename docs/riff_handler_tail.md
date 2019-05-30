## riff handler tail

watch handler logs

### Synopsis

<todo>

```
riff handler tail [flags]
```

### Examples

```
riff handler tail my-handler
riff handler tail my-handler --since 1h
```

### Options

```
  -h, --help             help for tail
  -n, --namespace name   kubernetes namespace (defaulted from kube config)
      --since duration   time duration to start reading logs from
```

### Options inherited from parent commands

```
      --config file        config file (default is $HOME/.riff.yaml)
      --kube-config file   kubectl config file (default is $HOME/.kube/config)
      --no-color           disable color output in terminals
```

### SEE ALSO

* [riff handler](riff_handler.md)	 - handlers map HTTP requests to applications, functions or images

