## riff function create

create a new function resource, with optional input binding

### Synopsis

create a new function resource, with optional input binding

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
      --artifact path                  path to the function artifact, source code or jar file. Attempts detection if not specified.
      --bus name                       the name of the bus to create the channel in.
      --cluster-bus name               the name of the cluster bus to create the channel in.
  -f, --force                          force writing of files if they already exist.
      --git-repo URL                   the URL for the git repo hosting the function source.
      --git-revision ref-spec          the git ref-spec to build. (default "master")
      --handler method or class        name of method or class to invoke. See specific invoker for detail.
  -h, --help                           help for create
      --image repository/image[:tag]   the name of the image to build. Must be a writable repository/image[:tag] with write credentials configured.
  -i, --input channel                  name of the input channel to subscribe the function to.
  -n, --namespace namespace            the namespace to use when interacting with resources.
  -w, --write                          whether to write yaml files for created resources.
```

### Options inherited from parent commands

```
      --kubeconfig path   path to a kubeconfig. (default "~/.kube/config")
      --master address    the address of the Kubernetes API server. Overrides any value in kubeconfig.
```

### SEE ALSO

* [riff function](riff_function.md)	 - interact with function related resources

