---
id: riff-streaming
title: "riff streaming"
---
## riff streaming

(experimental) streaming runtime for riff functions

### Synopsis

The streaming runtime uses riff functions, processor and stream custom resources
to deploy streaming workloads. 

Functions can accept several input and/or output streams.

### Options

```
  -h, --help   help for streaming
```

### Options inherited from parent commands

```
      --config file       config file (default is $HOME/.riff.yaml)
      --kubeconfig file   kubectl config file (default is $HOME/.kube/config)
      --no-color          disable color output in terminals
```

### SEE ALSO

* [riff](riff.md)	 - riff is for functions
* [riff streaming gateway](riff_streaming_gateway.md)	 - (experimental) stream gateway
* [riff streaming inmemory-gateway](riff_streaming_inmemory-gateway.md)	 - (experimental) in-memory stream gateway
* [riff streaming kafka-gateway](riff_streaming_kafka-gateway.md)	 - (experimental) kafka stream gateway
* [riff streaming processor](riff_streaming_processor.md)	 - (experimental) processors apply functions to messages on streams
* [riff streaming pulsar-gateway](riff_streaming_pulsar-gateway.md)	 - (experimental) pulsar stream gateway
* [riff streaming stream](riff_streaming_stream.md)	 - (experimental) streams of messages

