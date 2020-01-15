# RFC-0011: Streaming Gateways

**Authors:** @scothis, @markfisher, @ericbottard

**Status:**

**Pull Request URL:** [#1376](https://github.com/projectriff/riff/pull/1376)

**Superseded by:** N/A

**Supersedes:** N/A

**Related:** N/A


## Problem

The relationships between a `Stream` and a `(InMemory|Kafka|Pulsar)Provider` is fragile. The stream contains a hostname for the provider's provisioner, which the stream controller calls when reconciling the stream. This has a few problems:
- the `Stream` resource doesn't know when the `Provider` is unavailable, or deleted
- the `Stream` has no way to automatically obtain a new address for the provisioner
- the provisioner name is competing in the `Service` namespace. The likelihood of a name collision is reduced by appending a suffix to the service name. However, the user must include this suffix when creating each stream
- as new providers are created, there is a significant amount of reconciliation logic that is duplicated with trivial differences

The problem is derived from [projectriff/system#86](https://github.com/projectriff/system/issues/86).

### Anti-Goals

This RFC will not prevent naming collisions between different provisioners with the same name. The first gateway of a given name wins.

## Solution

1. Rename `*Provider` to `*Gateway`

   The high level role of the whole component is that of a gateway. 

1. Introduce a `Gateway` resource which creates a single `Deployment` and `Service`

   This reduces the toil of adding a new `*Gateway` reconciler as the boilerplate code will be encapsulated. It is recommend that users not create this resource directly.
   
   1. Cohabitate the gateway (liiklus) and provisioner `Deployment`s and `Service`s

      A single `Deployment` with two containers and a single `Service` with two ports can replace the current two Deployments and two Services per provider.

   1. Update each existing `*Gateway` to reconcile to a `Gateway`

1. Update `Stream` resource with a LocalObjectReference to its `Gateway`

   The `Stream` reconciler can track the `Gateway` and include the `Gateway`'s Ready condition as one of its conditions. `Processor`s already track each input and output `Stream`'s Ready condition to scale down the processor when a stream goes red.

1. [maybe] Pin a `Stream` to a `Gateway` instance

   `Stream`s are inherently stateful. A new `Gateway`, even if it has the same name, is unlikely to contain the same state. The gateway's UID could be saved on the stream and compared. if the UID changes, the stream's Ready condition could flip False to prevent usage of the stream.

1. Update Streaming CRDs status for new relationships

  Changes to the reconciliation object graph invalidate may of the existing status references.

### User Impact

The CRDs and CLI commands for streaming providers will replace `provider` with `gateway`. The stream create command will rename the `--provider` flag to `--gateway` with the value as the gateway's name (with no suffix needed).

### Backwards Compatibility and Upgrade Path

All existing behavior is being preserved, since the streaming runtime has not been released, there is no release to be backwards incompatible with.

## FAQ

none yet
