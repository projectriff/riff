## riff application create

create an application from source

### Synopsis

<todo>

```
riff application create [flags]
```

### Examples

```
riff application create my-app --image registry.example.com/image --git-repo https://example.com/my-app.git
riff application create my-app --image registry.example.com/image --local-path ./my-app
```

### Options

```
      --cache-size size        size of persistent volume to cache resources between builds
      --git-repo url           git url to remote source code
      --git-revision refspec   refspec within the git repo to checkout (default "master")
  -h, --help                   help for create
      --image repository       repository where the built images are pushed (default "_")
      --local-path directory   path to directory containing source code on the local machine
  -n, --namespace name         kubernetes namespace (defaulted from kube config)
      --sub-path directory     path to directory within the git repo to checkout
      --tail                   watch build logs
```

### Options inherited from parent commands

```
      --config file        config file (default is $HOME/.riff.yaml)
      --kube-config file   kubectl config file (default is $HOME/.kube/config)
      --no-color           disable color output in terminals
```

### SEE ALSO

* [riff application](riff_application.md)	 - applications built from source using application buildpacks

