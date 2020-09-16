# RFC-0015: Unopinionated Build

**Authors:** @scothis

**Status:** Draft

**Pull Request URL:** link

**Superseded by:** N/A

**Supersedes:** N/A

**Related:** N/A


## Problem

`Function` and `Application` builds in project riff are currently an opinionated wrapper around kpack. A fixed function and application builder image is shipped with riff, that is not easy to change. If a user wants to add a new language, or a different version of any buildpack they need to rebuild the builder image and install it into the cluster. This custom builder will be overwritten the next time riff is upgraded. Some users will prefer to use a commercial build service, or an alternative oss build.

riff Build was never intended to be *the* sole mechanism to build riff functions, but to provide an opinionated path. 

### Anti-Goals

We will not ship an alternative to riff Build, but should show through examples how tools like kpack can build riff functions that are deployed to Knative and the riff Streaming runtime.

## Solution

Project riff should no longer provide an opinionated build. Users should be free to use whatever build toolchain they prefer.

The `Function` and `Application` resource will be deprecated and removed. References to these resources in other resources will also be deprecated and removed.

The `Container` resource currently in build should move beside the `ImageBinding` resource. References to this resource (e.g. `Deployer`s and `Processor`s)  will be deprecated and removed.

### User Impact

This is a major change to the scope of riff. We will still provide function buildpacks and a function invoker, but it will be up to the user to choose when and how to consume those capabilities.

Changes to kpack between v0.0.x and v0.1.x would fundamentally change the riff Build flow. So a breaking change within riff Build is inevitable. We’re taking this opportunity to refine the scope of riff and increase developer choice.

### Backwards Compatibility and Upgrade Path

The existing riff Build behavior can be implemented directly with kpack and the `ImageBinding` resource. Users who don’t want to use kpack can substitute any other mechanism for building a container image.

There is no direct upgrade path.

## FAQ
