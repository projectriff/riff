---
id: riff-function
title: "riff function"
---
## riff function

functions built from source using function buildpacks

### Synopsis

Functions are a mechanism for converting language idiomatic units of logic into
container images that can be invoked over HTTP or used to process streams of
messages. Cloud Native Buildpacks are provided to detect the language, provide a
language runtime, install build and runtime dependencies, compile the function,
and packaging everything as a container.

The function resource is only responsible for converting source code into a
container. The function container image may then be deployed to one of the
runtimes.

Functions are distinct from applications in the scope and responsibilities of
the source code. Unlike applications, functions:

- no main method
- practice Inversion of Control (we'll call you)
- invocations are decoupled from networking protocols, no HTTP specifics
- limited to a single responsibility

### Options

```
  -h, --help   help for function
```

### Options inherited from parent commands

```
      --config file        config file (default is $HOME/.riff.yaml)
      --kube-config file   kubectl config file (default is $HOME/.kube/config)
      --no-color           disable color output in terminals
```

### SEE ALSO

* [riff](riff.md)	 - riff is for functions
* [riff function create](riff_function_create.md)	 - create a function from source
* [riff function delete](riff_function_delete.md)	 - delete function(s)
* [riff function list](riff_function_list.md)	 - table listing of functions
* [riff function status](riff_function_status.md)	 - show function status
* [riff function tail](riff_function_tail.md)	 - watch build logs

