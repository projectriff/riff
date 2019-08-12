---
id: riff
title: "riff"
---
## riff

riff is for functions

### Synopsis

The riff CLI combines with the projectriff system CRDs to build, run and wire
workloads (functions, applications and containers). The CRDs provide the riff
API of which this CLI is a client.

Before running riff, please install the projectriff system and its dependencies.
See https://projectriff.io/docs/getting-started/

The application, function and container commands define build plans and the
credential commands to authenticate builds to container registries.

Runtimes provide ways to execute the workloads. Different runtimes provide
alternate execution models and capabilities.

The core runtime uses core Kubernetes resources like Deployment and Service to
expose the workload over HTTP.

The Knative runtime uses Knative Serving to expose the workload over HTTP with
zero-to-n autoscaling and managed ingress.

### Options

```
      --config file        config file (default is $HOME/.riff.yaml)
  -h, --help               help for riff
      --kube-config file   kubectl config file (default is $HOME/.kube/config)
      --no-color           disable color output in terminals
```

### SEE ALSO

* [riff application](riff_application.md)	 - applications built from source using application buildpacks
* [riff completion](riff_completion.md)	 - generate shell completion script
* [riff container](riff_container.md)	 - containers resolve the latest image
* [riff core](riff_core.md)	 - core runtime for riff workloads
* [riff credential](riff_credential.md)	 - credentials for container registries
* [riff doctor](riff_doctor.md)	 - check riff's requirements are installed
* [riff function](riff_function.md)	 - functions built from source using function buildpacks
* [riff knative](riff_knative.md)	 - Knative runtime for riff workloads

