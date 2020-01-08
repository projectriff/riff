# RFC-#: Drop Core Runtime

**Authors:** @scothis

**Status:**

**Pull Request URL:** [#1373](https://github.com/projectriff/riff/pull/1373)

**Superseded by:** N/A

**Supersedes:** N/A

**Related:** N/A

## Problem

riff currently has two ways to deploy HTTP triggered workload, the Core and Knative runtimes. This dichotomy causes users extra cognitive load. For every operation they have to think about which runtime they are targeting. Moreover, the capabilities and behavior of each runtime are subtly different.

As the Knative runtime is a super-set of the Core runtime, we should consolidate on the Knative runtime. Users who don't want to run Knative, can take a riff built container image and deploy it however they like.

## Solution

Remove the Core runtime from riff System and CLI.

As a half measure, we could deprecate the Core runtime in riff 0.5, remove documentation from projectriff.io and hide the CLI commands. With full removal happening in riff 0.6.

### User Impact

Users would no longer be able to deploy workloads to the Core runtime. Of course a user could take a built image and create their own Deployment and Service to provide the equivalent of the Core runtime.

### Backwards Compatibility and Upgrade Path

This is a breaking change, users will need to create the equivalent Knative runtime Deployer.

## FAQ

*none so far*
