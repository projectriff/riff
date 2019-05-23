## riff function delete

delete a function, handlers and processors that reference this function will stop updating

### Synopsis


<todo>


```
riff function delete [flags]
```

### Examples

```
riff function delete my-function
riff function delete --all 
```

### Options

```
      --all                delete all functions within the namespace
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

* [riff function](riff_function.md)	 - functions built from source using riff function buildpacks

