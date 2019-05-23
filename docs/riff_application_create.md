## riff application create

build an application from source

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
      --cache-size string     size of persistent volume to cache resources between builds
      --git-repo string       git url to remote source code
      --git-revision string   refspec within the git repo to checkout (default "master")
  -h, --help                  help for create
      --image string          repository where the built images are pushed
      --local-path string     path to source code on the local machine
  -n, --namespace string      kubernetes namespace (defaulted from kube config)
      --sub-path string       path within the git repo to checkout
```

### Options inherited from parent commands

```
      --config string        config file (default is $HOME/.riff.yaml)
      --kube-config string   kubectl config file (default is $HOME/.kube/config)
      --no-color             disable color output in terminals
```

### SEE ALSO

* [riff application](riff_application.md)	 - applications are built from source using Cloud Foundry buildpacks

