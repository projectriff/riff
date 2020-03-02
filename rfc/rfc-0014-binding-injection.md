# RFC-0014: Binding Injection

**Authors:** @scothis

**Status:**

**Pull Request URL:** [#1382](https://github.com/projectriff/riff/pull/1382)

**Superseded by:** N/A

**Supersedes:** N/A

**Related:** [RFC 0002 Bindings](./rfc-0002-bindings.md), [RFC 0008 Pod Templates](./rfc-0008-pod-templates.md)


## Problem

[RFC 0002 Bindings](./rfc-0002-bindings.md) defines a interface for a resource to expose metadata and secrets to be consumed by another resource. It leverages the [CNB Binding](https://github.com/buildpack/spec/blob/master/extensions/bindings.md) spec for how the service metadata and secrets are exposed inside a container. While a binding can be consumed manually, it's difficult to manage as it requires defining two volumes, and two volume mounts and a environment variable per container.

### Anti-Goals

This RFC has no opinion as to the type of binding being injected, or the consumer of the binding.

## Solution

[Knative Bindings](https://docs.google.com/document/d/1t5WVrj2KQZ2u5s0LvIUtfHnSonBv5Vcv8Gl2k5NXrCQ/edit#heading=h.lnql658xmg9p) [Knative membership required to view] (yea, the term is overloaded) define an abstract way for reconcilers to operate on resources by duck typing known elements of the resource without requiring pre-knowledge or full type knowledge at runtime. By leveraging resources that use a PodTemplateSpec, it's possible to inject volumes, volume mounts and environment variables into an existing resource (even if owned and controlled by another resource). With [RFC 0008 Pod Templates](./rfc-0008-pod-templates.md) the riff Deployers and Processor resources implement the PodSpecable duck type.

A duck type for exposing a binding is defined in RFC 0002. While not as wide spread, the Stream resource currently implements this duck type. Combining the Binding and PodSpecable duck types allows for a generic means to inject binding metadata and secrets into a consuming workload. A new CRD can express this injection:

```yaml
apiVersion: bindings.projectriff.io/v1alpha1
kind: Injection
metadata:
  name: my-binding-injection
spec:
  binding:
    apiVersion: streaming.projectriff.io/v1alpha1
    kind: Stream
    name: my-stream
  containers:
  - user-container
  subject:
    apiVersion: knative.projectriff.io/v1alpha1
    kind: Deployer
    name: my-deployer
status:
  observedGeneration: 1
  conditions: []
```

- `.spec.binding` is an ObjectReference to a resource implementing the binding duck type
- `.spec.containers` optionally restricts which containers are augmented, defaults to all containers
- `.spec.subject` is an ObjectReference to a resource implementing the podspecable duck type
- the subject and binding target much be in the same namespace as the injection resource

The subject resource will be injected with the binding metadata and secret.

```yaml
apiVersion: streaming.projectriff.io/v1alpha1
kind: Stream
metadata:
  name: my-stream
spec: {}
status:
  binding:
    metadataRef:
      name: my-stream-binding-metadata
    secretRef:
      name: my-stream-binding-secret
```

```yaml
apiVersion: knative.projectriff.io/v1alpha1
kind: Deployer
metadata:
  name: before
spec:
  template:
    spec:
      containers:
      - name: user-container
        image: square
```

```yaml
apiVersion: knative.projectriff.io/v1alpha1
kind: Deployer
metadata:
  name: after
spec:
  template:
    spec:
      containers:
      - name: user-container
        image: square
        env:
        - name: CNB_BINDINGS
          value: /var/riff/bindings
        volumeMounts:
        - mountPath: /var/riff/bindings/my-binding-injection/metadata
          name: my-binding-injection-metadata
          readOnly: true
        - mountPath: /var/riff/bindings/my-binding-injection/secret
          name: my-binding-injection-secret
          readOnly: true
    volumes:
    - name: my-binding-injection-metadata
      configMap:
        name: my-stream-binding-metadata
    - name: my-binding-injection-secret
      secret:
        secretName: my-stream-binding-secret
```

Notes:
- if the `CNB_BINDINGS` env var is already defined, it must be respected
  - otherwise it must be defined
- volume names must not collide with an existing volume
- volume mount paths must be under the CNB_BINDINGS value
- the mount directory within CNB_BINDINGS must be unique

### User Impact

Users will be able to define and maintain bindings to workloads. Higher level resources can provide more opinionated means to connect bindings to workloads, for example, the Processor resource consumes Streams directly.

### Backwards Compatibility and Upgrade Path

This is a new resource that does not modify any existing resources.

## FAQ
*Answers to common questions that may arise and those that you’ve commonly been asked after requesting comments for this proposal.*
