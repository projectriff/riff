# Riff Helm Chart

[riff](https://github.com/projectriff/riff) is for functions - a FaaS built for Kubernetes


## Installing the Chart

To install the chart with the release name `projectriff` in the `riff-system` namespace:

```bash
$ helm repo add projectriff https://riff-charts.storage.googleapis.com
$ helm repo update
$ kubectl create namespace riff-system
$ helm install --name projectriff --namespace riff-system projectriff/riff
```

If you are using a cluster that does not have a load balancer (like Minikube) then you can install using a NodePort:

```bash
$ helm install --name projectriff --namespace riff-system --set httpGateway.service.type=NodePort projectriff/riff
```

## Configuration

The following lists the configurable parameters and their default values.

| Parameter               | Description                            | Default                   |
| ----------------------- | -------------------------------------- | ------------------------- |
| `rbac.create`|Specifies whether RBAC resources should be created|true|
| `kafka.create`|Specifies whether the single-node development Kafka chart should be installed|false|
| `functionController.image.tag`|The image tag for the function-controller|latest|
| `functionController.image.pullPolicy`|The imagePullPolicy for the function-controller|IfNotPresent|
| `functionController.sidecar.image.tag`|The image tag for the sidecar used|latest|
| `functionController.service.type`|The service type used for the function-controller|ClusterIP|
| `topicController.image.tag`|The image tag for the topic-controller|latest|
| `topicController.image.pullPolicy`|The imagePullPolicy for the topic-controller|IfNotPresent|
| `httpGateway.image.tag`|The image tag for the http-gateway|latest|
| `httpGateway.image.pullPolicy`|The imagePullPolicy for the http-gateway|IfNotPresent|
| `httpGateway.service.type`|The service type used for the http-gateway|LoadBalancer|
| `httpGateway.httpHeadersWhitelist`|A comma separated list of HTTP request headers to whitelist|(none)|

## Uninstalling the Release

To remove the chart release with the name `projectriff` and purge all the release info use:

```bash
$ helm delete --purge my-release
```
