# RFC-#: Child Resource Names

**Authors:** Scott Andrews

**Status:**

**Pull Request URL:** link

**Superseded by:** N/A

**Supersedes:** N/A

**Related:** N/A


## Problem
When a controller creates a new resource, it need to either pick a static name, or let kubernetes generate a name. Static names are predictable, but may already be in use. Generated names cannot be predicated, but are guaranteed to be available.

### Anti-Goals
This RFC does not prescribe which naming scheme to use for any specific resource.

## Solution
While there will not be a cut and dry answer for each situation, there are heuristics to guide the decision:
- if a resource is not consumed directly by a non-controlled resource by name, generate the name (e.g. Pods in a ReplicaSet)
- if the resource can be queried via a label selector instead of by name, generate the name (e.g. Pods fronted by a Service)
- if a resource's name surfaces in outward facing way that is part of the public API, consider using a static name (e.g. Services in DNS)
- if the name needs to be stable between different instances of the same resource, use a static name (e.g. immutable resources that cannot be updated)
- generate the name

Static names may use templates, but must be deterministic. based only on stable values of the parent resource (e.g. `sprintf('%s-suffix', current.name)`).

If a static name collides with an existing resource, the parent resource should reflect the collision in its conditions. 

The name generally should not be chosen by the creator of the parent resource, as it introduces complexity when reconciling updates for names changes as Kubernetes cannot move, or rename a resource.

### User Impact
No public interface is changing, as the names of child resources are reflected on the parent resource's status. That will continue to be true.

### Backwards Compatibility and Upgrade Path
riff generally uses generated names today, so current uses expect to use the name assigned by the system.

## FAQ
*Answers to common questions that may arise and those that you’ve commonly been asked after requesting comments for this proposal.*
