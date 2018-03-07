## riff create shell

Create a shell script function

### Synopsis


Create the function based on the function script specified as the filename, using the name
  and version specified for the function image repository and tag. 

For example, from a directory named 'echo' containing a function 'echo.sh', you can simply type :

    riff create shell -f echo

  or

    riff create shell

to create the required Dockerfile and resource definitions, and apply the resources, using sensible defaults.

```
riff create shell [flags]
```

### Options

```
  -h, --help   help for shell
```

### Options inherited from parent commands

```
  -a, --artifact string       path to the function artifact, source code or jar file
      --config string         config file (default is $HOME/.riff.yaml)
      --dry-run               print generated function artifacts content to stdout only
  -f, --filepath string       path or directory used for the function resources (defaults to the current directory)
      --force                 overwrite existing functions artifacts
  -i, --input string          the name of the input topic (defaults to function name)
  -n, --name string           the name of the function (defaults to the name of the current directory)
      --namespace string      the namespace used for the deployed resources (default "default")
  -o, --output string         the name of the output topic (optional)
  -p, --protocol string       the protocol to use for function invocations
      --push                  push the image to Docker registry
      --riff-version string   the version of riff to use when building containers (default "latest")
  -u, --useraccount string    the Docker user account to be used for the image repository (default "current OS user")
  -v, --version string        the version of the function image (default "0.0.1")
```

### SEE ALSO
* [riff create](riff_create.md)	 - Create a function

