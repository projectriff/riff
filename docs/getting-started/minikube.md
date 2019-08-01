---
id: minikube
title: Getting started on Minikube
sidebar_label: Minikube
---

The following will help you get started running a riff function with Knative on Minikube.

## TL;DR

1. install docker, kubectl, and minikube
2. install the latest riff CLI
3. create a minikube cluster for Knative
4. install Knative using the riff CLI
5. create a function
6. invoke the function

### install docker

Installing [Docker Community Edition](https://store.docker.com/search?type=edition&offering=community) is the easiest way get started with docker. Since minikube includes its own docker daemon, you actually only need the docker CLI to build function containers for riff. This means that if you want to, you can shut down the Docker (server) app, and turn off automatic startup of Docker on login.

### install kubectl

[Kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/) is the Kubernetes CLI. It is used to manage minikube as well as hosted Kubernetes clusters. If you already have the Google Cloud Platform SDK, use: `gcloud components install kubectl`.

### install minikube

[Minikube](https://kubernetes.io/docs/tasks/tools/install-minikube/) is a Kubernetes environment which runs in a single virtual machine. See the [latest release](https://github.com/kubernetes/minikube/releases) for installation, and the [readme](https://github.com/kubernetes/minikube/blob/master/README.md) for more detailed information.

For macOS we recommend using Hyperkit as the vm driver. To install Hyperkit, first install [Docker Desktop (Mac)](https://store.docker.com/editions/community/docker-ce-desktop-mac), then run:

```sh
curl -LO https://storage.googleapis.com/minikube/releases/latest/docker-machine-driver-hyperkit \
&& sudo install -o root -g wheel -m 4755 docker-machine-driver-hyperkit /usr/local/bin/
```

For Linux we suggest using the [kvm2](https://github.com/kubernetes/minikube/blob/master/docs/drivers.md#kvm2-driver) driver.

For additional details see the minikube [driver installation](https://github.com/kubernetes/minikube/blob/master/docs/drivers.md#hyperkit-driver) docs.

## create a Minikube cluster

```sh
minikube start --memory=4096 --cpus=4 \
--kubernetes-version=v1.14.0 \
--vm-driver=hyperkit \
--bootstrapper=kubeadm \
--extra-config=apiserver.enable-admission-plugins="LimitRanger,NamespaceExists,NamespaceLifecycle,ResourceQuota,ServiceAccount,DefaultStorageClass,MutatingAdmissionWebhook"
```

To use the kvm2 driver for Linux specify `--vm-driver=kvm2`. Omitting the `--vm-driver` option will use the default driver.

Confirm that your kubectl context is pointing to the new cluster

```sh
kubectl config current-context
```

## install the riff CLI

The [riff CLI](https://github.com/projectriff/riff/) is available to download from our GitHub [releases](https://github.com/projectriff/riff/releases) page. Once installed, check that the riff CLI version is 0.3.0 or later.

```sh
riff version
```
```
Version
  riff cli: 0.3.0 (4e474f57a463d4d2c1159af64d562532fcb3ac1b)
```

At this point it is useful to monitor your cluster using a utility like `watch`. To install on a Mac

```sh
brew install watch
```

Watch pods in a separate terminal.

```sh
watch -n 1 kubectl get pod --all-namespaces
```

## install Knative using the riff CLI

Install Knative, watching the pods until everything is running (this could take a couple of minutes). The `--node-port` option replaces LoadBalancer type services with NodePort.

```sh
riff system install --node-port
```

You should see pods running in namespaces istio-system, knative-build, knative-serving, and knative-eventing as well as kube-system when the system is fully operational. 

```sh
NAMESPACE          NAME                                            READY   STATUS      RESTARTS   AGE
istio-system       cluster-local-gateway-547467ccf6-xbh9m          1/1     Running     0          3m34s
istio-system       istio-citadel-7d64db8bcf-ljd5r                  1/1     Running     0          3m35s
istio-system       istio-cleanup-secrets-pw842                     0/1     Completed   0          3m36s
istio-system       istio-egressgateway-6ddf4c8bd6-k7bjr            1/1     Running     0          3m35s
istio-system       istio-galley-7dd996474-467xc                    1/1     Running     0          3m35s
istio-system       istio-ingressgateway-84b89d647f-76z5g           1/1     Running     0          3m35s
istio-system       istio-pilot-54b76645df-xdszt                    2/2     Running     0          3m21s
istio-system       istio-policy-5c4d9ff96b-htd5h                   2/2     Running     0          3m35s
istio-system       istio-sidecar-injector-6977b5cf5b-fh7mr         1/1     Running     0          3m35s
istio-system       istio-statsd-prom-bridge-b44b96d7b-htrgk        1/1     Running     0          3m35s
istio-system       istio-telemetry-7676df547f-b4vdw                2/2     Running     0          3m35s
knative-build      build-controller-7b8987d675-8vph5               1/1     Running     0          59s
knative-build      build-webhook-74795c8696-xwwld                  1/1     Running     0          59s
knative-eventing   eventing-controller-864657d8d4-hj7xz            1/1     Running     0          57s
knative-eventing   in-memory-channel-controller-f794cc9d8-nb59s    1/1     Running     0          56s
knative-eventing   in-memory-channel-dispatcher-8595c7f8d7-qzn9c   2/2     Running     1          56s
knative-eventing   webhook-5d76776d55-jb56d                        1/1     Running     0          57s
knative-serving    activator-7c8b59d78-2jrpk                       2/2     Running     1          58s
knative-serving    autoscaler-666c9bfcc6-vwcrq                     2/2     Running     1          58s
knative-serving    controller-799cd5c6dc-sbpzr                     1/1     Running     0          58s
knative-serving    webhook-5b66fdf6b9-kqvjh                        1/1     Running     0          58s
kube-system        coredns-86c58d9df4-dtf4v                        1/1     Running     0          9m17s
kube-system        coredns-86c58d9df4-hpzlx                        1/1     Running     0          9m17s
kube-system        etcd-minikube                                   1/1     Running     0          8m30s
kube-system        kube-addon-manager-minikube                     1/1     Running     0          8m15s
kube-system        kube-apiserver-minikube                         1/1     Running     0          8m20s
kube-system        kube-controller-manager-minikube                1/1     Running     0          8m29s
kube-system        kube-proxy-fcbqc                                1/1     Running     0          9m17s
kube-system        kube-scheduler-minikube                         1/1     Running     0          8m9s
kube-system        storage-provisioner                             1/1     Running     0          9m16s
```

### initialize the namespace and provide credentials for pushing images to DockerHub

Use the riff CLI to initialize your namespace (if you plan on using a namespace other than `default` then substitute the name you want to use). This will create a serviceaccount and a secret with the provided credentials and install a buildtemplate. Replace the ??? with your docker username.

```sh
export DOCKER_ID=???
```

```sh
riff namespace init default --docker-hub $DOCKER_ID
```

You will be prompted to provide the password.

## create a function

This step will pull the source code for a function from a GitHub repo, build a container image based on the node function invoker, and push the resulting image to your dockerhub repo.

```sh
riff function create square \
  --git-repo https://github.com/projectriff-samples/node-square  \
  --artifact square.js \
  --verbose
```

If you're still watching pods, you should see something like the following

```sh
NAMESPACE       NAME                         READY   STATUS      RESTARTS   AGE
default         square-rqmsf-pod-2cd1ef      0/1     Init:3/7    0          20s
```

The 7 "Init" containers may take a while to complete the first time a function is built, but eventually that pod should show a status of completed, and a new square deployment pod should be running 3/3 containers.

```sh
NAMESPACE      NAME                                         READY   STATUS      RESTARTS   AGE
default        square-5ksdq-deployment-6d875d87bf-64fz4     3/3     Running     0          47s
default        square-rqmsf-pod-2cd1ef                      0/1     Completed   0          2m30s
```

## invoke the function

```sh
riff service invoke square --json -- -w '\n' -d 8
```

#### result

```
curl http://192.168.64.46:31380/ -H 'Host: square.default.example.com' -H 'Content-Type: application/json' -w '\n' -d 8
64
```

## delete the function

```sh
riff service delete square
```
