# RFC-0008: Pod Templates

**Authors:** Scott Andrews

**Status:**

**Pull Request URL:** [#1371](https://github.com/projectriff/riff/pull/1371)

**Superseded by:** N/A

**Supersedes:** N/A

**Related:** N/A


## Problem
Many Kubernetes resources that decompose to Pods (like Deployment, ReplicaSet, Job, etc) embed a [PodTemplateSpec](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.11/#podtemplatespec-v1-core) at `.spec.template`. This not only makes these resources easier for users and tools to consume, it facilitates duck typed operations on resources that follow this pattern. riff doesn't follow this pattern, as our resources embed a [PodSpec](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.11/#podspec-v1-core) at `.spec.template` instead. The PodTemplateSpec contains inlined ObjectMeta in addition to the PodSpec; this enables adding custom labels and/or annotations to the resulting Pod.

## Solution
riff should switch all usage of PodSpec at `.spec.template` to be a PodTemplateSpec.

### User Impact
This is a breaking change. Users will need to move the content of `.spec.template` to `.spec.template.spec`.

### Backwards Compatibility and Upgrade Path
This is a breaking change for our CRDs. The impact can be minimized if the CLI is updated at the same time as the system and charts. No automatic migration will be performed.

Since our APIs are still alpha breaking changes are permitted without a version bump.

## FAQ
*Answers to common questions that may arise and those that you’ve commonly been asked after requesting comments for this proposal.*
