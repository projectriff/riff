## riff apply

Apply function resource definitions

### Synopsis


Apply the resource definition[s] included in the path. A resource will be created if it doesn't exist yet.

```
riff apply [flags]
```

### Examples

```
  riff apply -f some/function/path
  riff apply -f some/function/path/some.yaml
```

### Options

```
      --dry-run            print generated function artifacts content to stdout only
  -f, --filepath string    path or directory used for the function resources (defaults to the current directory)
  -h, --help               help for apply
      --namespace string   the namespace used for the deployed resources (default "default")
```

### Options inherited from parent commands

```
      --config string   config file (default is $HOME/.riff.yaml)
```

### SEE ALSO
* [riff](riff.md)	 - Commands for creating and managing function resources

