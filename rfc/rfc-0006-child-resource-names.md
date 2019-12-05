# RFC-0006: Child Resource Names

**Authors:** Scott Andrews

**Status:**

**Pull Request URL:** [#1369](https://github.com/projectriff/riff/pull/1369)

**Superseded by:** N/A

**Supersedes:** N/A

**Related:** N/A


## Problem
When a controller creates a new resource, it needs to either pick a static name or let Kubernetes generate a name. Static names are predictable, but may already be in use. Generated names cannot be predicated, but are guaranteed to be available. As the authors of reconcilers, we must define how each controlled resource is named.

### Anti-Goals
This RFC does not prescribe which naming scheme to use for any specific resource.

## Solution
While there will not be a cut and dried answer for each situation, there are heuristics to guide the decision:
1. if the name needs to be stable between different instances of the "same" resource, use a static name (e.g. immutable resources that cannot be updated instead they must be deleted and recreated)
1. if a resource is not consumed directly by a non-controlled resource by name, generate the name (e.g. Pods in a ReplicaSet)
1. if the resource can be queried- like via a label selector -instead of by name, generate the name (e.g. Pods fronted by a Service)
1. if a resource's name surfaces in an outward facing way that is part of the public contract, consider using a static name (e.g. DNS records for Services)
1. if none of the above criteria apply, generate the name

Static names may use templates, but must be deterministic, based only on stable values of the parent resource (e.g. `sprintf('%s-suffix', obj.name)`).

If a static name collides with an existing resource, the parent resource should reflect the collision in its conditions. 

The name generally should not be chosen by the creator of the parent resource, as it introduces complexity when reconciling updates for names changes as Kubernetes cannot move, or rename a resource. If the name of the child resource is essential, it should likely be a sibling resource rather than a child resource (e.g. Deployment and Service)

### User Impact
While the name of new child resources will change as this RFC is applied to each reconciler, no public interface is changing. The public contract for child resource is that the name of child resource is reflected on the parent resource's status. That will continue to be true.

## FAQ
*Answers to common questions that may arise and those that you’ve commonly been asked after requesting comments for this proposal.*
