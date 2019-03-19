## riff application build

Build a application container from local source

### Synopsis

Build a application container from local source

```
riff application build [flags]
```

### Options

```
  -h, --help                           help for build
      --image repository/image[:tag]   the name of the image to build; must be a writable repository/image[:tag] with credentials configured
  -l, --local-path path                path to local source to build the image from; only build-pack builds are supported at this time
```

### Options inherited from parent commands

```
      --kubeconfig path   the path of a kubeconfig (default "~/.kube/config")
      --master address    the address of the Kubernetes API server; overrides any value in kubeconfig
```

### SEE ALSO

* [riff application](riff_application.md)	 - Interact with application related resources

