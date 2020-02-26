---
id: riff-application
title: "riff application"
---
## riff application

applications built from source using application buildpacks

### Synopsis

Applications are a mechanism to convert web application source code into
container images that can be invoked over HTTP. Cloud Native Buildpacks are
provided to detect the language, provide a language runtime, install build and
runtime dependencies, compile the application, and packaging everything as a
container.

The application resource is only responsible for converting source code into a
container. The application container image may then be deployed on the core or
knative runtime.

### Options

```
  -h, --help   help for application
```

### Options inherited from parent commands

```
      --config file       config file (default is $HOME/.riff.yaml)
      --kubeconfig file   kubectl config file (default is $HOME/.kube/config)
      --no-color          disable color output in terminals
```

### SEE ALSO

* [riff](riff.md)	 - riff is for functions
* [riff application create](riff_application_create.md)	 - create an application from source
* [riff application delete](riff_application_delete.md)	 - delete application(s)
* [riff application list](riff_application_list.md)	 - table listing of applications
* [riff application status](riff_application_status.md)	 - show application status
* [riff application tail](riff_application_tail.md)	 - watch build logs

