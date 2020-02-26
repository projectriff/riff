---
id: riff-streaming-processor-create
title: "riff streaming processor create"
---
## riff streaming processor create

create a processor to apply a function to messages on streams

### Synopsis

Creates a processor within a namespace.

The processor is configured with a function or container reference and multiple
input and/or output streams.

```
riff streaming processor create <name> [flags]
```

### Examples

```
riff streaming processor create my-processor --function-ref my-func --input my-input-stream
riff streaming processor create my-processor --function-ref my-func --input input:my-input-stream --input my-join-stream@earliest --output out:my-output-stream
```

### Options

```
      --container-ref name      name of container to deploy
      --dry-run                 print kubernetes resources to stdout rather than apply them to the cluster, messages normally on stdout will be sent to stderr
      --env variable            environment variable defined as a key value pair separated by an equals sign, example "--env MY_VAR=my-value" (may be set multiple times)
      --env-from variable       environment variable from a config map or secret, example "--env-from MY_SECRET_VALUE=secretKeyRef:my-secret-name:key-in-secret", "--env-from MY_CONFIG_MAP_VALUE=configMapKeyRef:my-config-map-name:key-in-config-map" (may be set multiple times)
      --function-ref name       name of function to deploy
  -h, --help                    help for create
      --image image             container image to deploy
      --input name              name of stream to read messages from (or [<alias>:]<stream>[@<earliest|latest>], may be set multiple times)
  -n, --namespace name          kubernetes namespace (defaulted from kube config)
      --output name             name of stream to write messages to (or [<alias>:]<stream>, may be set multiple times)
      --tail                    watch processor logs
      --wait-timeout duration   duration to wait for the processor to become ready when watching logs (default "10m")
```

### Options inherited from parent commands

```
      --config file       config file (default is $HOME/.riff.yaml)
      --kubeconfig file   kubectl config file (default is $HOME/.kube/config)
      --no-color          disable color output in terminals
```

### SEE ALSO

* [riff streaming processor](riff_streaming_processor.md)	 - (experimental) processors apply functions to messages on streams

