# RFC-0016: Unopinionated HTTP Triggers

**Authors:** @scothis

**Status:**

**Pull Request URL:** [#1389](https://github.com/projectriff/riff/pull/1389)

**Superseded by:** N/A

**Supersedes:** N/A

**Related:** N/A


## Problem

riff currently has two HTTP based triggers: the Core and Knative runtimes. This creates complexity for users who need to pick a runtime and essentially forks guides with an "if knative ..., else if core ..." that is easy to mix up.

riff functions are able to handle both HTTP requests and streaming gRPC connections. The intent of the Core and Knative runtimes was to provide an opinionated way to run a function, but was never intended to be the only way to run a riff function. Offering an http runtime implicitly means riff could also run applications and pre-built containers, distracting from its stated purpose of being “for functions”.

The Core and Knative runtimes are both thin layers on top of other resources that expose many knobs and advanced capabilities that are not exposed on the `Deployer` resources. Moreover, riff did not offer configuration settings for essential production capabilities like vanity host names, external DNS or SSL.

### Anti-Goals

riff will not bundle a particular HTTP runtime. Functions will continue to accept HTTP requests and the built container images will continue to be runnable in other HTTP based container runtimes.

## Solution

We will deprecate and remove the riff Core and riff Knative runtimes. Users who want to deploy to Knative can use any distribution of Knative they like. Likewise, Core runtime users can create a Deployment and Service for their workload. The full power of each of those options will be available to users.

### User Impact

This is a major change in scope for riff and is a breaking change for users. 

### Backwards Compatibility and Upgrade Path

The riff Knative runtime `Deployer` can be replaced with a Knative `Service` and an `ImageBinding`. The riff Core runtime can be replaced with a `Deployment`, k8s `Service` (and optionally Ingress) and an `ImageBinding`.

There is no direct upgrade path.

## FAQ
