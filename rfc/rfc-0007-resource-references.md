# RFC-0007: Resource References

**Authors:** Scott Andrews

**Status:**

**Pull Request URL:** link

**Superseded by:** N/A

**Supersedes:** N/A

**Related:** [RFC 0006: Guidance for child resource names](https://github.com/projectriff/riff/pull/1369)


## Problem
It's common for one resource to reference another. For example, a Deployer can reference a Function build to obtain the latest image, or a stream provider references the gateway Service it created. riff should have a common way to express relationships between Kubernetes resources.

### Anti-Goals
This RFC is not define how specific riff resources reference other resources, it is only providing guidance.

## Solution
The Kubernetes API project provides two mechanisms for referencing other resources. ObjectReference and LocalObjectReference.

[ObjectReference](https://godoc.org/k8s.io/api/core/v1#ObjectReference) has full support for referencing any k8s resource, commonly by apiGroup, kind, namespace, and name.

[LocalObjectReference](https://godoc.org/k8s.io/api/core/v1#LocalObjectReference) contains only a resource's name. The apiGroup, kind and namespace must be inferred.

riff should favor using ObjectReferences when the kind or namespace may vary or is unknown. LocalObjectReferences should be used when only the name of the resource will vary. `nil` should be replace reference if the relationship is not defined.

### User Impact
Other tools in the k8s ecosystem are more likely to understand and able to traverse (Local)ObjectReferences than ad hoc reference. Replacing a string with a LocalObjectReference incurs a trivial performance overhead to marshal an object instead.

### Backwards Compatibility and Upgrade Path
Changing existing ad hoc references to use a (Local)ObjectReference is a breaking change. The impact of such a change should be considered before implementation.

While the riff APIs are in alpha breaking changes without incrementing the API version are permissible.

## FAQ
*Answers to common questions that may arise and those that you’ve commonly been asked after requesting comments for this proposal.*
