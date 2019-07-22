---
title: "Announcing riff v0.3.0"
---

We are happy to announce the release of [riff v0.3.0](https://github.com/projectriff/riff/releases/tag/v0.3.0). Thank you all riff, Buildpacks, and Knative contributors.

<!--truncate-->

The riff CLI can be downloaded from our [releases page](https://github.com/projectriff/riff/releases/tag/v0.3.0) on GitHub. Please follow one of the [getting started](/docs) guides, to create a new cluster on GKE, Minikube, Docker Desktop for Mac, or Docker Desktop for Windows.

## Notable changes

### Cloud Native Buildpacks

- update to [pack v0.1.0](https://github.com/buildpack/pack/releases/tag/v0.1.0)
- [builder](https://github.com/projectriff/builder/blob/v0.2.0/builder.toml) uses separate buildpack per function invoker 
- The Java buildpack includes support for Java 11.  Please add a source version of Java 8 or later to your Maven pom or Gradle build.

  #### Maven
  ```xml
  <properties>
    <maven.compiler.source>1.8</maven.compiler.source>
    <maven.compiler.target>1.8</maven.compiler.target>
  </properties>
  ```

  #### Gradle
  ```
  sourceCompatibility = 1.8
  ```

### Knative

- update to [Knative Serving v0.5.1](https://github.com/knative/serving/releases/tag/v0.5.1) and Istio [v1.0.7](https://github.com/istio/istio/releases/tag/1.0.7)
- update to [Knative Build v0.5.0](https://github.com/knative/build/releases/tag/v0.5.0)
- update to [Knative Eventing v0.4.0](https://github.com/knative/eventing/releases/tag/v0.4.0)

### riff CLI

- support for Windows
- easy install via [brew](https://formulae.brew.sh/formula/riff) and [chocolatey](https://chocolatey.org/packages/riff/0.3.0)
- `riff system install` and `riff system uninstall` with improved robustness
- `riff namespace init` with basic auth for alternative registries
- `riff namespace cleanup` similar to system uninstall, but for a namespace
- `riff function build` to build an image without creating a service
- `riff function create`
    - `--sub-path` for `--git-repo` builds from a subdirectory
    - `--image` auto-inferred from registry prefix and function name
    - `riff.toml` optional alternative to CLI parameters
- `riff service invoke` with improved error handling
- `riff channel create` will automatically use a default channel provisioner

## More modular Buildpacks
This release introduces a new buildpacks structure with a buildpack per language. The modular approach will help restore support for [custom language invokers](https://github.com/projectriff/riff/issues/1093) in a future riff release.
Here is an updated map of the buildpack-related repos on Github.

![](/img/builders2.svg)

[riff Builder](https://github.com/projectriff/builder) is the container for function builds using buildpacks.

#### Java group
- [OpenJDK buildpack](https://github.com/cloudfoundry/openjdk-buildpack): contributes OpenJDK JREs and JDKs
- [Build System buildpack](https://github.com/cloudfoundry/build-system-buildpack): performs Java based builds
- [Java function buildpack](https://github.com/projectriff/java-function-buildpack): contributes [Java invoker](https://github.com/projectriff/java-function-invoker)

#### Node group
- [NodeJS buildpack](https://github.com/cloudfoundry/nodejs-cnb): contributes node.js runtime
- [NPM buildpack](https://github.com/cloudfoundry/npm-cnb): performs npm based builds
- [Node function buildpack](https://github.com/projectriff/node-function-buildpack): contributes the [Node  invoker](https://github.com/projectriff/node-function-invoker) for running JavaScript functions

#### Command group
- [Command function buildpack](https://github.com/projectriff/command-function-buildpack): contributes the [Command invoker](https://github.com/projectriff/command-function-invoker) for running Linux commands.


## Our plans for stream processing

Since we anticipate replacing the existing Channels and Subscriptions with new Stream and Processor resources, aligned with stream-oriented Function Invokers, we are [deprecating](https://github.com/projectriff/riff/pull/1237) the use of Channel and Subscription resources in this release.

You can follow our progress in this [roadmap for serverless stream processing](https://github.com/projectriff/riff/issues/1159).
