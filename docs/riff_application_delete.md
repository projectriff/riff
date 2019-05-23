## riff application delete

delete an application, handlers that reference this app will stop updating

### Synopsis


<todo>


```
riff application delete [flags]
```

### Examples

```
riff application delete my-application
riff application delete --all
```

### Options

```
      --all                delete all applications within the namespace
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

* [riff application](riff_application.md)	 - applications built from source using Cloud Foundry buildpacks

