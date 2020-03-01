# RFC-0014: Service Binding

**Authors:** @scothis

**Status:**

**Pull Request URL:** link

**Superseded by:** N/A

**Supersedes:** [RFC 0002 Bindings](./rfc-0002-bindings.md)

**Related:** [RFC 0008 Pod Templates](./rfc-0008-pod-templates.md), [#1382](https://github.com/projectriff/riff/pull/1382)


## Problem

[RFC-0002 Bindings](./rfc-0002-bindings.md) defined a interface for a resource to expose metadata and secrets to be consumed by another resource. It leveraged the [Cloud Native Buildpacks Binding](https://github.com/buildpack/spec/blob/master/extensions/bindings.md) (CNB bindings) spec for how the provisioned metadata and secret are exposed inside a container. The upstream CNB spec has been deprecated in favor of the [Service Binding specification for Kubernetes](https://github.com/k8s-service-bindings/spec).

### Anti-Goals

This RFC has no opinion as to the semantic source, content, or consumer of the service binding,

## Solution

RFC-0002 Bindings is superseded by this RFC.

Many riff resources are PodSpec-able, and may be the target of a `ServiceBinding`. Function invokers and other riff components are encouraged to consume bindings by looking for the `SERVICE_BINDING_ROOT` environment variable and reading bound services as defined by [Application Projection](https://github.com/k8s-service-bindings/spec#application-projection).

riff resources that can be bound (like Streams) should implement the Service Binding [Provisioned Service](https://github.com/k8s-service-bindings/spec#provisioned-service) duck-type.

While riff users are encouraged to use the `ServiceBinding` resource to bind resources like a database to a running workload, riff components should avoid creating a dependency on a `ServiceBinding` implementation. For example, `Processor`s and `Stream`s can implement the projection and provisioning sub-specs respectively, but because they have tight knowledge of each other, they do not need a `ServiceBinding` resource to complete the binding.

### User Impact

Function authors may leverage Service Bindings to inject other services into a riff workload. Moreover, workloads can use Service Bindings to consume `Stream`s outside of a `Processor`. All bindings are limited in scope to the current namespace.

### Backwards Compatibility and Upgrade Path

The `Stream`'s status will switch from exposing a CNB Binding to exposing a Service Binding for Kubernetes. Non-Processor consumers of the streams (like dev-utils) will need to be updated. The two binding concepts are quite similar, with the Service Binding being simpler, yet just as capable.

Old:

```yaml
...
status:
  ...
  binding:
    metadataRef:
      name: provider-binding-metadata
    secretRef:
      name: provider-binding-secret
```

New:

```yaml
...
status:
  ...
  binding:
    name: provider-binding-secret
```

## FAQ
