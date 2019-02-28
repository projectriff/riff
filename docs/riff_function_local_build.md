## riff function local build

Build a function container from local source

### Synopsis

Build a function container from local source

```
riff function local build [flags]
```

### Options

```
      --artifact path                  path to the function source code or jar file; auto-detected if not specified
      --handler method or class        the name of the method or class to invoke, depending on the invoker used
  -h, --help                           help for build
      --image repository/image[:tag]   the name of the image to build; must be a writable repository/image[:tag] with credentials configured
      --invoker language               invoker runtime to override language detected by buildpack
  -l, --local-path path                path to local source to build the image from; only build-pack builds are supported at this time
```

### Options inherited from parent commands

```
      --kubeconfig path   the path of a kubeconfig (default "~/.kube/config")
      --master address    the address of the Kubernetes API server; overrides any value in kubeconfig
```

### SEE ALSO

* [riff function local](riff_function_local.md)	 - Interact with functions outside of a kubernetes cluster

