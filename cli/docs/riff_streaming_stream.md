---
id: riff-streaming-stream
title: "riff streaming stream"
---
## riff streaming stream

(experimental) streams of messages

### Synopsis

A stream encapsulates an addressable message channel (typically a message 
broker's topic). It can be mapped to a function input or output stream.

Streams are managed by an associated streaming gateway and define a content 
type that its messages adhere to.

### Options

```
  -h, --help   help for stream
```

### Options inherited from parent commands

```
      --config file       config file (default is $HOME/.riff.yaml)
      --kubeconfig file   kubectl config file (default is $HOME/.kube/config)
      --no-color          disable color output in terminals
```

### SEE ALSO

* [riff streaming](riff_streaming.md)	 - (experimental) streaming runtime for riff functions
* [riff streaming stream create](riff_streaming_stream_create.md)	 - create a stream of messages
* [riff streaming stream delete](riff_streaming_stream_delete.md)	 - delete stream(s)
* [riff streaming stream list](riff_streaming_stream_list.md)	 - table listing of streams
* [riff streaming stream status](riff_streaming_stream_status.md)	 - show stream status

