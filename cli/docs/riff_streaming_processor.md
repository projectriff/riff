---
id: riff-streaming-processor
title: "riff streaming processor"
---
## riff streaming processor

(experimental) processors apply functions to messages on streams

### Synopsis

Processors coordinate reading from input streams and writing to output streams
with a function or container.

Function-based processors continuously watch for the latest built image and will
deploy new images. If the underlying build resource is deleted, the processor
will continue to run, but will no longer self update. Container-based processors
must be manually updated to trigger the rollout of an updated image.

### Options

```
  -h, --help   help for processor
```

### Options inherited from parent commands

```
      --config file       config file (default is $HOME/.riff.yaml)
      --kubeconfig file   kubectl config file (default is $HOME/.kube/config)
      --no-color          disable color output in terminals
```

### SEE ALSO

* [riff streaming](riff_streaming.md)	 - (experimental) streaming runtime for riff functions
* [riff streaming processor create](riff_streaming_processor_create.md)	 - create a processor to apply a function to messages on streams
* [riff streaming processor delete](riff_streaming_processor_delete.md)	 - delete processor(s)
* [riff streaming processor list](riff_streaming_processor_list.md)	 - table listing of processors
* [riff streaming processor status](riff_streaming_processor_status.md)	 - show processor status
* [riff streaming processor tail](riff_streaming_processor_tail.md)	 - watch processor logs

