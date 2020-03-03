# RFC-0014: Binding Injection

**Authors:** @scothis

**Status:**

**Pull Request URL:** [#1382](https://github.com/projectriff/riff/pull/1382)

**Superseded by:** N/A

**Supersedes:** N/A

**Related:** [RFC 0002 Bindings](./rfc-0002-bindings.md), [RFC 0008 Pod Templates](./rfc-0008-pod-templates.md)


## Problem

[RFC 0002 Bindings](./rfc-0002-bindings.md) defines a interface for a resource to expose metadata and secrets to be consumed by another resource. It leverages the [Cloud Native Buildpacks Binding](https://github.com/buildpack/spec/blob/master/extensions/bindings.md) (CNB bindings) spec for how the provisioned metadata and secret are exposed inside a container. While a CNB binding can be consumed manually, it's difficult to manage as it requires defining two volumes, and two volume mounts and a environment variable per container.

### Anti-Goals

This RFC has no opinion as to the semantic source, content, or consumer of the provisioned CNB binding,

## Solution

[Knative Bindings](https://docs.google.com/document/d/1t5WVrj2KQZ2u5s0LvIUtfHnSonBv5Vcv8Gl2k5NXrCQ/edit#heading=h.lnql658xmg9p) [Knative membership required to view] (yea, the term is overloaded) define an abstract way for reconcilers to operate on resources by duck typing known elements of a resource without requiring pre-knowledge or full type knowledge at runtime. By leveraging resources that use a PodTemplateSpec, it's possible to inject volumes, volume mounts and environment variables into an existing resource (even if owned and controlled by another resource). As of [RFC 0008 Pod Templates](./rfc-0008-pod-templates.md), shipped in v0.5, the riff Deployers and Processor resources implement the PodSpecable duck type.

A duck type for exposing a CNB binding is defined in RFC 0002. While not wide spread, the Stream resource currently implements this duck type. Combining the CNB binding and Knative PodSpecable duck types allows for a generic means to inject CNB binding metadata and secrets into a container workload. A new CRD can express this injection:

```yaml
apiVersion: bindings.projectriff.io/v1alpha1
kind: Injection
metadata:
  name: my-binding-injection
spec:
  provider:
    apiVersion: streaming.projectriff.io/v1alpha1
    kind: Stream
    name: my-stream
  subject:
    apiVersion: knative.projectriff.io/v1alpha1
    kind: Deployer
    name: my-deployer
  allowedContainers:
  - user-container
status:
  {}
```

- `.spec.provider` is an ObjectReference to a resource implementing the CNB binding duck type
- `.spec.subject` is an ObjectReference to a resource implementing the podspecable duck type
- `.spec.allowedContainers` optionally restricts which containers are augmented, defaults to all containers
- the `subject` and `provider` targets must be in the same namespace as the injection resource

The subject resource will be injected with the CNB binding's metadata and secret.

```yaml
apiVersion: streaming.projectriff.io/v1alpha1
kind: Stream
metadata:
  name: my-stream
spec: {}
status:
  # CNB binding duck type defined by riff RFC 0002
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
          value: /var/bindings
        volumeMounts:
        - mountPath: /var/bindings/my-binding-injection/metadata
          name: my-binding-injection-metadata
          readOnly: true
        - mountPath: /var/bindings/my-binding-injection/secret
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
- each container may only define the `CNB_BINDINGS` env var once, if already defined the value must be respected
- injected volume names must not collide with an existing volume
- injected volume mount paths must not collide with an existing volume
- injected volume mount paths must compile with RFC 0002

### User Impact

Users will be able to define and maintain workload mounting of provisioned CNB bindings. Higher level resources can provide more opinionated means, for example, the Processor resource consumes Streams directly, but uses CNB bindings under the hood.

### Backwards Compatibility and Upgrade Path

This is a new resource that leverages existing riff, Knative and k8s resources without modification.

The Knative Binding tool chain uses the Knative reconciler infrastructure which is not compatible with controller-runtime used by riff-system. Implementing the behavior defined by this RFC will require a new repo and deployment unit. The new repo and deployment unit can be shared with other Knative Bindings that emerge from riff.

## FAQ
*Answers to common questions that may arise and those that you’ve commonly been asked after requesting comments for this proposal.*
