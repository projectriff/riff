---
id: riff-function-create
title: "riff function create"
---
## riff function create

create a function from source

### Synopsis

Create a function from source using the function Cloud Native Buildpack builder.

Function source can be specified either as a Git repository or as a local
directory. Builds from Git are run in the cluster while builds from a local
directory are run inside a local Docker daemon and are orchestrated by this
command (in the future, builds from local source may also be run in the
cluster).

In addition to the source code, functions are defined by these properties:

- invoker - language runtime that should host the function, the invoker is often
    auto-detected, but may need to be specified in cases of ambiguity.
- artifact - file in the source that contains the function.
- handler - invoker specific, typically the method or class within the artifact.

These values can be versioned with the source code in a riff.toml file, or
specified here to override the source. Versioning with the source is preferred
as changed can be deployed as a unit. Overriding is necessary when deploying
multiple functions from a single code base.

The riff.toml file takes the form:

    override = "<invoker name>"
	artifact = "<path to artifact>"
	handler = "<function handler>"

```
riff function create <name> [flags]
```

### Examples

```
riff function create my-func --image registry.example.com/image --git-repo https://example.com/my-func.git
riff function create my-func --image registry.example.com/image --local-path ./my-func
```

### Options

```
      --artifact file           file containing the function within the build workspace (detected by default)
      --cache-size size         size of persistent volume to cache resources between builds
      --dry-run                 print kubernetes resources to stdout rather than apply them to the cluster, messages normally on stdout will be sent to stderr
      --git-repo url            git url to remote source code
      --git-revision refspec    refspec within the git repo to checkout (default "master")
      --handler name            name of the method or class to invoke, depends on the invoker (detected by default)
  -h, --help                    help for create
      --image repository        repository where the built images are pushed (default "_")
      --invoker name            language runtime invoker name (detected by default)
      --local-path directory    path to directory containing source code on the local machine
  -n, --namespace name          kubernetes namespace (defaulted from kube config)
      --sub-path directory      path to directory within the git repo to checkout
      --tail                    watch build logs
      --wait-timeout duration   duration to wait for the function to become ready when watching logs (default "10m")
```

### Options inherited from parent commands

```
      --config file        config file (default is $HOME/.riff.yaml)
      --kube-config file   kubectl config file (default is $HOME/.kube/config)
      --no-color           disable color output in terminals
```

### SEE ALSO

* [riff function](riff_function.md)	 - functions built from source using function buildpacks

