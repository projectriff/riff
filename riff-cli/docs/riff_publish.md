## riff publish

Publish data to a topic using the http-gateway

### Synopsis


Publish data to a topic using the http-gateway. For example:

    riff publish -i greetings -d hello -r

will post 'hello' to the 'greetings' topic and wait for a reply.


```
riff publish [flags]
```

### Options

```
  -c, --count int          the number of times to post the data (default 1)
  -d, --data string        the data to post to the http-gateway using the input topic
  -h, --help               help for publish
  -i, --input string       the name of the input topic, defaults to name of current directory (default "riff-cli")
      --namespace string   the namespace of the http-gateway (default "default")
  -p, --pause int          the number of seconds to wait between postings
  -r, --reply              wait for a reply containing the results of the function execution
```

### Options inherited from parent commands

```
      --config string   config file (default is $HOME/.riff.yaml)
```

### SEE ALSO
* [riff](riff.md)	 - Commands for creating and managing function resources

