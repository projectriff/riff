## riff create

Create a function

### Synopsis

Create the function based on the function source code specified as the filename, using the name
and version specified for the function image repository and tag. 

For example, from a directory named 'square' containing a function 'square.js', you can simply type :

    riff create -f square

  or

    riff create

to create the resource definitions, and apply the resources, using sensible defaults.

```
riff create [flags]
```

### Options

```
  -a, --artifact string          path to the function artifact, source code or jar file
      --dry-run                  print generated function artifacts content to stdout only
  -f, --filepath string          path or directory used for the function resources (defaults to the current directory)
      --force                    overwrite existing functions artifacts
  -h, --help                     help for create
  -i, --input string             the name of the input topic (defaults to function name)
      --invoker-version string   the version of the invoker to use when building containers
  -n, --name string              the name of the function (defaults to the name of the current directory)
      --namespace string         the namespace used for the deployed resources (defaults to kubectl's default)
  -o, --output string            the name of the output topic (optional)
      --push                     push the image to Docker registry
  -u, --useraccount string       the Docker user account to be used for the image repository (default "current OS user")
  -v, --version string           the version of the function image (default "0.0.1")
```

### Options inherited from parent commands

```
      --config string   config file (default is $HOME/.riff.yaml)
```

### SEE ALSO

* [riff](riff.md)	 - Commands for creating and managing function resources

