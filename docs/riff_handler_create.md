## riff handler create

create a handler to map HTTP requests to an application, function or image

### Synopsis

<todo>

```
riff handler create [flags]
```

### Examples

```
riff handler create my-app-handler --application-ref my-app
riff handler create my-func-handler --function-ref my-func
riff handler create my-image-handler --image registry.example.com/my-image:latest
```

### Options

```
      --application-ref name   name of application to deploy
      --env variable           environment variable defined as a key value pair separated by an equals sign, example "--env MY_VAR=my-value" (may be set multiple times)
      --env-from variable      environment variable from a config map or secret, example "--env-from MY_SECRET_VALUE=secretKeyRef:my-secret-name:key-in-secret", "--env-from MY_CONFIG_MAP_VALUE=configMapKeyRef:my-config-map-name:key-in-config-map" (may be set multiple times)
      --function-ref name      name of function to deploy
  -h, --help                   help for create
      --image image            container image to deploy
  -n, --namespace name         kubernetes namespace (defaulted from kube config)
      --tail                   watch handler logs
```

### Options inherited from parent commands

```
      --config file        config file (default is $HOME/.riff.yaml)
      --kube-config file   kubectl config file (default is $HOME/.kube/config)
      --no-color           disable color output in terminals
```

### SEE ALSO

* [riff handler](riff_handler.md)	 - handlers map HTTP requests to applications, functions or images

