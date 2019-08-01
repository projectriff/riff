---
id: gke
title: Getting started on GKE
sidebar_label: GKE
---

The following will help you get started running a riff function with Knative on GKE.

## TL;DR

1. select a Project, install and configure gcloud and kubectl
1. create a GKE cluster for Knative
1. install the latest riff CLI
1. install Knative using the riff CLI
1. create a function
1. invoke the function


## create a Google Cloud project

A project is required to consume any Google Cloud services, including GKE clusters. When you log into the [console](https://console.cloud.google.com/) you can select or create a project from the dropdown at the top. 

### install gcloud

Follow the [quickstart instructions](https://cloud.google.com/sdk/docs/quickstarts) to install the [Google Cloud SDK](https://cloud.google.com/sdk/) which includes the `gcloud` CLI. You may need to add the `google-cloud-sdk/bin` directory to your path. Once installed, `gcloud init` will open a browser to start an oauth flow and configure gcloud to use your project.

```sh
gcloud init
```

### install kubectl
[Kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/) is the Kubernetes CLI. It is used to manage minikube as well as hosted Kubernetes clusters like GKE. If you don't already have kubectl on your machine, you can use gcloud to install it.

```sh
gcloud components install kubectl
```

### configure gcloud

Create an environment variable, replacing ??? with your project ID (not to be confused with your project name; use `gcloud projects list` to find your project ID). 

```sh
export GCP_PROJECT_ID=???
```

Check your default project.

```sh
gcloud config list
```

If necessary change the default project.

```sh
gcloud config set project $GCP_PROJECT_ID
```

List the available compute zones and also regions with quotas.

```sh
gcloud compute zones list
gcloud compute regions list
```

Choose a zone, preferably in a region with higher CPU quota.

```sh
export GCP_ZONE=us-central1-b
```

Enable the necessary APIs for gcloud. You also need to [enable billing](https://cloud.google.com/billing/docs/how-to/manage-billing-account) for your new project.

```sh
gcloud services enable \
  cloudapis.googleapis.com \
  container.googleapis.com \
  containerregistry.googleapis.com
```

## create a GKE cluster

Choose a new unique lowercase cluster name and create the cluster. For this demo, three nodes should be sufficient.

```sh
# replace ??? below with your own cluster name
export CLUSTER_NAME=???
```

```sh
gcloud container clusters create $CLUSTER_NAME \
  --cluster-version=latest \
  --machine-type=n1-standard-2 \
  --enable-autoscaling --min-nodes=1 --max-nodes=3 \
  --enable-autorepair \
  --scopes=service-control,service-management,compute-rw,storage-ro,cloud-platform,logging-write,monitoring-write,pubsub,datastore \
  --num-nodes=3 \
  --zone=$GCP_ZONE
```

For additional details see [Knative Install on Google Kubernetes Engine](https://github.com/knative/docs/blob/master/install/Knative-with-GKE.md).

Confirm that your kubectl context is pointing to the new cluster

```sh
kubectl config current-context
```

To list contexts:

```sh
kubectl config get-contexts
```

You should also be able to find the cluster the [Kubernetes Engine](https://console.cloud.google.com/kubernetes/) console.

## grant yourself cluster-admin permissions

```sh
kubectl create clusterrolebinding cluster-admin-binding \
--clusterrole=cluster-admin \
--user=$(gcloud config get-value core/account)
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

Install Knative, watching the pods until everything is running (this could take a couple of minutes).

```sh
riff system install
```

You should see pods running in namespaces istio-system, knative-build, knative-serving, and knative-eventing as well as kube-system when the system is fully operational. 

```
NAMESPACE          NAME                                                             READY   STATUS      RESTARTS   AGE
istio-system       cluster-local-gateway-6c785b8db7-nrfdp                           1/1     Running     0          7m11s
istio-system       istio-citadel-6959fcfb88-rrrw5                                   1/1     Running     0          7m28s
istio-system       istio-cleanup-secrets-ns7tp                                      0/1     Completed   0          7m54s
istio-system       istio-egressgateway-5b765869bf-dmf2j                             1/1     Running     0          7m30s
istio-system       istio-galley-7fccb9bbd9-wpbtm                                    1/1     Running     0          7m30s
istio-system       istio-ingressgateway-69b597b6bd-7qmgp                            1/1     Running     0          7m30s
istio-system       istio-pilot-78679fcc74-bf6zn                                     2/2     Running     0          7m4s
istio-system       istio-policy-59b7f4ccd5-9cwmv                                    2/2     Running     0          7m29s
istio-system       istio-sidecar-injector-5c4b6cb6bc-b8kwt                          1/1     Running     0          7m28s
istio-system       istio-statsd-prom-bridge-67bbcc746c-mqhvj                        1/1     Running     0          7m32s
istio-system       istio-telemetry-7686cd76bd-c622z                                 2/2     Running     0          7m29s
knative-build      build-controller-755f6dd8b4-knjzx                                1/1     Running     0          6m31s
knative-build      build-webhook-588dcc4f7f-hhmvb                                   1/1     Running     0          6m31s
knative-eventing   eventing-controller-6554f9cbcf-4mcsm                             1/1     Running     0          6m15s
knative-eventing   in-memory-channel-controller-7888dfffd7-klx5d                    1/1     Running     0          6m6s
knative-eventing   in-memory-channel-dispatcher-56d6f99dc6-mlgkv                    2/2     Running     2          6m5s
knative-eventing   webhook-654b696b9b-hj89q                                         1/1     Running     0          6m14s
knative-serving    activator-5f8c9678bd-qc49k                                       2/2     Running     2          6m24s
knative-serving    autoscaler-7486469d84-7l4lp                                      2/2     Running     1          6m24s
knative-serving    controller-677598fdff-q56wf                                      1/1     Running     0          6m20s
knative-serving    webhook-5bb858fc5-mls79                                          1/1     Running     0          6m20s
kube-system        event-exporter-v0.2.3-f9c896d75-7hs5v                            2/2     Running     0          87m
kube-system        fluentd-gcp-scaler-69d79984cb-f4qp6                              1/1     Running     0          87m
kube-system        fluentd-gcp-v3.2.0-cttdq                                         2/2     Running     0          86m
kube-system        fluentd-gcp-v3.2.0-x7xmk                                         2/2     Running     0          86m
kube-system        heapster-v1.6.0-beta.1-577d766b74-97wzl                          3/3     Running     0          86m
kube-system        kube-dns-5d8cd9fcb6-hkgrf                                        4/4     Running     0          87m
kube-system        kube-dns-5d8cd9fcb6-zvqck                                        4/4     Running     0          86m
kube-system        kube-dns-autoscaler-76fcd5f658-h65gx                             1/1     Running     0          87m
kube-system        kube-proxy-gke-jldec-riff-030-us-ea-default-pool-4b303b9f-dz56   1/1     Running     0          87m
kube-system        kube-proxy-gke-jldec-riff-030-us-ea-default-pool-4b303b9f-fpm4   1/1     Running     0          87m
kube-system        l7-default-backend-6f8697844f-8wfn5                              1/1     Running     0          87m
kube-system        metrics-server-v0.3.1-54699c9cc8-fl2n9                           2/2     Running     0          86m
kube-system        prometheus-to-sd-bdbnk                                           1/1     Running     0          87m
kube-system        prometheus-to-sd-gb9nf                                           1/1     Running     0          87m
```

## create a Kubernetes secret for pushing images to GCR

Create a [GCP Service Account](https://cloud.google.com/iam/docs/creating-managing-service-accounts) in the GCP console or using the gcloud CLI

```sh
gcloud iam service-accounts create push-image
```

Grant the service account a "storage.admin" role using the [IAM manager](https://cloud.google.com/iam/docs/granting-roles-to-service-accounts) or using gcloud.

```sh
gcloud projects add-iam-policy-binding $GCP_PROJECT_ID \
    --member serviceAccount:push-image@$GCP_PROJECT_ID.iam.gserviceaccount.com \
    --role roles/storage.admin
```

Create a new [authentication key](https://cloud.google.com/container-registry/docs/advanced-authentication#json_key_file) for the service account and save it in `gcr-storage-admin.json`.

```sh
gcloud iam service-accounts keys create \
  --iam-account "push-image@$GCP_PROJECT_ID.iam.gserviceaccount.com" \
  gcr-storage-admin.json
```

### initialize the namespace

Use the riff CLI to initialize your namespace (if you plan on using a namespace other than `default` then substitute the name you want to use). This creates a serviceaccount that uses the secret saved above, installs a buildtemplate and labels the namespace for automatic Istio sidecar injection.

```sh
riff namespace init default --gcr gcr-storage-admin.json
```

## create a function

This step will pull the source code for a function from a GitHub repo, build a container image based on the node function invoker, and push the resulting image to GCR.

```sh
riff function create square \
  --git-repo https://github.com/projectriff-samples/node-square \
  --artifact square.js \
  --verbose
```

If you're still watching pods, you should see something like the following

```
NAMESPACE    NAME                  READY     STATUS      RESTARTS   AGE
default      square-00001-jk9vj    0/1       Init:0/4    0          24s
```

The 4 "Init" containers may take a while to complete the first time a function is built, but eventually that pod should show a status of completed, and a new square deployment pod should be running 3/3 containers.

```
NAMESPACE   NAME                                       READY     STATUS      RESTARTS   AGE
default     square-00001-deployment-679bffb58c-cpzz8   3/3       Running     0          4m
default     square-00001-jk9vj                         0/1       Completed   0          5m
```

## invoke the function

```sh
riff service invoke square --json -- -w '\n' -d 8
```

#### result

```
curl 35.236.212.232/ -H 'Host: square.default.example.com' -H 'Content-Type: text/plain' -w '\n' -d 8
64
```

## delete the function

```sh
riff service delete square
```
