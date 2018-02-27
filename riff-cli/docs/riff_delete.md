## riff delete

Delete function resources

### Synopsis


Delete the resource[s] for the function or path specified.

```
riff delete [flags]
```

### Examples

```
  riff delete -n square
    or
  riff delete -f function/square
```

### Options

```
      --all                delete all resources including topics, not just the function resource
      --dry-run            print generated function artifacts content to stdout only
  -f, --filepath string    path or directory used for the function resources (defaults to the current directory)
  -h, --help               help for delete
  -n, --name string        the name of the function (defaults to the name of the current directory)
      --namespace string   the namespace used for the deployed resources (default "default")
```

### Options inherited from parent commands

```
      --config string   config file (default is $HOME/.riff.yaml)
```

### SEE ALSO
* [riff](riff.md)	 - Commands for creating and managing function resources

