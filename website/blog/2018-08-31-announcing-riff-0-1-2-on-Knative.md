---
title: "Announcing riff v0.1.2 on Knative"
---

[riff v0.1.2](https://github.com/projectriff/riff/releases/tag/v0.1.2) on Knative is now available.  
Many thanks to all the riff and Knative contributors.

<!--truncate-->

## install
We recommend installing riff v0.1.2 on a fresh Kubernetes cluster. The riff CLI can be downloaded from our [releases page](https://github.com/projectriff/riff/releases/tag/v0.1.2) on GitHub. Please follow one of the [getting started](/docs) guides, to create a new cluster on GKE or minikube. Remember that you can also use the CLI to uninstall everything.

#### uninstall knative and istio without prompting
```sh
riff system uninstall --istio --force
```

#### install on minikube (for GKE omit `--node-port`)
```sh
riff system install --manifest stable --node-port
```

Riff will install "stable" release builds of Knative serving, eventing, and build components. You can opt to install the latest nightly builds using the new `--manifest latest` option.

Remember that after installing you also need to install credentials for builds to push to a docker registry, and configure a namespace. See the [getting started docs](/docs) for more details

#### install dockerhub push credentials and initialize the default namespace
```sh
kubectl apply -f dockerhub-push-credentials.yaml
riff namespace init default --secret push-credentials
```


## feedback during builds
There are two new options to provide feedback during builds. In previous releases, `riff function create` would return immediately after creating the build resources, without waiting for the build to succeed or fail.

- `--wait` or `-w` waits until the build status is known
- `--verbose` or `-v` is like wait, but also relays logs to the output

For example, if you have not [initialized](/docs/getting-started-with-knative-riff-on-minikube/#initialize-the-namespace) the default namespace, creating a function with `--wait` will produce an error message.

#### create square and push image to docker
```sh
riff function create node square \
  --git-repo https://github.com/trisberg/node-fun-square.git \
  --artifact square.js \
  --image $DOCKER_ID/node-fun-square:v1 \
  --wait
```
```
Error: function creation failed: RevisionMissing: Configuration "square" does not have any ready Revision.; Revision creation failed with message: "Internal error occurred: admission webhook \"webhook.build.knative.dev\" denied the request: mutation failed: serviceaccounts \"riff-build\" not found".
```

Using `--verbose` instead of `--wait` is useful in situations where the cause of an error can be deduced from container logs.

## function chaining over channels

[Knative/eventing](https://github.com/knative/eventing/pull/325) can now relay the reponse of a function to a another channel. This can be configured through the riff CLI by specifying the output channel with `--output` or `-o` when creating a subscription.

![function chain](/img/function-chain.png)

In this example, we will build a chain consisting of 3 functions and 2 channels. The random function posts numbers to the square function via the numbers channel, which forwards the output of square to the squares channel, for processing by the hello function. 

To keep things interesting, we'll build the 3 functions in 3 different ways.
- **square** uses a Knative build, pulling source code from github, and pushing the image to a registry.
- **hello** is built manually with the docker CLI and a Dockerfile.
- **random** uses a prebuilt image pulled from dockerhub.

## square function
The command below runs an in-cluster build of the square function, and pushes the image to dockerhub. For gcr, replace $DOCKER_ID with your gcr.io/$GCP_PROJECT.

#### create square and push image to docker
```sh
riff function create node square \
  --git-repo https://github.com/trisberg/node-fun-square.git \
  --artifact square.js \
  --image $DOCKER_ID/node-fun-square:v1 \
  --wait
```

#### use watch to monitor pods
```sh
watch -n 1 kubectl get pod --all-namespaces
```
For the first build, you may see the `square-00001-xxxx` build pod show a status of `Init:0/4` for several minutes. Once built, the square function will show up as a pod called `square-00001-deployment-xxxxxxxxx-xxxxx`.

When the square function is running, you should be able to invoke it.
```sh
riff service invoke square -- -w '\n' \
  -H 'Content-Type: text/plain' \
  -d 7
```

## hello function
In this case, we'll start with a javascript function and a Dockerfile in a directory. Notice that the function logs its ouput in addition to returning it. We'll use this to monitor the output of the function chain. 

#### hello.js  
```js
module.exports = x => {
  var out = 'hello ' + x
  console.log(out)
  return out
}
```

#### Dockerfile  
```dockerfile
FROM projectriff/node-function-invoker:0.0.8
ENV FUNCTION_URI /functions/hello.js
ADD hello.js ${FUNCTION_URI}
```

Build the function image, and use it create a Knative Service.

#### build locally for minikube
```sh
eval $(minikube docker-env)
```
```sh
docker build -t dev.local/hello:v1 .
riff service create hello --image dev.local/hello:v1
```
The `dev.local` prefix tells Knative to use the local docker daemon instead of pulling an image from a remote container registry.

#### build for dockerhub
```sh
docker build -t $DOCKER_ID/hello:v1 .
docker push $DOCKER_ID/hello:v1
riff service create hello --image $DOCKER_ID/hello:v1
```
For gcr, replace $DOCKER_ID with your gcr.io/$GCP_PROJECT.

Using a tool like kail makes it easy to watch the function container log.

#### start kail in a separate terminal window
```sh
kail -d hello-00001-deployment -c user-container
```

#### invoke hello
```sh
riff service invoke hello -- -w '\n' \
  -H 'Content-Type: text/plain' \
  -d riff
```

#### kail output
```
default/hello...[user-container]: hello riff
```

## random function
We have published an image on [dockerhub](https://hub.docker.com/r/jldec/random/tags/) for the random function. The source can be found on [GitHub](https://github.com/jldec/random). This function posts random numbers between 0 and 999 to a channel or to another function.

Create the random function using the image from dockerhub.
```sh
riff service create random --image jldec/random:v0.0.2
```

Invoke the function to send posts to hello.
```sh
riff service invoke random -- -w '\n' \
  -H 'Content-Type:application/json' \
  -d '{"url":"http://hello.default.svc.cluster.local"}'
```

The kail log of the hello function from above should show the numbers as they are generated
```
default/hello...[user-container]: hello riff
default/hello...[user-container]: hello 315
default/hello...[user-container]: hello 980
default/hello...[user-container]: hello 122
default/hello...[user-container]: hello 891
```

## wiring everything together

#### create the numbers and squares channels
```sh
riff channel create numbers --cluster-bus stub
riff channel create squares --cluster-bus stub
```

#### create two subscriptions.
```sh
riff service subscribe square --input numbers --output squares
riff service subscribe hello --input squares
```

#### configure the random function to post to the numbers channel.
```sh
riff service invoke random -- -w '\n' \
  -H 'Content-Type:application/json' \
  -d '{"url":"http://numbers-channel.default.svc.cluster.local"}'
```

Now the hello function should show the output of square and hello chained together.
```
default/hello...[user-container]: hello 549081
default/hello...[user-container]: hello 88804
default/hello...[user-container]: hello 786769
default/hello...[user-container]: hello 1225
default/hello...[user-container]: hello 525625
```
