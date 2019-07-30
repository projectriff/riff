---
title: "Announcing riff v0.2.0"
---

We are happy to announce the release of [riff v0.2.0](https://github.com/projectriff/riff/releases/tag/v0.2.0). Thank you once again, all riff, Buildpacks, and Knative contributors.

<!--truncate-->

The riff CLI can be downloaded from our [releases page](https://github.com/projectriff/riff/releases/tag/v0.2.0) on GitHub. Please follow one of the [getting started](/docs) guides, to create a new cluster on GKE or minikube.

Notable changes in this release include:
- All builds now use [Buildpacks](https://buildpacks.io/)
- No more special-case builds of images starting with `dev.local`
- `riff function create`
  - no `<invoker>` argument before the `<name>` argument
  - new optional `--invoker` flag
- `riff function build` has been renamed `riff function update`
- `riff service revise` has been renamed `riff service update`
- `riff namespace init` has a new `--no-secret` flag

### Buildpacks everywhere!
This release extends the use of buildpacks across all of our currently supported invokers: Java, JavaScript, and Command. 

Here is a map of the buildpack-related repos on Github.

![](/img/builders.svg)

- [riff builder](https://github.com/projectriff/riff-buildpack-group) creates the `projectriff/builder` container. 
- [riff buildpack](https://github.com/projectriff/riff-buildpack) contributes invokers for running functions
  - [Node invoker](https://github.com/projectriff/node-function-invoker) runs JavaScript functions 
  - [Java invoker](https://github.com/projectriff/java-function-invoker) runs Java functions
  - [Command invoker](https://github.com/projectriff/command-function-invoker) runs command functions
- [OpenJDK buildpack](https://github.com/cloudfoundry/openjdk-buildpack) contributes OpenJDK JREs and JDKs
- [Build System buildpack](https://github.com/cloudfoundry/build-system-buildpack) performs Java based builds
- [NodeJS buildpack](https://github.com/cloudfoundry/nodejs-cnb) contributes node.js runtime
- [NPM buildpack](https://github.com/cloudfoundry/npm-cnb) performs npm based builds

### Function Create
 
Since buildpacks do detection, we have simplified what used to be:  
 `riff function create <invoker> <name>`

The new syntax is:  
 `riff function create <name>`

A new `--invoker <invokername>` flag allows overrides.

### Detection Logic

* The presence of a `pom.xml` or `build.gradle` file will trigger compilation and building of an image for running a [Java function](https://github.com/projectriff/java-function-invoker).
* A `package.json` file or an `--artifact` flag pointing to a `.js` file will build the image for running a [JavaScript function](https://github.com/projectriff/node-function-invoker).
* An `--artifact` flag pointing to a file with execute permissions will generate an image for running a [Command function](https://github.com/projectriff/command-function-invoker).

For example, say you have a directory containing just one file, `wordcount.sh`:
```sh
#!/bin/bash

tr ' ' '\n' | sort | uniq -c | sort -n
```

Make the file executable to use it as a command function.
```sh
chmod +x wordcount.sh
```

Call `riff function create` providing the name of the image. E.g. with your DockerHub repo ID.
```sh
riff function create wordcount \
  --local-path . \
  --artifact wordcount.sh \
  --image $DOCKER_ID/wordcount:v1
```

When the function is running:
```sh
riff service invoke wordcount --text -- \
  -d 'yo yo yo version 0.2.0' \
  -w '\w'
```
```
      1 0.2.0
      1 version
      3 yo
```

##### Extra!
Try running the following command to invoke your wordcount function on something a little more interesting. 
```sh
curl -s https://www.constitution.org/usdeclar.txt  \
 | riff service invoke wordcount --text -- -d @-
```

##### Build from GitHub
For in-cluster builds using a GitHub repo, e.g. in a hosted riff environment, replace the `--local-path .` with `--git-repo <url>`.

```sh
riff function create wordcount \
  --git-repo https://github.com/projectriff-samples/command-wordcount \
  --artifact wordcount.sh \
  --image $DOCKER_ID/wordcount \
  --verbose 
```

### No more dev.local

Note that we have removed support for the special `dev.local` image name prefix for local builds. All image names need to be prefixed with a registry, and you'll need to configure your riff namespace with credentials to push images to that registry. A new `--no-secret` flag has been added to `riff namespace init` if your registry does not require authentication.

For more details please see the help for `riff namespace init -h` or refer to one of the [Getting Started Guides](/docs).