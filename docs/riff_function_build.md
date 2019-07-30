---
id: riff-function-build
title: "riff function build"
---
## riff function build

Build a function container from local source

### Synopsis

Build a function container from local source

```
riff function build [flags]
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

* [riff function](riff_function.md)	 - Interact with function related resources

