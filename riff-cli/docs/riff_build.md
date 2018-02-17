## riff build

Build a function container

### Synopsis


Build the function based on the code available in the path directory, using the name
  and version specified for the image that is built.

```
riff build [flags]
```

### Examples

```
  riff build -n <name> -v <version> -f <path> [--push]
```

### Options

```
      --dry-run               print generated function artifacts content to stdout only
  -f, --filepath string       path or directory used for the function resources (defaults to the current directory)
  -h, --help                  help for build
  -n, --name string           the name of the function (defaults to the name of the current directory)
      --push                  push the image to Docker registry
      --riff-version string   the version of riff to use when building containers (default "latest")
  -u, --useraccount string    the Docker user account to be used for the image repository (default "current OS user")
  -v, --version string        the version of the function image (default "0.0.1")
```

### Options inherited from parent commands

```
      --config string   config file (default is $HOME/.riff.yaml)
```

### SEE ALSO
* [riff](riff.md)	 - Commands for creating and managing function resources

