## riff function create

create a function from source

### Synopsis


<todo>


```
riff function create [flags]
```

### Examples

```
riff function create my-func --image registry.example.com/image --git-repo https://example.com/my-func.git
riff function create my-func --image registry.example.com/image --local-path ./my-func
```

### Options

```
      --artifact string       file containing the function within the build workspace (detected by default)
      --cache-size string     size of persistent volume to cache resources between builds
      --git-repo string       git url to remote source code
      --git-revision string   refspec within the git repo to checkout (default "master")
      --handler string        name of the method or class to invoke, depends on the invoker (detected by default)
  -h, --help                  help for create
      --image string          repository where the built images are pushed
      --invoker string        language runtime invoker (detected by default)
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

* [riff function](riff_function.md)	 - functions built from source using function buildpacks

