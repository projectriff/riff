## riff processor create

create a processor to apply a function to messages on streams

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
      --function-ref name   name of function build to deploy
  -h, --help                help for create
      --input name          name of stream to read messages from (may be set multiple times)
  -n, --namespace name      kubernetes namespace (defaulted from kube config)
      --output name         name of stream to write messages to (may be set multiple times)
```

### Options inherited from parent commands

```
      --config file        config file (default is $HOME/.riff.yaml)
      --kube-config file   kubectl config file (default is $HOME/.kube/config)
      --no-color           disable color output in terminals
```

### SEE ALSO

* [riff processor](riff_processor.md)	 - processors apply functions to messages on streams

