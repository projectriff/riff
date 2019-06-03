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
      --artifact file           file containing the function within the build workspace (detected by default)
      --cache-size size         size of persistent volume to cache resources between builds
      --git-repo url            git url to remote source code
      --git-revision refspec    refspec within the git repo to checkout (default "master")
      --handler name            name of the method or class to invoke, depends on the invoker (detected by default)
  -h, --help                    help for create
      --image repository        repository where the built images are pushed (default "_")
      --invoker name            language runtime invoker name (detected by default)
      --local-path directory    path to directory containing source code on the local machine
  -n, --namespace name          kubernetes namespace (defaulted from kube config)
      --sub-path directory      path to directory within the git repo to checkout
      --tail                    watch build logs
      --wait-timeout duration   duration to wait for the function to become ready when watching logs (default "10m")
```

### Options inherited from parent commands

```
      --config file        config file (default is $HOME/.riff.yaml)
      --kube-config file   kubectl config file (default is $HOME/.kube/config)
      --no-color           disable color output in terminals
```

### SEE ALSO

* [riff function](riff_function.md)	 - functions built from source using function buildpacks

