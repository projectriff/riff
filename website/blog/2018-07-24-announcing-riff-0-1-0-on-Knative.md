---
title: "Announcing riff v0.1.0 on Knative"
---

We are excited to announce that the riff team has re-architected the riff core, bringing essential aspects of riff to Knative. This is the first release of riff on Knative.

<!--truncate-->

[Knative](https://github.com/knative/docs) is a new open source project announced at [Google Cloud Next '18](https://cloud.withgoogle.com/next18).

Knative provides Kubernetes-native APIs for deploying serverless-style functions, applications, and containers.

## The riff CLI will install Knative

Please follow one of our new [getting started](/docs/) guides to download the latest riff CLI and install Knative onto a Kubernetes cluster. We currently support GKE and minikube.

## Functions build on riff Invokers and run on Knative

The riff CLI creates functions using a Knative [build](https://github.com/knative/build) template based on [Kaniko](https://github.com/GoogleContainerTools/kaniko).

Once built, the Functions are deployed and run as Knative [Services](https://github.com/knative/serving/blob/master/docs/spec/overview.md#resource-types) with support for autoscaling, revisions, and traffic routing.

The getting started guides provide step by step instructions to create a sample function based on a riff invoker.

## Buses, Channels, Subscriptions

The riff sidecar architecture from earlier releases has changed in order to leverage the Knative revisions and routes.

This preserves the biggest differentiator of riff, which was the ability of riff functions to consume and produce event streams from topics on message brokers.

![riff Knative pubsub resources](/img/riff-knative-pubsub-resources.png)

## Next steps

Since this is an early preview, we are working hard to fill in some of the gaps and rough edges. 

We are looking forward to working with the Knative community, and continuing to make contributions to the Knative project, especially [Knative/eventing](https://github.com/knative/eventing/tree/master/pkg).