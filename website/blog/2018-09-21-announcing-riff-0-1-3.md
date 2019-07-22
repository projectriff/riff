---
title: "Announcing riff v0.1.3"
---

We are pleased to announce the release of [riff v0.1.3](https://github.com/projectriff/riff/releases/tag/v0.1.3). Thank you riff and Knative contributors.

<!--truncate-->

The riff CLI can be downloaded from our [releases page](https://github.com/projectriff/riff/releases/tag/v0.1.3) on GitHub. Please follow one of the [getting started](/docs) guides, to create a new cluster on GKE or minikube. To update an existing riff install first run `riff system uninstall --istio`. This release includes new manifests for the latest Knative and Istio.

Here's an overview of some of the new features in riff v0.1.3:

## credential helpers
Initializing a namespace with your push credentials is easier now. 

#### point to your [GCR key](/docs/getting-started-with-knative-riff-on-gke/#create-a-kubernetes-secret-for-pushing-images-to-gcr)
```sh
riff namespace init default --gcr <path-to-json-file>
```

#### or enter the password for your $DOCKER_ID when prompted
```sh
riff namespace init default --dockerhub $DOCKER_ID
```

## riff buildpack for java
This release supports building java functions from source, either locally or on-cluster. Both variants use a new riff buildpack for java. 

> NOTE: to preserve the old behavior of building containers with a pre-compiled jar file, use `riff function create jar`. 

All you need in your directory is the code with a maven pom, and the name of the handler class in a file called `riff.toml`. The example below uses a sample [java-hello](https://github.com/projectriff-samples/java-hello) function available on GitHub.

#### riff.toml
```toml
handler = "functions.Hello"
```

To build from code in a directory and push to local docker:

```sh
riff function create java hello \
  --local-path . \
  --image dev.local/java-hello:v1
```
Using a `--local-path` builds code directly from your machine. The `dev.local` prefix exports the image to your docker environment. Remember to run `eval $(minikube docker-env)` for minikube.


> NOTE: pre-existing images with tags matching `--image` will result in an error "Reading information from previous image for possible re-use". Remove those images first.

You can iterate on your code by rebuilding locally, triggering a new Knative Revision for each build.
```sh
riff function build hello --local-path path/to/function/source
```
To build from code on GitHub and push to DockerHub: 
```sh
riff function create java hello \
    --git-repo https://github.com/projectriff-samples/java-hello.git \
    --image $DOCKER_ID/java-hello \
    --verbose
```
Using `--verbose` shows the progress of the build as it's happening in the cluster. For GCR, replace `$DOCKER_ID` with your `gcr.io/$GCP_PROJECT`. 

## simpler riff service invoke 
You can now call `riff service invoke` with `--text` or `--json` to set the `Content-Type` header.

#### invoke the hello function with text input
```sh
riff service invoke hello --text -- -w '\n' -d world
```

## riff subscription commands
Subscriptions now have their own separate CLI commands. The corresponding options on `riff function` and `riff service` have been removed.

#### create
```
Usage:
  riff subscription create [SUBSCRIPTION_NAME] [flags]

Examples:
  riff subscription create --channel tweets --subscriber tweets-logger
  riff subscription create my-subscription --channel tweets --subscriber tweets-logger
  riff subscription create --channel tweets --subscriber tweets-logger --reply-to logged-tweets

Flags:
  -c, --channel string      the input channel of the subscription
  -r, --reply-to string     the optional output channel of the subscription
  -s, --subscriber string   the subscriber of the subscription
```

#### delete
```
Usage:
  riff subscription delete SUBSCRIPTION_NAME [flags]

Example:
  riff subscription delete my-subscription --namespace joseph-ns
```

#### list
```
Usage:
  riff subscription list [flags]
```