---
id: riff-knative-deployer-create
title: "riff knative deployer create"
---
## riff knative deployer create

create a deployer to map HTTP requests to a workload

### Synopsis

Create a Knative deployer.

Build references are resolved within the same namespace as the deployer. As the
build produces new images, the image will roll out automatically. Image based
deployers must be updated manually to roll out new images. Rolling out an image
means creating a Knative Configuration with a pod referencing the image and a
Knative Route referencing the Configuration.

The runtime environment can be configured by --env for static key-value pairs
and --env-from to map values from a ConfigMap or Secret.

```
riff knative deployer create <name> [flags]
```

### Examples

```
riff knative deployer create my-app-deployer --application-ref my-app
riff knative deployer create my-func-deployer --function-ref my-func
riff knative deployer create my-func-deployer --container-ref my-container
riff knative deployer create my-image-deployer --image registry.example.com/my-image:latest
```

### Options

```
      --application-ref name           name of application to deploy
      --container-concurrency number   the maximum number of concurrent requests to send to a replica at one time
      --container-ref name             name of container to deploy
      --dry-run                        print kubernetes resources to stdout rather than apply them to the cluster, messages normally on stdout will be sent to stderr
      --env variable                   environment variable defined as a key value pair separated by an equals sign, example "--env MY_VAR=my-value" (may be set multiple times)
      --env-from variable              environment variable from a config map or secret, example "--env-from MY_SECRET_VALUE=secretKeyRef:my-secret-name:key-in-secret", "--env-from MY_CONFIG_MAP_VALUE=configMapKeyRef:my-config-map-name:key-in-config-map" (may be set multiple times)
      --function-ref name              name of function to deploy
  -h, --help                           help for create
      --image image                    container image to deploy
      --ingress-policy policy          ingress policy for network access to the workload, one of "ClusterLocal" or "External" (default "ClusterLocal")
      --limit-cpu cores                the maximum amount of cpu allowed, in CPU cores (500m = .5 cores)
      --limit-memory bytes             the maximum amount of memory allowed, in bytes (500Mi = 500MiB = 500 * 1024 * 1024)
      --max-scale number               maximum number of replicas (default unbounded)
      --min-scale number               minimum number of replicas (default 0)
  -n, --namespace name                 kubernetes namespace (defaulted from kube config)
      --tail                           watch deployer logs
      --target-port port               port that the workload listens on for traffic. The value is exposed to the workload as the PORT environment variable
      --wait-timeout duration          duration to wait for the deployer to become ready when watching logs (default "10m")
```

### Options inherited from parent commands

```
      --config file       config file (default is $HOME/.riff.yaml)
      --kubeconfig file   kubectl config file (default is $HOME/.kube/config)
      --no-color          disable color output in terminals
```

### SEE ALSO

* [riff knative deployer](riff_knative_deployer.md)	 - deployers map HTTP requests to a workload

