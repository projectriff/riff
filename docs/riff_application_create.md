## riff application create

Create a new application resource

### Synopsis

Create a new application resource from the content of the provided Git repo/revision or local source.

Images will be pushed to the registry specified in the image name. If a default image prefix was specified during namespace init, the image flag is optional. The application name is combined with the default prefix to define the image. Instead of using the application name, a custom repository can be specified with the image set like `--image _/custom-name` which would resolve to `docker.io/example/custom-name`.

From then on you can use the sub-commands for the `service` command to interact with the service created for the application.

If `--env-from` is specified the source reference can be `configMapKeyRef` to select a key from a ConfigMap or `secretKeyRef` to select a key from a Secret. The following formats are supported:

    --env-from configMapKeyRef:{config-map-name}:{key-to-select}
    --env-from secretKeyRef:{secret-name}:{key-to-select}


```
riff application create [flags]
```

### Examples

```
  riff application create square --git-repo https://github.com/acme/square --artifact square.js --image acme/square --namespace joseph-ns
  riff application create tweets-logger --git-repo https://github.com/acme/tweets --image acme/tweets-logger:1.0.0
```

### Options

```
      --dry-run                        don't create resources but print yaml representation on stdout
      --env stringArray                environment variable expressed in a 'key=value' format
      --env-from stringArray           environment variable created from a source reference; see command help for supported formats
      --git-repo URL                   the URL for a git repository hosting the application code
      --git-revision ref-spec          the git ref-spec of the application code to use (default "master")
  -h, --help                           help for create
      --image repository/image[:tag]   the name of the image to build; must be a writable repository/image[:tag] with credentials configured
  -l, --local-path path                path to local source to build the image from; only build-pack builds are supported at this time
  -n, --namespace namespace            the namespace of the service
      --sub-path string                the directory within the git repo to expose, files outside of this directory will not be available during the build
  -v, --verbose                        print details of command progress
  -w, --wait                           wait until the created resource reaches either a successful or an error state (automatic with --verbose)
```

### Options inherited from parent commands

```
      --kubeconfig path   the path of a kubeconfig (default "~/.kube/config")
      --master address    the address of the Kubernetes API server; overrides any value in kubeconfig
```

### SEE ALSO

* [riff application](riff_application.md)	 - Interact with application related resources

