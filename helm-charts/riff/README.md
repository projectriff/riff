# Riff Helm Chart

[riff](https://github.com/projectriff/riff) is for functions - a FaaS built for Kubernetes


## Installing the Chart

To install the chart with the release name `my-release`:

```bash
$ helm repo add riffrepo https://riff-charts.storage.googleapis.com
$ helm repo update
$ helm install --name my-release riffrepo/riff
```

If you are using a cluster that does not have a load balancer (like Minikube) then you can install using a NodePort:

```bash
$ helm install --name my-release --set httpGateway.service.type=NodePort riffrepo/riff
```

## Configuration

The following lists the configurable parameters and their default values.

| Parameter               | Description                            | Default                   |
| ----------------------- | -------------------------------------- | ------------------------- |
| `functionController.image.tag`|The image tag for the function-controller|0.0.1|
| `functionController.image.pullPolicy`|The imagePullPolicy for the function-controller|IfNotPresent|
| `functionController.sidecar.image.tag`|The image tag for the sidecar used|0.0.1|
| `functionController.service.type`|The service type used for the function-controller|ClusterIP|
| `topicController.image.tag`|The image tag for the topic-controller|0.0.1|
| `topicController.image.pullPolicy`|The imagePullPolicy for the topic-controller|IfNotPresent|
| `httpGateway.image.tag`|The image tag for the http-gateway|0.0.1|
| `httpGateway.image.pullPolicy`|The imagePullPolicy for the http-gateway|IfNotPresent|
| `httpGateway.service.type`|The service type used for the http-gateway|LoadBalancer|

## Uninstalling the Release

To remove the chart release with the name `my-release` and purge all the release info use:

```bash
$ helm delete --purge my-release
```
