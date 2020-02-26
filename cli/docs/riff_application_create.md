---
id: riff-application-create
title: "riff application create"
---
## riff application create

create an application from source

### Synopsis

Create an application from source using the application Cloud Native Buildpack
builder.

Application source can be specified either as a Git repository or as a local
directory. Builds from Git are run in the cluster while builds from a local
directory are run inside a local Docker daemon and are orchestrated by this
command (in the future, builds from local source may also be run in the
cluster).

```
riff application create <name> [flags]
```

### Examples

```
riff application create my-app --image registry.example.com/image --git-repo https://example.com/my-app.git
riff application create my-app --image registry.example.com/image --local-path ./my-app
```

### Options

```
      --cache-size size         size of persistent volume to cache resources between builds
      --dry-run                 print kubernetes resources to stdout rather than apply them to the cluster, messages normally on stdout will be sent to stderr
      --env variable            environment variable defined as a key value pair separated by an equals sign, example "--env MY_VAR=my-value" (may be set multiple times)
      --git-repo url            git url to remote source code
      --git-revision refspec    refspec within the git repo to checkout (default "master")
  -h, --help                    help for create
      --image repository        repository where the built images are pushed (default "_")
      --limit-cpu cores         the maximum amount of cpu allowed, in CPU cores (500m = .5 cores)
      --limit-memory bytes      the maximum amount of memory allowed, in bytes (500Mi = 500MiB = 500 * 1024 * 1024)
      --local-path directory    path to directory containing source code on the local machine
  -n, --namespace name          kubernetes namespace (defaulted from kube config)
      --sub-path directory      path to directory within the git repo to checkout
      --tail                    watch build logs
      --wait-timeout duration   duration to wait for the application to become ready when watching logs (default "10m")
```

### Options inherited from parent commands

```
      --config file       config file (default is $HOME/.riff.yaml)
      --kubeconfig file   kubectl config file (default is $HOME/.kube/config)
      --no-color          disable color output in terminals
```

### SEE ALSO

* [riff application](riff_application.md)	 - applications built from source using application buildpacks

