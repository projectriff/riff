# RFC-0002: Bindings

**Authors:** @scothis

**Status:** 

**Pull Request URL:** [#1360](https://github.com/projectriff/riff/pull/1360)

**Superseded by:** N/A

**Supersedes:** N/A

**Related:** N/A


## Problem
Non-trivial workloads often need to access resources, like databases, message brokers or any credentialed resource. In Cloud Foundry, these are called services and they are provisioned from a broker and then bound to applications at both build and runtime. As Kubernetes has a different notion of a Service, this RFC will refer to "service bindings" as simply bindings.

The binding contains both metadata and credentials to provide a workload the details it needs to connect and consume the resource.

### Anti-Goals
A generic mechanism to provision bindings, and the specific structure of the metadata and credentials are out of scope for this RFC.

## Solution
The Cloud Native Buildpacks project has a [spec for Bindings](https://github.com/buildpack/spec/blob/master/extensions/bindings.md) which defines the form bindings should take at both build and runtime. In short, each binding is represented as a directory inside a well known base directory, and each binding directory contains two sub-directories: `metadata` for general non-sensitive information about the binding, and `secret` for credentials and other sensitive information. Each attribute for the binding is a file in one of those two directories. The metadata is available at buildtime and both the metadata and secret are available at runtime.

riff should adopt the CNB Bindings spec for bindings at both buildtime and runtime.

Kubernetes uses [Volumes](https://kubernetes.io/docs/concepts/storage/volumes/) to inject files from an external source into a running container. While there is a diverse spectrum of volume implementations available, in order to maximize both compatibility with standard Kubernetes and simplify the implementation, we should focus on a [configMap volume](https://kubernetes.io/docs/concepts/storage/volumes/#configmap) for binding metadata and a [secret volume](https://kubernetes.io/docs/concepts/storage/volumes/#secret) for binding secrets, which map directly to Kubernetes [ConfigMaps](https://kubernetes.io/docs/tasks/configure-pod-container/configure-pod-configmap/) and [Secrets](https://kubernetes.io/docs/concepts/configuration/secret/) respectively.

As a binding is backed by a ConfigMap and Secret a mechanism is needed to specify which ConfigMap and Secret. [LocalObjectReferences](https://godoc.org/k8s.io/api/core/v1#LocalObjectReference) are commonly used to define the coordinates to an arbitrary resource. Therefore, a binding reference consists of a `metadata` LocalObjectReference to a ConfigMap and a `secret` LocalObjectReference to a Secret.

Any resource that provides a binding should include the binding reference on its status, for example:

```yaml
name: provider
...
status:
  ...
  binding:
    metadataRef:
      name: provider-binding-metadata
    secretRef:
      name: provider-binding-secret
```

The provider may change the values within the metadata and secret as needed, but should avoid changing the keys. The CNB Binding spec defines some well known keys that should be defined in the metadata.

A workload consuming the binding can use the provider's binding, For example:

```yaml
name: consumer
...
spec:
  ...
  bindings:
  - name: my-binding
    metadataRef:
      name: provider-binding-metadata
    secretRef:
      name: provider-binding-secret
```

It is common to consume multiple bindings within a single workload. Each binding must have a unique name for the workload. The binding name is the directory the metadata and secret volumes are mounted within. While the name has no impact on the binding per se, workloads may use the name to convey semantic meaning about the intent of the binding (e.g. a primary database vs a replica).

Note: The ConfigMap and Secret for the binding must be in the same namespace as the workload Pod.

The reconciler for the resource consuming a binding is responsible for mapping the ConfigMap and Secret to volume mounts within the pod.

kpack, as of v0.0.5, does not provide any mechanism for injecting volumes into builds, or any other means to satisfy the CNB Bindings spec. Further work will be required for proper integration of bindings at buildtime.

### User Impact
This RFC does not require users to take any action. They remain free to use any other mechanism provided by Kubernetes to discover and/or inject credentials into their workloads. riff resources like Deployers and Processors currently contain a PodSpec where custom volumes, environment variables and arguments may be defined. Individual riff resources may start to produce and/or consume bindings, which users may then consume.

While a user may continue to use lower-level Kubernetes idioms, a higher-level experience can provide users a simpler and more maintainable experience.

### Backwards Compatibility and Upgrade Path
There are no direct backwards compatibility concerns as riff does not currently provide any support for bindings. Cloud Foundry services are exposed to applications via the `VCAP_SERVICES` environment variable which is a different structure than the CNB Bindings spec.

It is expected that the bindings space will actively evolve within the Kubernetes ecosystem. As the space matures this RFC will almost certainly be superseded to align with the broader community.

## FAQ
*How do functions and applications consume the binding?*

It's anticipated that frameworks and libraries will manage the binding within the container without the developer needing to directly read from the filesystem. This RFC has no opinion how a container should consume the binding metadata or secret. At this time, there are no known such libraries. It would be logical for riff function invokers to adopt an appropriate library to provide a higher level experience to function authors.

*Does the binding name have any meaning?*

Maybe. The name of the binding appears in the filesystem as the directory containing the `metadata` and `secret` directories that comprise the binding. A workload may take the name of a binding into consideration when consuming the binding, or it may ignore it. The binding name is the only aspect of the binding that is under the control of the consumer. All other elements of the binding are defined by the binding producer. Additional metadata may be injected into the workload about the binding out of band, but is not part of this RFC.

*How do consumers discover bindings?*

This RFC does not prescribe a mechanism for discovering bindings. It is assumed that the consumer has foreknowledge. Another RFC may define a higher level experience.

*How do CNB Bindings relate to a [Service Catalog ServiceBinding](https://svc-cat.io/docs/resources/#servicebinding)?*

Service Catalog is a client to an Open Service Broker. A Service Catalog ServiceBinding is a mechanism to bind a provisioned "service" to the cluster by exposing the credentials as a Secret. CNB Bindings have no mechanism to create "service" instances, instead, its focus is binding metadata and secrets to a workload. The output of the ServiceBinding is not directly consumable by CNB Bindings. A mechanism to convert a ServiceBinding to a format consumable by CNB Bindings is out of scope for this RFC.
