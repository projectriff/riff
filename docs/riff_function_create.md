## riff function create

Create a new function resource

### Synopsis

Create a new function resource from the content of the provided Git repo/revision or local source.

The RUNTIME arg defines the language runtime that is added to the function code in the build step. The resulting image is then used to create a Knative Service (`service.serving.knative.dev`) instance of the name specified for the function. The following runtimes are available:

- 'java': uses riff's java-function-invoker (aliased as java-invoker)
- 'node': uses riff's node-function-invoker (aliased as node-invoker)
- 'command': uses riff's command-function-invoker (aliased as command-invoker)
- 'java-buildpack': uses the riff Buildpack 
- 'detect': uses the riff Buildpack's detection (currently limited to Java functions) 

Classic riff Invoker runtimes are available in addition to experimental Buildpack runtimes.

Buildpack based runtimes support building from local source in addition to within the cluster. Locally built images prefixed with 'dev.local/' are saved to the local Docker daemon while all other images are pushed to the registry specified in the image name.

From then on you can use the sub-commands for the `service` command to interact with the service created for the function.

If `--env-from` is specified the source reference can be `configMapKeyRef` to select a key from a ConfigMap or `secretKeyRef` to select a key from a Secret. The following formats are supported:

    --env-from configMapKeyRef:{config-map-name}:{key-to-select}
    --env-from secretKeyRef:{secret-name}:{key-to-select}


```
riff function create [flags]
```

### Examples

```
  riff function create node square --git-repo https://github.com/acme/square --image acme/square --namespace joseph-ns
  riff function create java tweets-logger --git-repo https://github.com/acme/tweets --image acme/tweets-logger:1.0.0
```

### Options

```
      --artifact path                  path to the function source code or jar file; auto-detected if not specified
      --dry-run                        don't create resources but print yaml representation on stdout
      --env stringArray                environment variable expressed in a 'key=value' format
      --env-from stringArray           environment variable created from a source reference; see command help for supported formats
      --git-repo URL                   the URL for a git repository hosting the function code
      --git-revision ref-spec          the git ref-spec of the function code to use (default "master")
      --handler method or class        the name of the method or class to invoke, depending on the runtime used
  -h, --help                           help for create
      --image repository/image[:tag]   the name of the image to build; must be a writable repository/image[:tag] with credentials configured
  -l, --local-path path                path to local source to build the image from
  -n, --namespace namespace            the namespace of the subscription, channel, and function
  -v, --verbose                        print details of command progress
  -w, --wait                           wait until the created resource reaches either a successful or an error state (automatic with --verbose)
```

### Options inherited from parent commands

```
      --kubeconfig path   the path of a kubeconfig (default "~/.kube/config")
      --master address    the address of the Kubernetes API server; overrides any value in kubeconfig
```

### SEE ALSO

* [riff function](riff_function.md)	 - Interact with function related resources

