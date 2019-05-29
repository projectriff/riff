## riff application tail

watch build logs

### Synopsis

<todo>

```
riff application tail [flags]
```

### Examples

```
riff application tail my-application
riff application tail my-application --since 1h
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

* [riff application](riff_application.md)	 - applications built from source using application buildpacks

