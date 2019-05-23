## riff handler create

deploy an application, function or image to handle http requests

### Synopsis

deploy an application, function or image to handle http requests

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
      --application-ref string   application build to deploy
      --env stringArray          environment variable defined as a key value pair separated by an equals sign, example "--env MY_VAR=my-value"
      --env-from stringArray     environment variable from a config map or secret, example "--env-from MY_SECRET_VALUE=secretKeyRef:my-secret-name:key-in-secret", "--env-from MY_CONFIG_MAP_VALUE=configMapKeyRef:my-config-map-name:key-in-config-map"
      --function-ref string      function build to deploy
  -h, --help                     help for create
      --image string             container image to deploy
  -n, --namespace string         kubernetes namespace (defaulted from kube config)
```

### Options inherited from parent commands

```
      --config string        config file (default is $HOME/.riff.yaml)
      --kube-config string   kubectl config file (default is $HOME/.kube/config)
      --no-color             disable color output in terminals
```

### SEE ALSO

* [riff handler](riff_handler.md)	 - handle http requests with an application, function or image
