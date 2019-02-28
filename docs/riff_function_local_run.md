## riff function local run

Run a function from source locally

### Synopsis

Run a function from source locally

```
riff function local run [flags]
```

### Options

```
      --artifact path             path to the function source code or jar file; auto-detected if not specified
      --handler method or class   the name of the method or class to invoke, depending on the invoker used
  -h, --help                      help for run
      --invoker language          invoker runtime to override language detected by buildpack
  -l, --local-path path           path to local source to build the image from; only build-pack builds are supported at this time
  -p, --port strings              Port to publish (defaults to port(s) exposed by container)
```

### Options inherited from parent commands

```
      --kubeconfig path   the path of a kubeconfig (default "~/.kube/config")
      --master address    the address of the Kubernetes API server; overrides any value in kubeconfig
```

### SEE ALSO

* [riff function local](riff_function_local.md)	 - Interact with functions outside of a kubernetes cluster

