## riff function create

Create a new function resource, with optional input binding

### Synopsis

Create a new function resource from the content of the provided Git repo/revision.

The INVOKER arg defines the language invoker that is added to the function code in the build step. The resulting image is 
then used to create a Knative Service (service.serving.knative.dev) instance of the name specified for the function. 
From then on you can use the sub-commands for the 'service' command to interact with the service created for the function. 

If an input channel and bus are specified, create the channel in the bus and subscribe the service to the channel.

If an env-from flag is specified the source reference can be 'configMapKeyRef' to select a key from a ConfigMap
or 'secretKeyRef' to select a key from a Secret. The following formats are supported:
  --env-from configMapKeyRef:{config-map-name}:{key-to-select}
  --env-from secretKeyRef:{secret-name}:{key-to-select}


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
      --bus name                       the name of the bus to create the channel in.
      --cluster-bus name               the name of the cluster bus to create the channel in.
      --dry-run                        don't create resources but print yaml representation on stdout
      --env stringArray                environment variable expressed in a 'key=value' format
      --env-from stringArray           environment variable created from a source reference; see command help for supported formats
      --git-repo URL                   the URL for a git repository hosting the function code
      --git-revision ref-spec          the git ref-spec of the function code to use (default "master")
      --handler method or class        the name of the method or class to invoke, depending on the invoker used
  -h, --help                           help for create
      --image repository/image[:tag]   the name of the image to build; must be a writable repository/image[:tag] with credentials configured
  -i, --input channel                  name of the function's input channel, if any
  -n, --namespace namespace            the namespace of the subscription, channel, and function
```

### Options inherited from parent commands

```
      --kubeconfig path   the path of a kubeconfig (default "~/.kube/config")
      --master address    the address of the Kubernetes API server; overrides any value in kubeconfig
```

### SEE ALSO

* [riff function](riff_function.md)	 - Interact with function related resources

