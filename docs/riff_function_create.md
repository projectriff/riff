## riff function create

Create a new function resource, with optional input binding

### Synopsis

Create a new function resource, with optional input binding

```
riff function create [flags]
```

### Examples

```
  riff function create node square --git-repo https://github.com/acme/square --image acme/square --namespace joseph-ns
  riff function create java tweets-logger --git-repo https://github.com/acme/tweets --image acme/tweets-logger:1.0.0 --input tweets --bus kafka
```

### Options

```
      --artifact path                  path to the function source code or jar file; auto-detected if not specified
      --bus name                       the name of a bus for the channel
      --cluster-bus name               the name of a cluster bus for the channel
  -f, --force                          whether to force writing of files if they already exist.
      --git-repo URL                   the URL for a git repository hosting the function code
      --git-revision ref-spec          the git ref-spec of the function code to use (default "master")
      --handler method or class        the name of the method or class to invoke, depending on the invoker used
  -h, --help                           help for create
      --image repository/image[:tag]   the name of the image to build; must be a writable repository/image[:tag] with credentials configured
  -i, --input channel                  name of the function's input channel, if any
  -n, --namespace namespace            the namespace of the function and the specified resources
  -w, --write                          whether to write yaml files for created resources
```

### Options inherited from parent commands

```
      --kubeconfig path   the path of a kubeconfig (default "~/.kube/config")
      --master address    the address of the Kubernetes API server; overrides any value in kubeconfig
```

### SEE ALSO

* [riff function](riff_function.md)	 - Interact with function related resources

