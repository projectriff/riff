# RFC-0014: Source Code Upload

**Authors:** Emily Casey

**Status:**

**Pull Request URL:** [#1374](https://github.com/projectriff/riff/pull/1374)

**Superseded by:** N/A

**Supersedes:** N/A

**Related:** N/A


## Problem
The way we currently handle local source is less than ideal in a few ways:
1. Images built from local source will not get autmatically rebased or rebuilt based on stack or buildpack updates,
 and are therefore susceptible to more security vulnerability
1. Because build from source uses `pack`, it requires users to have a docker daemon
1. It requires users to wait for large builder images to pull
1. Results produced by pack and kpack should be identical but difference introduces risk

## Solution

#### Registry Type SourceResolver
When a user specifies the `--local-path <dir>` flag during `riff function create` or `riff application create` the source
code in the specified directory will be packaged as a OCI image containing a single layer and uploaded to a docker registry.
`riff` will create a `kpack` `image` with a `registry` type `sourceresolver` https://github.com/pivotal/kpack/blob/master/docs/image.md#source-configuration.
`kpack` can then build the image as it does in the git source type case.

#### Source Image Tag
This will require a repository to which riff can upload the source image. By default `riff` can upload the source image to a predictable 
derivation of the target image tag, by appending a suffix to the tag. For example if the generated image will be published to `docker.io/my/app:latest`,
the source image will be uploaded to `docker.io/my/app:latest-source`

The following flags will allow for configurability:

* `--source-image repository`

#### Source Image Push Auth
When pushing the source image `riff` will use the default docker keychain and will not try to use the registry secrets configured in `riff`

#### Source Image Pull Auth
Credentials for the source image must be present in the riff service account so that kpack can pull the source code

#### Include/Exclude
`riff` will respect include/exclude configuration from a [CNB project descriptor](https://github.com/buildpacks/rfcs/blob/master/text/0019-project-descriptor.md#buildinclude-and-buildexclude)
if present 

### User Impact

#### Pros
By making builds from local source behave more similarly to builds from a git source the behavior of `riff` will be less
surprising to users. Users will also be able to receive CVE fixes for images built from local source through a declarative
model, rather than requiring an imperative rebuild or rebase like the current model.

#### Cons
Upload source introduces risks, especially if the include/exclude model is opaque or unexpected. Users may accidentally upload
files containing secrets, ignored by git.

Defining source tags and the associated credentials adds complexity to the `--local-path` case.

### Backwards Compatibility and Upgrade Path

## FAQ
