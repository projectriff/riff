## riff init

Initialize a function

### Synopsis


Generate the function based on the function source code specified as the filename, using the name
  and version specified for the function image repository and tag. 

For example, from a directory named 'square' containing a function 'square.js', you can simply type :

    riff init node -f square

  or

    riff  init node

to generate the required Dockerfile and resource definitions using sensible defaults.

```
riff init [language] [flags]
```

### Options

```
  -a, --artifact string       path to the function artifact, source code or jar file
      --dry-run               print generated function artifacts content to stdout only
  -f, --filepath string       path or directory used for the function resources (defaults to the current directory)
      --force                 overwrite existing functions artifacts
  -h, --help                  help for init
  -i, --input string          the name of the input topic (defaults to function name)
  -n, --name string           the name of the function (defaults to the name of the current directory)
  -o, --output string         the name of the output topic (optional)
  -p, --protocol string       the protocol to use for function invocations
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
* [riff init java](riff_init_java.md)	 - Initialize a Java function
* [riff init node](riff_init_node.md)	 - Initialize a node.js function
* [riff init python](riff_init_python.md)	 - Initialize a Python function
* [riff init shell](riff_init_shell.md)	 - Initialize a shell script function

