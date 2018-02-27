# Riff Helm Chart

Single-node Kafka for riff

## Installing the Chart

To install the chart with the release name `my-release`:

```bash
$ helm repo add riffrepo https://riff-charts.storage.googleapis.com
$ helm repo update
$ helm install --name my-release riffrepo/kafka
```

## Uninstalling the Release

To remove the chart release with the name `my-release` and purge all the release info use:

```bash
$ helm delete --purge my-release
```
