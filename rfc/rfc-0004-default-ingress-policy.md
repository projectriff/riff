# RFC-0004: Default Ingress Policy

**Authors:** @scothis

**Status:** Accepted

**Pull Request URL:** [#1366](https://github.com/projectriff/riff/pull/1366)

**Superseded by:** N/A

**Supersedes:** N/A

**Related:** N/A


## Problem
Core and Knative runtime `Deployer`s are exposed outside the cluster via ingress by default. Especially for functions, workloads may not be designed to be exposed to a hostile network. riff should strive to provide secure defaults that can be intentionally overridden to expose functionality and accept risk.

### Anti-Goals
This RFC does not define a new Ingress Policy, or resource.

## Solution
The current default Ingress Policy is `External`, this RFC changes that default to `ClusterLocal`. ClusterLocal ingress exposes the workload to traffic that is already inside the cluster, while External exposes the workload to traffic outside of the cluster. 

This default is applied in the system's admission webhook and via the CLI. Both of those defaults should change.

This change should be rolled out consciously as it may break existing users and FATS tests:
1. expose the ingress policy via the riff System and CLI (complete)
1. define an explicit ingress policy everywhere that currently consumes the default
1. change the default for riff System
1. change the default for riff CLI

### User Impact
Users creating new Deployers will either need to accept that the workload is not exposed outside the cluster, or they will need to add `--ingress-policy External` to the creation command.

The Knative runtime Adapter resource is not changing as part of this RFC as it does not create a Knative workload. The ingress policy of the adapted resource is preserved.

### Backwards Compatibility and Upgrade Path
While this is a breaking change from main for both the Core and Knative runtimes, this is restoring the existing behavior from riff 0.4 for the core runtime.

## FAQ
**Is ClusterLocal really "ingress"?**

Maybe. While the core runtime will create an Ingress resource - or not -depending on the ingress policy, the Knative runtime (at least with Istio) defines an ingress gateway and a cluster-local gateway. The workload is always registered on the cluster-local gateway and selectively registered on the ingress gateway.
