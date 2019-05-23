## riff credential delete

delete credential(s)

### Synopsis

<todo>

```
riff credential delete [flags]
```

### Examples

```
riff credential delete my-creds
riff credential delete --all 
```

### Options

```
      --all              delete all credentials within the namespace
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

* [riff credential](riff_credential.md)	 - credentials for container registries

