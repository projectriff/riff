## riff processor create

process messages with a function

### Synopsis


<todo>


```
riff processor create [flags]
```

### Examples

```
riff processor create my-processor --function-ref my-func --input my-input-stream
riff processor create my-processor --function-ref my-func --input my-input-stream --input my-join-stream --output my-output-stream
```

### Options

```
      --function-ref string   function build to deploy
  -h, --help                  help for create
      --input stringArray     stream to read messages from
  -n, --namespace string      kubernetes namespace (defaulted from kube config)
      --output stringArray    stream to write messages to
```

### Options inherited from parent commands

```
      --config string        config file (default is $HOME/.riff.yaml)
      --kube-config string   kubectl config file (default is $HOME/.kube/config)
      --no-color             disable color output in terminals
```

### SEE ALSO

* [riff processor](riff_processor.md)	 - processors apply functions to messages on streams

